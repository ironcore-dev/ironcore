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
	v1alpha1 "github.com/onmetal/onmetal-api/apis/ipam/v1alpha1"
	"k8s.io/apimachinery/pkg/api/errors"
	"k8s.io/apimachinery/pkg/labels"
	"k8s.io/client-go/tools/cache"
)

// PrefixLister helps list Prefixes.
// All objects returned here must be treated as read-only.
type PrefixLister interface {
	// List lists all Prefixes in the indexer.
	// Objects returned here must be treated as read-only.
	List(selector labels.Selector) (ret []*v1alpha1.Prefix, err error)
	// Prefixes returns an object that can list and get Prefixes.
	Prefixes(namespace string) PrefixNamespaceLister
	PrefixListerExpansion
}

// prefixLister implements the PrefixLister interface.
type prefixLister struct {
	indexer cache.Indexer
}

// NewPrefixLister returns a new PrefixLister.
func NewPrefixLister(indexer cache.Indexer) PrefixLister {
	return &prefixLister{indexer: indexer}
}

// List lists all Prefixes in the indexer.
func (s *prefixLister) List(selector labels.Selector) (ret []*v1alpha1.Prefix, err error) {
	err = cache.ListAll(s.indexer, selector, func(m interface{}) {
		ret = append(ret, m.(*v1alpha1.Prefix))
	})
	return ret, err
}

// Prefixes returns an object that can list and get Prefixes.
func (s *prefixLister) Prefixes(namespace string) PrefixNamespaceLister {
	return prefixNamespaceLister{indexer: s.indexer, namespace: namespace}
}

// PrefixNamespaceLister helps list and get Prefixes.
// All objects returned here must be treated as read-only.
type PrefixNamespaceLister interface {
	// List lists all Prefixes in the indexer for a given namespace.
	// Objects returned here must be treated as read-only.
	List(selector labels.Selector) (ret []*v1alpha1.Prefix, err error)
	// Get retrieves the Prefix from the indexer for a given namespace and name.
	// Objects returned here must be treated as read-only.
	Get(name string) (*v1alpha1.Prefix, error)
	PrefixNamespaceListerExpansion
}

// prefixNamespaceLister implements the PrefixNamespaceLister
// interface.
type prefixNamespaceLister struct {
	indexer   cache.Indexer
	namespace string
}

// List lists all Prefixes in the indexer for a given namespace.
func (s prefixNamespaceLister) List(selector labels.Selector) (ret []*v1alpha1.Prefix, err error) {
	err = cache.ListAllByNamespace(s.indexer, s.namespace, selector, func(m interface{}) {
		ret = append(ret, m.(*v1alpha1.Prefix))
	})
	return ret, err
}

// Get retrieves the Prefix from the indexer for a given namespace and name.
func (s prefixNamespaceLister) Get(name string) (*v1alpha1.Prefix, error) {
	obj, exists, err := s.indexer.GetByKey(s.namespace + "/" + name)
	if err != nil {
		return nil, err
	}
	if !exists {
		return nil, errors.NewNotFound(v1alpha1.Resource("prefix"), name)
	}
	return obj.(*v1alpha1.Prefix), nil
}
