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

// VirtualIPInformer provides access to a shared informer and lister for
// VirtualIPs.
type VirtualIPInformer interface {
	Informer() cache.SharedIndexInformer
	Lister() networkingv1alpha1.VirtualIPLister
}

type virtualIPInformer struct {
	factory          internalinterfaces.SharedInformerFactory
	tweakListOptions internalinterfaces.TweakListOptionsFunc
	namespace        string
}

// NewVirtualIPInformer constructs a new informer for VirtualIP type.
// Always prefer using an informer factory to get a shared informer instead of getting an independent
// one. This reduces memory footprint and number of connections to the server.
func NewVirtualIPInformer(client versioned.Interface, namespace string, resyncPeriod time.Duration, indexers cache.Indexers) cache.SharedIndexInformer {
	return NewFilteredVirtualIPInformer(client, namespace, resyncPeriod, indexers, nil)
}

// NewFilteredVirtualIPInformer constructs a new informer for VirtualIP type.
// Always prefer using an informer factory to get a shared informer instead of getting an independent
// one. This reduces memory footprint and number of connections to the server.
func NewFilteredVirtualIPInformer(client versioned.Interface, namespace string, resyncPeriod time.Duration, indexers cache.Indexers, tweakListOptions internalinterfaces.TweakListOptionsFunc) cache.SharedIndexInformer {
	return cache.NewSharedIndexInformer(
		&cache.ListWatch{
			ListFunc: func(options v1.ListOptions) (runtime.Object, error) {
				if tweakListOptions != nil {
					tweakListOptions(&options)
				}
				return client.NetworkingV1alpha1().VirtualIPs(namespace).List(context.Background(), options)
			},
			WatchFunc: func(options v1.ListOptions) (watch.Interface, error) {
				if tweakListOptions != nil {
					tweakListOptions(&options)
				}
				return client.NetworkingV1alpha1().VirtualIPs(namespace).Watch(context.Background(), options)
			},
			ListWithContextFunc: func(ctx context.Context, options v1.ListOptions) (runtime.Object, error) {
				if tweakListOptions != nil {
					tweakListOptions(&options)
				}
				return client.NetworkingV1alpha1().VirtualIPs(namespace).List(ctx, options)
			},
			WatchFuncWithContext: func(ctx context.Context, options v1.ListOptions) (watch.Interface, error) {
				if tweakListOptions != nil {
					tweakListOptions(&options)
				}
				return client.NetworkingV1alpha1().VirtualIPs(namespace).Watch(ctx, options)
			},
		},
		&apinetworkingv1alpha1.VirtualIP{},
		resyncPeriod,
		indexers,
	)
}

func (f *virtualIPInformer) defaultInformer(client versioned.Interface, resyncPeriod time.Duration) cache.SharedIndexInformer {
	return NewFilteredVirtualIPInformer(client, f.namespace, resyncPeriod, cache.Indexers{cache.NamespaceIndex: cache.MetaNamespaceIndexFunc}, f.tweakListOptions)
}

func (f *virtualIPInformer) Informer() cache.SharedIndexInformer {
	return f.factory.InformerFor(&apinetworkingv1alpha1.VirtualIP{}, f.defaultInformer)
}

func (f *virtualIPInformer) Lister() networkingv1alpha1.VirtualIPLister {
	return networkingv1alpha1.NewVirtualIPLister(f.Informer().GetIndexer())
}
