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

package ip

import (
	"context"
	"fmt"

	"github.com/onmetal/onmetal-api/api"
	"github.com/onmetal/onmetal-api/apis/ipam"
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
	ip, ok := obj.(*ipam.IP)
	if !ok {
		return nil, nil, fmt.Errorf("given object is not a IP")
	}
	return ip.Labels, SelectableFields(ip), nil
}

func MatchIP(label labels.Selector, field fields.Selector) apisrvstorage.SelectionPredicate {
	return apisrvstorage.SelectionPredicate{
		Label:    label,
		Field:    field,
		GetAttrs: GetAttrs,
	}
}

func SelectableFields(ip *ipam.IP) fields.Set {
	return generic.ObjectMetaFieldsSet(&ip.ObjectMeta, true)
}

type ipStrategy struct {
	runtime.ObjectTyper
	names.NameGenerator
}

var Strategy = ipStrategy{api.Scheme, names.SimpleNameGenerator}

func (ipStrategy) NamespaceScoped() bool {
	return true
}

func (ipStrategy) PrepareForCreate(ctx context.Context, obj runtime.Object) {
}

func (ipStrategy) PrepareForUpdate(ctx context.Context, obj, old runtime.Object) {
}

func (ipStrategy) Validate(ctx context.Context, obj runtime.Object) field.ErrorList {
	return field.ErrorList{}
}

func (ipStrategy) WarningsOnCreate(ctx context.Context, obj runtime.Object) []string {
	return nil
}

func (ipStrategy) AllowCreateOnUpdate() bool {
	return false
}

func (ipStrategy) AllowUnconditionalUpdate() bool {
	return false
}

func (ipStrategy) Canonicalize(obj runtime.Object) {
}

func (ipStrategy) ValidateUpdate(ctx context.Context, obj, old runtime.Object) field.ErrorList {
	return field.ErrorList{}
}

func (ipStrategy) WarningsOnUpdate(ctx context.Context, obj, old runtime.Object) []string {
	return nil
}

type ipStatusStrategy struct {
	ipStrategy
}

var StatusStrategy = ipStatusStrategy{Strategy}

func (ipStatusStrategy) GetResetFields() map[fieldpath.APIVersion]*fieldpath.Set {
	return map[fieldpath.APIVersion]*fieldpath.Set{
		"ipam.api.onmetal.de/v1alpha1": fieldpath.NewSet(
			fieldpath.MakePathOrDie("spec"),
		),
	}
}

func (ipStatusStrategy) PrepareForUpdate(ctx context.Context, obj, old runtime.Object) {
	newIP := obj.(*ipam.IP)
	oldIP := old.(*ipam.IP)
	newIP.Spec = oldIP.Spec
}

func (ipStatusStrategy) ValidateUpdate(ctx context.Context, obj, old runtime.Object) field.ErrorList {
	return nil
}

func (ipStatusStrategy) WarningsOnUpdate(cxt context.Context, obj, old runtime.Object) []string {
	return nil
}
