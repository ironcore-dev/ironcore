// SPDX-FileCopyrightText: 2023 SAP SE or an SAP affiliate company and IronCore contributors
// SPDX-License-Identifier: Apache-2.0

package network

import (
	"context"
	"fmt"

	"github.com/ironcore-dev/ironcore/internal/api"
	"github.com/ironcore-dev/ironcore/internal/apis/networking"
	"github.com/ironcore-dev/ironcore/internal/apis/networking/validation"
	"github.com/ironcore-dev/ironcore/utils/equality"
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
	network, ok := obj.(*networking.Network)
	if !ok {
		return nil, nil, fmt.Errorf("given object is not a Network")
	}
	return network.Labels, SelectableFields(network), nil
}

func MatchNetwork(label labels.Selector, field fields.Selector) apisrvstorage.SelectionPredicate {
	return apisrvstorage.SelectionPredicate{
		Label:    label,
		Field:    field,
		GetAttrs: GetAttrs,
	}
}

func SelectableFields(network *networking.Network) fields.Set {
	return generic.ObjectMetaFieldsSet(&network.ObjectMeta, true)
}

type networkStrategy struct {
	runtime.ObjectTyper
	names.NameGenerator
}

var Strategy = networkStrategy{api.Scheme, names.SimpleNameGenerator}

func (networkStrategy) NamespaceScoped() bool {
	return true
}

func (networkStrategy) PrepareForCreate(ctx context.Context, obj runtime.Object) {
	machinePool := obj.(*networking.Network)
	machinePool.Status = networking.NetworkStatus{}
	machinePool.Generation = 1
}

func (networkStrategy) PrepareForUpdate(ctx context.Context, obj, old runtime.Object) {
	newNetwork, oldNetwork := obj.(*networking.Network), old.(*networking.Network)
	newNetwork.Status = oldNetwork.Status

	if !equality.Semantic.DeepEqual(newNetwork.Spec, oldNetwork.Spec) {
		newNetwork.Generation = oldNetwork.Generation + 1
	}
}

func (networkStrategy) Validate(ctx context.Context, obj runtime.Object) field.ErrorList {
	network := obj.(*networking.Network)
	return validation.ValidateNetwork(network)
}

func (networkStrategy) WarningsOnCreate(ctx context.Context, obj runtime.Object) []string {
	return nil
}

func (networkStrategy) AllowCreateOnUpdate() bool {
	return false
}

func (networkStrategy) AllowUnconditionalUpdate() bool {
	return false
}

func (networkStrategy) Canonicalize(obj runtime.Object) {
}

func (networkStrategy) ValidateUpdate(ctx context.Context, obj, old runtime.Object) field.ErrorList {
	newNetwork, oldNetwork := obj.(*networking.Network), old.(*networking.Network)
	return validation.ValidateNetworkUpdate(newNetwork, oldNetwork)
}

func (networkStrategy) WarningsOnUpdate(ctx context.Context, obj, old runtime.Object) []string {
	return nil
}

type networkStatusStrategy struct {
	networkStrategy
}

var StatusStrategy = networkStatusStrategy{Strategy}

func (networkStatusStrategy) GetResetFields() map[fieldpath.APIVersion]*fieldpath.Set {
	return map[fieldpath.APIVersion]*fieldpath.Set{
		"networking.ironcore.dev/v1alpha1": fieldpath.NewSet(
			fieldpath.MakePathOrDie("spec"),
		),
	}
}

func (networkStatusStrategy) PrepareForUpdate(ctx context.Context, obj, old runtime.Object) {
	newNetwork, oldNetwork := obj.(*networking.Network), old.(*networking.Network)
	newNetwork.Spec = oldNetwork.Spec
}

func (networkStatusStrategy) ValidateUpdate(ctx context.Context, obj, old runtime.Object) field.ErrorList {
	newNetwork := obj.(*networking.Network)
	oldNetwork := old.(*networking.Network)
	return validation.ValidateNetworkUpdate(newNetwork, oldNetwork)
}

func (networkStatusStrategy) WarningsOnUpdate(cxt context.Context, obj, old runtime.Object) []string {
	return nil
}
