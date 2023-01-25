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

package networkinterface

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
		"networking.api.onmetal.de/v1alpha1": fieldpath.NewSet(
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
