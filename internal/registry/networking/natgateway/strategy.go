// SPDX-FileCopyrightText: 2023 SAP SE or an SAP affiliate company and IronCore contributors
// SPDX-License-Identifier: Apache-2.0

package natgateway

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
	natGateway, ok := obj.(*networking.NATGateway)
	if !ok {
		return nil, nil, fmt.Errorf("given object is not a NATGateway")
	}
	return natGateway.Labels, SelectableFields(natGateway), nil
}

func MatchNATGateway(label labels.Selector, field fields.Selector) apisrvstorage.SelectionPredicate {
	return apisrvstorage.SelectionPredicate{
		Label:    label,
		Field:    field,
		GetAttrs: GetAttrs,
	}
}

func SelectableFields(natGateway *networking.NATGateway) fields.Set {
	return generic.ObjectMetaFieldsSet(&natGateway.ObjectMeta, true)
}

type natGatewayStrategy struct {
	runtime.ObjectTyper
	names.NameGenerator
}

var Strategy = natGatewayStrategy{api.Scheme, names.SimpleNameGenerator}

func (natGatewayStrategy) NamespaceScoped() bool {
	return true
}

func (natGatewayStrategy) PrepareForCreate(ctx context.Context, obj runtime.Object) {
}

func (natGatewayStrategy) PrepareForUpdate(ctx context.Context, obj, old runtime.Object) {
}

func (natGatewayStrategy) Validate(ctx context.Context, obj runtime.Object) field.ErrorList {
	natGateway := obj.(*networking.NATGateway)
	return validation.ValidateNATGateway(natGateway)
}

func (natGatewayStrategy) WarningsOnCreate(ctx context.Context, obj runtime.Object) []string {
	return nil
}

func (natGatewayStrategy) AllowCreateOnUpdate() bool {
	return false
}

func (natGatewayStrategy) AllowUnconditionalUpdate() bool {
	return false
}

func (natGatewayStrategy) Canonicalize(obj runtime.Object) {
}

func (natGatewayStrategy) ValidateUpdate(ctx context.Context, obj, old runtime.Object) field.ErrorList {
	newNATGateway := obj.(*networking.NATGateway)
	oldNATGateway := old.(*networking.NATGateway)
	return validation.ValidateNATGatewayUpdate(newNATGateway, oldNATGateway)
}

func (natGatewayStrategy) WarningsOnUpdate(ctx context.Context, obj, old runtime.Object) []string {
	return nil
}

type natGatewayStatusStrategy struct {
	natGatewayStrategy
}

var StatusStrategy = natGatewayStatusStrategy{Strategy}

func (natGatewayStatusStrategy) GetResetFields() map[fieldpath.APIVersion]*fieldpath.Set {
	return map[fieldpath.APIVersion]*fieldpath.Set{
		"networking.ironcore.dev/v1alpha1": fieldpath.NewSet(
			fieldpath.MakePathOrDie("spec"),
		),
	}
}

func (natGatewayStatusStrategy) PrepareForUpdate(ctx context.Context, obj, old runtime.Object) {
}

func (natGatewayStatusStrategy) ValidateUpdate(ctx context.Context, obj, old runtime.Object) field.ErrorList {
	newNATGateway := obj.(*networking.NATGateway)
	oldNATGateway := old.(*networking.NATGateway)
	return validation.ValidateNATGatewayUpdate(newNATGateway, oldNATGateway)
}

func (natGatewayStatusStrategy) WarningsOnUpdate(cxt context.Context, obj, old runtime.Object) []string {
	return nil
}
