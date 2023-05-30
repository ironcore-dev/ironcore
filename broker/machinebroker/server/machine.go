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
	networkingv1alpha1 "github.com/onmetal/onmetal-api/api/networking/v1alpha1"
	storagev1alpha1 "github.com/onmetal/onmetal-api/api/storage/v1alpha1"
	"github.com/onmetal/onmetal-api/broker/machinebroker/apiutils"
	ori "github.com/onmetal/onmetal-api/ori/apis/machine/v1alpha1"
	corev1 "k8s.io/api/core/v1"
)

type AggregateOnmetalMachine struct {
	IgnitionSecret *corev1.Secret
	Machine        *computev1alpha1.Machine
	// NetworkInterfaces is a mapping of machine network interface name to actual network interface.
	NetworkInterfaces map[string]*AggregateOnmetalNetworkInterface
	// Volumes is a mapping of machine volume name to actual volume.
	Volumes map[string]*AggregateOnmetalVolume
}

type AggregateOnmetalVolume struct {
	Volume       *storagev1alpha1.Volume
	AccessSecret *corev1.Secret
}

type AggregateOnmetalNetworkInterface struct {
	Network          *networkingv1alpha1.Network
	NetworkInterface *networkingv1alpha1.NetworkInterface
}

var onmetalMachineStateToMachineState = map[computev1alpha1.MachineState]ori.MachineState{
	computev1alpha1.MachineStatePending:    ori.MachineState_MACHINE_PENDING,
	computev1alpha1.MachineStateRunning:    ori.MachineState_MACHINE_RUNNING,
	computev1alpha1.MachineStateShutdown:   ori.MachineState_MACHINE_SUSPENDED,
	computev1alpha1.MachineStateTerminated: ori.MachineState_MACHINE_TERMINATED,
}

func (s *Server) convertOnmetalMachineState(state computev1alpha1.MachineState) (ori.MachineState, error) {
	if res, ok := onmetalMachineStateToMachineState[state]; ok {
		return res, nil
	}
	return 0, fmt.Errorf("unknown onmetal machine state %q", state)
}

var onmetalNetworkInterfaceStateToNetworkInterfaceAttachmentState = map[computev1alpha1.NetworkInterfaceState]ori.NetworkInterfaceState{
	computev1alpha1.NetworkInterfaceStatePending:  ori.NetworkInterfaceState_NETWORK_INTERFACE_PENDING,
	computev1alpha1.NetworkInterfaceStateAttached: ori.NetworkInterfaceState_NETWORK_INTERFACE_ATTACHED,
}

func (s *Server) convertOnmetalNetworkInterfaceState(state computev1alpha1.NetworkInterfaceState) (ori.NetworkInterfaceState, error) {
	if res, ok := onmetalNetworkInterfaceStateToNetworkInterfaceAttachmentState[state]; ok {
		return res, nil
	}
	return 0, fmt.Errorf("unknown onmetal network interface attachment state %q", state)
}

func (s *Server) convertOnmetalNetworkInterfaceStatus(status computev1alpha1.NetworkInterfaceStatus) (*ori.NetworkInterfaceStatus, error) {
	state, err := s.convertOnmetalNetworkInterfaceState(status.State)
	if err != nil {
		return nil, err
	}

	return &ori.NetworkInterfaceStatus{
		Name:   status.Name,
		Handle: status.Handle,
		State:  state,
	}, nil
}

var onmetalVolumeStateToVolumeAttachmentState = map[computev1alpha1.VolumeState]ori.VolumeState{
	computev1alpha1.VolumeStatePending:  ori.VolumeState_VOLUME_PENDING,
	computev1alpha1.VolumeStateAttached: ori.VolumeState_VOLUME_ATTACHED,
}

func (s *Server) convertOnmetalVolumeState(state computev1alpha1.VolumeState) (ori.VolumeState, error) {
	if res, ok := onmetalVolumeStateToVolumeAttachmentState[state]; ok {
		return res, nil
	}
	return 0, fmt.Errorf("unknown onmetal volume attachment state %q", state)
}

func (s *Server) convertOnmetalVolumeStatus(status computev1alpha1.VolumeStatus) (*ori.VolumeStatus, error) {
	state, err := s.convertOnmetalVolumeState(status.State)
	if err != nil {
		return nil, err
	}

	return &ori.VolumeStatus{
		Name:   status.Name,
		Handle: status.Handle,
		State:  state,
	}, nil
}

func (s *Server) convertOnmetalVolume(
	onmetalMachineVolume computev1alpha1.Volume,
	onmetalVolume *AggregateOnmetalVolume,
) (*ori.Volume, error) {
	var (
		connection *ori.VolumeConnection
		emptyDisk  *ori.EmptyDisk
	)
	switch {
	case onmetalMachineVolume.VolumeRef != nil:
		if access := onmetalVolume.Volume.Status.Access; access != nil {
			var secretData map[string][]byte
			if access.SecretRef != nil {
				secretData = onmetalVolume.AccessSecret.Data
			}

			connection = &ori.VolumeConnection{
				Driver:     access.Driver,
				Handle:     access.Handle,
				Attributes: access.VolumeAttributes,
				SecretData: secretData,
			}
		}
	case onmetalMachineVolume.EmptyDisk != nil:
		var sizeBytes uint64
		if sizeLimit := onmetalMachineVolume.EmptyDisk.SizeLimit; sizeLimit != nil {
			sizeBytes = sizeLimit.AsDec().UnscaledBig().Uint64()
		}
		emptyDisk = &ori.EmptyDisk{
			SizeBytes: sizeBytes,
		}
	default:
		return nil, fmt.Errorf("machine volume %#v does neither specify volume ref nor empty disk", onmetalMachineVolume)
	}

	return &ori.Volume{
		Name:       onmetalMachineVolume.Name,
		Device:     *onmetalMachineVolume.Device,
		EmptyDisk:  emptyDisk,
		Connection: connection,
	}, nil
}

