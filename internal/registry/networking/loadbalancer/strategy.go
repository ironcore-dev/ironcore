// SPDX-FileCopyrightText: 2023 SAP SE or an SAP affiliate company and IronCore contributors
// SPDX-License-Identifier: Apache-2.0

package loadbalancer

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
	loadBalancer := obj.(*networking.LoadBalancer)
	loadBalancer.Status = networking.LoadBalancerStatus{}
	dropTypeDependentFields(loadBalancer)
}

func (loadBalancerStrategy) PrepareForUpdate(ctx context.Context, obj, old runtime.Object) {
	newLoadBalancer := obj.(*networking.LoadBalancer)
	oldLoadBalancer := old.(*networking.LoadBalancer)
	newLoadBalancer.Status = oldLoadBalancer.Status
	dropTypeDependentFields(newLoadBalancer)
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
		"networking.ironcore.dev/v1alpha1": fieldpath.NewSet(
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

func needsIPs(loadBalancer *networking.LoadBalancer) bool {
	switch loadBalancer.Spec.Type {
	case networking.LoadBalancerTypeInternal:
		return true
	default:
		return false
	}
}

func dropTypeDependentFields(loadBalancer *networking.LoadBalancer) {
	if !needsIPs(loadBalancer) {
		loadBalancer.Spec.IPs = nil
	}
}
