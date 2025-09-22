// SPDX-FileCopyrightText: 2023 SAP SE or an SAP affiliate company and IronCore contributors
// SPDX-License-Identifier: Apache-2.0

package volumesnapshot

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
	"sigs.k8s.io/structured-merge-diff/v4/fieldpath"
)

func GetAttrs(obj runtime.Object) (labels.Set, fields.Set, error) {
	volumeSnapshot, ok := obj.(*storage.VolumeSnapshot)
	if !ok {
		return nil, nil, fmt.Errorf("given object is not a VolumeSnapshot")
	}
	return volumeSnapshot.Labels, SelectableFields(volumeSnapshot), nil
}

func MatchVolumeSnapshot(label labels.Selector, field fields.Selector) apisrvstorage.SelectionPredicate {
	return apisrvstorage.SelectionPredicate{
		Label:    label,
		Field:    field,
		GetAttrs: GetAttrs,
	}
}

func SelectableFields(volumesnapshot *storage.VolumeSnapshot) fields.Set {
	return generic.ObjectMetaFieldsSet(&volumesnapshot.ObjectMeta, false)
}

type volumeSnapshotStrategy struct {
	runtime.ObjectTyper
	names.NameGenerator
}

var Strategy = volumeSnapshotStrategy{api.Scheme, names.SimpleNameGenerator}

func (volumeSnapshotStrategy) NamespaceScoped() bool {
	return true
}

func (volumeSnapshotStrategy) PrepareForCreate(ctx context.Context, obj runtime.Object) {
}

func (volumeSnapshotStrategy) PrepareForUpdate(ctx context.Context, obj, old runtime.Object) {
}

func (volumeSnapshotStrategy) Validate(ctx context.Context, obj runtime.Object) field.ErrorList {
	volumeSnapshot := obj.(*storage.VolumeSnapshot)
	return validation.ValidateVolumeSnapshot(volumeSnapshot)
}

func (volumeSnapshotStrategy) WarningsOnCreate(ctx context.Context, obj runtime.Object) []string {
	return nil
}

func (volumeSnapshotStrategy) AllowCreateOnUpdate() bool {
	return false
}

func (volumeSnapshotStrategy) AllowUnconditionalUpdate() bool {
	return false
}

func (volumeSnapshotStrategy) Canonicalize(obj runtime.Object) {
}

func (volumeSnapshotStrategy) ValidateUpdate(ctx context.Context, obj, old runtime.Object) field.ErrorList {
	newVolumeSnapshot, oldVolumeSnapshot := obj.(*storage.VolumeSnapshot), old.(*storage.VolumeSnapshot)
	return validation.ValidateVolumeSnapshotUpdate(newVolumeSnapshot, oldVolumeSnapshot)
}

func (volumeSnapshotStrategy) WarningsOnUpdate(ctx context.Context, obj, old runtime.Object) []string {
	return nil
}

type volumeSnapshotStatusStrategy struct {
	volumeSnapshotStrategy
}

var StatusStrategy = volumeSnapshotStatusStrategy{Strategy}

func (volumeSnapshotStatusStrategy) GetResetFields() map[fieldpath.APIVersion]*fieldpath.Set {
	return map[fieldpath.APIVersion]*fieldpath.Set{
		"storage.ironcore.dev/v1alpha1": fieldpath.NewSet(
			fieldpath.MakePathOrDie("spec"),
		),
	}
}

func (volumeSnapshotStatusStrategy) PrepareForUpdate(ctx context.Context, obj, old runtime.Object) {
	newVolumeSnapshot := obj.(*storage.VolumeSnapshot)
	oldVolumeSnapshot := old.(*storage.VolumeSnapshot)
	newVolumeSnapshot.Spec = oldVolumeSnapshot.Spec
}

func (volumeSnapshotStatusStrategy) ValidateUpdate(ctx context.Context, obj, old runtime.Object) field.ErrorList {
	return nil
}

func (volumeSnapshotStatusStrategy) WarningsOnUpdate(cxt context.Context, obj, old runtime.Object) []string {
	return nil
}
