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

package volumepool

import (
	"context"
	"fmt"

	"github.com/onmetal/onmetal-api/api"
	"github.com/onmetal/onmetal-api/apis/storage"
	"github.com/onmetal/onmetal-api/apis/storage/validation"
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
	volumePool, ok := obj.(*storage.VolumePool)
	if !ok {
		return nil, nil, fmt.Errorf("given object is not a VolumePoolRef")
	}
	return volumePool.Labels, SelectableFields(volumePool), nil
}

func MatchVolumePool(label labels.Selector, field fields.Selector) apisrvstorage.SelectionPredicate {
	return apisrvstorage.SelectionPredicate{
		Label:    label,
		Field:    field,
		GetAttrs: GetAttrs,
	}
}

func SelectableFields(volumePool *storage.VolumePool) fields.Set {
	return generic.ObjectMetaFieldsSet(&volumePool.ObjectMeta, false)
}

type volumePoolStrategy struct {
	runtime.ObjectTyper
	names.NameGenerator
}

var Strategy = volumePoolStrategy{api.Scheme, names.SimpleNameGenerator}

func (volumePoolStrategy) NamespaceScoped() bool {
	return false
}

func (volumePoolStrategy) GetResetFields() map[fieldpath.APIVersion]*fieldpath.Set {
	return map[fieldpath.APIVersion]*fieldpath.Set{
		"storage.api.onmetal.de/v1alpha1": fieldpath.NewSet(
			fieldpath.MakePathOrDie("status"),
		),
	}
}

func (volumePoolStrategy) PrepareForCreate(ctx context.Context, obj runtime.Object) {
	volumePool := obj.(*storage.VolumePool)
	volumePool.Status = storage.VolumePoolStatus{}
	volumePool.Generation = 1
}

func (volumePoolStrategy) PrepareForUpdate(ctx context.Context, obj, old runtime.Object) {
	newVolumePool := obj.(*storage.VolumePool)
	oldVolumePool := old.(*storage.VolumePool)
	newVolumePool.Status = oldVolumePool.Status

	if !equality.Semantic.DeepEqual(newVolumePool.Spec, oldVolumePool.Spec) {
		newVolumePool.Generation = oldVolumePool.Generation + 1
	}
}

func (volumePoolStrategy) Validate(ctx context.Context, obj runtime.Object) field.ErrorList {
	volumePool := obj.(*storage.VolumePool)
	return validation.ValidateVolumePool(volumePool)
}

func (volumePoolStrategy) WarningsOnCreate(ctx context.Context, obj runtime.Object) []string {
	return nil
}

func (volumePoolStrategy) AllowCreateOnUpdate() bool {
	return false
}

func (volumePoolStrategy) AllowUnconditionalUpdate() bool {
	return false
}

func (volumePoolStrategy) Canonicalize(obj runtime.Object) {
}

func (volumePoolStrategy) ValidateUpdate(ctx context.Context, obj, old runtime.Object) field.ErrorList {
	return field.ErrorList{}
}

func (volumePoolStrategy) WarningsOnUpdate(ctx context.Context, obj, old runtime.Object) []string {
	return nil
}

type volumePoolStatusStrategy struct {
	volumePoolStrategy
}

var StatusStrategy = volumePoolStatusStrategy{Strategy}

func (volumePoolStatusStrategy) GetResetFields() map[fieldpath.APIVersion]*fieldpath.Set {
	return map[fieldpath.APIVersion]*fieldpath.Set{
		"storage.api.onmetal.de/v1alpha1": fieldpath.NewSet(
			fieldpath.MakePathOrDie("spec"),
		),
	}
}

func (volumePoolStatusStrategy) PrepareForUpdate(ctx context.Context, obj, old runtime.Object) {
	newVolumePool := obj.(*storage.VolumePool)
	oldVolumePool := old.(*storage.VolumePool)
	newVolumePool.Spec = oldVolumePool.Spec
}

func (volumePoolStatusStrategy) ValidateUpdate(ctx context.Context, obj, old runtime.Object) field.ErrorList {
	newVolumePool := obj.(*storage.VolumePool)
	oldVolumePool := old.(*storage.VolumePool)
	return validation.ValidateVolumePoolUpdate(newVolumePool, oldVolumePool)
}

func (volumePoolStatusStrategy) WarningsOnUpdate(cxt context.Context, obj, old runtime.Object) []string {
	return nil
}
