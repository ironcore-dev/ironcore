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

package virtualipclaim

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
	"sigs.k8s.io/structured-merge-diff/v4/fieldpath"
)

func GetAttrs(obj runtime.Object) (labels.Set, fields.Set, error) {
	virtualIPClaim, ok := obj.(*networking.VirtualIPClaim)
	if !ok {
		return nil, nil, fmt.Errorf("given object is not a VirtualIPClaim")
	}
	return virtualIPClaim.Labels, SelectableFields(virtualIPClaim), nil
}

func MatchVirtualIPClaim(label labels.Selector, field fields.Selector) apisrvstorage.SelectionPredicate {
	return apisrvstorage.SelectionPredicate{
		Label:    label,
		Field:    field,
		GetAttrs: GetAttrs,
	}
}

func SelectableFields(virtualIPClaim *networking.VirtualIPClaim) fields.Set {
	return generic.ObjectMetaFieldsSet(&virtualIPClaim.ObjectMeta, true)
}

type virtualIPClaimStrategy struct {
	runtime.ObjectTyper
	names.NameGenerator
}

var Strategy = virtualIPClaimStrategy{api.Scheme, names.SimpleNameGenerator}

func (virtualIPClaimStrategy) NamespaceScoped() bool {
	return true
}

func (virtualIPClaimStrategy) PrepareForCreate(ctx context.Context, obj runtime.Object) {
}

func (virtualIPClaimStrategy) PrepareForUpdate(ctx context.Context, obj, old runtime.Object) {
}

func (virtualIPClaimStrategy) Validate(ctx context.Context, obj runtime.Object) field.ErrorList {
	virtualIPClaim := obj.(*networking.VirtualIPClaim)
	return validation.ValidateVirtualIPClaim(virtualIPClaim)
}

func (virtualIPClaimStrategy) WarningsOnCreate(ctx context.Context, obj runtime.Object) []string {
	return nil
}

func (virtualIPClaimStrategy) AllowCreateOnUpdate() bool {
	return false
}

func (virtualIPClaimStrategy) AllowUnconditionalUpdate() bool {
	return false
}

func (virtualIPClaimStrategy) Canonicalize(obj runtime.Object) {
}

func (virtualIPClaimStrategy) ValidateUpdate(ctx context.Context, obj, old runtime.Object) field.ErrorList {
	return field.ErrorList{}
}

func (virtualIPClaimStrategy) WarningsOnUpdate(ctx context.Context, obj, old runtime.Object) []string {
	return nil
}

type virtualIPClaimStatusStrategy struct {
	virtualIPClaimStrategy
}

var StatusStrategy = virtualIPClaimStatusStrategy{Strategy}

func (virtualIPClaimStatusStrategy) GetResetFields() map[fieldpath.APIVersion]*fieldpath.Set {
	return map[fieldpath.APIVersion]*fieldpath.Set{
		"networking.api.onmetal.de/v1alpha1": fieldpath.NewSet(
			fieldpath.MakePathOrDie("spec"),
		),
	}
}

func (virtualIPClaimStatusStrategy) PrepareForUpdate(ctx context.Context, obj, old runtime.Object) {
	newVirtualIPClaim := obj.(*networking.VirtualIPClaim)
	oldVirtualIPClaim := old.(*networking.VirtualIPClaim)
	newVirtualIPClaim.Spec = oldVirtualIPClaim.Spec
}

func (virtualIPClaimStatusStrategy) ValidateUpdate(ctx context.Context, obj, old runtime.Object) field.ErrorList {
	newVirtualIPClaim := obj.(*networking.VirtualIPClaim)
	oldVirtualIPClaim := old.(*networking.VirtualIPClaim)
	return validation.ValidateVirtualIPClaimUpdate(newVirtualIPClaim, oldVirtualIPClaim)
}

func (virtualIPClaimStatusStrategy) WarningsOnUpdate(cxt context.Context, obj, old runtime.Object) []string {
	return nil
}
