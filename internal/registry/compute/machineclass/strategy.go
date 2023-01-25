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

package machineclass

import (
	"context"
	"fmt"

	"github.com/onmetal/onmetal-api/internal/api"
	"github.com/onmetal/onmetal-api/internal/apis/compute"
	"github.com/onmetal/onmetal-api/internal/apis/compute/validation"
	"github.com/onmetal/onmetal-api/utils/equality"
	"k8s.io/apimachinery/pkg/fields"
	"k8s.io/apimachinery/pkg/labels"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apimachinery/pkg/util/validation/field"
	"k8s.io/apiserver/pkg/registry/generic"
	"k8s.io/apiserver/pkg/storage"
	"k8s.io/apiserver/pkg/storage/names"
)

func GetAttrs(obj runtime.Object) (labels.Set, fields.Set, error) {
	machineClass, ok := obj.(*compute.MachineClass)
	if !ok {
		return nil, nil, fmt.Errorf("given object is not a MachineClassRef")
	}
	return machineClass.Labels, SelectableFields(machineClass), nil
}

func MatchMachineClass(label labels.Selector, field fields.Selector) storage.SelectionPredicate {
	return storage.SelectionPredicate{
		Label:    label,
		Field:    field,
		GetAttrs: GetAttrs,
	}
}

func SelectableFields(machine *compute.MachineClass) fields.Set {
	return generic.ObjectMetaFieldsSet(&machine.ObjectMeta, false)
}

type machineClassStrategy struct {
	runtime.ObjectTyper
	names.NameGenerator
}

var Strategy = machineClassStrategy{api.Scheme, names.SimpleNameGenerator}

func (machineClassStrategy) NamespaceScoped() bool {
	return false
}

func (machineClassStrategy) PrepareForCreate(ctx context.Context, obj runtime.Object) {
	machineClass := obj.(*compute.MachineClass)
	machineClass.Generation = 1
}

func (machineClassStrategy) PrepareForUpdate(ctx context.Context, obj, old runtime.Object) {
	newMachineClass := obj.(*compute.MachineClass)
	oldMachineClass := old.(*compute.MachineClass)

	if !equality.Semantic.DeepEqual(newMachineClass.Capabilities, oldMachineClass.Capabilities) {
		newMachineClass.Generation = oldMachineClass.Generation + 1
	}
}

func (machineClassStrategy) Validate(ctx context.Context, obj runtime.Object) field.ErrorList {
	machineClass := obj.(*compute.MachineClass)
	return validation.ValidateMachineClass(machineClass)
}

func (machineClassStrategy) WarningsOnCreate(ctx context.Context, obj runtime.Object) []string {
	return nil
}

func (machineClassStrategy) AllowCreateOnUpdate() bool {
	return false
}

func (machineClassStrategy) AllowUnconditionalUpdate() bool {
	return false
}

func (machineClassStrategy) Canonicalize(obj runtime.Object) {
}

func (machineClassStrategy) ValidateUpdate(ctx context.Context, obj, old runtime.Object) field.ErrorList {
	newMachineClass := obj.(*compute.MachineClass)
	oldMachineClass := old.(*compute.MachineClass)
	return validation.ValidateMachineClassUpdate(newMachineClass, oldMachineClass)
}

func (machineClassStrategy) WarningsOnUpdate(ctx context.Context, obj, old runtime.Object) []string {
	return nil
}
