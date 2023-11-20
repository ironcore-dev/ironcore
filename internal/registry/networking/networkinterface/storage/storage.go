// SPDX-FileCopyrightText: 2023 SAP SE or an SAP affiliate company and IronCore contributors
// SPDX-License-Identifier: Apache-2.0

package storage

import (
	"context"

	"github.com/ironcore-dev/ironcore/internal/apis/networking"
	"github.com/ironcore-dev/ironcore/internal/registry/networking/networkinterface"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apiserver/pkg/registry/generic"
	genericregistry "k8s.io/apiserver/pkg/registry/generic/registry"
	"k8s.io/apiserver/pkg/registry/rest"
	"sigs.k8s.io/structured-merge-diff/v4/fieldpath"
)

type NetworkInterfaceStorage struct {
	NetworkInterface *REST
	Status           *StatusREST
}

type REST struct {
	*genericregistry.Store
}

func (REST) ShortNames() []string {
	return []string{"nic"}
}

func NewStorage(optsGetter generic.RESTOptionsGetter) (NetworkInterfaceStorage, error) {
	store := &genericregistry.Store{
		NewFunc: func() runtime.Object {
			return &networking.NetworkInterface{}
		},
		NewListFunc: func() runtime.Object {
			return &networking.NetworkInterfaceList{}
		},
		PredicateFunc:             networkinterface.MatchNetworkInterface,
		DefaultQualifiedResource:  networking.Resource("networkinterfaces"),
		SingularQualifiedResource: networking.Resource("networkinterface"),

		CreateStrategy: networkinterface.Strategy,
		UpdateStrategy: networkinterface.Strategy,
		DeleteStrategy: networkinterface.Strategy,

		TableConvertor: newTableConvertor(),
	}

	options := &generic.StoreOptions{RESTOptions: optsGetter, AttrFunc: networkinterface.GetAttrs}
	if err := store.CompleteWithOptions(options); err != nil {
		return NetworkInterfaceStorage{}, err
	}

	statusStore := *store
	statusStore.UpdateStrategy = networkinterface.StatusStrategy
	statusStore.ResetFieldsStrategy = networkinterface.StatusStrategy

	return NetworkInterfaceStorage{
		NetworkInterface: &REST{store},
		Status:           &StatusREST{&statusStore},
	}, nil
}

type StatusREST struct {
	store *genericregistry.Store
}

func (r *StatusREST) New() runtime.Object {
	return &networking.NetworkInterface{}
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
