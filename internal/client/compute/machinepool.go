// Copyright 2023 IronCore authors
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

package compute

import (
	"context"

	computev1alpha1 "github.com/ironcore-dev/ironcore/api/compute/v1alpha1"
	"sigs.k8s.io/controller-runtime/pkg/client"
)

const (
	MachinePoolAvailableMachineClassesField = "machinepool-available-machine-classes"
)

func SetupMachinePoolAvailableMachineClassesFieldIndexer(ctx context.Context, indexer client.FieldIndexer) error {
	return indexer.IndexField(ctx, &computev1alpha1.MachinePool{}, MachinePoolAvailableMachineClassesField, func(object client.Object) []string {
		machinePool := object.(*computev1alpha1.MachinePool)

		names := make([]string, 0, len(machinePool.Status.AvailableMachineClasses))
		for _, availableMachineClass := range machinePool.Status.AvailableMachineClasses {
			names = append(names, availableMachineClass.Name)
		}

		if len(names) == 0 {
			return []string{""}
		}
		return names
	})
}
