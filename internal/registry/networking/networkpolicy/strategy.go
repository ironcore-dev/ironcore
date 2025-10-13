// SPDX-FileCopyrightText: 2023 SAP SE or an SAP affiliate company and IronCore contributors
// SPDX-License-Identifier: Apache-2.0

package networkpolicy

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
		"networking.ironcore.dev/v1alpha1": fieldpath.NewSet(
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
