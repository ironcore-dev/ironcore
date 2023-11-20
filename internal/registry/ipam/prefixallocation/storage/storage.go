// SPDX-FileCopyrightText: 2023 SAP SE or an SAP affiliate company and IronCore contributors
// SPDX-License-Identifier: Apache-2.0

package storage

import (
	"context"

	"github.com/ironcore-dev/ironcore/internal/apis/ipam"
	"github.com/ironcore-dev/ironcore/internal/registry/ipam/prefixallocation"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apiserver/pkg/registry/generic"
	genericregistry "k8s.io/apiserver/pkg/registry/generic/registry"
	"k8s.io/apiserver/pkg/registry/rest"
	"sigs.k8s.io/structured-merge-diff/v4/fieldpath"
)

type PrefixAllocationStorage struct {
	PrefixAllocation *REST
	Status           *StatusREST
}

type REST struct {
	*genericregistry.Store
}

func NewStorage(optsGetter generic.RESTOptionsGetter) (PrefixAllocationStorage, error) {
	store := &genericregistry.Store{
		NewFunc: func() runtime.Object {
			return &ipam.PrefixAllocation{}
		},
		NewListFunc: func() runtime.Object {
			return &ipam.PrefixAllocationList{}
		},
		PredicateFunc:             prefixallocation.MatchPrefixAllocation,
		DefaultQualifiedResource:  ipam.Resource("prefixallocations"),
		SingularQualifiedResource: ipam.Resource("prefixallocation"),

		CreateStrategy: prefixallocation.Strategy,
		UpdateStrategy: prefixallocation.Strategy,
		DeleteStrategy: prefixallocation.Strategy,

		TableConvertor: newTableConvertor(),
	}

	options := &generic.StoreOptions{RESTOptions: optsGetter, AttrFunc: prefixallocation.GetAttrs}
	if err := store.CompleteWithOptions(options); err != nil {
		return PrefixAllocationStorage{}, err
	}

	statusStore := *store
	statusStore.UpdateStrategy = prefixallocation.StatusStrategy
	statusStore.ResetFieldsStrategy = prefixallocation.StatusStrategy

	return PrefixAllocationStorage{
		PrefixAllocation: &REST{store},
		Status:           &StatusREST{&statusStore},
	}, nil
}

type StatusREST struct {
	store *genericregistry.Store
}

func (r *StatusREST) New() runtime.Object {
	return &ipam.PrefixAllocation{}
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
