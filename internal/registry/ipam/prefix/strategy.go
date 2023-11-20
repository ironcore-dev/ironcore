// SPDX-FileCopyrightText: 2023 SAP SE or an SAP affiliate company and IronCore contributors
// SPDX-License-Identifier: Apache-2.0

package prefix

import (
	"context"
	"fmt"

	"github.com/ironcore-dev/ironcore/internal/api"
	"github.com/ironcore-dev/ironcore/internal/apis/ipam"
	"github.com/ironcore-dev/ironcore/internal/apis/ipam/validation"
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
	prefix, ok := obj.(*ipam.Prefix)
	if !ok {
		return nil, nil, fmt.Errorf("given object is not a Prefix")
	}
	return prefix.Labels, SelectableFields(prefix), nil
}

func MatchPrefix(label labels.Selector, field fields.Selector) apisrvstorage.SelectionPredicate {
	return apisrvstorage.SelectionPredicate{
		Label:    label,
		Field:    field,
		GetAttrs: GetAttrs,
	}
}

func SelectableFields(prefix *ipam.Prefix) fields.Set {
	return generic.ObjectMetaFieldsSet(&prefix.ObjectMeta, true)
}

type prefixStrategy struct {
	runtime.ObjectTyper
	names.NameGenerator
}

var Strategy = prefixStrategy{api.Scheme, names.SimpleNameGenerator}

func (prefixStrategy) NamespaceScoped() bool {
	return true
}

func (prefixStrategy) PrepareForCreate(ctx context.Context, obj runtime.Object) {
}

func (prefixStrategy) PrepareForUpdate(ctx context.Context, obj, old runtime.Object) {
}

func (prefixStrategy) Validate(ctx context.Context, obj runtime.Object) field.ErrorList {
	prefix := obj.(*ipam.Prefix)
	return validation.ValidatePrefix(prefix)
}

func (prefixStrategy) WarningsOnCreate(ctx context.Context, obj runtime.Object) []string {
	return nil
}

func (prefixStrategy) AllowCreateOnUpdate() bool {
	return false
}

func (prefixStrategy) AllowUnconditionalUpdate() bool {
	return false
}

func (prefixStrategy) Canonicalize(obj runtime.Object) {
}

func (prefixStrategy) ValidateUpdate(ctx context.Context, obj, old runtime.Object) field.ErrorList {
	newPrefix := obj.(*ipam.Prefix)
	oldPrefix := old.(*ipam.Prefix)
	return validation.ValidatePrefixUpdate(newPrefix, oldPrefix)
}

func (prefixStrategy) WarningsOnUpdate(ctx context.Context, obj, old runtime.Object) []string {
	return nil
}

type prefixStatusStrategy struct {
	prefixStrategy
}

var StatusStrategy = prefixStatusStrategy{Strategy}

func (prefixStatusStrategy) GetResetFields() map[fieldpath.APIVersion]*fieldpath.Set {
	return map[fieldpath.APIVersion]*fieldpath.Set{
		"ipam.ironcore.dev/v1alpha1": fieldpath.NewSet(
			fieldpath.MakePathOrDie("spec"),
		),
	}
}

func (prefixStatusStrategy) PrepareForUpdate(ctx context.Context, obj, old runtime.Object) {
	newPrefix := obj.(*ipam.Prefix)
	oldPrefix := old.(*ipam.Prefix)
	newPrefix.Spec = oldPrefix.Spec
}

func (prefixStatusStrategy) ValidateUpdate(ctx context.Context, obj, old runtime.Object) field.ErrorList {
	newPrefix := obj.(*ipam.Prefix)
	oldPrefix := old.(*ipam.Prefix)
	return validation.ValidatePrefixStatusUpdate(newPrefix, oldPrefix)
}

func (prefixStatusStrategy) WarningsOnUpdate(cxt context.Context, obj, old runtime.Object) []string {
	return nil
}
