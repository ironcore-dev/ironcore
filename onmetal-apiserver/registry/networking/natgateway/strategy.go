/*
 * Copyright (c) 2022 by the OnMetal authors.
 *
 * Licensed under the Apache License, Version 2.0 (the "License");
 * you may not use this file except in compliance with the License.
 * You may obtain a copy of the License at
 *
 *     http://www.apache.org/licenses/LICENSE-2.0
 *
 * Unless required by applicable law or agreed to in writing, software
 * distributed under the License is distributed on an "AS IS" BASIS,
 * WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
 * See the License for the specific language governing permissions and
 * limitations under the License.
 */

package natgateway

import (
	"context"
	"fmt"

	"github.com/onmetal/onmetal-api/apis/networking"
	"github.com/onmetal/onmetal-api/apis/networking/validation"
	"github.com/onmetal/onmetal-api/onmetal-apiserver/api"
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
		"networking.api.onmetal.de/v1alpha1": fieldpath.NewSet(
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
