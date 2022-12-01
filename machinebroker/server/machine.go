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
	"fmt"

	computev1alpha1 "github.com/onmetal/onmetal-api/api/compute/v1alpha1"
	"github.com/onmetal/onmetal-api/machinebroker/apiutils"
	ori "github.com/onmetal/onmetal-api/ori/apis/machine/v1alpha1"
	corev1 "k8s.io/api/core/v1"
)

type AggregateOnmetalMachine struct {
	IgnitionSecret *corev1.Secret
	Machine        *computev1alpha1.Machine
}

var onmetalMachineStateToMachineState = map[computev1alpha1.MachineState]ori.MachineState{
	computev1alpha1.MachineStatePending:  ori.MachineState_MACHINE_PENDING,
	computev1alpha1.MachineStateRunning:  ori.MachineState_MACHINE_RUNNING,
	computev1alpha1.MachineStateShutdown: ori.MachineState_MACHINE_SUSPENDED,
}

func (s *Server) convertOnmetalMachineState(state computev1alpha1.MachineState) (ori.MachineState, error) {
	if res, ok := onmetalMachineStateToMachineState[state]; ok {
		return res, nil
	}
	return 0, fmt.Errorf("unknown onmetal machine state %q", state)
}

var onmetalNetworkInterfaceStateToNetworkInterfaceAttachmentState = map[computev1alpha1.NetworkInterfaceState]ori.NetworkInterfaceAttachmentState{
	computev1alpha1.NetworkInterfaceStatePending:  ori.NetworkInterfaceAttachmentState_NETWORK_INTERFACE_ATTACHMENT_PENDING,
	computev1alpha1.NetworkInterfaceStateAttached: ori.NetworkInterfaceAttachmentState_NETWORK_INTERFACE_ATTACHMENT_ATTACHED,
	computev1alpha1.NetworkInterfaceStateDetached: ori.NetworkInterfaceAttachmentState_NETWORK_INTERFACE_ATTACHMENT_DETACHED,
}

func (s *Server) convertOnmetalNetworkInterfaceState(state computev1alpha1.NetworkInterfaceState) (ori.NetworkInterfaceAttachmentState, error) {
	if res, ok := onmetalNetworkInterfaceStateToNetworkInterfaceAttachmentState[state]; ok {
		return res, nil
	}
	return 0, fmt.Errorf("unknown onmetal network interface attachment state %q", state)
}

func (s *Server) convertOnmetalNetworkInterfaceStatus(status computev1alpha1.NetworkInterfaceStatus) (*ori.NetworkInterfaceAttachmentStatus, error) {
	state, err := s.convertOnmetalNetworkInterfaceState(status.State)
	if err != nil {
		return nil, err
	}

	return &ori.NetworkInterfaceAttachmentStatus{
		Name:                   status.Name,
		NetworkInterfaceHandle: status.Handle,
		State:                  state,
	}, nil
}

var onmetalVolumeStateToVolumeAttachmentState = map[computev1alpha1.VolumeState]ori.VolumeAttachmentState{
	computev1alpha1.VolumeStatePending:  ori.VolumeAttachmentState_VOLUME_ATTACHMENT_PENDING,
	computev1alpha1.VolumeStateAttached: ori.VolumeAttachmentState_VOLUME_ATTACHMENT_ATTACHED,
	computev1alpha1.VolumeStateDetached: ori.VolumeAttachmentState_VOLUME_ATTACHMENT_DETACHED,
}

func (s *Server) convertOnmetalVolumeState(state computev1alpha1.VolumeState) (ori.VolumeAttachmentState, error) {
	if res, ok := onmetalVolumeStateToVolumeAttachmentState[state]; ok {
		return res, nil
	}
	return 0, fmt.Errorf("unknown onmetal volume attachment state %q", state)
}

func (s *Server) convertOnmetalVolumeStatus(status computev1alpha1.VolumeStatus) (*ori.VolumeAttachmentStatus, error) {
	state, err := s.convertOnmetalVolumeState(status.State)
	if err != nil {
		return nil, err
	}

	return &ori.VolumeAttachmentStatus{
		Name:         status.Name,
		VolumeHandle: status.Handle,
		State:        state,
	}, nil
}

