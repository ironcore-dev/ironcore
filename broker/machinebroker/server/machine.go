// SPDX-FileCopyrightText: 2023 SAP SE or an SAP affiliate company and IronCore contributors
// SPDX-License-Identifier: Apache-2.0

package server

import (
	"fmt"

	computev1alpha1 "github.com/ironcore-dev/ironcore/api/compute/v1alpha1"
	networkingv1alpha1 "github.com/ironcore-dev/ironcore/api/networking/v1alpha1"
	storagev1alpha1 "github.com/ironcore-dev/ironcore/api/storage/v1alpha1"
	"github.com/ironcore-dev/ironcore/broker/machinebroker/apiutils"
	iri "github.com/ironcore-dev/ironcore/iri/apis/machine/v1alpha1"
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

var ironcoreMachineStateToMachineState = map[computev1alpha1.MachineState]iri.MachineState{
	computev1alpha1.MachineStatePending:     iri.MachineState_MACHINE_PENDING,
	computev1alpha1.MachineStateRunning:     iri.MachineState_MACHINE_RUNNING,
	computev1alpha1.MachineStateShutdown:    iri.MachineState_MACHINE_SUSPENDED,
	computev1alpha1.MachineStateTerminated:  iri.MachineState_MACHINE_TERMINATED,
	computev1alpha1.MachineStateTerminating: iri.MachineState_MACHINE_TERMINATING,
}

func (s *Server) convertIronCoreMachineState(state computev1alpha1.MachineState) (iri.MachineState, error) {
	if res, ok := ironcoreMachineStateToMachineState[state]; ok {
		return res, nil
	}
	return 0, fmt.Errorf("unknown ironcore machine state %q", state)
}

func (s *Server) convertIronCorePowerState(power computev1alpha1.Power) (iri.Power, error) {
	switch power {
	case computev1alpha1.PowerOn:
		return iri.Power_POWER_ON, nil
	case computev1alpha1.PowerOff:
		return iri.Power_POWER_OFF, nil
	default:
		return 0, fmt.Errorf("unknown power state: %q", power)
	}
}

var ironcoreNetworkInterfaceStateToNetworkInterfaceAttachmentState = map[computev1alpha1.NetworkInterfaceState]iri.NetworkInterfaceState{
	computev1alpha1.NetworkInterfaceStatePending:  iri.NetworkInterfaceState_NETWORK_INTERFACE_PENDING,
	computev1alpha1.NetworkInterfaceStateAttached: iri.NetworkInterfaceState_NETWORK_INTERFACE_ATTACHED,
}

func (s *Server) convertIronCoreNetworkInterfaceState(state computev1alpha1.NetworkInterfaceState) (iri.NetworkInterfaceState, error) {
	if res, ok := ironcoreNetworkInterfaceStateToNetworkInterfaceAttachmentState[state]; ok {
		return res, nil
	}
	return 0, fmt.Errorf("unknown ironcore network interface attachment state %q", state)
}

func (s *Server) convertIronCoreNetworkInterfaceStatus(status computev1alpha1.NetworkInterfaceStatus) (*iri.NetworkInterfaceStatus, error) {
	state, err := s.convertIronCoreNetworkInterfaceState(status.State)
	if err != nil {
		return nil, err
	}

	return &iri.NetworkInterfaceStatus{
		Name:   status.Name,
		Handle: status.Handle,
		State:  state,
	}, nil
}

var ironcoreVolumeStateToVolumeAttachmentState = map[computev1alpha1.VolumeState]iri.VolumeState{
	computev1alpha1.VolumeStatePending:  iri.VolumeState_VOLUME_PENDING,
	computev1alpha1.VolumeStateAttached: iri.VolumeState_VOLUME_ATTACHED,
}

func (s *Server) convertIronCoreVolumeState(state computev1alpha1.VolumeState) (iri.VolumeState, error) {
	if res, ok := ironcoreVolumeStateToVolumeAttachmentState[state]; ok {
		return res, nil
	}
	return 0, fmt.Errorf("unknown ironcore volume attachment state %q", state)
}

func (s *Server) convertIronCoreVolumeStatus(status computev1alpha1.VolumeStatus) (*iri.VolumeStatus, error) {
	state, err := s.convertIronCoreVolumeState(status.State)
	if err != nil {
		return nil, err
	}

	return &iri.VolumeStatus{
		Name:   status.Name,
		Handle: status.Handle,
		State:  state,
	}, nil
}

func (s *Server) convertIronCoreVolume(
	ironcoreMachineVolume computev1alpha1.Volume,
	ironcoreVolume *AggregateIronCoreVolume,
) (*iri.Volume, error) {
	var (
		connection *iri.VolumeConnection
		localDisk  *iri.LocalDisk
	)
	switch {
	case ironcoreMachineVolume.VolumeRef != nil:
		effectiveStorageBytes := ironcoreVolume.Volume.Spec.Resources.Storage().Value()
		if ironcoreVolume.Volume.Status.Resources != nil {
			effectiveStorageBytes = ironcoreVolume.Volume.Status.Resources.Storage().Value()
		}
		if access := ironcoreVolume.Volume.Status.Access; access != nil {
			var secretData map[string][]byte
			if access.SecretRef != nil {
				secretData = ironcoreVolume.AccessSecret.Data
			}

			connection = &iri.VolumeConnection{
				Driver:                access.Driver,
				Handle:                access.Handle,
				Attributes:            access.VolumeAttributes,
				SecretData:            secretData,
				EffectiveStorageBytes: effectiveStorageBytes,
			}
		}
	case ironcoreMachineVolume.LocalDisk != nil:
		var sizeBytes int64
		if sizeLimit := ironcoreMachineVolume.LocalDisk.SizeLimit; sizeLimit != nil {
			sizeBytes = sizeLimit.Value()
		}
		localDisk = &iri.LocalDisk{
			SizeBytes: sizeBytes,
			Image: &iri.ImageSpec{
				Image: ironcoreMachineVolume.LocalDisk.Image,
			},
		}
	default:
		return nil, fmt.Errorf("machine volume %#v does neither specify volume ref nor empty disk, nor local disk", ironcoreMachineVolume)
	}

	return &iri.Volume{
		Name:       ironcoreMachineVolume.Name,
		Device:     *ironcoreMachineVolume.Device,
		LocalDisk:  localDisk,
		Connection: connection,
	}, nil
}

func (s *Server) convertIronCoreNetworkInterfaceAttachment(
	ironcoreMachineNic computev1alpha1.NetworkInterface,
	ironcoreNic *AggregateIronCoreNetworkInterface,
) (*iri.NetworkInterface, error) {
	switch {
	case ironcoreMachineNic.NetworkInterfaceRef != nil:
		ips, err := s.convertIronCoreIPSourcesToIPs(ironcoreNic.NetworkInterface.Spec.IPs)
		if err != nil {
			return nil, err
		}

		return &iri.NetworkInterface{
			Name:       ironcoreMachineNic.Name,
			NetworkId:  ironcoreNic.Network.Spec.ProviderID,
			Ips:        ips,
			Attributes: ironcoreNic.NetworkInterface.Spec.Attributes,
		}, nil
	default:
		return nil, fmt.Errorf("unrecognized ironcore machine network interface %#v", ironcoreMachineNic)
	}
}

func (s *Server) convertAggregateIronCoreMachine(aggIronCoreMachine *AggregateIronCoreMachine) (*iri.Machine, error) {
	metadata, err := apiutils.GetObjectMetadata(aggIronCoreMachine.Machine)
	if err != nil {
		return nil, err
	}

	var ignitionData []byte
	if ignitionSecret := aggIronCoreMachine.IgnitionSecret; ignitionSecret != nil {
		ignitionData = ignitionSecret.Data[computev1alpha1.DefaultIgnitionKey]
	}

	power, err := s.convertIronCorePowerState(aggIronCoreMachine.Machine.Spec.Power)
	if err != nil {
		return nil, fmt.Errorf("error converting power state: %w", err)
	}

	volumes := make([]*iri.Volume, len(aggIronCoreMachine.Machine.Spec.Volumes))
	for i, ironcoreMachineVolume := range aggIronCoreMachine.Machine.Spec.Volumes {
		ironcoreVolume := aggIronCoreMachine.Volumes[ironcoreMachineVolume.Name]
		volume, err := s.convertIronCoreVolume(ironcoreMachineVolume, ironcoreVolume)
		if err != nil {
			return nil, fmt.Errorf("error converting machine volume %s: %w", *ironcoreMachineVolume.Device, err)
		}

		volumes[i] = volume
	}

	nics := make([]*iri.NetworkInterface, len(aggIronCoreMachine.Machine.Spec.NetworkInterfaces))
	for i, ironcoreMachineNic := range aggIronCoreMachine.Machine.Spec.NetworkInterfaces {
		ironcoreNic := aggIronCoreMachine.NetworkInterfaces[ironcoreMachineNic.Name]
		nic, err := s.convertIronCoreNetworkInterfaceAttachment(ironcoreMachineNic, ironcoreNic)
		if err != nil {
			return nil, fmt.Errorf("error converting machine network interface %s: %w", ironcoreMachineNic.Name, err)
		}

		nics[i] = nic
	}

	volumeAttachmentStates := make([]*iri.VolumeStatus, len(aggIronCoreMachine.Machine.Status.Volumes))
	for i, volume := range aggIronCoreMachine.Machine.Status.Volumes {
		volumeAttachmentStatus, err := s.convertIronCoreVolumeStatus(volume)
		if err != nil {
			return nil, fmt.Errorf("error converting machine volume status %s: %w", volume.Name, err)
		}

		volumeAttachmentStates[i] = volumeAttachmentStatus
	}

	networkInterfaceAttachmentStates := make([]*iri.NetworkInterfaceStatus, len(aggIronCoreMachine.Machine.Status.NetworkInterfaces))
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

	return &iri.Machine{
		Metadata: metadata,
		Spec: &iri.MachineSpec{
			Power:             power,
			Class:             aggIronCoreMachine.Machine.Spec.MachineClassRef.Name,
			IgnitionData:      ignitionData,
			Volumes:           volumes,
			NetworkInterfaces: nics,
		},
		Status: &iri.MachineStatus{
			ObservedGeneration: aggIronCoreMachine.Machine.Status.ObservedGeneration,
			State:              state,
			ImageRef:           "", // TODO: Fill
			Volumes:            volumeAttachmentStates,
			NetworkInterfaces:  networkInterfaceAttachmentStates,
		},
	}, nil
}
