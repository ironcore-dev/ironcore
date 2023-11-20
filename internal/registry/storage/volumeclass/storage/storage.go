// Copyright 2022 IronCore authors
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
	"github.com/ironcore-dev/ironcore/internal/apis/storage"
	registryvolumeclass "github.com/ironcore-dev/ironcore/internal/registry/storage/volumeclass"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apiserver/pkg/registry/generic"
	genericregistry "k8s.io/apiserver/pkg/registry/generic/registry"
)

type VolumeClassStorage struct {
	VolumeClass *REST
}

type REST struct {
	*genericregistry.Store
}

func NewStorage(optsGetter generic.RESTOptionsGetter) (VolumeClassStorage, error) {
	store := &genericregistry.Store{
		NewFunc: func() runtime.Object {
			return &storage.VolumeClass{}
		},
		NewListFunc: func() runtime.Object {
			return &storage.VolumeClassList{}
		},
		PredicateFunc:             registryvolumeclass.MatchVolumeClass,
		DefaultQualifiedResource:  storage.Resource("volumeclasses"),
		SingularQualifiedResource: storage.Resource("volumeclass"),

		CreateStrategy: registryvolumeclass.Strategy,
		UpdateStrategy: registryvolumeclass.Strategy,
		DeleteStrategy: registryvolumeclass.Strategy,

		TableConvertor: newTableConvertor(),
	}

	options := &generic.StoreOptions{RESTOptions: optsGetter, AttrFunc: registryvolumeclass.GetAttrs}
	if err := store.CompleteWithOptions(options); err != nil {
		return VolumeClassStorage{}, err
	}
	return VolumeClassStorage{
		VolumeClass: &REST{store},
	}, nil
}
