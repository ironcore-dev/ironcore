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

package controllers

import (
	"context"
	"errors"
	"fmt"

	commonv1alpha1 "github.com/onmetal/onmetal-api/apis/common/v1alpha1"
	computev1alpha1 "github.com/onmetal/onmetal-api/apis/compute/v1alpha1"
	networkingv1alpha1 "github.com/onmetal/onmetal-api/apis/networking/v1alpha1"
	storagev1alpha1 "github.com/onmetal/onmetal-api/apis/storage/v1alpha1"
	ori "github.com/onmetal/onmetal-api/ori/apis/runtime/v1alpha1"
	corev1 "k8s.io/api/core/v1"
	apierrors "k8s.io/apimachinery/pkg/api/errors"
	"k8s.io/apimachinery/pkg/runtime/schema"
	"sigs.k8s.io/controller-runtime/pkg/client"
)

type dependencyNotReadyError struct {
	group    string
	resource string
	name     string

	cause error
}

func (d *dependencyNotReadyError) Error() string {
	return fmt.Sprintf("dependency %s %s not ready: %v",
		schema.GroupResource{
			Group:    d.group,
			Resource: d.resource,
		},
		d.name,
		d.cause,
	)
}

func (d *dependencyNotReadyError) Unwrap() error {
	return d.cause
}

func IsDependencyNotReadyError(err error) bool {
	return errors.As(err, new(*dependencyNotReadyError))
}

func IgnoreDependencyNotReadyError(err error) error {
	if IsDependencyNotReadyError(err) {
		return nil
	}
	return err
}

func NewDependencyNotReadyError(gr schema.GroupResource, name string, cause error) error {
	return &dependencyNotReadyError{
		group:    gr.Group,
		resource: gr.Resource,
		name:     name,
		cause:    cause,
	}
}

func (r *MachineReconciler) getORIMachineMetadata(machine *computev1alpha1.Machine) *ori.MachineMetadata {
	return &ori.MachineMetadata{
		Namespace: machine.Namespace,
		Name:      machine.Name,
		Uid:       string(machine.UID),
	}
}

func (r *MachineReconciler) getORIMachineResources(ctx context.Context, machine *computev1alpha1.Machine) (*ori.MachineResources, error) {
	machineClass := &computev1alpha1.MachineClass{}
	machineClassKey := client.ObjectKey{Name: machine.Spec.MachineClassRef.Name}
	if err := r.Get(ctx, machineClassKey, machineClass); err != nil {
		err = fmt.Errorf("error getting machine class %s: %w", machineClassKey, err)
		if !apierrors.IsNotFound(err) {
			return nil, err
		}

		return nil, NewDependencyNotReadyError(
			computev1alpha1.Resource("machineclasses"),
			machineClassKey.Name,
			err,
		)
	}

	cpuCount := machineClass.Capabilities.Cpu().Value()
	memoryBytes := machineClass.Capabilities.Memory().Value()

	return &ori.MachineResources{
		CpuCount:    int32(cpuCount),
		MemoryBytes: uint64(memoryBytes),
	}, nil
}

func (r *MachineReconciler) getORIIgnitionConfig(ctx context.Context, machine *computev1alpha1.Machine, ignitionRef *commonv1alpha1.SecretKeySelector) (*ori.IgnitionConfig, error) {
	ignitionSecret := &corev1.Secret{}
	ignitionSecretKey := client.ObjectKey{Namespace: machine.Namespace, Name: ignitionRef.Name}
	if err := r.Get(ctx, ignitionSecretKey, ignitionSecret); err != nil {
		err = fmt.Errorf("error getting ignition secret %s: %w", ignitionSecretKey, err)
		if !apierrors.IsNotFound(err) {
			return nil, err
		}

		return nil, NewDependencyNotReadyError(
			corev1.Resource("secrets"),
			ignitionSecretKey.String(),
			err,
		)
	}

	ignitionKey := ignitionRef.Key
	if ignitionKey == "" {
		ignitionKey = computev1alpha1.DefaultIgnitionKey
	}

	data, ok := ignitionSecret.Data[ignitionKey]
	if !ok {
		err := fmt.Errorf("ignition secret %s has no data at key %s", ignitionSecretKey, ignitionKey)
		return nil, NewDependencyNotReadyError(
			corev1.Resource("secrets"),
			ignitionSecretKey.String(),
			err,
		)
	}

	return &ori.IgnitionConfig{
		Data: data,
	}, nil
}

