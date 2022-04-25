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
	"github.com/onmetal/onmetal-api/apis/compute"
	"github.com/onmetal/onmetal-api/apis/networking"
	"github.com/onmetal/onmetal-api/registry/networking/virtualiprouting"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apiserver/pkg/registry/generic"
	genericregistry "k8s.io/apiserver/pkg/registry/generic/registry"
)

type VirtualIPRoutingStorage struct {
	VirtualIPRouting *REST
}

type REST struct {
	*genericregistry.Store
}

func (REST) ShortNames() []string {
	return []string{"vipr", "viprouting"}
}

func NewStorage(optsGetter generic.RESTOptionsGetter) (VirtualIPRoutingStorage, error) {
	store := &genericregistry.Store{
		NewFunc: func() runtime.Object {
			return &networking.VirtualIPRouting{}
		},
		NewListFunc: func() runtime.Object {
			return &networking.VirtualIPRoutingList{}
		},
		PredicateFunc:            virtualiprouting.MatchVirtualIPRouting,
		DefaultQualifiedResource: compute.Resource("virtualiproutings"),

		CreateStrategy: virtualiprouting.Strategy,
		UpdateStrategy: virtualiprouting.Strategy,
		DeleteStrategy: virtualiprouting.Strategy,

		TableConvertor: newTableConvertor(),
	}

	options := &generic.StoreOptions{RESTOptions: optsGetter, AttrFunc: virtualiprouting.GetAttrs}
	if err := store.CompleteWithOptions(options); err != nil {
		return VirtualIPRoutingStorage{}, err
	}

	return VirtualIPRoutingStorage{
		VirtualIPRouting: &REST{store},
	}, nil
}
