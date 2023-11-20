// SPDX-FileCopyrightText: 2023 SAP SE or an SAP affiliate company and IronCore contributors
// SPDX-License-Identifier: Apache-2.0

package storage

import (
	"context"

	"github.com/ironcore-dev/ironcore/internal/apis/networking"
	"github.com/ironcore-dev/ironcore/internal/registry/networking/loadbalancer"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apiserver/pkg/registry/generic"
	genericregistry "k8s.io/apiserver/pkg/registry/generic/registry"
	"k8s.io/apiserver/pkg/registry/rest"
	"sigs.k8s.io/structured-merge-diff/v4/fieldpath"
)

type LoadBalancerStorage struct {
	LoadBalancer *REST
	Status       *StatusREST
}

type REST struct {
	*genericregistry.Store
}

func (REST) ShortNames() []string {
	return []string{"lb"}
}

func NewStorage(optsGetter generic.RESTOptionsGetter) (LoadBalancerStorage, error) {
	store := &genericregistry.Store{
		NewFunc: func() runtime.Object {
			return &networking.LoadBalancer{}
		},
		NewListFunc: func() runtime.Object {
			return &networking.LoadBalancerList{}
		},
		PredicateFunc:             loadbalancer.MatchLoadBalancer,
		DefaultQualifiedResource:  networking.Resource("loadbalancers"),
		SingularQualifiedResource: networking.Resource("loadbalancer"),

		CreateStrategy: loadbalancer.Strategy,
		UpdateStrategy: loadbalancer.Strategy,
		DeleteStrategy: loadbalancer.Strategy,

		TableConvertor: newTableConvertor(),
	}

	options := &generic.StoreOptions{RESTOptions: optsGetter, AttrFunc: loadbalancer.GetAttrs}
	if err := store.CompleteWithOptions(options); err != nil {
		return LoadBalancerStorage{}, err
	}

	statusStore := *store
	statusStore.UpdateStrategy = loadbalancer.StatusStrategy
	statusStore.ResetFieldsStrategy = loadbalancer.StatusStrategy

	return LoadBalancerStorage{
		LoadBalancer: &REST{store},
		Status:       &StatusREST{&statusStore},
	}, nil
}

type StatusREST struct {
	store *genericregistry.Store
}

func (r *StatusREST) New() runtime.Object {
	return &networking.LoadBalancer{}
}

func (r *StatusREST) Get(ctx context.Context, name string, options *metav1.GetOptions) (runtime.Object, error) {
	return r.store.Get(ctx, name, options)
}

func (r *StatusREST) Update(ctx context.Context, name string, objInfo rest.UpdatedObjectInfo, createValidation rest.ValidateObjectFunc, updateValidation rest.ValidateObjectUpdateFunc, forceAllowCreate bool, options *metav1.UpdateOptions) (runtime.Object, bool, error) {
	return r.store.Update(ctx, name, objInfo, createValidation, updateValidation, false, options)
}

func (r *StatusREST) GetResetFields() map[fieldpath.APIVersion]*fieldpath.Set {
	return r.store.GetResetFields()
}

func (r *StatusREST) Destroy() {}