func (r *MachineReconciler) getORIVolumeAccessConfig(ctx context.Context, volume *storagev1alpha1.Volume) (*ori.VolumeAccessConfig, error) {
	if volume.Status.State != storagev1alpha1.VolumeStateAvailable {
		return nil, NewDependencyNotReadyError(
			storagev1alpha1.Resource("volumes"),
			client.ObjectKeyFromObject(volume).String(),
			fmt.Errorf("volume is not yet in state available"),
		)
	}

	access := volume.Status.Access
	if access == nil {
		return nil, NewDependencyNotReadyError(
			storagev1alpha1.Resource("volumes"),
			client.ObjectKeyFromObject(volume).String(),
			fmt.Errorf("volume access is not yet populated"),
		)
	}

	var secretData map[string][]byte
	if secretRef := access.SecretRef; secretRef != nil {
		secret := &corev1.Secret{}
		secretKey := client.ObjectKey{Namespace: volume.Namespace, Name: secretRef.Name}
		if err := r.Get(ctx, secretKey, secret); err != nil {
			err = fmt.Errorf("error getting volume secret %s: %w", secretKey, err)
			if !apierrors.IsNotFound(err) {
				return nil, err
			}

			return nil, NewDependencyNotReadyError(
				corev1.Resource("secrets"),
				secretKey.String(),
				err,
			)
		}

		secretData = secret.Data
	}

	return &ori.VolumeAccessConfig{
		Driver:     access.Driver,
		Handle:     access.Handle,
		Attributes: access.VolumeAttributes,
		SecretData: secretData,
	}, nil
}

// checkReferencedVolumeBoundToMachine checks if the referenced volume is bound to the machine.
//
// It is assumed that the caller verified that the machine points to the volume.
func (r *MachineReconciler) checkReferencedVolumeBoundToMachine(machine *computev1alpha1.Machine, machineVolumeName string, referencedVolume *storagev1alpha1.Volume) error {
	if volumePhase := referencedVolume.Status.Phase; volumePhase != storagev1alpha1.VolumePhaseBound {
		return NewDependencyNotReadyError(
			storagev1alpha1.Resource("volumes"),
			client.ObjectKeyFromObject(referencedVolume).String(),
			fmt.Errorf("volume phase %q is not bound", volumePhase),
		)
	}

	claimRef := referencedVolume.Spec.ClaimRef
	if claimRef == nil {
		return NewDependencyNotReadyError(
			storagev1alpha1.Resource("volumes"),
			client.ObjectKeyFromObject(referencedVolume).String(),
			fmt.Errorf("volume does not reference any claimer"),
		)
	}

	if claimRef.Name != machine.Name || claimRef.UID != machine.UID {
		return NewDependencyNotReadyError(
			storagev1alpha1.Resource("volumes"),
			client.ObjectKeyFromObject(referencedVolume).String(),
			fmt.Errorf("volume references different claimer %s (uid %s)", claimRef.Name, claimRef.UID),
		)
	}

	for _, status := range machine.Status.Volumes {
		if status.Name == machineVolumeName {
			if status.Phase != computev1alpha1.VolumePhaseBound {
				return NewDependencyNotReadyError(
					computev1alpha1.Resource("machines"),
					client.ObjectKeyFromObject(machine).String(),
					fmt.Errorf("machine volume phase %q is not bound", status.Phase),
				)
			}
			return nil
		}
	}
	return NewDependencyNotReadyError(
		computev1alpha1.Resource("machines"),
		client.ObjectKeyFromObject(machine).String(),
		fmt.Errorf("machine is not yet bound to volume"),
	)
}

