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

package storage

import (
	"context"

	"github.com/onmetal/onmetal-api/internal/apis/storage"
	"github.com/onmetal/onmetal-api/internal/rbac"
	"github.com/onmetal/onmetal-api/internal/registry/storage/volumepool"
	metainternalversion "k8s.io/apimachinery/pkg/apis/meta/internalversion"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apiserver/pkg/registry/generic"
	genericregistry "k8s.io/apiserver/pkg/registry/generic/registry"
	"k8s.io/apiserver/pkg/registry/rest"
	"sigs.k8s.io/structured-merge-diff/v4/fieldpath"
)

type REST struct {
	*genericregistry.Store
	groupsToShowPoolResources []string
}

func (e *REST) Get(ctx context.Context, name string, options *metav1.GetOptions) (runtime.Object, error) {
	obj, err := e.Store.Get(ctx, name, options)
	if err != nil {
		return nil, err
	}

	if rbac.UserIsMemberOf(ctx, e.groupsToShowPoolResources) {
		return obj, err
	}

	pool, ok := obj.(*storage.VolumePool)
	if !ok {
		return nil, err
	}
	pool.Status.Allocatable = nil
	pool.Status.Capacity = nil

	return pool, nil
}

func (e *REST) List(ctx context.Context, options *metainternalversion.ListOptions) (runtime.Object, error) {
	objList, err := e.Store.List(ctx, options)

	if rbac.UserIsMemberOf(ctx, e.groupsToShowPoolResources) {
		return objList, err
	}

	pools, ok := objList.(*storage.VolumePoolList)
	if !ok {
		return nil, err
	}

	for index := range pools.Items {
		pools.Items[index].Status.Allocatable = nil
		pools.Items[index].Status.Capacity = nil
	}

	return pools, nil
}

type VolumePoolStorage struct {
	VolumePool *REST
	Status     *StatusREST
}

func NewStorage(optsGetter generic.RESTOptionsGetter, groupsToShowPoolResources []string) (VolumePoolStorage, error) {
	store := &genericregistry.Store{
		NewFunc: func() runtime.Object {
			return &storage.VolumePool{}
		},
		NewListFunc: func() runtime.Object {
			return &storage.VolumePoolList{}
		},
		PredicateFunc:             volumepool.MatchVolumePool,
		DefaultQualifiedResource:  storage.Resource("volumepools"),
		SingularQualifiedResource: storage.Resource("volumepool"),

		CreateStrategy: volumepool.Strategy,
		UpdateStrategy: volumepool.Strategy,
		DeleteStrategy: volumepool.Strategy,

		TableConvertor: newTableConvertor(),
	}

	options := &generic.StoreOptions{RESTOptions: optsGetter, AttrFunc: volumepool.GetAttrs}
	if err := store.CompleteWithOptions(options); err != nil {
		return VolumePoolStorage{}, err
	}

	statusStore := *store
	statusStore.UpdateStrategy = volumepool.StatusStrategy
	statusStore.ResetFieldsStrategy = volumepool.StatusStrategy

	return VolumePoolStorage{
		VolumePool: &REST{
			Store:                     store,
			groupsToShowPoolResources: groupsToShowPoolResources,
		},
		Status: &StatusREST{&statusStore},
	}, nil
}

type StatusREST struct {
	store *genericregistry.Store
}

func (r *StatusREST) New() runtime.Object {
	return &storage.VolumePool{}
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
