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
	"k8s.io/apimachinery/pkg/util/sets"
	controllerruntime "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/client"
)

const (
	MachineNetworkInterfaceNamesField = "machine-network-interface-names"
	MachineSpecVolumeNamesField       = "machine-volume-names"
)

func MachineEphemeralNetworkInterfaceName(machineName, ifaceName string) string {
	return fmt.Sprintf("%s-%s", machineName, ifaceName)
}

func MachineNetworkInterfaceNames(machine *computev1alpha1.Machine) []string {
	var names []string
	for _, iface := range machine.Spec.NetworkInterfaces {
		switch {
		case iface.NetworkInterfaceRef != nil:
			names = append(names, iface.NetworkInterfaceRef.Name)
		case iface.Ephemeral != nil:
			names = append(names, MachineEphemeralNetworkInterfaceName(machine.Name, iface.Name))
		}
	}
	return names
}

func MachineSpecVolumeNames(machine *computev1alpha1.Machine) sets.String {
	names := sets.NewString()
	for _, volume := range machine.Spec.Volumes {
		switch {
		case volume.VolumeRef != nil:
			names.Insert(volume.VolumeRef.Name)
		}
	}
	return names
}

func SetupMachineNetworkInterfaceNamesFieldIndexer(mgr controllerruntime.Manager) error {
	ctx := context.Background()
	return mgr.GetFieldIndexer().IndexField(ctx, &computev1alpha1.Machine{}, MachineNetworkInterfaceNamesField, func(obj client.Object) []string {
		machine := obj.(*computev1alpha1.Machine)
		if names := MachineNetworkInterfaceNames(machine); len(names) > 0 {
			return names
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
