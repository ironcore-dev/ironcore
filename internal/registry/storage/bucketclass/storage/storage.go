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
	"github.com/onmetal/onmetal-api/internal/apis/storage"
	registrybucketclass "github.com/onmetal/onmetal-api/internal/registry/storage/bucketclass"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apiserver/pkg/registry/generic"
	genericregistry "k8s.io/apiserver/pkg/registry/generic/registry"
)

type BucketClassStorage struct {
	BucketClass *REST
}

type REST struct {
	*genericregistry.Store
}

func NewStorage(optsGetter generic.RESTOptionsGetter) (BucketClassStorage, error) {
	store := &genericregistry.Store{
		NewFunc: func() runtime.Object {
			return &storage.BucketClass{}
		},
		NewListFunc: func() runtime.Object {
			return &storage.BucketClassList{}
		},
		PredicateFunc:             registrybucketclass.MatchBucketClass,
		DefaultQualifiedResource:  storage.Resource("bucketclasses"),
		SingularQualifiedResource: storage.Resource("bucketclass"),

		CreateStrategy: registrybucketclass.Strategy,
		UpdateStrategy: registrybucketclass.Strategy,
		DeleteStrategy: registrybucketclass.Strategy,

		TableConvertor: newTableConvertor(),
	}

	options := &generic.StoreOptions{RESTOptions: optsGetter, AttrFunc: registrybucketclass.GetAttrs}
	if err := store.CompleteWithOptions(options); err != nil {
		return BucketClassStorage{}, err
	}
	return BucketClassStorage{
		BucketClass: &REST{store},
	}, nil
}
