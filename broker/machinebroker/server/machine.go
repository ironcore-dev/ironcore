// Copyright 2022 IronCore authors
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

	computev1alpha1 "github.com/ironcore-dev/ironcore/api/compute/v1alpha1"
	networkingv1alpha1 "github.com/ironcore-dev/ironcore/api/networking/v1alpha1"
	storagev1alpha1 "github.com/ironcore-dev/ironcore/api/storage/v1alpha1"
	"github.com/ironcore-dev/ironcore/broker/machinebroker/apiutils"
	ori "github.com/ironcore-dev/ironcore/ori/apis/machine/v1alpha1"
	corev1 "k8s.io/api/core/v1"
)

type AggregateIronCoreMachine struct {
	IgnitionSecret *corev1.Secret
	Machine        *computev1alpha1.Machine
	// NetworkInterfaces is a mapping of machine network interface name to actual network interface.
	NetworkInterfaces map[string]*AggregateIronCoreNetworkInterface
	// Volumes is a mapping of machine volume name to actual volume.
	Volumes map[string]*AggregateIronCoreVolume
}

type AggregateIronCoreVolume struct {
	Volume       *storagev1alpha1.Volume
	AccessSecret *corev1.Secret
}

type AggregateIronCoreNetworkInterface struct {
	Network          *networkingv1alpha1.Network
	NetworkInterface *networkingv1alpha1.NetworkInterface
}

var ironcoreMachineStateToMachineState = map[computev1alpha1.MachineState]ori.MachineState{
	computev1alpha1.MachineStatePending:    ori.MachineState_MACHINE_PENDING,
	computev1alpha1.MachineStateRunning:    ori.MachineState_MACHINE_RUNNING,
	computev1alpha1.MachineStateShutdown:   ori.MachineState_MACHINE_SUSPENDED,
	computev1alpha1.MachineStateTerminated: ori.MachineState_MACHINE_TERMINATED,
}

func (s *Server) convertIronCoreMachineState(state computev1alpha1.MachineState) (ori.MachineState, error) {
	if res, ok := ironcoreMachineStateToMachineState[state]; ok {
		return res, nil
	}
	return 0, fmt.Errorf("unknown ironcore machine state %q", state)
}

var ironcoreNetworkInterfaceStateToNetworkInterfaceAttachmentState = map[computev1alpha1.NetworkInterfaceState]ori.NetworkInterfaceState{
	computev1alpha1.NetworkInterfaceStatePending:  ori.NetworkInterfaceState_NETWORK_INTERFACE_PENDING,
	computev1alpha1.NetworkInterfaceStateAttached: ori.NetworkInterfaceState_NETWORK_INTERFACE_ATTACHED,
}

func (s *Server) convertIronCoreNetworkInterfaceState(state computev1alpha1.NetworkInterfaceState) (ori.NetworkInterfaceState, error) {
	if res, ok := ironcoreNetworkInterfaceStateToNetworkInterfaceAttachmentState[state]; ok {
		return res, nil
	}
	return 0, fmt.Errorf("unknown ironcore network interface attachment state %q", state)
}

func (s *Server) convertIronCoreNetworkInterfaceStatus(status computev1alpha1.NetworkInterfaceStatus) (*ori.NetworkInterfaceStatus, error) {
	state, err := s.convertIronCoreNetworkInterfaceState(status.State)
	if err != nil {
		return nil, err
	}

	return &ori.NetworkInterfaceStatus{
		Name:   status.Name,
		Handle: status.Handle,
		State:  state,
	}, nil
}

var ironcoreVolumeStateToVolumeAttachmentState = map[computev1alpha1.VolumeState]ori.VolumeState{
	computev1alpha1.VolumeStatePending:  ori.VolumeState_VOLUME_PENDING,
	computev1alpha1.VolumeStateAttached: ori.VolumeState_VOLUME_ATTACHED,
}

func (s *Server) convertIronCoreVolumeState(state computev1alpha1.VolumeState) (ori.VolumeState, error) {
	if res, ok := ironcoreVolumeStateToVolumeAttachmentState[state]; ok {
		return res, nil
	}
	return 0, fmt.Errorf("unknown ironcore volume attachment state %q", state)
}

func (s *Server) convertIronCoreVolumeStatus(status computev1alpha1.VolumeStatus) (*ori.VolumeStatus, error) {
	state, err := s.convertIronCoreVolumeState(status.State)
	if err != nil {
		return nil, err
	}

	return &ori.VolumeStatus{
		Name:   status.Name,
		Handle: status.Handle,
		State:  state,
	}, nil
}

func (s *Server) convertIronCoreVolume(
	ironcoreMachineVolume computev1alpha1.Volume,
	ironcoreVolume *AggregateIronCoreVolume,
) (*ori.Volume, error) {
	var (
		connection *ori.VolumeConnection
		emptyDisk  *ori.EmptyDisk
	)
	switch {
	case ironcoreMachineVolume.VolumeRef != nil:
		if access := ironcoreVolume.Volume.Status.Access; access != nil {
			var secretData map[string][]byte
			if access.SecretRef != nil {
				secretData = ironcoreVolume.AccessSecret.Data
			}

			connection = &ori.VolumeConnection{
				Driver:     access.Driver,
				Handle:     access.Handle,
				Attributes: access.VolumeAttributes,
				SecretData: secretData,
			}
		}
	case ironcoreMachineVolume.EmptyDisk != nil:
		var sizeBytes int64
		if sizeLimit := ironcoreMachineVolume.EmptyDisk.SizeLimit; sizeLimit != nil {
			sizeBytes = sizeLimit.Value()
		}
		emptyDisk = &ori.EmptyDisk{
			SizeBytes: sizeBytes,
		}
	default:
		return nil, fmt.Errorf("machine volume %#v does neither specify volume ref nor empty disk", ironcoreMachineVolume)
	}

	return &ori.Volume{
		Name:       ironcoreMachineVolume.Name,
		Device:     *ironcoreMachineVolume.Device,
		EmptyDisk:  emptyDisk,
		Connection: connection,
	}, nil
}

