// SPDX-FileCopyrightText: 2023 SAP SE or an SAP affiliate company and IronCore contributors
// SPDX-License-Identifier: Apache-2.0

package bucketclass

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
	bucketClass, ok := obj.(*storage.BucketClass)
	if !ok {
		return nil, nil, fmt.Errorf("given object is not a BucketClass")
	}
	return bucketClass.Labels, SelectableFields(bucketClass), nil
}

func MatchBucketClass(label labels.Selector, field fields.Selector) apisrvstorage.SelectionPredicate {
	return apisrvstorage.SelectionPredicate{
		Label:    label,
		Field:    field,
		GetAttrs: GetAttrs,
	}
}

func SelectableFields(bucketClass *storage.BucketClass) fields.Set {
	return generic.ObjectMetaFieldsSet(&bucketClass.ObjectMeta, false)
}

type bucketClassStrategy struct {
	runtime.ObjectTyper
	names.NameGenerator
}

var Strategy = bucketClassStrategy{api.Scheme, names.SimpleNameGenerator}

func (bucketClassStrategy) NamespaceScoped() bool {
	return false
}

func (bucketClassStrategy) PrepareForCreate(ctx context.Context, obj runtime.Object) {
}

func (bucketClassStrategy) PrepareForUpdate(ctx context.Context, obj, old runtime.Object) {
}

func (bucketClassStrategy) Validate(ctx context.Context, obj runtime.Object) field.ErrorList {
	bucketClass := obj.(*storage.BucketClass)
	return validation.ValidateBucketClass(bucketClass)
}

func (bucketClassStrategy) WarningsOnCreate(ctx context.Context, obj runtime.Object) []string {
	return nil
}

func (bucketClassStrategy) AllowCreateOnUpdate() bool {
	return false
}

func (bucketClassStrategy) AllowUnconditionalUpdate() bool {
	return false
}

func (bucketClassStrategy) Canonicalize(obj runtime.Object) {
}

func (bucketClassStrategy) ValidateUpdate(ctx context.Context, obj, old runtime.Object) field.ErrorList {
	newBucketClass := obj.(*storage.BucketClass)
	oldBucketClass := old.(*storage.BucketClass)
	return validation.ValidateBucketClassUpdate(newBucketClass, oldBucketClass)
}

func (bucketClassStrategy) WarningsOnUpdate(ctx context.Context, obj, old runtime.Object) []string {
	return nil
}