func (s *Server) convertOnmetalVolumeAttachment(volume computev1alpha1.Volume) (*ori.VolumeAttachment, error) {
	var (
		volumeID  string
		emptyDisk *ori.EmptyDiskSpec
	)
	switch {
	case volume.VolumeRef != nil:
		volumeID = volume.VolumeRef.Name
	case volume.EmptyDisk != nil:
		var sizeBytes uint64
		if sizeLimit := volume.EmptyDisk.SizeLimit; sizeLimit != nil {
			sizeBytes = sizeLimit.AsDec().UnscaledBig().Uint64()
		}
		emptyDisk = &ori.EmptyDiskSpec{
			SizeBytes: sizeBytes,
		}
	default:
		return nil, fmt.Errorf("volume %#v does neither specify volume ref nor empty disk", volume)
	}

	return &ori.VolumeAttachment{
		Device:    *volume.Device,
		VolumeId:  volumeID,
		EmptyDisk: emptyDisk,
	}, nil
}

func (s *Server) convertOnmetalNetworkInterfaceAttachment(networkInterface computev1alpha1.NetworkInterface) (*ori.NetworkInterfaceAttachment, error) {
	switch {
	case networkInterface.NetworkInterfaceRef != nil:
		return &ori.NetworkInterfaceAttachment{
			Name:               networkInterface.Name,
			NetworkInterfaceId: networkInterface.NetworkInterfaceRef.Name,
		}, nil
	default:
		return nil, fmt.Errorf("network interface %#v does not specify network interface ref", networkInterface)
	}
}

func (s *Server) convertAggregateOnmetalMachine(machine *AggregateOnmetalMachine) (*ori.Machine, error) {
	metadata, err := apiutils.GetObjectMetadata(machine.Machine)
	if err != nil {
		return nil, err
	}

	var ignitionSpec *ori.IgnitionSpec
	if ignitionSecret := machine.IgnitionSecret; ignitionSecret != nil {
		ignitionSpec = &ori.IgnitionSpec{
			Data: ignitionSecret.Data[computev1alpha1.DefaultIgnitionKey],
		}
	}

	var imageSpec *ori.ImageSpec
	if image := machine.Machine.Spec.Image; image != "" {
		imageSpec = &ori.ImageSpec{
			Image: image,
		}
	}

	volumeAttachments := make([]*ori.VolumeAttachment, len(machine.Machine.Spec.Volumes))
	for i, volume := range machine.Machine.Spec.Volumes {
		volumeAttachment, err := s.convertOnmetalVolumeAttachment(volume)
		if err != nil {
			return nil, fmt.Errorf("error converting machine volume %s: %w", *volume.Device, err)
		}

		volumeAttachments[i] = volumeAttachment
	}

	networkInterfaceAttachments := make([]*ori.NetworkInterfaceAttachment, len(machine.Machine.Spec.NetworkInterfaces))
	for i, networkInterface := range machine.Machine.Spec.NetworkInterfaces {
		networkInterfaceAttachment, err := s.convertOnmetalNetworkInterfaceAttachment(networkInterface)
		if err != nil {
			return nil, fmt.Errorf("error converting machine network interface %s: %w", networkInterface.Name, err)
		}

		networkInterfaceAttachments[i] = networkInterfaceAttachment
	}

	volumeAttachmentStates := make([]*ori.VolumeAttachmentStatus, len(machine.Machine.Status.Volumes))
	for i, volume := range machine.Machine.Status.Volumes {
		volumeAttachmentStatus, err := s.convertOnmetalVolumeStatus(volume)
		if err != nil {
			return nil, fmt.Errorf("error converting machine volume status %s: %w", volume.Name, err)
		}

		volumeAttachmentStates[i] = volumeAttachmentStatus
	}

	networkInterfaceAttachmentStates := make([]*ori.NetworkInterfaceAttachmentStatus, len(machine.Machine.Status.NetworkInterfaces))
	for i, networkInterface := range machine.Machine.Status.NetworkInterfaces {
		networkInterfaceAttachmentStatus, err := s.convertOnmetalNetworkInterfaceStatus(networkInterface)
		if err != nil {
			return nil, fmt.Errorf("error converting machine network interface status %s: %w", networkInterface.Name, err)
		}

		networkInterfaceAttachmentStates[i] = networkInterfaceAttachmentStatus
	}

	state, err := s.convertOnmetalMachineState(machine.Machine.Status.State)
	if err != nil {
		return nil, err
	}

	return &ori.Machine{
		Metadata: metadata,
		Spec: &ori.MachineSpec{
			Image:             imageSpec,
			Class:             machine.Machine.Spec.MachineClassRef.Name,
			Ignition:          ignitionSpec,
			Volumes:           volumeAttachments,
			NetworkInterfaces: networkInterfaceAttachments,
		},
		Status: &ori.MachineStatus{
			ObservedGeneration: machine.Machine.Status.MachinePoolObservedGeneration,
			State:              state,
			ImageRef:           "", // TODO: Fill
			Volumes:            volumeAttachmentStates,
			NetworkInterfaces:  networkInterfaceAttachmentStates,
		},
	}, nil
}
