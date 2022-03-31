/*
 * Copyright (c) 2021 by the OnMetal authors.
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
// Code generated by lister-gen. DO NOT EDIT.

package v1alpha1

import (
	v1alpha1 "github.com/onmetal/onmetal-api/apis/compute/v1alpha1"
	"k8s.io/apimachinery/pkg/api/errors"
	"k8s.io/apimachinery/pkg/labels"
	"k8s.io/client-go/tools/cache"
)

// MachinePoolLister helps list MachinePools.
// All objects returned here must be treated as read-only.
type MachinePoolLister interface {
	// List lists all MachinePools in the indexer.
	// Objects returned here must be treated as read-only.
	List(selector labels.Selector) (ret []*v1alpha1.MachinePool, err error)
	// MachinePools returns an object that can list and get MachinePools.
	MachinePools(namespace string) MachinePoolNamespaceLister
	MachinePoolListerExpansion
}

// machinePoolLister implements the MachinePoolLister interface.
type machinePoolLister struct {
	indexer cache.Indexer
}

// NewMachinePoolLister returns a new MachinePoolLister.
func NewMachinePoolLister(indexer cache.Indexer) MachinePoolLister {
	return &machinePoolLister{indexer: indexer}
}

// List lists all MachinePools in the indexer.
func (s *machinePoolLister) List(selector labels.Selector) (ret []*v1alpha1.MachinePool, err error) {
	err = cache.ListAll(s.indexer, selector, func(m interface{}) {
		ret = append(ret, m.(*v1alpha1.MachinePool))
	})
	return ret, err
}

// MachinePools returns an object that can list and get MachinePools.
func (s *machinePoolLister) MachinePools(namespace string) MachinePoolNamespaceLister {
	return machinePoolNamespaceLister{indexer: s.indexer, namespace: namespace}
}

// MachinePoolNamespaceLister helps list and get MachinePools.
// All objects returned here must be treated as read-only.
type MachinePoolNamespaceLister interface {
	// List lists all MachinePools in the indexer for a given namespace.
	// Objects returned here must be treated as read-only.
	List(selector labels.Selector) (ret []*v1alpha1.MachinePool, err error)
	// Get retrieves the MachinePool from the indexer for a given namespace and name.
	// Objects returned here must be treated as read-only.
	Get(name string) (*v1alpha1.MachinePool, error)
	MachinePoolNamespaceListerExpansion
}

// machinePoolNamespaceLister implements the MachinePoolNamespaceLister
// interface.
type machinePoolNamespaceLister struct {
	indexer   cache.Indexer
	namespace string
}

// List lists all MachinePools in the indexer for a given namespace.
func (s machinePoolNamespaceLister) List(selector labels.Selector) (ret []*v1alpha1.MachinePool, err error) {
	err = cache.ListAllByNamespace(s.indexer, s.namespace, selector, func(m interface{}) {
		ret = append(ret, m.(*v1alpha1.MachinePool))
	})
	return ret, err
}

// Get retrieves the MachinePool from the indexer for a given namespace and name.
func (s machinePoolNamespaceLister) Get(name string) (*v1alpha1.MachinePool, error) {
	obj, exists, err := s.indexer.GetByKey(s.namespace + "/" + name)
	if err != nil {
		return nil, err
	}
	if !exists {
		return nil, errors.NewNotFound(v1alpha1.Resource("machinepool"), name)
	}
	return obj.(*v1alpha1.MachinePool), nil
}
