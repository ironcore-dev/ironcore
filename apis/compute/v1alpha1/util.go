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
