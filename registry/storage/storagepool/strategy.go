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

package storagepool

import (
	"context"
	"fmt"

	"github.com/onmetal/onmetal-api/api"
	"github.com/onmetal/onmetal-api/apis/storage"
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
	storagePool, ok := obj.(*storage.StoragePool)
	if !ok {
		return nil, nil, fmt.Errorf("given object is not a StoragePool")
	}
	return storagePool.Labels, SelectableFields(storagePool), nil
}

func MatchStoragePool(label labels.Selector, field fields.Selector) apisrvstorage.SelectionPredicate {
	return apisrvstorage.SelectionPredicate{
		Label:    label,
		Field:    field,
		GetAttrs: GetAttrs,
	}
}

func SelectableFields(storagePool *storage.StoragePool) fields.Set {
	return generic.ObjectMetaFieldsSet(&storagePool.ObjectMeta, false)
}

type storagePoolStrategy struct {
	runtime.ObjectTyper
	names.NameGenerator
}

var Strategy = storagePoolStrategy{api.Scheme, names.SimpleNameGenerator}

func (storagePoolStrategy) NamespaceScoped() bool {
	return false
}

func (storagePoolStrategy) GetResetFields() map[fieldpath.APIVersion]*fieldpath.Set {
	return map[fieldpath.APIVersion]*fieldpath.Set{
		"storage.onmetal.de/v1alpha1": fieldpath.NewSet(
			fieldpath.MakePathOrDie("status"),
		),
	}
}

func (storagePoolStrategy) PrepareForCreate(ctx context.Context, obj runtime.Object) {
	storagePool := obj.(*storage.StoragePool)
	storagePool.Status = storage.StoragePoolStatus{}
	storagePool.Generation = 1
}

func (storagePoolStrategy) PrepareForUpdate(ctx context.Context, obj, old runtime.Object) {
	newStoragePool := obj.(*storage.StoragePool)
	oldStoragePool := old.(*storage.StoragePool)
	newStoragePool.Status = oldStoragePool.Status

	if !equality.Semantic.DeepEqual(&newStoragePool.Spec, oldStoragePool.Spec) {
		newStoragePool.Generation = oldStoragePool.Generation + 1
	}
}

func (storagePoolStrategy) Validate(ctx context.Context, obj runtime.Object) field.ErrorList {
	return field.ErrorList{}
}

func (storagePoolStrategy) WarningsOnCreate(ctx context.Context, obj runtime.Object) []string {
	return nil
}

func (storagePoolStrategy) AllowCreateOnUpdate() bool {
	return false
}

func (storagePoolStrategy) AllowUnconditionalUpdate() bool {
	return false
}

func (storagePoolStrategy) Canonicalize(obj runtime.Object) {
}

func (storagePoolStrategy) ValidateUpdate(ctx context.Context, obj, old runtime.Object) field.ErrorList {
	return field.ErrorList{}
}

func (storagePoolStrategy) WarningsOnUpdate(ctx context.Context, obj, old runtime.Object) []string {
	return nil
}

type storagePoolStatusStrategy struct {
	storagePoolStrategy
}

var StatusStrategy = storagePoolStatusStrategy{Strategy}

func (storagePoolStatusStrategy) GetResetFields() map[fieldpath.APIVersion]*fieldpath.Set {
	return map[fieldpath.APIVersion]*fieldpath.Set{
		"storage.onmetal.de/v1alpha1": fieldpath.NewSet(
			fieldpath.MakePathOrDie("spec"),
		),
	}
}

func (storagePoolStatusStrategy) PrepareForUpdate(ctx context.Context, obj, old runtime.Object) {
	newStoragePool := obj.(*storage.StoragePool)
	oldStoragePool := old.(*storage.StoragePool)
	newStoragePool.Spec = oldStoragePool.Spec
}

func (storagePoolStatusStrategy) ValidateUpdate(ctx context.Context, obj, old runtime.Object) field.ErrorList {
	return nil
}

func (storagePoolStatusStrategy) WarningsOnUpdate(cxt context.Context, obj, old runtime.Object) []string {
	return nil
}
