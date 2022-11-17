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

package loadbalancer

import (
	"context"
	"fmt"

	"github.com/onmetal/onmetal-api/onmetal-apiserver/internal/api"
	"github.com/onmetal/onmetal-api/onmetal-apiserver/internal/apis/networking"
	"github.com/onmetal/onmetal-api/onmetal-apiserver/internal/apis/networking/validation"
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
	loadBalancer, ok := obj.(*networking.LoadBalancer)
	if !ok {
		return nil, nil, fmt.Errorf("given object is not a LoadBalancer")
	}
	return loadBalancer.Labels, SelectableFields(loadBalancer), nil
}

func MatchLoadBalancer(label labels.Selector, field fields.Selector) apisrvstorage.SelectionPredicate {
	return apisrvstorage.SelectionPredicate{
		Label:    label,
		Field:    field,
		GetAttrs: GetAttrs,
	}
}

func SelectableFields(loadBalancer *networking.LoadBalancer) fields.Set {
	return generic.ObjectMetaFieldsSet(&loadBalancer.ObjectMeta, true)
}

type loadBalancerStrategy struct {
	runtime.ObjectTyper
	names.NameGenerator
}

var Strategy = loadBalancerStrategy{api.Scheme, names.SimpleNameGenerator}

func (loadBalancerStrategy) NamespaceScoped() bool {
	return true
}

func (loadBalancerStrategy) PrepareForCreate(ctx context.Context, obj runtime.Object) {
}

func (loadBalancerStrategy) PrepareForUpdate(ctx context.Context, obj, old runtime.Object) {
}

func (loadBalancerStrategy) Validate(ctx context.Context, obj runtime.Object) field.ErrorList {
	loadBalancer := obj.(*networking.LoadBalancer)
	return validation.ValidateLoadBalancer(loadBalancer)
}

func (loadBalancerStrategy) WarningsOnCreate(ctx context.Context, obj runtime.Object) []string {
	return nil
}

func (loadBalancerStrategy) AllowCreateOnUpdate() bool {
	return false
}

func (loadBalancerStrategy) AllowUnconditionalUpdate() bool {
	return false
}

func (loadBalancerStrategy) Canonicalize(obj runtime.Object) {
}

func (loadBalancerStrategy) ValidateUpdate(ctx context.Context, obj, old runtime.Object) field.ErrorList {
	newLoadBalancer := obj.(*networking.LoadBalancer)
	oldLoadBalancer := old.(*networking.LoadBalancer)
	return validation.ValidateLoadBalancerUpdate(newLoadBalancer, oldLoadBalancer)
}

func (loadBalancerStrategy) WarningsOnUpdate(ctx context.Context, obj, old runtime.Object) []string {
	return nil
}

type loadBalancerStatusStrategy struct {
	loadBalancerStrategy
}

var StatusStrategy = loadBalancerStatusStrategy{Strategy}

func (loadBalancerStatusStrategy) GetResetFields() map[fieldpath.APIVersion]*fieldpath.Set {
	return map[fieldpath.APIVersion]*fieldpath.Set{
		"networking.api.onmetal.de/v1alpha1": fieldpath.NewSet(
			fieldpath.MakePathOrDie("spec"),
		),
	}
}

func (loadBalancerStatusStrategy) PrepareForUpdate(ctx context.Context, obj, old runtime.Object) {
}

func (loadBalancerStatusStrategy) ValidateUpdate(ctx context.Context, obj, old runtime.Object) field.ErrorList {
	newLoadBalancer := obj.(*networking.LoadBalancer)
	oldLoadBalancer := old.(*networking.LoadBalancer)
	return validation.ValidateLoadBalancerUpdate(newLoadBalancer, oldLoadBalancer)
}

func (loadBalancerStatusStrategy) WarningsOnUpdate(cxt context.Context, obj, old runtime.Object) []string {
	return nil
}
