// SPDX-FileCopyrightText: 2023 SAP SE or an SAP affiliate company and IronCore contributors
// SPDX-License-Identifier: Apache-2.0

package bucketpool

import (
	"context"
	"fmt"

	"github.com/ironcore-dev/ironcore/internal/api"
	"github.com/ironcore-dev/ironcore/internal/apis/storage"
	"github.com/ironcore-dev/ironcore/internal/apis/storage/validation"
	"github.com/ironcore-dev/ironcore/utils/equality"
	"k8s.io/apimachinery/pkg/fields"
	"k8s.io/apimachinery/pkg/labels"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apimachinery/pkg/util/validation/field"
	"k8s.io/apiserver/pkg/registry/generic"
	apisrvstorage "k8s.io/apiserver/pkg/storage"
	"k8s.io/apiserver/pkg/storage/names"
	"sigs.k8s.io/structured-merge-diff/v6/fieldpath"
)

func GetAttrs(obj runtime.Object) (labels.Set, fields.Set, error) {
	bucketPool, ok := obj.(*storage.BucketPool)
	if !ok {
		return nil, nil, fmt.Errorf("given object is not a BucketPool")
	}
	return bucketPool.Labels, SelectableFields(bucketPool), nil
}

func MatchBucketPool(label labels.Selector, field fields.Selector) apisrvstorage.SelectionPredicate {
	return apisrvstorage.SelectionPredicate{
		Label:    label,
		Field:    field,
		GetAttrs: GetAttrs,
	}
}

func SelectableFields(bucketPool *storage.BucketPool) fields.Set {
	return generic.ObjectMetaFieldsSet(&bucketPool.ObjectMeta, false)
}

type bucketPoolStrategy struct {
	runtime.ObjectTyper
	names.NameGenerator
}

var Strategy = bucketPoolStrategy{api.Scheme, names.SimpleNameGenerator}

func (bucketPoolStrategy) NamespaceScoped() bool {
	return false
}

func (bucketPoolStrategy) GetResetFields() map[fieldpath.APIVersion]*fieldpath.Set {
	return map[fieldpath.APIVersion]*fieldpath.Set{
		"storage.ironcore.dev/v1alpha1": fieldpath.NewSet(
			fieldpath.MakePathOrDie("status"),
		),
	}
}

func (bucketPoolStrategy) PrepareForCreate(ctx context.Context, obj runtime.Object) {
	bucketPool := obj.(*storage.BucketPool)
	bucketPool.Status = storage.BucketPoolStatus{}
	bucketPool.Generation = 1
}

func (bucketPoolStrategy) PrepareForUpdate(ctx context.Context, obj, old runtime.Object) {
	newBucketPool := obj.(*storage.BucketPool)
	oldBucketPool := old.(*storage.BucketPool)
	newBucketPool.Status = oldBucketPool.Status

	if !equality.Semantic.DeepEqual(newBucketPool.Spec, oldBucketPool.Spec) {
		newBucketPool.Generation = oldBucketPool.Generation + 1
	}
}

func (bucketPoolStrategy) Validate(ctx context.Context, obj runtime.Object) field.ErrorList {
	bucketPool := obj.(*storage.BucketPool)
	return validation.ValidateBucketPool(bucketPool)
}

func (bucketPoolStrategy) WarningsOnCreate(ctx context.Context, obj runtime.Object) []string {
	return nil
}

func (bucketPoolStrategy) AllowCreateOnUpdate() bool {
	return false
}

func (bucketPoolStrategy) AllowUnconditionalUpdate() bool {
	return false
}

func (bucketPoolStrategy) Canonicalize(obj runtime.Object) {
}

func (bucketPoolStrategy) ValidateUpdate(ctx context.Context, obj, old runtime.Object) field.ErrorList {
	return field.ErrorList{}
}

func (bucketPoolStrategy) WarningsOnUpdate(ctx context.Context, obj, old runtime.Object) []string {
	return nil
}

type bucketPoolStatusStrategy struct {
	bucketPoolStrategy
}

var StatusStrategy = bucketPoolStatusStrategy{Strategy}

func (bucketPoolStatusStrategy) GetResetFields() map[fieldpath.APIVersion]*fieldpath.Set {
	return map[fieldpath.APIVersion]*fieldpath.Set{
		"storage.ironcore.dev/v1alpha1": fieldpath.NewSet(
			fieldpath.MakePathOrDie("spec"),
		),
	}
}

func (bucketPoolStatusStrategy) PrepareForUpdate(ctx context.Context, obj, old runtime.Object) {
	newBucketPool := obj.(*storage.BucketPool)
	oldBucketPool := old.(*storage.BucketPool)
	newBucketPool.Spec = oldBucketPool.Spec
}

func (bucketPoolStatusStrategy) ValidateUpdate(ctx context.Context, obj, old runtime.Object) field.ErrorList {
	newBucketPool := obj.(*storage.BucketPool)
	oldBucketPool := old.(*storage.BucketPool)
	return validation.ValidateBucketPoolUpdate(newBucketPool, oldBucketPool)
}

func (bucketPoolStatusStrategy) WarningsOnUpdate(cxt context.Context, obj, old runtime.Object) []string {
	return nil
}
