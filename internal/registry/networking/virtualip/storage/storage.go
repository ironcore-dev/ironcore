// SPDX-FileCopyrightText: 2023 SAP SE or an SAP affiliate company and IronCore contributors
// SPDX-License-Identifier: Apache-2.0

package storage

import (
	"context"

	"github.com/ironcore-dev/ironcore/internal/apis/networking"
	"github.com/ironcore-dev/ironcore/internal/registry/networking/virtualip"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apiserver/pkg/registry/generic"
	genericregistry "k8s.io/apiserver/pkg/registry/generic/registry"
	"k8s.io/apiserver/pkg/registry/rest"
	"sigs.k8s.io/structured-merge-diff/v4/fieldpath"
)

type VirtualIPStorage struct {
	VirtualIP *REST
	Status    *StatusREST
}

type REST struct {
	*genericregistry.Store
}

func (REST) ShortNames() []string {
	return []string{"vip"}
}

func NewStorage(optsGetter generic.RESTOptionsGetter) (VirtualIPStorage, error) {
	store := &genericregistry.Store{
		NewFunc: func() runtime.Object {
			return &networking.VirtualIP{}
		},
		NewListFunc: func() runtime.Object {
			return &networking.VirtualIPList{}
		},
		PredicateFunc:             virtualip.MatchVirtualIP,
		DefaultQualifiedResource:  networking.Resource("virtualips"),
		SingularQualifiedResource: networking.Resource("virtualip"),

		CreateStrategy: virtualip.Strategy,
		UpdateStrategy: virtualip.Strategy,
		DeleteStrategy: virtualip.Strategy,

		TableConvertor: newTableConvertor(),
	}

	options := &generic.StoreOptions{RESTOptions: optsGetter, AttrFunc: virtualip.GetAttrs}
	if err := store.CompleteWithOptions(options); err != nil {
		return VirtualIPStorage{}, err
	}

	statusStore := *store
	statusStore.UpdateStrategy = virtualip.StatusStrategy
	statusStore.ResetFieldsStrategy = virtualip.StatusStrategy

	return VirtualIPStorage{
		VirtualIP: &REST{store},
		Status:    &StatusREST{&statusStore},
	}, nil
}

type StatusREST struct {
	store *genericregistry.Store
}

func (r *StatusREST) New() runtime.Object {
	return &networking.VirtualIP{}
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
