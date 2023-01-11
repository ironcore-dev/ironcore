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

package machinepool

import (
	"context"
	"fmt"

	"github.com/onmetal/onmetal-api/onmetal-apiserver/internal/api"
	"github.com/onmetal/onmetal-api/onmetal-apiserver/internal/apis/compute"
	"github.com/onmetal/onmetal-api/onmetal-apiserver/internal/apis/compute/validation"
	"github.com/onmetal/onmetal-api/utils/equality"
	"k8s.io/apimachinery/pkg/fields"
	"k8s.io/apimachinery/pkg/labels"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apimachinery/pkg/util/validation/field"
	"k8s.io/apiserver/pkg/registry/generic"
	apisrvstorage "k8s.io/apiserver/pkg/storage"
	"k8s.io/apiserver/pkg/storage/names"
	"sigs.k8s.io/structured-merge-diff/v4/fieldpath"
)

func GetAttrs(obj runtime.Object) (labels.Set, fields.Set, error) {
	machinePool, ok := obj.(*compute.MachinePool)
	if !ok {
		return nil, nil, fmt.Errorf("given object is not a MachinePoolRef")
	}
	return machinePool.Labels, SelectableFields(machinePool), nil
}

func MatchMachinePool(label labels.Selector, field fields.Selector) apisrvstorage.SelectionPredicate {
	return apisrvstorage.SelectionPredicate{
		Label:    label,
		Field:    field,
		GetAttrs: GetAttrs,
	}
}

func SelectableFields(machinePool *compute.MachinePool) fields.Set {
	return generic.ObjectMetaFieldsSet(&machinePool.ObjectMeta, false)
}

type machinePoolStrategy struct {
	runtime.ObjectTyper
	names.NameGenerator
}

var Strategy = machinePoolStrategy{api.Scheme, names.SimpleNameGenerator}

func (machinePoolStrategy) NamespaceScoped() bool {
	return false
}

func (machinePoolStrategy) GetResetFields() map[fieldpath.APIVersion]*fieldpath.Set {
	return map[fieldpath.APIVersion]*fieldpath.Set{
		"compute.api.onmetal.de/v1alpha1": fieldpath.NewSet(
			fieldpath.MakePathOrDie("status"),
		),
	}
}

func (machinePoolStrategy) PrepareForCreate(ctx context.Context, obj runtime.Object) {
	machinePool := obj.(*compute.MachinePool)
	machinePool.Status = compute.MachinePoolStatus{}
	machinePool.Generation = 1
}

func (machinePoolStrategy) PrepareForUpdate(ctx context.Context, obj, old runtime.Object) {
	newMachinePool := obj.(*compute.MachinePool)
	oldMachinePool := old.(*compute.MachinePool)
	newMachinePool.Status = oldMachinePool.Status

	if !equality.Semantic.DeepEqual(newMachinePool.Spec, oldMachinePool.Spec) {
		newMachinePool.Generation = oldMachinePool.Generation + 1
	}
}

func (machinePoolStrategy) Validate(ctx context.Context, obj runtime.Object) field.ErrorList {
	machinePool := obj.(*compute.MachinePool)
	return validation.ValidateMachinePool(machinePool)
}

func (machinePoolStrategy) WarningsOnCreate(ctx context.Context, obj runtime.Object) []string {
	return nil
}

func (machinePoolStrategy) AllowCreateOnUpdate() bool {
	return false
}

func (machinePoolStrategy) AllowUnconditionalUpdate() bool {
	return false
}

func (machinePoolStrategy) Canonicalize(obj runtime.Object) {
}

func (machinePoolStrategy) ValidateUpdate(ctx context.Context, obj, old runtime.Object) field.ErrorList {
	newMachinePool := obj.(*compute.MachinePool)
	oldMachinePool := old.(*compute.MachinePool)
	return validation.ValidateMachinePoolUpdate(newMachinePool, oldMachinePool)
}

func (machinePoolStrategy) WarningsOnUpdate(ctx context.Context, obj, old runtime.Object) []string {
	return nil
}

type machinePoolStatusStrategy struct {
	machinePoolStrategy
}

var StatusStrategy = machinePoolStatusStrategy{Strategy}

func (machinePoolStatusStrategy) GetResetFields() map[fieldpath.APIVersion]*fieldpath.Set {
	return map[fieldpath.APIVersion]*fieldpath.Set{
		"compute.api.onmetal.de/v1alpha1": fieldpath.NewSet(
			fieldpath.MakePathOrDie("spec"),
		),
	}
}

func (machinePoolStatusStrategy) PrepareForUpdate(ctx context.Context, obj, old runtime.Object) {
	newMachinePool := obj.(*compute.MachinePool)
	oldMachinePool := old.(*compute.MachinePool)
	newMachinePool.Spec = oldMachinePool.Spec
}

func (machinePoolStatusStrategy) ValidateUpdate(ctx context.Context, obj, old runtime.Object) field.ErrorList {
	newMachinePool := obj.(*compute.MachinePool)
	oldMachinePool := old.(*compute.MachinePool)
	return validation.ValidateMachinePoolUpdate(newMachinePool, oldMachinePool)
}

func (machinePoolStatusStrategy) WarningsOnUpdate(cxt context.Context, obj, old runtime.Object) []string {
	return nil
}
