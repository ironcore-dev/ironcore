// SPDX-FileCopyrightText: 2023 SAP SE or an SAP affiliate company and IronCore contributors
// SPDX-License-Identifier: Apache-2.0

package virtualip

import (
	"context"
	"fmt"

	"github.com/ironcore-dev/ironcore/internal/api"
	"github.com/ironcore-dev/ironcore/internal/apis/networking"
	"github.com/ironcore-dev/ironcore/internal/apis/networking/validation"
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
	virtualIP, ok := obj.(*networking.VirtualIP)
	if !ok {
		return nil, nil, fmt.Errorf("given object is not a VirtualIP")
	}
	return virtualIP.Labels, SelectableFields(virtualIP), nil
}

func MatchVirtualIP(label labels.Selector, field fields.Selector) apisrvstorage.SelectionPredicate {
	return apisrvstorage.SelectionPredicate{
		Label:    label,
		Field:    field,
		GetAttrs: GetAttrs,
	}
}

func SelectableFields(virtualIP *networking.VirtualIP) fields.Set {
	return generic.ObjectMetaFieldsSet(&virtualIP.ObjectMeta, true)
}

type virtualIPStrategy struct {
	runtime.ObjectTyper
	names.NameGenerator
}

var Strategy = virtualIPStrategy{api.Scheme, names.SimpleNameGenerator}

func (virtualIPStrategy) NamespaceScoped() bool {
	return true
}

func (virtualIPStrategy) PrepareForCreate(ctx context.Context, obj runtime.Object) {
}

func (virtualIPStrategy) PrepareForUpdate(ctx context.Context, obj, old runtime.Object) {
}

func (virtualIPStrategy) Validate(ctx context.Context, obj runtime.Object) field.ErrorList {
	virtualIP := obj.(*networking.VirtualIP)
	return validation.ValidateVirtualIP(virtualIP)
}

func (virtualIPStrategy) WarningsOnCreate(ctx context.Context, obj runtime.Object) []string {
	return nil
}

func (virtualIPStrategy) AllowCreateOnUpdate() bool {
	return false
}

func (virtualIPStrategy) AllowUnconditionalUpdate() bool {
	return false
}

func (virtualIPStrategy) Canonicalize(obj runtime.Object) {
}

func (virtualIPStrategy) ValidateUpdate(ctx context.Context, obj, old runtime.Object) field.ErrorList {
	newVirtualIP := obj.(*networking.VirtualIP)
	oldVirtualIP := old.(*networking.VirtualIP)
	return validation.ValidateVirtualIPUpdate(newVirtualIP, oldVirtualIP)
}

func (virtualIPStrategy) WarningsOnUpdate(ctx context.Context, obj, old runtime.Object) []string {
	return nil
}

type virtualIPStatusStrategy struct {
	virtualIPStrategy
}

var StatusStrategy = virtualIPStatusStrategy{Strategy}

func (virtualIPStatusStrategy) GetResetFields() map[fieldpath.APIVersion]*fieldpath.Set {
	return map[fieldpath.APIVersion]*fieldpath.Set{
		"networking.ironcore.dev/v1alpha1": fieldpath.NewSet(
			fieldpath.MakePathOrDie("spec"),
		),
	}
}

func (virtualIPStatusStrategy) PrepareForUpdate(ctx context.Context, obj, old runtime.Object) {
	newVirtualIP := obj.(*networking.VirtualIP)
	oldVirtualIP := old.(*networking.VirtualIP)
	newVirtualIP.Spec = oldVirtualIP.Spec
}

func (virtualIPStatusStrategy) ValidateUpdate(ctx context.Context, obj, old runtime.Object) field.ErrorList {
	newVirtualIP := obj.(*networking.VirtualIP)
	oldVirtualIP := old.(*networking.VirtualIP)
	return validation.ValidateVirtualIPUpdate(newVirtualIP, oldVirtualIP)
}

func (virtualIPStatusStrategy) WarningsOnUpdate(cxt context.Context, obj, old runtime.Object) []string {
	return nil
}