func (r *MachineReconciler) getORIVolumeConfig(ctx context.Context, machine *computev1alpha1.Machine, machineVolume *computev1alpha1.Volume) (*ori.VolumeConfig, error) {
	var (
		emptyDiskConfig    *ori.EmptyDiskConfig
		volumeAccessConfig *ori.VolumeAccessConfig
	)

	getReferencedVolumeORIAccessConfig := func(volumeKey client.ObjectKey) (*ori.VolumeAccessConfig, error) {
		volume := &storagev1alpha1.Volume{}
		if err := r.Get(ctx, volumeKey, volume); err != nil {
			return nil, fmt.Errorf("error getting volume %s: %w", volumeKey, err)
		}

		if err := r.checkReferencedVolumeBoundToMachine(machine, machineVolume.Name, volume); err != nil {
			return nil, err
		}

		volumeAccessConfig, err := r.getORIVolumeAccessConfig(ctx, volume)
		if err != nil {
			return nil, fmt.Errorf("error getting volume %s access config: %w", volumeKey, err)
		}
		return volumeAccessConfig, nil
	}

	switch {
	case machineVolume.Ephemeral != nil:
		volumeKey := client.ObjectKey{Namespace: machine.Namespace, Name: computev1alpha1.MachineEphemeralVolumeName(machine.Name, machineVolume.Name)}

		var err error
		volumeAccessConfig, err = getReferencedVolumeORIAccessConfig(volumeKey)
		if err != nil {
			return nil, err
		}
	case machineVolume.VolumeRef != nil:
		volumeKey := client.ObjectKey{Namespace: machine.Namespace, Name: machineVolume.VolumeRef.Name}

		var err error
		volumeAccessConfig, err = getReferencedVolumeORIAccessConfig(volumeKey)
		if err != nil {
			return nil, err
		}
	case machineVolume.EmptyDisk != nil:
		var sizeLimitBytes uint64
		if sizeLimit := machineVolume.EmptyDisk.SizeLimit; sizeLimit != nil {
			sizeLimitBytes = uint64(sizeLimit.Value())
		}
		emptyDiskConfig = &ori.EmptyDiskConfig{SizeLimitBytes: sizeLimitBytes}
	default:
		return nil, fmt.Errorf("unsupported volume %#v", machineVolume)
	}

	return &ori.VolumeConfig{
		Name:      machineVolume.Name,
		Device:    machineVolume.Device,
		Access:    volumeAccessConfig,
		EmptyDisk: emptyDiskConfig,
	}, nil
}

// checkReferencedNetworkInterfaceBoundToMachine checks if the referenced network interface is bound to the machine.
//
// It is assumed that the caller verified that the machine points to the network interface.
func (r *MachineReconciler) checkReferencedNetworkInterfaceBoundToMachine(machine *computev1alpha1.Machine, machineNetworkInterfaceName string, referencedNetworkInterface *networkingv1alpha1.NetworkInterface) error {
	if networkInterfacePhase := referencedNetworkInterface.Status.Phase; networkInterfacePhase != networkingv1alpha1.NetworkInterfacePhaseBound {
		return NewDependencyNotReadyError(
			networkingv1alpha1.Resource("networkinterfaces"),
			client.ObjectKeyFromObject(referencedNetworkInterface).String(),
			fmt.Errorf("network interface phase %q is not bound", networkInterfacePhase),
		)
	}

	claimRef := referencedNetworkInterface.Spec.MachineRef
	if claimRef == nil {
		return NewDependencyNotReadyError(
			networkingv1alpha1.Resource("networkinterfaces"),
			client.ObjectKeyFromObject(referencedNetworkInterface).String(),
			fmt.Errorf("network interface does not reference any claimer"),
		)
	}

	if claimRef.Name != machine.Name || claimRef.UID != machine.UID {
		return NewDependencyNotReadyError(
			networkingv1alpha1.Resource("networkinterfaces"),
			client.ObjectKeyFromObject(referencedNetworkInterface).String(),
			fmt.Errorf("network interface references different claimer %s (uid %s)", claimRef.Name, claimRef.UID),
		)
	}

	for _, status := range machine.Status.NetworkInterfaces {
		if status.Name == machineNetworkInterfaceName {
			if status.Phase != computev1alpha1.NetworkInterfacePhaseBound {
				return NewDependencyNotReadyError(
					computev1alpha1.Resource("machines"),
					client.ObjectKeyFromObject(machine).String(),
					fmt.Errorf("machine network interface phase %q is not bound", status.Phase),
				)
			}
			return nil
		}
	}
	return NewDependencyNotReadyError(
		computev1alpha1.Resource("machines"),
		client.ObjectKeyFromObject(machine).String(),
		fmt.Errorf("machine is not yet bound to network interface"),
	)
}

