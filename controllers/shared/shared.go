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

package shared

import (
	"context"
	"fmt"

	computev1alpha1 "github.com/onmetal/onmetal-api/apis/compute/v1alpha1"
	v1alpha12 "github.com/onmetal/onmetal-api/apis/networking/v1alpha1"
	"k8s.io/apimachinery/pkg/util/sets"
	"sigs.k8s.io/controller-runtime/pkg/client"
)

const (
	MachineSpecNetworkInterfaceNamesField = "machine-spec-network-interface-names"
	MachineSpecVolumeNamesField           = "machine-spec-volume-names"

	NetworkInterfaceVirtualIPNames = "networkinterface-virtual-ip-names"
)

func MachineEphemeralNetworkInterfaceName(machineName, machineNicName string) string {
	return fmt.Sprintf("%s-%s", machineName, machineNicName)
}

func MachineEphemeralVolumeName(machineName, machineVolumeName string) string {
	return fmt.Sprintf("%s-%s", machineName, machineVolumeName)
}

func MachineSpecNetworkInterfaceNames(machine *computev1alpha1.Machine) sets.String {
	names := sets.NewString()
	for _, machineNic := range machine.Spec.NetworkInterfaces {
		switch {
		case machineNic.NetworkInterfaceRef != nil:
			names.Insert(machineNic.NetworkInterfaceRef.Name)
		case machineNic.Ephemeral != nil:
			names.Insert(MachineEphemeralNetworkInterfaceName(machine.Name, machineNic.Name))
		}
	}
	return names
}

func MachineSpecVolumeNames(machine *computev1alpha1.Machine) sets.String {
	names := sets.NewString()
	for _, machineVolume := range machine.Spec.Volumes {
		switch {
		case machineVolume.VolumeRef != nil:
			names.Insert(machineVolume.VolumeRef.Name)
		case machineVolume.Ephemeral != nil:
			names.Insert(MachineEphemeralVolumeName(machine.Name, machineVolume.Name))
		}
	}
	return names
}

func SetupMachineSpecNetworkInterfaceNamesFieldIndexer(ctx context.Context, indexer client.FieldIndexer) error {
	return indexer.IndexField(ctx, &computev1alpha1.Machine{}, MachineSpecNetworkInterfaceNamesField, func(obj client.Object) []string {
		machine := obj.(*computev1alpha1.Machine)
		if names := MachineSpecNetworkInterfaceNames(machine); len(names) > 0 {
			return names.UnsortedList()
		}
		return []string{""}
	})
}

func SetupMachineSpecVolumeNamesFieldIndexer(ctx context.Context, indexer client.FieldIndexer) error {
	return indexer.IndexField(ctx, &computev1alpha1.Machine{}, MachineSpecVolumeNamesField, func(obj client.Object) []string {
		machine := obj.(*computev1alpha1.Machine)
		if names := MachineSpecVolumeNames(machine); len(names) > 0 {
			return names.UnsortedList()
		}
		return []string{""}
	})
}

func NetworkInterfaceVirtualIPName(nicName string, vipSource v1alpha12.VirtualIPSource) string {
	switch {
	case vipSource.VirtualIPRef != nil:
		return vipSource.VirtualIPRef.Name
	case vipSource.Ephemeral != nil:
		return nicName
	default:
		return ""
	}
}

func NetworkInterfaceEphemeralIPName(nicName string, idx int) string {
	return fmt.Sprintf("%s-%d", nicName, idx)
}

func SetupNetworkInterfaceVirtualIPNameFieldIndexer(ctx context.Context, indexer client.FieldIndexer) error {
	return indexer.IndexField(ctx, &v1alpha12.NetworkInterface{}, NetworkInterfaceVirtualIPNames, func(obj client.Object) []string {
		nic := obj.(*v1alpha12.NetworkInterface)

		virtualIP := nic.Spec.VirtualIP
		if virtualIP == nil {
			return nil
		}

		virtualIPName := NetworkInterfaceVirtualIPName(nic.Name, *nic.Spec.VirtualIP)
		if virtualIPName == "" {
			return nil
		}

		return []string{virtualIPName}
	})
}
