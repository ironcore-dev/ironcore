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

package networkinterfacebinding

import (
	"context"
	"fmt"

	"github.com/onmetal/onmetal-api/api"
	"github.com/onmetal/onmetal-api/apis/networking"
	"github.com/onmetal/onmetal-api/apis/networking/validation"
	"k8s.io/apimachinery/pkg/fields"
	"k8s.io/apimachinery/pkg/labels"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apimachinery/pkg/util/validation/field"
	"k8s.io/apiserver/pkg/registry/generic"
	apisrvstorage "k8s.io/apiserver/pkg/storage"
	"k8s.io/apiserver/pkg/storage/names"
)

func GetAttrs(obj runtime.Object) (labels.Set, fields.Set, error) {
	networkInterfaceBinding, ok := obj.(*networking.NetworkInterfaceBinding)
	if !ok {
		return nil, nil, fmt.Errorf("given object is not a NetworkInterfaceBinding")
	}
	return networkInterfaceBinding.Labels, SelectableFields(networkInterfaceBinding), nil
}

func MatchNetworkInterfaceBinding(label labels.Selector, field fields.Selector) apisrvstorage.SelectionPredicate {
	return apisrvstorage.SelectionPredicate{
		Label:    label,
		Field:    field,
		GetAttrs: GetAttrs,
	}
}

func SelectableFields(networkInterfaceBinding *networking.NetworkInterfaceBinding) fields.Set {
	return generic.ObjectMetaFieldsSet(&networkInterfaceBinding.ObjectMeta, true)
}

type networkInterfaceBindingStrategy struct {
	runtime.ObjectTyper
	names.NameGenerator
}

var Strategy = networkInterfaceBindingStrategy{api.Scheme, names.SimpleNameGenerator}

func (networkInterfaceBindingStrategy) NamespaceScoped() bool {
	return true
}

func (networkInterfaceBindingStrategy) PrepareForCreate(ctx context.Context, obj runtime.Object) {
}

func (networkInterfaceBindingStrategy) PrepareForUpdate(ctx context.Context, obj, old runtime.Object) {
}

func (networkInterfaceBindingStrategy) Validate(ctx context.Context, obj runtime.Object) field.ErrorList {
	networkInterfaceBinding := obj.(*networking.NetworkInterfaceBinding)
	return validation.ValidateNetworkInterfaceBinding(networkInterfaceBinding)
}

func (networkInterfaceBindingStrategy) WarningsOnCreate(ctx context.Context, obj runtime.Object) []string {
	return nil
}

func (networkInterfaceBindingStrategy) AllowCreateOnUpdate() bool {
	return false
}

func (networkInterfaceBindingStrategy) AllowUnconditionalUpdate() bool {
	return false
}

func (networkInterfaceBindingStrategy) Canonicalize(obj runtime.Object) {
}

func (networkInterfaceBindingStrategy) ValidateUpdate(ctx context.Context, obj, old runtime.Object) field.ErrorList {
	newNetworkInterfaceBinding := obj.(*networking.NetworkInterfaceBinding)
	oldNetworkInterfaceBinding := old.(*networking.NetworkInterfaceBinding)
	return validation.ValidateNetworkInterfaceBindingUpdate(newNetworkInterfaceBinding, oldNetworkInterfaceBinding)
}

func (networkInterfaceBindingStrategy) WarningsOnUpdate(ctx context.Context, obj, old runtime.Object) []string {
	return nil
}
