// SPDX-FileCopyrightText: 2023 SAP SE or an SAP affiliate company and IronCore contributors
// SPDX-License-Identifier: Apache-2.0

package storage

import (
	"context"

	"github.com/ironcore-dev/ironcore/internal/apis/networking"
	"github.com/ironcore-dev/ironcore/internal/registry/networking/natgateway"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apiserver/pkg/registry/generic"
	genericregistry "k8s.io/apiserver/pkg/registry/generic/registry"
	"k8s.io/apiserver/pkg/registry/rest"
	"sigs.k8s.io/structured-merge-diff/v6/fieldpath"
)

type NATGatewayStorage struct {
	NATGateway *REST
	Status     *StatusREST
}

type REST struct {
	*genericregistry.Store
}

func (REST) ShortNames() []string {
	return []string{"ngw"}
}

func NewStorage(optsGetter generic.RESTOptionsGetter) (NATGatewayStorage, error) {
	store := &genericregistry.Store{
		NewFunc: func() runtime.Object {
			return &networking.NATGateway{}
		},
		NewListFunc: func() runtime.Object {
			return &networking.NATGatewayList{}
		},
		PredicateFunc:             natgateway.MatchNATGateway,
		DefaultQualifiedResource:  networking.Resource("natgateways"),
		SingularQualifiedResource: networking.Resource("natgateway"),

		CreateStrategy: natgateway.Strategy,
		UpdateStrategy: natgateway.Strategy,
		DeleteStrategy: natgateway.Strategy,

		TableConvertor: newTableConvertor(),
	}

	options := &generic.StoreOptions{RESTOptions: optsGetter, AttrFunc: natgateway.GetAttrs}
	if err := store.CompleteWithOptions(options); err != nil {
		return NATGatewayStorage{}, err
	}

	statusStore := *store
	statusStore.UpdateStrategy = natgateway.StatusStrategy
	statusStore.ResetFieldsStrategy = natgateway.StatusStrategy

	return NATGatewayStorage{
		NATGateway: &REST{store},
		Status:     &StatusREST{&statusStore},
	}, nil
}

type StatusREST struct {
	store *genericregistry.Store
}

func (r *StatusREST) New() runtime.Object {
	return &networking.NATGateway{}
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
