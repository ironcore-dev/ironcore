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

	"github.com/onmetal/onmetal-api/apis/compute"
	"github.com/onmetal/onmetal-api/registry/compute/console"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apiserver/pkg/registry/generic"
	genericregistry "k8s.io/apiserver/pkg/registry/generic/registry"
	"k8s.io/apiserver/pkg/registry/rest"
	"sigs.k8s.io/structured-merge-diff/v4/fieldpath"
)

type ConsoleStorage struct {
	Console *REST
	Status  *StatusREST
}

type REST struct {
	*genericregistry.Store
}

func NewStorage(optsGetter generic.RESTOptionsGetter) (ConsoleStorage, error) {
	store := &genericregistry.Store{
		NewFunc: func() runtime.Object {
			return &compute.Console{}
		},
		NewListFunc: func() runtime.Object {
			return &compute.ConsoleList{}
		},
		PredicateFunc:            console.MatchConsole,
		DefaultQualifiedResource: compute.Resource("consoles"),

		CreateStrategy: console.Strategy,
		UpdateStrategy: console.Strategy,
		DeleteStrategy: console.Strategy,

		TableConvertor: newTableConvertor(),
	}

	options := &generic.StoreOptions{RESTOptions: optsGetter, AttrFunc: console.GetAttrs}
	if err := store.CompleteWithOptions(options); err != nil {
		return ConsoleStorage{}, err
	}

	statusStore := *store
	statusStore.UpdateStrategy = console.StatusStrategy
	statusStore.ResetFieldsStrategy = console.StatusStrategy

	return ConsoleStorage{
		Console: &REST{store},
		Status:  &StatusREST{&statusStore},
	}, nil
}

type StatusREST struct {
	store *genericregistry.Store
}

func (r *StatusREST) New() runtime.Object {
	return &compute.Console{}
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