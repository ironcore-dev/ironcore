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

package prefixallocation

import (
	"context"
	"fmt"

	"github.com/onmetal/onmetal-api/onmetal-apiserver/internal/api"
	"github.com/onmetal/onmetal-api/onmetal-apiserver/internal/apis/ipam"
	"github.com/onmetal/onmetal-api/onmetal-apiserver/internal/apis/ipam/validation"
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
	prefixAllocation, ok := obj.(*ipam.PrefixAllocation)
	if !ok {
		return nil, nil, fmt.Errorf("given object is not a PrefixAllocation")
	}
	return prefixAllocation.Labels, SelectableFields(prefixAllocation), nil
}

func MatchPrefixAllocation(label labels.Selector, field fields.Selector) apisrvstorage.SelectionPredicate {
	return apisrvstorage.SelectionPredicate{
		Label:    label,
		Field:    field,
		GetAttrs: GetAttrs,
	}
}

func SelectableFields(prefixAllocation *ipam.PrefixAllocation) fields.Set {
	return generic.ObjectMetaFieldsSet(&prefixAllocation.ObjectMeta, true)
}

type prefixAllocationStrategy struct {
	runtime.ObjectTyper
	names.NameGenerator
}

var Strategy = prefixAllocationStrategy{api.Scheme, names.SimpleNameGenerator}

func (prefixAllocationStrategy) NamespaceScoped() bool {
	return true
}

func (prefixAllocationStrategy) PrepareForCreate(ctx context.Context, obj runtime.Object) {
}

func (prefixAllocationStrategy) PrepareForUpdate(ctx context.Context, obj, old runtime.Object) {
}

func (prefixAllocationStrategy) Validate(ctx context.Context, obj runtime.Object) field.ErrorList {
	prefixAllocation := obj.(*ipam.PrefixAllocation)
	return validation.ValidatePrefixAllocation(prefixAllocation)
}

func (prefixAllocationStrategy) WarningsOnCreate(ctx context.Context, obj runtime.Object) []string {
	return nil
}

func (prefixAllocationStrategy) AllowCreateOnUpdate() bool {
	return false
}

func (prefixAllocationStrategy) AllowUnconditionalUpdate() bool {
	return false
}

func (prefixAllocationStrategy) Canonicalize(obj runtime.Object) {
}

func (prefixAllocationStrategy) ValidateUpdate(ctx context.Context, obj, old runtime.Object) field.ErrorList {
	newPrefixAllocation := obj.(*ipam.PrefixAllocation)
	oldPrefixAllocation := old.(*ipam.PrefixAllocation)
	return validation.ValidatePrefixAllocationUpdate(newPrefixAllocation, oldPrefixAllocation)
}

func (prefixAllocationStrategy) WarningsOnUpdate(ctx context.Context, obj, old runtime.Object) []string {
	return nil
}

type prefixAllocationStatusStrategy struct {
	prefixAllocationStrategy
}

var StatusStrategy = prefixAllocationStatusStrategy{Strategy}

func (prefixAllocationStatusStrategy) GetResetFields() map[fieldpath.APIVersion]*fieldpath.Set {
	return map[fieldpath.APIVersion]*fieldpath.Set{
		"ipam.api.onmetal.de/v1alpha1": fieldpath.NewSet(
			fieldpath.MakePathOrDie("spec"),
		),
	}
}

func (prefixAllocationStatusStrategy) PrepareForUpdate(ctx context.Context, obj, old runtime.Object) {
	newPrefixAllocation := obj.(*ipam.PrefixAllocation)
	oldPrefixAllocation := old.(*ipam.PrefixAllocation)
	newPrefixAllocation.Spec = oldPrefixAllocation.Spec
}

func (prefixAllocationStatusStrategy) ValidateUpdate(ctx context.Context, obj, old runtime.Object) field.ErrorList {
	newPrefixAllocation := obj.(*ipam.PrefixAllocation)
	oldPrefixAllocation := old.(*ipam.PrefixAllocation)
	return validation.ValidatePrefixAllocationStatusUpdate(newPrefixAllocation, oldPrefixAllocation)
}

func (prefixAllocationStatusStrategy) WarningsOnUpdate(cxt context.Context, obj, old runtime.Object) []string {
	return nil
}
