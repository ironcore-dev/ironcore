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

package loadbalancerrouting

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
)

func GetAttrs(obj runtime.Object) (labels.Set, fields.Set, error) {
	loadBalancerRouting, ok := obj.(*networking.LoadBalancerRouting)
	if !ok {
		return nil, nil, fmt.Errorf("given object is not a LoadBalancerRouting")
	}
	return loadBalancerRouting.Labels, SelectableFields(loadBalancerRouting), nil
}

func MatchLoadBalancerRouting(label labels.Selector, field fields.Selector) apisrvstorage.SelectionPredicate {
	return apisrvstorage.SelectionPredicate{
		Label:    label,
		Field:    field,
		GetAttrs: GetAttrs,
	}
}

func SelectableFields(loadBalancerRouting *networking.LoadBalancerRouting) fields.Set {
	return generic.ObjectMetaFieldsSet(&loadBalancerRouting.ObjectMeta, true)
}

type loadBalancerRoutingStrategy struct {
	runtime.ObjectTyper
	names.NameGenerator
}

var Strategy = loadBalancerRoutingStrategy{api.Scheme, names.SimpleNameGenerator}

func (loadBalancerRoutingStrategy) NamespaceScoped() bool {
	return true
}

func (loadBalancerRoutingStrategy) PrepareForCreate(ctx context.Context, obj runtime.Object) {
}

func (loadBalancerRoutingStrategy) PrepareForUpdate(ctx context.Context, obj, old runtime.Object) {
}

func (loadBalancerRoutingStrategy) Validate(ctx context.Context, obj runtime.Object) field.ErrorList {
	loadBalancerRouting := obj.(*networking.LoadBalancerRouting)
	return validation.ValidateLoadBalancerRouting(loadBalancerRouting)
}

func (loadBalancerRoutingStrategy) WarningsOnCreate(ctx context.Context, obj runtime.Object) []string {
	return nil
}

func (loadBalancerRoutingStrategy) AllowCreateOnUpdate() bool {
	return false
}

func (loadBalancerRoutingStrategy) AllowUnconditionalUpdate() bool {
	return false
}

func (loadBalancerRoutingStrategy) Canonicalize(obj runtime.Object) {
}

func (loadBalancerRoutingStrategy) ValidateUpdate(ctx context.Context, obj, old runtime.Object) field.ErrorList {
	newLoadBalancerRouting := obj.(*networking.LoadBalancerRouting)
	oldLoadBalancerRouting := old.(*networking.LoadBalancerRouting)
	return validation.ValidateLoadBalancerRoutingUpdate(newLoadBalancerRouting, oldLoadBalancerRouting)
}

func (loadBalancerRoutingStrategy) WarningsOnUpdate(ctx context.Context, obj, old runtime.Object) []string {
	return nil
}
