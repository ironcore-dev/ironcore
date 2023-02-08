/*
 * Copyright (c) 2022 by the OnMetal authors.
 *
 * Licensed under the Apache License, Version 2.0 (the "License");
 * you may not use this file except in compliance with the License.
 * You may obtain a copy of the License at
 *
 *     http://www.apache.org/licenses/LICENSE-2.0
 *
 * Unless required by applicable law or agreed to in writing, software
 * distributed under the License is distributed on an "AS IS" BASIS,
 * WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
 * See the License for the specific language governing permissions and
 * limitations under the License.
 */

package storage

import (
	"github.com/onmetal/onmetal-api/internal/apis/networking"
	"github.com/onmetal/onmetal-api/internal/registry/networking/loadbalancerrouting"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apiserver/pkg/registry/generic"
	genericregistry "k8s.io/apiserver/pkg/registry/generic/registry"
)

type LoadBalancerRoutingStorage struct {
	LoadBalancerRouting *REST
}

type REST struct {
	*genericregistry.Store
}

func (REST) ShortNames() []string {
	return []string{"lbr", "lbrt"}
}

func NewStorage(optsGetter generic.RESTOptionsGetter) (LoadBalancerRoutingStorage, error) {
	store := &genericregistry.Store{
		NewFunc: func() runtime.Object {
			return &networking.LoadBalancerRouting{}
		},
		NewListFunc: func() runtime.Object {
			return &networking.LoadBalancerRoutingList{}
		},
		PredicateFunc:            loadbalancerrouting.MatchLoadBalancerRouting,
		DefaultQualifiedResource: networking.Resource("loadbalancerroutings"),

		CreateStrategy: loadbalancerrouting.Strategy,
		UpdateStrategy: loadbalancerrouting.Strategy,
		DeleteStrategy: loadbalancerrouting.Strategy,

		TableConvertor: newTableConvertor(),
	}

	options := &generic.StoreOptions{RESTOptions: optsGetter, AttrFunc: loadbalancerrouting.GetAttrs}
	if err := store.CompleteWithOptions(options); err != nil {
		return LoadBalancerRoutingStorage{}, err
	}

	return LoadBalancerRoutingStorage{
		LoadBalancerRouting: &REST{store},
	}, nil
}
