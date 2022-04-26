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
// Code generated by informer-gen. DO NOT EDIT.

package internalversion

import (
	"context"
	time "time"

	v1alpha1 "github.com/onmetal/onmetal-api/apis/networking/v1alpha1"
	clientsetinternalversion "github.com/onmetal/onmetal-api/generated/clientset/internalversion"
	internalinterfaces "github.com/onmetal/onmetal-api/generated/informers/internalversion/internalinterfaces"
	internalversion "github.com/onmetal/onmetal-api/generated/listers/v1alpha1/internalversion"
	v1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	runtime "k8s.io/apimachinery/pkg/runtime"
	watch "k8s.io/apimachinery/pkg/watch"
	cache "k8s.io/client-go/tools/cache"
)

// VirtualIPRoutingInformer provides access to a shared informer and lister for
// VirtualIPRoutings.
type VirtualIPRoutingInformer interface {
	Informer() cache.SharedIndexInformer
	Lister() internalversion.VirtualIPRoutingLister
}

type virtualIPRoutingInformer struct {
	factory          internalinterfaces.SharedInformerFactory
	tweakListOptions internalinterfaces.TweakListOptionsFunc
	namespace        string
}

// NewVirtualIPRoutingInformer constructs a new informer for VirtualIPRouting type.
// Always prefer using an informer factory to get a shared informer instead of getting an independent
// one. This reduces memory footprint and number of connections to the server.
func NewVirtualIPRoutingInformer(client clientsetinternalversion.Interface, namespace string, resyncPeriod time.Duration, indexers cache.Indexers) cache.SharedIndexInformer {
	return NewFilteredVirtualIPRoutingInformer(client, namespace, resyncPeriod, indexers, nil)
}

// NewFilteredVirtualIPRoutingInformer constructs a new informer for VirtualIPRouting type.
// Always prefer using an informer factory to get a shared informer instead of getting an independent
// one. This reduces memory footprint and number of connections to the server.
func NewFilteredVirtualIPRoutingInformer(client clientsetinternalversion.Interface, namespace string, resyncPeriod time.Duration, indexers cache.Indexers, tweakListOptions internalinterfaces.TweakListOptionsFunc) cache.SharedIndexInformer {
	return cache.NewSharedIndexInformer(
		&cache.ListWatch{
			ListFunc: func(options v1.ListOptions) (runtime.Object, error) {
				if tweakListOptions != nil {
					tweakListOptions(&options)
				}
				return client.Networking().VirtualIPRoutings(namespace).List(context.TODO(), options)
			},
			WatchFunc: func(options v1.ListOptions) (watch.Interface, error) {
				if tweakListOptions != nil {
					tweakListOptions(&options)
				}
				return client.Networking().VirtualIPRoutings(namespace).Watch(context.TODO(), options)
			},
		},
		&v1alpha1.VirtualIPRouting{},
		resyncPeriod,
		indexers,
	)
}

func (f *virtualIPRoutingInformer) defaultInformer(client clientsetinternalversion.Interface, resyncPeriod time.Duration) cache.SharedIndexInformer {
	return NewFilteredVirtualIPRoutingInformer(client, f.namespace, resyncPeriod, cache.Indexers{cache.NamespaceIndex: cache.MetaNamespaceIndexFunc}, f.tweakListOptions)
}

func (f *virtualIPRoutingInformer) Informer() cache.SharedIndexInformer {
	return f.factory.InformerFor(&v1alpha1.VirtualIPRouting{}, f.defaultInformer)
}

func (f *virtualIPRoutingInformer) Lister() internalversion.VirtualIPRoutingLister {
	return internalversion.NewVirtualIPRoutingLister(f.Informer().GetIndexer())
}
