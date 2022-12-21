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

package bucketpool

import (
	"context"
	"fmt"

	"github.com/onmetal/onmetal-api/apiutils/equality"
	"github.com/onmetal/onmetal-api/onmetal-apiserver/internal/api"
	"github.com/onmetal/onmetal-api/onmetal-apiserver/internal/apis/storage"
	"github.com/onmetal/onmetal-api/onmetal-apiserver/internal/apis/storage/validation"
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
	bucketPool, ok := obj.(*storage.BucketPool)
	if !ok {
		return nil, nil, fmt.Errorf("given object is not a BucketPoolRef")
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
		"storage.api.onmetal.de/v1alpha1": fieldpath.NewSet(
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
		"storage.api.onmetal.de/v1alpha1": fieldpath.NewSet(
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
