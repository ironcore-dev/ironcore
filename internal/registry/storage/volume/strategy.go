// Copyright 2022 IronCore authors
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

package volume

import (
	"context"
	"fmt"

	"github.com/ironcore-dev/ironcore/internal/api"
	"github.com/ironcore-dev/ironcore/internal/apis/storage"
	"github.com/ironcore-dev/ironcore/internal/apis/storage/validation"
	"k8s.io/apimachinery/pkg/fields"
	"k8s.io/apimachinery/pkg/labels"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apimachinery/pkg/util/validation/field"
	"k8s.io/apiserver/pkg/registry/generic"
	apisrvstorage "k8s.io/apiserver/pkg/storage"
	"k8s.io/apiserver/pkg/storage/names"
	"k8s.io/client-go/tools/cache"
	"sigs.k8s.io/structured-merge-diff/v4/fieldpath"
)

func GetAttrs(obj runtime.Object) (labels.Set, fields.Set, error) {
	volume, ok := obj.(*storage.Volume)
	if !ok {
		return nil, nil, fmt.Errorf("given object is not a Volume")
	}
	return volume.Labels, SelectableFields(volume), nil
}

func MatchVolume(label labels.Selector, field fields.Selector) apisrvstorage.SelectionPredicate {
	return apisrvstorage.SelectionPredicate{
		Label:       label,
		Field:       field,
		GetAttrs:    GetAttrs,
		IndexFields: []string{storage.VolumeVolumePoolRefNameField},
	}
}

func volumeVolumePoolRefName(volume *storage.Volume) string {
	if volumePoolRef := volume.Spec.VolumePoolRef; volumePoolRef != nil {
		return volumePoolRef.Name
	}
	return ""
}

func volumeVolumeClassRefName(volume *storage.Volume) string {
	if volumeClassRef := volume.Spec.VolumeClassRef; volumeClassRef != nil {
		return volumeClassRef.Name
	}
	return ""
}

func SelectableFields(volume *storage.Volume) fields.Set {
	fieldsSet := make(fields.Set)
	fieldsSet[storage.VolumeVolumePoolRefNameField] = volumeVolumePoolRefName(volume)
	fieldsSet[storage.VolumeVolumeClassRefNameField] = volumeVolumeClassRefName(volume)
	return generic.AddObjectMetaFieldsSet(fieldsSet, &volume.ObjectMeta, true)
}

func VolumePoolRefNameIndexFunc(obj any) ([]string, error) {
	volume, ok := obj.(*storage.Volume)
	if !ok {
		return nil, fmt.Errorf("not a volume")
	}
	return []string{volumeVolumePoolRefName(volume)}, nil
}

func VolumePoolRefNameTriggerFunc(obj runtime.Object) string {
	return volumeVolumePoolRefName(obj.(*storage.Volume))
}

func Indexers() *cache.Indexers {
	return &cache.Indexers{
		apisrvstorage.FieldIndex(storage.VolumeVolumePoolRefNameField): VolumePoolRefNameIndexFunc,
	}
}

type volumeStrategy struct {
	runtime.ObjectTyper
	names.NameGenerator
}

var Strategy = volumeStrategy{api.Scheme, names.SimpleNameGenerator}

func (volumeStrategy) NamespaceScoped() bool {
	return true
}

func (volumeStrategy) PrepareForCreate(ctx context.Context, obj runtime.Object) {
}

func (volumeStrategy) PrepareForUpdate(ctx context.Context, obj, old runtime.Object) {
}

func (volumeStrategy) Validate(ctx context.Context, obj runtime.Object) field.ErrorList {
	volume := obj.(*storage.Volume)
	return validation.ValidateVolume(volume)
}

func (volumeStrategy) WarningsOnCreate(ctx context.Context, obj runtime.Object) []string {
	return nil
}

func (volumeStrategy) AllowCreateOnUpdate() bool {
	return false
}

func (volumeStrategy) AllowUnconditionalUpdate() bool {
	return false
}

func (volumeStrategy) Canonicalize(obj runtime.Object) {
}

func (volumeStrategy) ValidateUpdate(ctx context.Context, obj, old runtime.Object) field.ErrorList {
	newVolume, oldVolume := obj.(*storage.Volume), old.(*storage.Volume)
	return validation.ValidateVolumeUpdate(newVolume, oldVolume)
}

func (volumeStrategy) WarningsOnUpdate(ctx context.Context, obj, old runtime.Object) []string {
	return nil
}

type volumeStatusStrategy struct {
	volumeStrategy
}

var StatusStrategy = volumeStatusStrategy{Strategy}

func (volumeStatusStrategy) GetResetFields() map[fieldpath.APIVersion]*fieldpath.Set {
	return map[fieldpath.APIVersion]*fieldpath.Set{
		"storage.ironcore.dev/v1alpha1": fieldpath.NewSet(
			fieldpath.MakePathOrDie("spec"),
		),
	}
}

func (volumeStatusStrategy) PrepareForUpdate(ctx context.Context, obj, old runtime.Object) {
	newVolume := obj.(*storage.Volume)
	oldVolume := old.(*storage.Volume)
	newVolume.Spec = oldVolume.Spec
}

func (volumeStatusStrategy) ValidateUpdate(ctx context.Context, obj, old runtime.Object) field.ErrorList {
	return nil
}

func (volumeStatusStrategy) WarningsOnUpdate(cxt context.Context, obj, old runtime.Object) []string {
	return nil
}
