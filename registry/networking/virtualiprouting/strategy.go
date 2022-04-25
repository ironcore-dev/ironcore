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

package virtualiprouting

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
	virtualIPRouting, ok := obj.(*networking.VirtualIPRouting)
	if !ok {
		return nil, nil, fmt.Errorf("given object is not a VirtualIPRouting")
	}
	return virtualIPRouting.Labels, SelectableFields(virtualIPRouting), nil
}

func MatchVirtualIPRouting(label labels.Selector, field fields.Selector) apisrvstorage.SelectionPredicate {
	return apisrvstorage.SelectionPredicate{
		Label:    label,
		Field:    field,
		GetAttrs: GetAttrs,
	}
}

func SelectableFields(virtualIPRouting *networking.VirtualIPRouting) fields.Set {
	return generic.ObjectMetaFieldsSet(&virtualIPRouting.ObjectMeta, true)
}

type virtualIPRoutingStrategy struct {
	runtime.ObjectTyper
	names.NameGenerator
}

var Strategy = virtualIPRoutingStrategy{api.Scheme, names.SimpleNameGenerator}

func (virtualIPRoutingStrategy) NamespaceScoped() bool {
	return true
}

func (virtualIPRoutingStrategy) PrepareForCreate(ctx context.Context, obj runtime.Object) {
}

func (virtualIPRoutingStrategy) PrepareForUpdate(ctx context.Context, obj, old runtime.Object) {
}

func (virtualIPRoutingStrategy) Validate(ctx context.Context, obj runtime.Object) field.ErrorList {
	virtualIPRouting := obj.(*networking.VirtualIPRouting)
	return validation.ValidateVirtualIPRouting(virtualIPRouting)
}

func (virtualIPRoutingStrategy) WarningsOnCreate(ctx context.Context, obj runtime.Object) []string {
	return nil
}

func (virtualIPRoutingStrategy) AllowCreateOnUpdate() bool {
	return false
}

func (virtualIPRoutingStrategy) AllowUnconditionalUpdate() bool {
	return false
}

func (virtualIPRoutingStrategy) Canonicalize(obj runtime.Object) {
}

func (virtualIPRoutingStrategy) ValidateUpdate(ctx context.Context, obj, old runtime.Object) field.ErrorList {
	newVirtualIPRouting := obj.(*networking.VirtualIPRouting)
	oldVirtualIPRouting := old.(*networking.VirtualIPRouting)
	return validation.ValidateVirtualIPRoutingUpdate(newVirtualIPRouting, oldVirtualIPRouting)
}

func (virtualIPRoutingStrategy) WarningsOnUpdate(ctx context.Context, obj, old runtime.Object) []string {
	return nil
}
