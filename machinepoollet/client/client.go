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

package client

import (
	"context"

	computev1alpha1 "github.com/onmetal/onmetal-api/apis/compute/v1alpha1"
	"sigs.k8s.io/controller-runtime/pkg/client"
)

const MachineSpecNetworkInterfaceNamesField = "machine-spec-network-interfaces"

func SetupMachineSpecNetworkInterfaceNamesField(ctx context.Context, indexer client.FieldIndexer) error {
	return indexer.IndexField(
		ctx,
		&computev1alpha1.Machine{},
		MachineSpecNetworkInterfaceNamesField,
		func(object client.Object) []string {
			machine := object.(*computev1alpha1.Machine)
			return computev1alpha1.MachineNetworkInterfaceNames(machine)
		},
	)
}

const MachineSpecVolumeNamesField = "machine-spec-volumes"

func SetupMachineSpecVolumeNamesField(ctx context.Context, indexer client.FieldIndexer) error {
	return indexer.IndexField(
		ctx,
		&computev1alpha1.Machine{},
		MachineSpecVolumeNamesField,
		func(object client.Object) []string {
			machine := object.(*computev1alpha1.Machine)
			return computev1alpha1.MachineVolumeNames(machine)
		},
	)
}

const MachineSpecSecretNamesField = "machine-spec-secrets"

func SetupMachineSpecSecretNamesField(ctx context.Context, indexer client.FieldIndexer) error {
	return indexer.IndexField(
		ctx,
		&computev1alpha1.Machine{},
		MachineSpecSecretNamesField,
		func(object client.Object) []string {
			machine := object.(*computev1alpha1.Machine)
			return computev1alpha1.MachineSecretNames(machine)
		},
	)
}
