// SPDX-FileCopyrightText: 2023 SAP SE or an SAP affiliate company and IronCore contributors
// SPDX-License-Identifier: Apache-2.0

package volumeclass

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
)

func GetAttrs(obj runtime.Object) (labels.Set, fields.Set, error) {
	volumeClass, ok := obj.(*storage.VolumeClass)
	if !ok {
		return nil, nil, fmt.Errorf("given object is not a VolumeClass")
	}
	return volumeClass.Labels, SelectableFields(volumeClass), nil
}

func MatchVolumeClass(label labels.Selector, field fields.Selector) apisrvstorage.SelectionPredicate {
	return apisrvstorage.SelectionPredicate{
		Label:    label,
		Field:    field,
		GetAttrs: GetAttrs,
	}
}

func SelectableFields(volumeClass *storage.VolumeClass) fields.Set {
	return generic.ObjectMetaFieldsSet(&volumeClass.ObjectMeta, false)
}

type volumeClassStrategy struct {
	runtime.ObjectTyper
	names.NameGenerator
}

var Strategy = volumeClassStrategy{api.Scheme, names.SimpleNameGenerator}

func (volumeClassStrategy) NamespaceScoped() bool {
	return false
}

func (volumeClassStrategy) PrepareForCreate(ctx context.Context, obj runtime.Object) {
}

func (volumeClassStrategy) PrepareForUpdate(ctx context.Context, obj, old runtime.Object) {
}

func (volumeClassStrategy) Validate(ctx context.Context, obj runtime.Object) field.ErrorList {
	volumeClass := obj.(*storage.VolumeClass)
	return validation.ValidateVolumeClass(volumeClass)
}

func (volumeClassStrategy) WarningsOnCreate(ctx context.Context, obj runtime.Object) []string {
	return nil
}

func (volumeClassStrategy) AllowCreateOnUpdate() bool {
	return false
}

func (volumeClassStrategy) AllowUnconditionalUpdate() bool {
	return false
}

func (volumeClassStrategy) Canonicalize(obj runtime.Object) {
}

func (volumeClassStrategy) ValidateUpdate(ctx context.Context, obj, old runtime.Object) field.ErrorList {
	newVolumeClass := obj.(*storage.VolumeClass)
	oldVolumeClass := old.(*storage.VolumeClass)
	return validation.ValidateVolumeClassUpdate(newVolumeClass, oldVolumeClass)
}

func (volumeClassStrategy) WarningsOnUpdate(ctx context.Context, obj, old runtime.Object) []string {
	return nil
}
