// Copyright 2022 OnMetal authors
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
//      http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.

package server

import (
	"context"
	"fmt"

	computev1alpha1 "github.com/onmetal/onmetal-api/api/compute/v1alpha1"
	"github.com/onmetal/onmetal-api/machinebroker/apiutils"
	ori "github.com/onmetal/onmetal-api/ori/apis/compute/v1alpha1"
)

var onmetalVolumeStateToVolumeState = map[computev1alpha1.VolumeState]ori.VolumeState{
	computev1alpha1.VolumeStatePending:  ori.VolumeState_VOLUME_PENDING,
	computev1alpha1.VolumeStateAttached: ori.VolumeState_VOLUME_ATTACHED,
	computev1alpha1.VolumeStateDetached: ori.VolumeState_VOLUME_DETACHED,
	computev1alpha1.VolumeStateError:    ori.VolumeState_VOLUME_ERROR,
}

func (s *Server) convertOnmetalVolumeState(state computev1alpha1.VolumeState) ori.VolumeState {
	return onmetalVolumeStateToVolumeState[state]
}

func (s *Server) convertOnmetalVolumeStatus(onmetalVolumeStatus computev1alpha1.VolumeStatus) (*ori.VolumeStatus, error) {
	var emptyDisk *ori.EmptyDiskStatus
	if onmetalEmptyDiskStatus := onmetalVolumeStatus.EmptyDisk; onmetalEmptyDiskStatus != nil {
		var sizeBytes uint64
		if size := onmetalEmptyDiskStatus.Size; size != nil {
			sizeBytes = uint64(size.Value())
		}

		emptyDisk = &ori.EmptyDiskStatus{
			SizeBytes: sizeBytes,
		}
	}

	var access *ori.VolumeAccessStatus
	if referenced := onmetalVolumeStatus.Referenced; referenced != nil {
		access = &ori.VolumeAccessStatus{
			Driver: referenced.Driver,
			Handle: referenced.Handle,
		}
	}

	return &ori.VolumeStatus{
		Name:      onmetalVolumeStatus.Name,
		Device:    onmetalVolumeStatus.Device,
		State:     s.convertOnmetalVolumeState(onmetalVolumeStatus.State),
		Access:    access,
		EmptyDisk: emptyDisk,
	}, nil
}

var onmetalNetworkInterfaceStateToNetworkInterfaceState = map[computev1alpha1.NetworkInterfaceState]ori.NetworkInterfaceState{
	computev1alpha1.NetworkInterfaceStatePending:  ori.NetworkInterfaceState_NETWORK_INTERFACE_PENDING,
	computev1alpha1.NetworkInterfaceStateAttached: ori.NetworkInterfaceState_NETWORK_INTERFACE_ATTACHED,
	computev1alpha1.NetworkInterfaceStateDetached: ori.NetworkInterfaceState_NETWORK_INTERFACE_DETACHED,
	computev1alpha1.NetworkInterfaceStateError:    ori.NetworkInterfaceState_NETWORK_INTERFACE_ERROR,
}

func (s *Server) convertOnmetalNetworkInterfaceState(state computev1alpha1.NetworkInterfaceState) ori.NetworkInterfaceState {
	return onmetalNetworkInterfaceStateToNetworkInterfaceState[state]
}

func (s *Server) convertOnmetalNetworkInterfaceStatus(onmetalNetworkInterfaceStatus computev1alpha1.NetworkInterfaceStatus) (*ori.NetworkInterfaceStatus, error) {
	var virtualIP *ori.VirtualIPStatus
	if onmetalVirtualIP := onmetalNetworkInterfaceStatus.VirtualIP; onmetalVirtualIP != nil {
		virtualIP = &ori.VirtualIPStatus{
			Ip: onmetalVirtualIP.String(),
		}
	}

	return &ori.NetworkInterfaceStatus{
		Name:      onmetalNetworkInterfaceStatus.Name,
		Network:   &ori.NetworkStatus{Handle: onmetalNetworkInterfaceStatus.NetworkHandle},
		Ips:       s.convertOnmetalIPs(onmetalNetworkInterfaceStatus.IPs),
		VirtualIp: virtualIP,
		State:     s.convertOnmetalNetworkInterfaceState(onmetalNetworkInterfaceStatus.State),
	}, nil
}

var onmetalMachineStateToORIState = map[computev1alpha1.MachineState]ori.MachineState{
	computev1alpha1.MachineStatePending:  ori.MachineState_MACHINE_PENDING,
	computev1alpha1.MachineStateRunning:  ori.MachineState_MACHINE_RUNNING,
	computev1alpha1.MachineStateShutdown: ori.MachineState_MACHINE_SHUTDOWN,
	computev1alpha1.MachineStateError:    ori.MachineState_MACHINE_ERROR,
	computev1alpha1.MachineStateUnknown:  ori.MachineState_MACHINE_UNKNOWN,
}

func (s *Server) convertOnmetalMachineState(state computev1alpha1.MachineState) ori.MachineState {
	if oriState, ok := onmetalMachineStateToORIState[state]; ok {
		return oriState
	}
	return ori.MachineState_MACHINE_UNKNOWN
}

func (s *Server) MachineStatus(ctx context.Context, req *ori.MachineStatusRequest) (*ori.MachineStatusResponse, error) {
	log := s.loggerFrom(ctx)
	id := req.MachineId
	log = log.WithValues("MachineID", id)

	log.V(1).Info("Getting machine")
	onmetalMachine, err := s.getOnmetalMachine(ctx, id)
	if err != nil {
		return nil, err
	}

	metadata, err := apiutils.GetMetadataAnnotation(onmetalMachine)
	if err != nil {
		return nil, fmt.Errorf("error getting metadata: %w", err)
	}

	annotations, err := apiutils.GetAnnotationsAnnotation(onmetalMachine)
	if err != nil {
		return nil, fmt.Errorf("error getting annotations: %w", err)
	}

	labels, err := apiutils.GetLabelsAnnotation(onmetalMachine)
	if err != nil {
		return nil, fmt.Errorf("error getting labels: %w", err)
	}

	state := s.convertOnmetalMachineState(onmetalMachine.Status.State)

	var volumeStates []*ori.VolumeStatus
	for _, onmetalVolumeStatus := range onmetalMachine.Status.Volumes {
		volumeStatus, err := s.convertOnmetalVolumeStatus(onmetalVolumeStatus)
		if err != nil {
			return nil, err
		}

		volumeStates = append(volumeStates, volumeStatus)
	}

	var networkInterfaceStates []*ori.NetworkInterfaceStatus
	for _, onmetalNetworkInterfaceStatus := range onmetalMachine.Status.NetworkInterfaces {
		networkInterfaceStatus, err := s.convertOnmetalNetworkInterfaceStatus(onmetalNetworkInterfaceStatus)
		if err != nil {
			return nil, err
		}

		networkInterfaceStates = append(networkInterfaceStates, networkInterfaceStatus)
	}

	return &ori.MachineStatusResponse{
		Status: &ori.MachineStatus{
			Id:                id,
			Metadata:          metadata,
			Image:             onmetalMachine.Spec.Image,
			ImageRef:          "", // TODO: Fill
			State:             state,
			Volumes:           volumeStates,
			NetworkInterfaces: networkInterfaceStates,
			Annotations:       annotations,
			Labels:            labels,
		},
	}, nil
}
