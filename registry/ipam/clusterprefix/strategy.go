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

package clusterprefix

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
	clusterPrefix, ok := obj.(*ipam.ClusterPrefix)
	if !ok {
		return nil, nil, fmt.Errorf("given object is not a ClusterPrefix")
	}
	return clusterPrefix.Labels, SelectableFields(clusterPrefix), nil
}

func MatchClusterPrefix(label labels.Selector, field fields.Selector) apisrvstorage.SelectionPredicate {
	return apisrvstorage.SelectionPredicate{
		Label:    label,
		Field:    field,
		GetAttrs: GetAttrs,
	}
}

func SelectableFields(clusterPrefix *ipam.ClusterPrefix) fields.Set {
	return generic.ObjectMetaFieldsSet(&clusterPrefix.ObjectMeta, false)
}

type clusterPrefixStrategy struct {
	runtime.ObjectTyper
	names.NameGenerator
}

var Strategy = clusterPrefixStrategy{api.Scheme, names.SimpleNameGenerator}

func (clusterPrefixStrategy) NamespaceScoped() bool {
	return false
}

func (clusterPrefixStrategy) GetResetFields() map[fieldpath.APIVersion]*fieldpath.Set {
	return map[fieldpath.APIVersion]*fieldpath.Set{
		"ipam.api.onmetal.de/v1alpha1": fieldpath.NewSet(
			fieldpath.MakePathOrDie("status"),
		),
	}
}

func (clusterPrefixStrategy) PrepareForCreate(ctx context.Context, obj runtime.Object) {
	clusterPrefix := obj.(*ipam.ClusterPrefix)
	clusterPrefix.Status = ipam.ClusterPrefixStatus{}
	clusterPrefix.Generation = 1
}

func (clusterPrefixStrategy) PrepareForUpdate(ctx context.Context, obj, old runtime.Object) {
	newClusterPrefix := obj.(*ipam.ClusterPrefix)
	oldClusterPrefix := old.(*ipam.ClusterPrefix)
	newClusterPrefix.Status = oldClusterPrefix.Status

	if !equality.Semantic.DeepEqual(&newClusterPrefix.Spec, oldClusterPrefix.Spec) {
		newClusterPrefix.Generation = oldClusterPrefix.Generation + 1
	}
}

func (clusterPrefixStrategy) Validate(ctx context.Context, obj runtime.Object) field.ErrorList {
	return field.ErrorList{}
}

func (clusterPrefixStrategy) WarningsOnCreate(ctx context.Context, obj runtime.Object) []string {
	return nil
}

func (clusterPrefixStrategy) AllowCreateOnUpdate() bool {
	return false
}

func (clusterPrefixStrategy) AllowUnconditionalUpdate() bool {
	return false
}

func (clusterPrefixStrategy) Canonicalize(obj runtime.Object) {
}

func (clusterPrefixStrategy) ValidateUpdate(ctx context.Context, obj, old runtime.Object) field.ErrorList {
	return field.ErrorList{}
}

func (clusterPrefixStrategy) WarningsOnUpdate(ctx context.Context, obj, old runtime.Object) []string {
	return nil
}

type clusterPrefixStatusStrategy struct {
	clusterPrefixStrategy
}

var StatusStrategy = clusterPrefixStatusStrategy{Strategy}

func (clusterPrefixStatusStrategy) GetResetFields() map[fieldpath.APIVersion]*fieldpath.Set {
	return map[fieldpath.APIVersion]*fieldpath.Set{
		"ipam.api.onmetal.de/v1alpha1": fieldpath.NewSet(
			fieldpath.MakePathOrDie("spec"),
		),
	}
}

func (clusterPrefixStatusStrategy) PrepareForUpdate(ctx context.Context, obj, old runtime.Object) {
	newClusterPrefix := obj.(*ipam.ClusterPrefix)
	oldClusterPrefix := old.(*ipam.ClusterPrefix)
	newClusterPrefix.Spec = oldClusterPrefix.Spec
}

func (clusterPrefixStatusStrategy) ValidateUpdate(ctx context.Context, obj, old runtime.Object) field.ErrorList {
	return nil
}

func (clusterPrefixStatusStrategy) WarningsOnUpdate(cxt context.Context, obj, old runtime.Object) []string {
	return nil
}
