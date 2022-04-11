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

package machine

import (
	"context"
	"fmt"

	"github.com/onmetal/onmetal-api/api"
	"github.com/onmetal/onmetal-api/apis/compute"
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
	machine, ok := obj.(*compute.Machine)
	if !ok {
		return nil, nil, fmt.Errorf("given object is not a Machine")
	}
	return machine.Labels, SelectableFields(machine), nil
}

func MatchMachine(label labels.Selector, field fields.Selector) apisrvstorage.SelectionPredicate {
	return apisrvstorage.SelectionPredicate{
		Label:    label,
		Field:    field,
		GetAttrs: GetAttrs,
	}
}

func SelectableFields(machine *compute.Machine) fields.Set {
	return generic.ObjectMetaFieldsSet(&machine.ObjectMeta, true)
}

type machineStrategy struct {
	runtime.ObjectTyper
	names.NameGenerator
}

var Strategy = machineStrategy{api.Scheme, names.SimpleNameGenerator}

func (machineStrategy) NamespaceScoped() bool {
	return true
}

func (machineStrategy) PrepareForCreate(ctx context.Context, obj runtime.Object) {
}

func (machineStrategy) PrepareForUpdate(ctx context.Context, obj, old runtime.Object) {
}

func (machineStrategy) Validate(ctx context.Context, obj runtime.Object) field.ErrorList {
	return field.ErrorList{}
}

func (machineStrategy) WarningsOnCreate(ctx context.Context, obj runtime.Object) []string {
	return nil
}

func (machineStrategy) AllowCreateOnUpdate() bool {
	return false
}

func (machineStrategy) AllowUnconditionalUpdate() bool {
	return false
}

func (machineStrategy) Canonicalize(obj runtime.Object) {
}

func (machineStrategy) ValidateUpdate(ctx context.Context, obj, old runtime.Object) field.ErrorList {
	return field.ErrorList{}
}

func (machineStrategy) WarningsOnUpdate(ctx context.Context, obj, old runtime.Object) []string {
	return nil
}

type machineStatusStrategy struct {
	machineStrategy
}

var StatusStrategy = machineStatusStrategy{Strategy}

func (machineStatusStrategy) GetResetFields() map[fieldpath.APIVersion]*fieldpath.Set {
	return map[fieldpath.APIVersion]*fieldpath.Set{
		"compute.onmetal.de/v1alpha1": fieldpath.NewSet(
			fieldpath.MakePathOrDie("spec"),
		),
	}
}

func (machineStatusStrategy) PrepareForUpdate(ctx context.Context, obj, old runtime.Object) {
	newMachine := obj.(*compute.Machine)
	oldMachine := old.(*compute.Machine)
	newMachine.Spec = oldMachine.Spec
}

func (machineStatusStrategy) ValidateUpdate(ctx context.Context, obj, old runtime.Object) field.ErrorList {
	return nil
}

func (machineStatusStrategy) WarningsOnUpdate(cxt context.Context, obj, old runtime.Object) []string {
	return nil
}