func (s *Server) convertOnmetalNetworkInterfaceAttachment(
	onmetalMachineNic computev1alpha1.NetworkInterface,
	onmetalNic *AggregateOnmetalNetworkInterface,
) (*ori.NetworkInterface, error) {
	switch {
	case onmetalMachineNic.NetworkInterfaceRef != nil:
		ips, err := s.convertOnmetalIPSourcesToIPs(onmetalNic.NetworkInterface.Spec.IPs)
		if err != nil {
			return nil, err
		}

		return &ori.NetworkInterface{
			Name:      onmetalMachineNic.Name,
			NetworkId: onmetalNic.Network.Spec.ProviderID,
			Ips:       ips,
		}, nil
	default:
		return nil, fmt.Errorf("unrecognized onmetal machine network interface %#v", onmetalMachineNic)
	}
}

func (s *Server) convertAggregateOnmetalMachine(aggOnmetalMachine *AggregateOnmetalMachine) (*ori.Machine, error) {
	metadata, err := apiutils.GetObjectMetadata(aggOnmetalMachine.Machine)
	if err != nil {
		return nil, err
	}

	var ignitionData []byte
	if ignitionSecret := aggOnmetalMachine.IgnitionSecret; ignitionSecret != nil {
		ignitionData = ignitionSecret.Data[computev1alpha1.DefaultIgnitionKey]
	}

	var imageSpec *ori.ImageSpec
	if image := aggOnmetalMachine.Machine.Spec.Image; image != "" {
		imageSpec = &ori.ImageSpec{
			Image: image,
		}
	}

	volumes := make([]*ori.Volume, len(aggOnmetalMachine.Machine.Spec.Volumes))
	for i, onmetalMachineVolume := range aggOnmetalMachine.Machine.Spec.Volumes {
		onmetalVolume := aggOnmetalMachine.Volumes[onmetalMachineVolume.Name]
		volume, err := s.convertOnmetalVolume(onmetalMachineVolume, onmetalVolume)
		if err != nil {
			return nil, fmt.Errorf("error converting machine volume %s: %w", *onmetalMachineVolume.Device, err)
		}

		volumes[i] = volume
	}

	nics := make([]*ori.NetworkInterface, len(aggOnmetalMachine.Machine.Spec.NetworkInterfaces))
	for i, onmetalMachineNic := range aggOnmetalMachine.Machine.Spec.NetworkInterfaces {
		onmetalNic := aggOnmetalMachine.NetworkInterfaces[onmetalMachineNic.Name]
		nic, err := s.convertOnmetalNetworkInterfaceAttachment(onmetalMachineNic, onmetalNic)
		if err != nil {
			return nil, fmt.Errorf("error converting machine network interface %s: %w", onmetalMachineNic.Name, err)
		}

		nics[i] = nic
	}

	volumeAttachmentStates := make([]*ori.VolumeStatus, len(aggOnmetalMachine.Machine.Status.Volumes))
	for i, volume := range aggOnmetalMachine.Machine.Status.Volumes {
		volumeAttachmentStatus, err := s.convertOnmetalVolumeStatus(volume)
		if err != nil {
			return nil, fmt.Errorf("error converting machine volume status %s: %w", volume.Name, err)
		}

		volumeAttachmentStates[i] = volumeAttachmentStatus
	}

	networkInterfaceAttachmentStates := make([]*ori.NetworkInterfaceStatus, len(aggOnmetalMachine.Machine.Status.NetworkInterfaces))
	for i, networkInterface := range aggOnmetalMachine.Machine.Status.NetworkInterfaces {
		networkInterfaceAttachmentStatus, err := s.convertOnmetalNetworkInterfaceStatus(networkInterface)
		if err != nil {
			return nil, fmt.Errorf("error converting machine network interface status %s: %w", networkInterface.Name, err)
		}

		networkInterfaceAttachmentStates[i] = networkInterfaceAttachmentStatus
	}

	state, err := s.convertOnmetalMachineState(aggOnmetalMachine.Machine.Status.State)
	if err != nil {
		return nil, err
	}

	return &ori.Machine{
		Metadata: metadata,
		Spec: &ori.MachineSpec{
			Image:             imageSpec,
			Class:             aggOnmetalMachine.Machine.Spec.MachineClassRef.Name,
			IgnitionData:      ignitionData,
			Volumes:           volumes,
			NetworkInterfaces: nics,
		},
		Status: &ori.MachineStatus{
			ObservedGeneration: aggOnmetalMachine.Machine.Status.ObservedGeneration,
			State:              state,
			ImageRef:           "", // TODO: Fill
			Volumes:            volumeAttachmentStates,
			NetworkInterfaces:  networkInterfaceAttachmentStates,
		},
	}, nil
}
