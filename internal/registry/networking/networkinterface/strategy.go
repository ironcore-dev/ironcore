// SPDX-FileCopyrightText: 2023 SAP SE or an SAP affiliate company and IronCore contributors
// SPDX-License-Identifier: Apache-2.0

package networkinterface

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
	"sigs.k8s.io/structured-merge-diff/v6/fieldpath"
)

func GetAttrs(obj runtime.Object) (labels.Set, fields.Set, error) {
	networkInterface, ok := obj.(*networking.NetworkInterface)
	if !ok {
		return nil, nil, fmt.Errorf("given object is not a NetworkInterface")
	}
	return networkInterface.Labels, SelectableFields(networkInterface), nil
}

func MatchNetworkInterface(label labels.Selector, field fields.Selector) apisrvstorage.SelectionPredicate {
	return apisrvstorage.SelectionPredicate{
		Label:    label,
		Field:    field,
		GetAttrs: GetAttrs,
	}
}

func SelectableFields(networkInterface *networking.NetworkInterface) fields.Set {
	return generic.ObjectMetaFieldsSet(&networkInterface.ObjectMeta, true)
}

type networkInterfaceStrategy struct {
	runtime.ObjectTyper
	names.NameGenerator
}

var Strategy = networkInterfaceStrategy{api.Scheme, names.SimpleNameGenerator}

func (networkInterfaceStrategy) NamespaceScoped() bool {
	return true
}

func (networkInterfaceStrategy) PrepareForCreate(ctx context.Context, obj runtime.Object) {
}

func (networkInterfaceStrategy) PrepareForUpdate(ctx context.Context, obj, old runtime.Object) {
}

func (networkInterfaceStrategy) Validate(ctx context.Context, obj runtime.Object) field.ErrorList {
	networkInterface := obj.(*networking.NetworkInterface)
	return validation.ValidateNetworkInterface(networkInterface)
}

func (networkInterfaceStrategy) WarningsOnCreate(ctx context.Context, obj runtime.Object) []string {
	return nil
}

func (networkInterfaceStrategy) AllowCreateOnUpdate() bool {
	return false
}

func (networkInterfaceStrategy) AllowUnconditionalUpdate() bool {
	return false
}

func (networkInterfaceStrategy) Canonicalize(obj runtime.Object) {
}

func (networkInterfaceStrategy) ValidateUpdate(ctx context.Context, obj, old runtime.Object) field.ErrorList {
	oldNetworkInterface := old.(*networking.NetworkInterface)
	newNetworkInterfae := obj.(*networking.NetworkInterface)
	return validation.ValidateNetworkInterfaceUpdate(newNetworkInterfae, oldNetworkInterface)
}

func (networkInterfaceStrategy) WarningsOnUpdate(ctx context.Context, obj, old runtime.Object) []string {
	return nil
}

type networkInterfaceStatusStrategy struct {
	networkInterfaceStrategy
}

var StatusStrategy = networkInterfaceStatusStrategy{Strategy}

func (networkInterfaceStatusStrategy) GetResetFields() map[fieldpath.APIVersion]*fieldpath.Set {
	return map[fieldpath.APIVersion]*fieldpath.Set{
		"networking.ironcore.dev/v1alpha1": fieldpath.NewSet(
			fieldpath.MakePathOrDie("spec"),
		),
	}
}

func (networkInterfaceStatusStrategy) PrepareForUpdate(ctx context.Context, obj, old runtime.Object) {
	newNetworkInterface := obj.(*networking.NetworkInterface)
	oldNetworkInterface := old.(*networking.NetworkInterface)
	newNetworkInterface.Spec = oldNetworkInterface.Spec
}

func (networkInterfaceStatusStrategy) ValidateUpdate(ctx context.Context, obj, old runtime.Object) field.ErrorList {
	newNetworkInterface := obj.(*networking.NetworkInterface)
	oldNetworkInterface := old.(*networking.NetworkInterface)
	return validation.ValidateNetworkInterfaceUpdate(newNetworkInterface, oldNetworkInterface)
}

func (networkInterfaceStatusStrategy) WarningsOnUpdate(cxt context.Context, obj, old runtime.Object) []string {
	return nil
}
