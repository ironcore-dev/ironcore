// SPDX-FileCopyrightText: 2023 SAP SE or an SAP affiliate company and IronCore contributors
// SPDX-License-Identifier: Apache-2.0

package loadbalancerrouting

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
