// SPDX-FileCopyrightText: 2023 SAP SE or an SAP affiliate company and IronCore contributors
// SPDX-License-Identifier: Apache-2.0

package client

import (
	"context"

	computev1alpha1 "github.com/ironcore-dev/ironcore/api/compute/v1alpha1"
	"sigs.k8s.io/controller-runtime/pkg/client"
)

func machineIsOnMachinePool(machine *computev1alpha1.Machine, machinePoolName string) bool {
	machinePoolRef := machine.Spec.MachinePoolRef
	if machinePoolRef == nil {
		return false
	}

	return machinePoolRef.Name == machinePoolName
}

const MachineSpecNetworkInterfaceNamesField = "machine-spec-network-interfaces"

func SetupMachineSpecNetworkInterfaceNamesField(ctx context.Context, indexer client.FieldIndexer, machinePoolName string) error {
	return indexer.IndexField(
		ctx,
		&computev1alpha1.Machine{},
		MachineSpecNetworkInterfaceNamesField,
		func(object client.Object) []string {
			machine := object.(*computev1alpha1.Machine)
			if !machineIsOnMachinePool(machine, machinePoolName) {
				return nil
			}

			return computev1alpha1.MachineNetworkInterfaceNames(machine)
		},
	)
}

const MachineSpecVolumeNamesField = "machine-spec-volumes"

func SetupMachineSpecVolumeNamesField(ctx context.Context, indexer client.FieldIndexer, machinePoolName string) error {
	return indexer.IndexField(
		ctx,
		&computev1alpha1.Machine{},
		MachineSpecVolumeNamesField,
		func(object client.Object) []string {
			machine := object.(*computev1alpha1.Machine)
			if !machineIsOnMachinePool(machine, machinePoolName) {
				return nil
			}
			return computev1alpha1.MachineVolumeNames(machine)
		},
	)
}

const MachineSpecSecretNamesField = "machine-spec-secrets"

func SetupMachineSpecSecretNamesField(ctx context.Context, indexer client.FieldIndexer, machinePoolName string) error {
	return indexer.IndexField(
		ctx,
		&computev1alpha1.Machine{},
		MachineSpecSecretNamesField,
		func(object client.Object) []string {
			machine := object.(*computev1alpha1.Machine)
			if !machineIsOnMachinePool(machine, machinePoolName) {
				return nil
			}

			return computev1alpha1.MachineSecretNames(machine)
		},
	)
}
