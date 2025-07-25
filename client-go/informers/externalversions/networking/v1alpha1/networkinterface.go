// SPDX-FileCopyrightText: 2025 SAP SE or an SAP affiliate company and IronCore contributors
// SPDX-License-Identifier: Apache-2.0

// Code generated by informer-gen. DO NOT EDIT.

package v1alpha1

import (
	context "context"
	time "time"

	apinetworkingv1alpha1 "github.com/ironcore-dev/ironcore/api/networking/v1alpha1"
	internalinterfaces "github.com/ironcore-dev/ironcore/client-go/informers/externalversions/internalinterfaces"
	versioned "github.com/ironcore-dev/ironcore/client-go/ironcore/versioned"
	networkingv1alpha1 "github.com/ironcore-dev/ironcore/client-go/listers/networking/v1alpha1"
	v1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	runtime "k8s.io/apimachinery/pkg/runtime"
	watch "k8s.io/apimachinery/pkg/watch"
	cache "k8s.io/client-go/tools/cache"
)

// NetworkInterfaceInformer provides access to a shared informer and lister for
// NetworkInterfaces.
type NetworkInterfaceInformer interface {
	Informer() cache.SharedIndexInformer
	Lister() networkingv1alpha1.NetworkInterfaceLister
}

type networkInterfaceInformer struct {
	factory          internalinterfaces.SharedInformerFactory
	tweakListOptions internalinterfaces.TweakListOptionsFunc
	namespace        string
}

// NewNetworkInterfaceInformer constructs a new informer for NetworkInterface type.
// Always prefer using an informer factory to get a shared informer instead of getting an independent
// one. This reduces memory footprint and number of connections to the server.
func NewNetworkInterfaceInformer(client versioned.Interface, namespace string, resyncPeriod time.Duration, indexers cache.Indexers) cache.SharedIndexInformer {
	return NewFilteredNetworkInterfaceInformer(client, namespace, resyncPeriod, indexers, nil)
}

// NewFilteredNetworkInterfaceInformer constructs a new informer for NetworkInterface type.
// Always prefer using an informer factory to get a shared informer instead of getting an independent
// one. This reduces memory footprint and number of connections to the server.
func NewFilteredNetworkInterfaceInformer(client versioned.Interface, namespace string, resyncPeriod time.Duration, indexers cache.Indexers, tweakListOptions internalinterfaces.TweakListOptionsFunc) cache.SharedIndexInformer {
	return cache.NewSharedIndexInformer(
		&cache.ListWatch{
			ListFunc: func(options v1.ListOptions) (runtime.Object, error) {
				if tweakListOptions != nil {
					tweakListOptions(&options)
				}
				return client.NetworkingV1alpha1().NetworkInterfaces(namespace).List(context.Background(), options)
			},
			WatchFunc: func(options v1.ListOptions) (watch.Interface, error) {
				if tweakListOptions != nil {
					tweakListOptions(&options)
				}
				return client.NetworkingV1alpha1().NetworkInterfaces(namespace).Watch(context.Background(), options)
			},
			ListWithContextFunc: func(ctx context.Context, options v1.ListOptions) (runtime.Object, error) {
				if tweakListOptions != nil {
					tweakListOptions(&options)
				}
				return client.NetworkingV1alpha1().NetworkInterfaces(namespace).List(ctx, options)
			},
			WatchFuncWithContext: func(ctx context.Context, options v1.ListOptions) (watch.Interface, error) {
				if tweakListOptions != nil {
					tweakListOptions(&options)
				}
				return client.NetworkingV1alpha1().NetworkInterfaces(namespace).Watch(ctx, options)
			},
		},
		&apinetworkingv1alpha1.NetworkInterface{},
		resyncPeriod,
		indexers,
	)
}

func (f *networkInterfaceInformer) defaultInformer(client versioned.Interface, resyncPeriod time.Duration) cache.SharedIndexInformer {
	return NewFilteredNetworkInterfaceInformer(client, f.namespace, resyncPeriod, cache.Indexers{cache.NamespaceIndex: cache.MetaNamespaceIndexFunc}, f.tweakListOptions)
}

func (f *networkInterfaceInformer) Informer() cache.SharedIndexInformer {
	return f.factory.InformerFor(&apinetworkingv1alpha1.NetworkInterface{}, f.defaultInformer)
}

func (f *networkInterfaceInformer) Lister() networkingv1alpha1.NetworkInterfaceLister {
	return networkingv1alpha1.NewNetworkInterfaceLister(f.Informer().GetIndexer())
}
