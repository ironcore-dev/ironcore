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
	"errors"
	"fmt"

	"github.com/onmetal/onmetal-api/machinebroker/apiutils"
	ori "github.com/onmetal/onmetal-api/ori/apis/runtime/v1alpha1"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

func (s *Server) MachineStatus(ctx context.Context, req *ori.MachineStatusRequest) (*ori.MachineStatusResponse, error) {
	log := s.loggerFrom(ctx)
	id := req.MachineId
	log = log.WithValues("MachineID", id)

	log.V(1).Info("Getting machine")
	onmetalMachine, err := s.getOnmetalMachine(ctx, id)
	if err != nil {
		if !errors.Is(err, ErrMachineNotFound) {
			return nil, fmt.Errorf("error getting onmetal machine: %w", err)
		}
		return nil, status.Error(codes.NotFound, fmt.Sprintf("Requested machine %s not found", id))
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

	state := OnmetalMachineStateToORIState(onmetalMachine.Status.State)

	var volumeStates []*ori.VolumeStatus
	for _, onmetalVolumeStatus := range onmetalMachine.Status.Volumes {
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

		volumeStates = append(volumeStates, &ori.VolumeStatus{
			Name:      onmetalVolumeStatus.Name,
			Device:    onmetalVolumeStatus.Device,
			State:     OnmetalVolumeStateToVolumeState(onmetalVolumeStatus.State),
			Access:    access,
			EmptyDisk: emptyDisk,
		})
	}

	var networkInterfaceStates []*ori.NetworkInterfaceStatus
	for _, onmetalNetworkInterfaceStatus := range onmetalMachine.Status.NetworkInterfaces {
		var virtualIP *ori.VirtualIPStatus
		if onmetalVirtualIP := onmetalNetworkInterfaceStatus.VirtualIP; onmetalVirtualIP != nil {
			virtualIP = &ori.VirtualIPStatus{
				Ip: onmetalVirtualIP.String(),
			}
		}

		networkInterfaceStates = append(networkInterfaceStates, &ori.NetworkInterfaceStatus{
			Name:      onmetalNetworkInterfaceStatus.Name,
			Network:   &ori.NetworkStatus{Handle: onmetalNetworkInterfaceStatus.NetworkHandle},
			Ips:       OnmetalIPsToIPs(onmetalNetworkInterfaceStatus.IPs),
			VirtualIp: virtualIP,
			State:     OnmetalNetworkInterfaceStateToNetworkInterfaceState(onmetalNetworkInterfaceStatus.State),
		})
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
