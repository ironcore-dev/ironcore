// SPDX-FileCopyrightText: 2023 SAP SE or an SAP affiliate company and IronCore contributors
// SPDX-License-Identifier: Apache-2.0

package v1alpha1

import (
	"fmt"
)

// MachineEphemeralNetworkInterfaceName returns the name of a NetworkInterface for an
// ephemeral machine network interface.
func MachineEphemeralNetworkInterfaceName(machineName, machineNicName string) string {
	return fmt.Sprintf("%s-%s", machineName, machineNicName)
}

// MachineNetworkInterfaceName returns the name of the NetworkInterface for a machine network interface.
func MachineNetworkInterfaceName(machineName string, networkInterface NetworkInterface) string {
	switch {
	case networkInterface.NetworkInterfaceRef != nil:
		return networkInterface.NetworkInterfaceRef.Name
	case networkInterface.Ephemeral != nil:
		return MachineEphemeralNetworkInterfaceName(machineName, networkInterface.Name)
	default:
		return ""
	}
}

// MachineNetworkInterfaceNames returns all NetworkInterface names of a machine.
func MachineNetworkInterfaceNames(machine *Machine) []string {
	var names []string
	for _, nic := range machine.Spec.NetworkInterfaces {
		if name := MachineNetworkInterfaceName(machine.Name, nic); name != "" {
			names = append(names, name)
		}
	}
	return names
}

// MachineEphemeralVolumeName returns the name of a Volume for an ephemeral machine volume.
func MachineEphemeralVolumeName(machineName, machineVolumeName string) string {
	return fmt.Sprintf("%s-%s", machineName, machineVolumeName)
}

// MachineVolumeName returns the name of the Volume for a machine volume.
func MachineVolumeName(machineName string, volume Volume) string {
	switch {
	case volume.VolumeRef != nil:
		return volume.VolumeRef.Name
	case volume.Ephemeral != nil:
		return MachineEphemeralVolumeName(machineName, volume.Name)
	default:
		return ""
	}
}

// MachineVolumeNames returns all Volume names of a machine.
func MachineVolumeNames(machine *Machine) []string {
	var names []string
	for _, volume := range machine.Spec.Volumes {
		if name := MachineVolumeName(machine.Name, volume); name != "" {
			names = append(names, name)
		}
	}
	return names
}

// MachineSecretNames returns all secret names of a machine.
func MachineSecretNames(machine *Machine) []string {
	var names []string

	if imagePullSecretRef := machine.Spec.ImagePullSecretRef; imagePullSecretRef != nil {
		names = append(names, imagePullSecretRef.Name)
	}

	if ignitionRef := machine.Spec.IgnitionRef; ignitionRef != nil {
		names = append(names, ignitionRef.Name)
	}

	return names
}
