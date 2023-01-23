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

package resourcequota

import (
	"context"
	"fmt"

	"github.com/onmetal/onmetal-api/onmetal-apiserver/internal/api"
	"github.com/onmetal/onmetal-api/onmetal-apiserver/internal/apis/core"
	"github.com/onmetal/onmetal-api/onmetal-apiserver/internal/apis/core/validation"
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
	resourceQuota, ok := obj.(*core.ResourceQuota)
	if !ok {
		return nil, nil, fmt.Errorf("given object is not a ResourceQuota")
	}
	return resourceQuota.Labels, SelectableFields(resourceQuota), nil
}

func MatchResourceQuota(label labels.Selector, field fields.Selector) apisrvstorage.SelectionPredicate {
	return apisrvstorage.SelectionPredicate{
		Label:    label,
		Field:    field,
		GetAttrs: GetAttrs,
	}
}

func SelectableFields(resourceQuota *core.ResourceQuota) fields.Set {
	return generic.ObjectMetaFieldsSet(&resourceQuota.ObjectMeta, true)
}

type resourceQuotaStrategy struct {
	runtime.ObjectTyper
	names.NameGenerator
}

var Strategy = resourceQuotaStrategy{api.Scheme, names.SimpleNameGenerator}

func (resourceQuotaStrategy) NamespaceScoped() bool {
	return true
}

func (resourceQuotaStrategy) PrepareForCreate(ctx context.Context, obj runtime.Object) {
}

func (resourceQuotaStrategy) PrepareForUpdate(ctx context.Context, obj, old runtime.Object) {
}

func (resourceQuotaStrategy) Validate(ctx context.Context, obj runtime.Object) field.ErrorList {
	resourceQuota := obj.(*core.ResourceQuota)
	return validation.ValidateResourceQuota(resourceQuota)
}

func (resourceQuotaStrategy) WarningsOnCreate(ctx context.Context, obj runtime.Object) []string {
	return nil
}

func (resourceQuotaStrategy) AllowCreateOnUpdate() bool {
	return false
}

func (resourceQuotaStrategy) AllowUnconditionalUpdate() bool {
	return false
}

func (resourceQuotaStrategy) Canonicalize(obj runtime.Object) {
}

func (resourceQuotaStrategy) ValidateUpdate(ctx context.Context, obj, old runtime.Object) field.ErrorList {
	newResourceQuota := obj.(*core.ResourceQuota)
	oldResourceQuota := old.(*core.ResourceQuota)
	return validation.ValidateResourceQuotaUpdate(newResourceQuota, oldResourceQuota)
}

func (resourceQuotaStrategy) WarningsOnUpdate(ctx context.Context, obj, old runtime.Object) []string {
	return nil
}

type resourceQuotaStatusStrategy struct {
	resourceQuotaStrategy
}

var StatusStrategy = resourceQuotaStatusStrategy{Strategy}

func (resourceQuotaStatusStrategy) GetResetFields() map[fieldpath.APIVersion]*fieldpath.Set {
	return map[fieldpath.APIVersion]*fieldpath.Set{
		"core.api.onmetal.de/v1alpha1": fieldpath.NewSet(
			fieldpath.MakePathOrDie("spec"),
		),
	}
}

func (resourceQuotaStatusStrategy) PrepareForUpdate(ctx context.Context, obj, old runtime.Object) {
	newResourceQuota := obj.(*core.ResourceQuota)
	oldResourceQuota := old.(*core.ResourceQuota)
	newResourceQuota.Spec = oldResourceQuota.Spec
}

func (resourceQuotaStatusStrategy) ValidateUpdate(ctx context.Context, obj, old runtime.Object) field.ErrorList {
	newResourceQuota := obj.(*core.ResourceQuota)
	oldResourceQuota := old.(*core.ResourceQuota)
	return validation.ValidateResourceQuotaStatusUpdate(newResourceQuota, oldResourceQuota)
}

func (resourceQuotaStatusStrategy) WarningsOnUpdate(cxt context.Context, obj, old runtime.Object) []string {
	return nil
}
