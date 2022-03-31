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
	"github.com/onmetal/onmetal-api/apis/storage"
	registrystorageclass "github.com/onmetal/onmetal-api/registry/storage/storageclass"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apiserver/pkg/registry/generic"
	genericregistry "k8s.io/apiserver/pkg/registry/generic/registry"
)

type StorageClassStorage struct {
	StorageClass *REST
}

type REST struct {
	*genericregistry.Store
}

func NewStorage(optsGetter generic.RESTOptionsGetter) (StorageClassStorage, error) {
	store := &genericregistry.Store{
		NewFunc: func() runtime.Object {
			return &storage.StorageClass{}
		},
		NewListFunc: func() runtime.Object {
			return &storage.StorageClassList{}
		},
		PredicateFunc:            registrystorageclass.MatchStorageClass,
		DefaultQualifiedResource: storage.Resource("storageclasses"),

		CreateStrategy: registrystorageclass.Strategy,
		UpdateStrategy: registrystorageclass.Strategy,
		DeleteStrategy: registrystorageclass.Strategy,

		TableConvertor: newTableConvertor(),
	}

	options := &generic.StoreOptions{RESTOptions: optsGetter, AttrFunc: registrystorageclass.GetAttrs}
	if err := store.CompleteWithOptions(options); err != nil {
		return StorageClassStorage{}, err
	}
	return StorageClassStorage{
		StorageClass: &REST{store},
	}, nil
}
