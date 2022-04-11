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

package volumeclaim

import (
	"context"
	"fmt"

	"github.com/onmetal/onmetal-api/api"
	"github.com/onmetal/onmetal-api/apis/storage"
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
	volumeClaim, ok := obj.(*storage.VolumeClaim)
	if !ok {
		return nil, nil, fmt.Errorf("given object is not a VolumeClaim")
	}
	return volumeClaim.Labels, SelectableFields(volumeClaim), nil
}

func MatchVolumeClaim(label labels.Selector, field fields.Selector) apisrvstorage.SelectionPredicate {
	return apisrvstorage.SelectionPredicate{
		Label:    label,
		Field:    field,
		GetAttrs: GetAttrs,
	}
}

func SelectableFields(volumeClaim *storage.VolumeClaim) fields.Set {
	return generic.ObjectMetaFieldsSet(&volumeClaim.ObjectMeta, true)
}

type volumeClaimStrategy struct {
	runtime.ObjectTyper
	names.NameGenerator
}

var Strategy = volumeClaimStrategy{api.Scheme, names.SimpleNameGenerator}

func (volumeClaimStrategy) NamespaceScoped() bool {
	return true
}

func (volumeClaimStrategy) PrepareForCreate(ctx context.Context, obj runtime.Object) {
}

func (volumeClaimStrategy) PrepareForUpdate(ctx context.Context, obj, old runtime.Object) {
}

func (volumeClaimStrategy) Validate(ctx context.Context, obj runtime.Object) field.ErrorList {
	return field.ErrorList{}
}

func (volumeClaimStrategy) WarningsOnCreate(ctx context.Context, obj runtime.Object) []string {
	return nil
}

func (volumeClaimStrategy) AllowCreateOnUpdate() bool {
	return false
}

func (volumeClaimStrategy) AllowUnconditionalUpdate() bool {
	return false
}

func (volumeClaimStrategy) Canonicalize(obj runtime.Object) {
}

func (volumeClaimStrategy) ValidateUpdate(ctx context.Context, obj, old runtime.Object) field.ErrorList {
	return field.ErrorList{}
}

func (volumeClaimStrategy) WarningsOnUpdate(ctx context.Context, obj, old runtime.Object) []string {
	return nil
}

type volumeClaimStatusStrategy struct {
	volumeClaimStrategy
}

var StatusStrategy = volumeClaimStatusStrategy{Strategy}

func (volumeClaimStatusStrategy) GetResetFields() map[fieldpath.APIVersion]*fieldpath.Set {
	return map[fieldpath.APIVersion]*fieldpath.Set{
		"storage.onmetal.de/v1alpha1": fieldpath.NewSet(
			fieldpath.MakePathOrDie("spec"),
		),
	}
}

func (volumeClaimStatusStrategy) PrepareForUpdate(ctx context.Context, obj, old runtime.Object) {
	newVolumeClaim := obj.(*storage.VolumeClaim)
	oldVolumeClaim := old.(*storage.VolumeClaim)
	newVolumeClaim.Spec = oldVolumeClaim.Spec
}

func (volumeClaimStatusStrategy) ValidateUpdate(ctx context.Context, obj, old runtime.Object) field.ErrorList {
	return nil
}

func (volumeClaimStatusStrategy) WarningsOnUpdate(cxt context.Context, obj, old runtime.Object) []string {
	return nil
}
