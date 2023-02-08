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
	"github.com/onmetal/onmetal-api/internal/apis/compute"
	"github.com/onmetal/onmetal-api/internal/registry/compute/machineclass"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apiserver/pkg/registry/generic"
	genericregistry "k8s.io/apiserver/pkg/registry/generic/registry"
)

type MachineClassStorage struct {
	MachineClass *REST
}

type REST struct {
	*genericregistry.Store
}

func NewStorage(optsGetter generic.RESTOptionsGetter) (MachineClassStorage, error) {
	store := &genericregistry.Store{
		NewFunc: func() runtime.Object {
			return &compute.MachineClass{}
		},
		NewListFunc: func() runtime.Object {
			return &compute.MachineClassList{}
		},
		PredicateFunc:            machineclass.MatchMachineClass,
		DefaultQualifiedResource: compute.Resource("machineclasses"),

		CreateStrategy: machineclass.Strategy,
		UpdateStrategy: machineclass.Strategy,
		DeleteStrategy: machineclass.Strategy,

		TableConvertor: newTableConvertor(),
	}

	options := &generic.StoreOptions{RESTOptions: optsGetter, AttrFunc: machineclass.GetAttrs}
	if err := store.CompleteWithOptions(options); err != nil {
		return MachineClassStorage{}, err
	}

	return MachineClassStorage{
		MachineClass: &REST{store},
	}, nil
}
