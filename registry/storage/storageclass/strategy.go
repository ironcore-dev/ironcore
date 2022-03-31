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

package storageclass

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
)

func GetAttrs(obj runtime.Object) (labels.Set, fields.Set, error) {
	storageClass, ok := obj.(*storage.StorageClass)
	if !ok {
		return nil, nil, fmt.Errorf("given object is not a StorageClass")
	}
	return storageClass.Labels, SelectableFields(storageClass), nil
}

func MatchStorageClass(label labels.Selector, field fields.Selector) apisrvstorage.SelectionPredicate {
	return apisrvstorage.SelectionPredicate{
		Label:    label,
		Field:    field,
		GetAttrs: GetAttrs,
	}
}

func SelectableFields(storageClass *storage.StorageClass) fields.Set {
	return generic.ObjectMetaFieldsSet(&storageClass.ObjectMeta, false)
}

type storageClassStrategy struct {
	runtime.ObjectTyper
	names.NameGenerator
}

var Strategy = storageClassStrategy{api.Scheme, names.SimpleNameGenerator}

func (storageClassStrategy) NamespaceScoped() bool {
	return false
}

func (storageClassStrategy) PrepareForCreate(ctx context.Context, obj runtime.Object) {
}

func (storageClassStrategy) PrepareForUpdate(ctx context.Context, obj, old runtime.Object) {
}

func (storageClassStrategy) Validate(ctx context.Context, obj runtime.Object) field.ErrorList {
	return field.ErrorList{}
}

func (storageClassStrategy) WarningsOnCreate(ctx context.Context, obj runtime.Object) []string {
	return nil
}

func (storageClassStrategy) AllowCreateOnUpdate() bool {
	return false
}

func (storageClassStrategy) AllowUnconditionalUpdate() bool {
	return false
}

func (storageClassStrategy) Canonicalize(obj runtime.Object) {
}

func (storageClassStrategy) ValidateUpdate(ctx context.Context, obj, old runtime.Object) field.ErrorList {
	return field.ErrorList{}
}

func (storageClassStrategy) WarningsOnUpdate(ctx context.Context, obj, old runtime.Object) []string {
	return nil
}
