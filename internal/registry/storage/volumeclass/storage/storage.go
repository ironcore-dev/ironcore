// SPDX-FileCopyrightText: 2023 SAP SE or an SAP affiliate company and IronCore contributors
// SPDX-License-Identifier: Apache-2.0

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
