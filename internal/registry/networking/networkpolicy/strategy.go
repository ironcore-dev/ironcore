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

package networkpolicy

import (
	"context"
	"fmt"

	"github.com/onmetal/onmetal-api/internal/api"
	"github.com/onmetal/onmetal-api/internal/apis/networking"
	"github.com/onmetal/onmetal-api/internal/apis/networking/validation"
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
	networkPolicy, ok := obj.(*networking.NetworkPolicy)
	if !ok {
		return nil, nil, fmt.Errorf("given object is not a NetworkPolicy")
	}
	return networkPolicy.Labels, SelectableFields(networkPolicy), nil
}

func MatchNetworkPolicy(label labels.Selector, field fields.Selector) apisrvstorage.SelectionPredicate {
	return apisrvstorage.SelectionPredicate{
		Label:    label,
		Field:    field,
		GetAttrs: GetAttrs,
	}
}

func SelectableFields(networkPolicy *networking.NetworkPolicy) fields.Set {
	return generic.ObjectMetaFieldsSet(&networkPolicy.ObjectMeta, true)
}

type networkPolicyStrategy struct {
	runtime.ObjectTyper
	names.NameGenerator
}

var Strategy = networkPolicyStrategy{api.Scheme, names.SimpleNameGenerator}

func (networkPolicyStrategy) NamespaceScoped() bool {
	return true
}

func (networkPolicyStrategy) PrepareForCreate(ctx context.Context, obj runtime.Object) {
	networkPolicy := obj.(*networking.NetworkPolicy)
	networkPolicy.Status = networking.NetworkPolicyStatus{}
}

func (networkPolicyStrategy) PrepareForUpdate(ctx context.Context, obj, old runtime.Object) {
	newNetworkPolicy := obj.(*networking.NetworkPolicy)
	oldNetworkPolicy := old.(*networking.NetworkPolicy)
	newNetworkPolicy.Status = oldNetworkPolicy.Status
}

func (networkPolicyStrategy) Validate(ctx context.Context, obj runtime.Object) field.ErrorList {
	networkPolicy := obj.(*networking.NetworkPolicy)
	return validation.ValidateNetworkPolicy(networkPolicy)
}

func (networkPolicyStrategy) WarningsOnCreate(ctx context.Context, obj runtime.Object) []string {
	return nil
}

func (networkPolicyStrategy) AllowCreateOnUpdate() bool {
	return false
}

func (networkPolicyStrategy) AllowUnconditionalUpdate() bool {
	return false
}

func (networkPolicyStrategy) Canonicalize(obj runtime.Object) {
}

func (networkPolicyStrategy) ValidateUpdate(ctx context.Context, obj, old runtime.Object) field.ErrorList {
	newNetworkPolicy := obj.(*networking.NetworkPolicy)
	oldNetworkPolicy := old.(*networking.NetworkPolicy)
	return validation.ValidateNetworkPolicyUpdate(newNetworkPolicy, oldNetworkPolicy)
}

func (networkPolicyStrategy) WarningsOnUpdate(ctx context.Context, obj, old runtime.Object) []string {
	return nil
}

type networkPolicyStatusStrategy struct {
	networkPolicyStrategy
}

var StatusStrategy = networkPolicyStatusStrategy{Strategy}

func (networkPolicyStatusStrategy) GetResetFields() map[fieldpath.APIVersion]*fieldpath.Set {
	return map[fieldpath.APIVersion]*fieldpath.Set{
		"networking.api.onmetal.de/v1alpha1": fieldpath.NewSet(
			fieldpath.MakePathOrDie("spec"),
		),
	}
}

func (networkPolicyStatusStrategy) PrepareForUpdate(ctx context.Context, obj, old runtime.Object) {
}

func (networkPolicyStatusStrategy) ValidateUpdate(ctx context.Context, obj, old runtime.Object) field.ErrorList {
	newNetworkPolicy := obj.(*networking.NetworkPolicy)
	oldNetworkPolicy := old.(*networking.NetworkPolicy)
	return validation.ValidateNetworkPolicyUpdate(newNetworkPolicy, oldNetworkPolicy)
}

func (networkPolicyStatusStrategy) WarningsOnUpdate(cxt context.Context, obj, old runtime.Object) []string {
	return nil
}