func (r *MachineReconciler) getORINetworkInterfaceConfig(
	ctx context.Context,
	machine *computev1alpha1.Machine,
	machineNetworkInterface *computev1alpha1.NetworkInterface,
) (*ori.NetworkInterfaceConfig, error) {
	getReferencedNetworkInterfaceORINetworkInterfaceConfig := func(networkInterfaceKey client.ObjectKey) (*ori.NetworkInterfaceConfig, error) {
		networkInterface := &networkingv1alpha1.NetworkInterface{}
		if err := r.Get(ctx, networkInterfaceKey, networkInterface); err != nil {
			if !apierrors.IsNotFound(err) {
				return nil, fmt.Errorf("error getting network interface %s: %w", networkInterfaceKey, err)
			}
			return nil, NewDependencyNotReadyError(
				networkingv1alpha1.Resource("networkinterfaces"),
				client.ObjectKeyFromObject(networkInterface).String(),
				err,
			)
		}

		if err := r.checkReferencedNetworkInterfaceBoundToMachine(machine, machineNetworkInterface.Name, networkInterface); err != nil {
			return nil, err
		}

		// TODO: The network interface should expose a ready state to avoid this lengthy check
		if len(networkInterface.Spec.IPFamilies) != len(networkInterface.Status.IPs) || networkInterface.Status.NetworkHandle == "" {
			return nil, NewDependencyNotReadyError(
				networkingv1alpha1.Resource("networkinterfaces"),
				client.ObjectKeyFromObject(networkInterface).String(),
				fmt.Errorf("not all ips have been allocated"),
			)
		}

		networkConfig := &ori.NetworkConfig{
			Handle: networkInterface.Status.NetworkHandle,
		}

		var virtualIPConfig *ori.VirtualIPConfig
		if virtualIP := networkInterface.Status.VirtualIP; virtualIP != nil {
			virtualIPConfig = &ori.VirtualIPConfig{
				Ip: virtualIP.Addr.String(),
			}
		}

		return &ori.NetworkInterfaceConfig{
			Name:      machineNetworkInterface.Name,
			Network:   networkConfig,
			Ips:       CommonV1Alpha1IPsToIPStrings(networkInterface.Status.IPs),
			VirtualIp: virtualIPConfig,
		}, nil
	}

	switch {
	case machineNetworkInterface.Ephemeral != nil:
		networkInterfaceKey := client.ObjectKey{Namespace: machine.Namespace, Name: computev1alpha1.MachineEphemeralNetworkInterfaceName(machine.Name, machineNetworkInterface.Name)}
		return getReferencedNetworkInterfaceORINetworkInterfaceConfig(networkInterfaceKey)
	case machineNetworkInterface.NetworkInterfaceRef != nil:
		networkInterfaceKey := client.ObjectKey{Namespace: machine.Namespace, Name: machineNetworkInterface.NetworkInterfaceRef.Name}
		return getReferencedNetworkInterfaceORINetworkInterfaceConfig(networkInterfaceKey)
	default:
		return nil, fmt.Errorf("unsupported network interface %#v", machineNetworkInterface)
	}
}

func CommonV1Alpha1IPsToIPStrings(ips []commonv1alpha1.IP) []string {
	res := make([]string, len(ips))
	for i, ip := range ips {
		res[i] = ip.Addr.String()
	}
	return res
}

var oriMachineStateToComputeV1Alpha1MachineState = map[ori.MachineState]computev1alpha1.MachineState{
	ori.MachineState_MACHINE_PENDING:  computev1alpha1.MachineStatePending,
	ori.MachineState_MACHINE_RUNNING:  computev1alpha1.MachineStateRunning,
	ori.MachineState_MACHINE_SHUTDOWN: computev1alpha1.MachineStateShutdown,
	ori.MachineState_MACHINE_ERROR:    computev1alpha1.MachineStateError,
	ori.MachineState_MACHINE_UNKNOWN:  computev1alpha1.MachineStateUnknown,
}

func (r *MachineReconciler) oriMachineStateToComputeV1Alpha1MachineState(oriState ori.MachineState) computev1alpha1.MachineState {
	if mapped, ok := oriMachineStateToComputeV1Alpha1MachineState[oriState]; ok {
		return mapped
	}
	return computev1alpha1.MachineStateUnknown
}

var oriVolumeStateToComputeV1Alpha1VolumeState = map[ori.VolumeState]computev1alpha1.VolumeState{
	ori.VolumeState_VOLUME_ATTACHED: computev1alpha1.VolumeStateAttached,
	ori.VolumeState_VOLUME_DETACHED: computev1alpha1.VolumeStateDetached,
	ori.VolumeState_VOLUME_ERROR:    computev1alpha1.VolumeStateError,
	ori.VolumeState_VOLUME_PENDING:  computev1alpha1.VolumeStatePending,
}

func (r *MachineReconciler) oriVolumeStateToComputeV1Alpha1VolumeState(oriState ori.VolumeState) computev1alpha1.VolumeState {
	return oriVolumeStateToComputeV1Alpha1VolumeState[oriState]
}
