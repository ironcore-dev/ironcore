// SPDX-FileCopyrightText: 2023 SAP SE or an SAP affiliate company and IronCore contributors
// SPDX-License-Identifier: Apache-2.0

package bucket

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
	bucket, ok := obj.(*storage.Bucket)
	if !ok {
		return nil, nil, fmt.Errorf("given object is not a Bucket")
	}
	return bucket.Labels, SelectableFields(bucket), nil
}

func MatchBucket(label labels.Selector, field fields.Selector) apisrvstorage.SelectionPredicate {
	return apisrvstorage.SelectionPredicate{
		Label:       label,
		Field:       field,
		GetAttrs:    GetAttrs,
		IndexFields: []string{storage.BucketBucketPoolRefNameField},
	}
}

func bucketBucketPoolRefName(bucket *storage.Bucket) string {
	if bucketPoolRef := bucket.Spec.BucketPoolRef; bucketPoolRef != nil {
		return bucketPoolRef.Name
	}
	return ""
}

func bucketBucketClassRefName(bucket *storage.Bucket) string {
	if bucketClassRef := bucket.Spec.BucketClassRef; bucketClassRef != nil {
		return bucketClassRef.Name
	}
	return ""
}

func SelectableFields(bucket *storage.Bucket) fields.Set {
	fieldsSet := make(fields.Set)
	fieldsSet[storage.BucketBucketPoolRefNameField] = bucketBucketPoolRefName(bucket)
	fieldsSet[storage.BucketBucketClassRefNameField] = bucketBucketClassRefName(bucket)
	return generic.AddObjectMetaFieldsSet(fieldsSet, &bucket.ObjectMeta, true)
}

func BucketPoolRefNameIndexFunc(obj any) ([]string, error) {
	bucket, ok := obj.(*storage.Bucket)
	if !ok {
		return nil, fmt.Errorf("not a bucket")
	}
	return []string{bucketBucketPoolRefName(bucket)}, nil
}

func BucketPoolRefNameTriggerFunc(obj runtime.Object) string {
	return bucketBucketPoolRefName(obj.(*storage.Bucket))
}

func Indexers() *cache.Indexers {
	return &cache.Indexers{
		apisrvstorage.FieldIndex(storage.BucketBucketPoolRefNameField): BucketPoolRefNameIndexFunc,
	}
}

type bucketStrategy struct {
	runtime.ObjectTyper
	names.NameGenerator
}

var Strategy = bucketStrategy{api.Scheme, names.SimpleNameGenerator}

func (bucketStrategy) NamespaceScoped() bool {
	return true
}

func (bucketStrategy) PrepareForCreate(ctx context.Context, obj runtime.Object) {
}

func (bucketStrategy) PrepareForUpdate(ctx context.Context, obj, old runtime.Object) {
}

func (bucketStrategy) Validate(ctx context.Context, obj runtime.Object) field.ErrorList {
	bucket := obj.(*storage.Bucket)
	return validation.ValidateBucket(bucket)
}

func (bucketStrategy) WarningsOnCreate(ctx context.Context, obj runtime.Object) []string {
	return nil
}

func (bucketStrategy) AllowCreateOnUpdate() bool {
	return false
}

func (bucketStrategy) AllowUnconditionalUpdate() bool {
	return false
}

func (bucketStrategy) Canonicalize(obj runtime.Object) {
}

func (bucketStrategy) ValidateUpdate(ctx context.Context, obj, old runtime.Object) field.ErrorList {
	newBucket, oldBucket := obj.(*storage.Bucket), old.(*storage.Bucket)
	return validation.ValidateBucketUpdate(newBucket, oldBucket)
}

func (bucketStrategy) WarningsOnUpdate(ctx context.Context, obj, old runtime.Object) []string {
	return nil
}

type bucketStatusStrategy struct {
	bucketStrategy
}

var StatusStrategy = bucketStatusStrategy{Strategy}

func (bucketStatusStrategy) GetResetFields() map[fieldpath.APIVersion]*fieldpath.Set {
	return map[fieldpath.APIVersion]*fieldpath.Set{
		"storage.ironcore.dev/v1alpha1": fieldpath.NewSet(
			fieldpath.MakePathOrDie("spec"),
		),
	}
}

func (bucketStatusStrategy) PrepareForUpdate(ctx context.Context, obj, old runtime.Object) {
	newBucket := obj.(*storage.Bucket)
	oldBucket := old.(*storage.Bucket)
	newBucket.Spec = oldBucket.Spec
}

func (bucketStatusStrategy) ValidateUpdate(ctx context.Context, obj, old runtime.Object) field.ErrorList {
	return nil
}

func (bucketStatusStrategy) WarningsOnUpdate(cxt context.Context, obj, old runtime.Object) []string {
	return nil
}
