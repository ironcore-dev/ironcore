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

package clusterprefixallocation

import (
	"context"
	"fmt"

	"github.com/onmetal/onmetal-api/api"
	"github.com/onmetal/onmetal-api/apis/ipam"
	"github.com/onmetal/onmetal-api/equality"
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
	machinePool, ok := obj.(*ipam.ClusterPrefixAllocation)
	if !ok {
		return nil, nil, fmt.Errorf("given object is not a ClusterPrefixAllocation")
	}
	return machinePool.Labels, SelectableFields(machinePool), nil
}

func MatchClusterPrefixAllocation(label labels.Selector, field fields.Selector) apisrvstorage.SelectionPredicate {
	return apisrvstorage.SelectionPredicate{
		Label:    label,
		Field:    field,
		GetAttrs: GetAttrs,
	}
}

func SelectableFields(machinePool *ipam.ClusterPrefixAllocation) fields.Set {
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
		"ipam.api.onmetal.de/v1alpha1": fieldpath.NewSet(
			fieldpath.MakePathOrDie("status"),
		),
	}
}

func (machinePoolStrategy) PrepareForCreate(ctx context.Context, obj runtime.Object) {
	machinePool := obj.(*ipam.ClusterPrefixAllocation)
	machinePool.Status = ipam.ClusterPrefixAllocationStatus{}
	machinePool.Generation = 1
}

func (machinePoolStrategy) PrepareForUpdate(ctx context.Context, obj, old runtime.Object) {
	newClusterPrefixAllocation := obj.(*ipam.ClusterPrefixAllocation)
	oldClusterPrefixAllocation := old.(*ipam.ClusterPrefixAllocation)
	newClusterPrefixAllocation.Status = oldClusterPrefixAllocation.Status

	if !equality.Semantic.DeepEqual(&newClusterPrefixAllocation.Spec, oldClusterPrefixAllocation.Spec) {
		newClusterPrefixAllocation.Generation = oldClusterPrefixAllocation.Generation + 1
	}
}

func (machinePoolStrategy) Validate(ctx context.Context, obj runtime.Object) field.ErrorList {
	return field.ErrorList{}
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
	return field.ErrorList{}
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
		"ipam.api.onmetal.de/v1alpha1": fieldpath.NewSet(
			fieldpath.MakePathOrDie("spec"),
		),
	}
}

func (machinePoolStatusStrategy) PrepareForUpdate(ctx context.Context, obj, old runtime.Object) {
	newClusterPrefixAllocation := obj.(*ipam.ClusterPrefixAllocation)
	oldClusterPrefixAllocation := old.(*ipam.ClusterPrefixAllocation)
	newClusterPrefixAllocation.Spec = oldClusterPrefixAllocation.Spec
}

func (machinePoolStatusStrategy) ValidateUpdate(ctx context.Context, obj, old runtime.Object) field.ErrorList {
	return nil
}

func (machinePoolStatusStrategy) WarningsOnUpdate(cxt context.Context, obj, old runtime.Object) []string {
	return nil
}
