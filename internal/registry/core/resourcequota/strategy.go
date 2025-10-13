// SPDX-FileCopyrightText: 2023 SAP SE or an SAP affiliate company and IronCore contributors
// SPDX-License-Identifier: Apache-2.0

package resourcequota

import (
	"context"
	"fmt"

	"github.com/ironcore-dev/ironcore/internal/api"
	"github.com/ironcore-dev/ironcore/internal/apis/core"
	"github.com/ironcore-dev/ironcore/internal/apis/core/validation"
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
		"core.ironcore.dev/v1alpha1": fieldpath.NewSet(
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
