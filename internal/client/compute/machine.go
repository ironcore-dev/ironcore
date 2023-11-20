// SPDX-FileCopyrightText: 2023 SAP SE or an SAP affiliate company and IronCore contributors
// SPDX-License-Identifier: Apache-2.0

package compute

import (
	"context"

	computev1alpha1 "github.com/ironcore-dev/ironcore/api/compute/v1alpha1"
	"sigs.k8s.io/controller-runtime/pkg/client"
)

const (
	MachineSpecMachinePoolRefNameField    = computev1alpha1.MachineMachinePoolRefNameField
	MachineSpecMachineClassRefNameField   = computev1alpha1.MachineMachineClassRefNameField
	MachineSpecNetworkInterfaceNamesField = "machine-spec-network-interface-names"
	MachineSpecVolumeNamesField           = "machine-spec-volume-names"
)

func SetupMachineSpecMachinePoolRefNameFieldIndexer(ctx context.Context, indexer client.FieldIndexer) error {
	return indexer.IndexField(ctx, &computev1alpha1.Machine{}, MachineSpecMachinePoolRefNameField, func(obj client.Object) []string {
		machine := obj.(*computev1alpha1.Machine)
		machinePoolRef := machine.Spec.MachinePoolRef
		if machinePoolRef == nil {
			return []string{""}
		}
		return []string{machinePoolRef.Name}
	})
}

func SetupMachineSpecMachineClassRefNameFieldIndexer(ctx context.Context, indexer client.FieldIndexer) error {
	return indexer.IndexField(ctx, &computev1alpha1.Machine{}, MachineSpecMachineClassRefNameField, func(obj client.Object) []string {
		machine := obj.(*computev1alpha1.Machine)
		return []string{machine.Spec.MachineClassRef.Name}
	})
}

func SetupMachineSpecNetworkInterfaceNamesFieldIndexer(ctx context.Context, indexer client.FieldIndexer) error {
	return indexer.IndexField(ctx, &computev1alpha1.Machine{}, MachineSpecNetworkInterfaceNamesField, func(obj client.Object) []string {
		machine := obj.(*computev1alpha1.Machine)
		if names := computev1alpha1.MachineNetworkInterfaceNames(machine); len(names) > 0 {
			return names
		}
		return []string{""}
	})
}

func SetupMachineSpecVolumeNamesFieldIndexer(ctx context.Context, indexer client.FieldIndexer) error {
	return indexer.IndexField(ctx, &computev1alpha1.Machine{}, MachineSpecVolumeNamesField, func(obj client.Object) []string {
		machine := obj.(*computev1alpha1.Machine)
		if names := computev1alpha1.MachineVolumeNames(machine); len(names) > 0 {
			return names
		}
		return []string{""}
	})
}