func (s *Server) convertIronCoreNetworkInterfaceAttachment(
	ironcoreMachineNic computev1alpha1.NetworkInterface,
	ironcoreNic *AggregateIronCoreNetworkInterface,
) (*ori.NetworkInterface, error) {
	switch {
	case ironcoreMachineNic.NetworkInterfaceRef != nil:
		ips, err := s.convertIronCoreIPSourcesToIPs(ironcoreNic.NetworkInterface.Spec.IPs)
		if err != nil {
			return nil, err
		}

		return &ori.NetworkInterface{
			Name:       ironcoreMachineNic.Name,
			NetworkId:  ironcoreNic.Network.Spec.ProviderID,
			Ips:        ips,
			Attributes: ironcoreNic.NetworkInterface.Spec.Attributes,
		}, nil
	default:
		return nil, fmt.Errorf("unrecognized ironcore machine network interface %#v", ironcoreMachineNic)
	}
}

func (s *Server) convertAggregateIronCoreMachine(aggIronCoreMachine *AggregateIronCoreMachine) (*ori.Machine, error) {
	metadata, err := apiutils.GetObjectMetadata(aggIronCoreMachine.Machine)
	if err != nil {
		return nil, err
	}

	var ignitionData []byte
	if ignitionSecret := aggIronCoreMachine.IgnitionSecret; ignitionSecret != nil {
		ignitionData = ignitionSecret.Data[computev1alpha1.DefaultIgnitionKey]
	}

	var imageSpec *ori.ImageSpec
	if image := aggIronCoreMachine.Machine.Spec.Image; image != "" {
		imageSpec = &ori.ImageSpec{
			Image: image,
		}
	}

	volumes := make([]*ori.Volume, len(aggIronCoreMachine.Machine.Spec.Volumes))
	for i, ironcoreMachineVolume := range aggIronCoreMachine.Machine.Spec.Volumes {
		ironcoreVolume := aggIronCoreMachine.Volumes[ironcoreMachineVolume.Name]
		volume, err := s.convertIronCoreVolume(ironcoreMachineVolume, ironcoreVolume)
		if err != nil {
			return nil, fmt.Errorf("error converting machine volume %s: %w", *ironcoreMachineVolume.Device, err)
		}

		volumes[i] = volume
	}

	nics := make([]*ori.NetworkInterface, len(aggIronCoreMachine.Machine.Spec.NetworkInterfaces))
	for i, ironcoreMachineNic := range aggIronCoreMachine.Machine.Spec.NetworkInterfaces {
		ironcoreNic := aggIronCoreMachine.NetworkInterfaces[ironcoreMachineNic.Name]
		nic, err := s.convertIronCoreNetworkInterfaceAttachment(ironcoreMachineNic, ironcoreNic)
		if err != nil {
			return nil, fmt.Errorf("error converting machine network interface %s: %w", ironcoreMachineNic.Name, err)
		}

		nics[i] = nic
	}

	volumeAttachmentStates := make([]*ori.VolumeStatus, len(aggIronCoreMachine.Machine.Status.Volumes))
	for i, volume := range aggIronCoreMachine.Machine.Status.Volumes {
		volumeAttachmentStatus, err := s.convertIronCoreVolumeStatus(volume)
		if err != nil {
			return nil, fmt.Errorf("error converting machine volume status %s: %w", volume.Name, err)
		}

		volumeAttachmentStates[i] = volumeAttachmentStatus
	}

	networkInterfaceAttachmentStates := make([]*ori.NetworkInterfaceStatus, len(aggIronCoreMachine.Machine.Status.NetworkInterfaces))
	for i, networkInterface := range aggIronCoreMachine.Machine.Status.NetworkInterfaces {
		networkInterfaceAttachmentStatus, err := s.convertIronCoreNetworkInterfaceStatus(networkInterface)
		if err != nil {
			return nil, fmt.Errorf("error converting machine network interface status %s: %w", networkInterface.Name, err)
		}

		networkInterfaceAttachmentStates[i] = networkInterfaceAttachmentStatus
	}

	state, err := s.convertIronCoreMachineState(aggIronCoreMachine.Machine.Status.State)
	if err != nil {
		return nil, err
	}

	return &ori.Machine{
		Metadata: metadata,
		Spec: &ori.MachineSpec{
			Image:             imageSpec,
			Class:             aggIronCoreMachine.Machine.Spec.MachineClassRef.Name,
			IgnitionData:      ignitionData,
			Volumes:           volumes,
			NetworkInterfaces: nics,
		},
		Status: &ori.MachineStatus{
			ObservedGeneration: aggIronCoreMachine.Machine.Status.ObservedGeneration,
			State:              state,
			ImageRef:           "", // TODO: Fill
			Volumes:            volumeAttachmentStates,
			NetworkInterfaces:  networkInterfaceAttachmentStates,
		},
	}, nil
}
