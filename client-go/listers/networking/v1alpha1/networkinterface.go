// SPDX-FileCopyrightText: 2025 SAP SE or an SAP affiliate company and IronCore contributors
// SPDX-License-Identifier: Apache-2.0

// Code generated by lister-gen. DO NOT EDIT.

package v1alpha1

import (
	networkingv1alpha1 "github.com/ironcore-dev/ironcore/api/networking/v1alpha1"
	labels "k8s.io/apimachinery/pkg/labels"
	listers "k8s.io/client-go/listers"
	cache "k8s.io/client-go/tools/cache"
)

// NetworkInterfaceLister helps list NetworkInterfaces.
// All objects returned here must be treated as read-only.
type NetworkInterfaceLister interface {
	// List lists all NetworkInterfaces in the indexer.
	// Objects returned here must be treated as read-only.
	List(selector labels.Selector) (ret []*networkingv1alpha1.NetworkInterface, err error)
	// NetworkInterfaces returns an object that can list and get NetworkInterfaces.
	NetworkInterfaces(namespace string) NetworkInterfaceNamespaceLister
	NetworkInterfaceListerExpansion
}

// networkInterfaceLister implements the NetworkInterfaceLister interface.
type networkInterfaceLister struct {
	listers.ResourceIndexer[*networkingv1alpha1.NetworkInterface]
}

// NewNetworkInterfaceLister returns a new NetworkInterfaceLister.
func NewNetworkInterfaceLister(indexer cache.Indexer) NetworkInterfaceLister {
	return &networkInterfaceLister{listers.New[*networkingv1alpha1.NetworkInterface](indexer, networkingv1alpha1.Resource("networkinterface"))}
}

// NetworkInterfaces returns an object that can list and get NetworkInterfaces.
func (s *networkInterfaceLister) NetworkInterfaces(namespace string) NetworkInterfaceNamespaceLister {
	return networkInterfaceNamespaceLister{listers.NewNamespaced[*networkingv1alpha1.NetworkInterface](s.ResourceIndexer, namespace)}
}

// NetworkInterfaceNamespaceLister helps list and get NetworkInterfaces.
// All objects returned here must be treated as read-only.
type NetworkInterfaceNamespaceLister interface {
	// List lists all NetworkInterfaces in the indexer for a given namespace.
	// Objects returned here must be treated as read-only.
	List(selector labels.Selector) (ret []*networkingv1alpha1.NetworkInterface, err error)
	// Get retrieves the NetworkInterface from the indexer for a given namespace and name.
	// Objects returned here must be treated as read-only.
	Get(name string) (*networkingv1alpha1.NetworkInterface, error)
	NetworkInterfaceNamespaceListerExpansion
}

// networkInterfaceNamespaceLister implements the NetworkInterfaceNamespaceLister
// interface.
type networkInterfaceNamespaceLister struct {
	listers.ResourceIndexer[*networkingv1alpha1.NetworkInterface]
}
