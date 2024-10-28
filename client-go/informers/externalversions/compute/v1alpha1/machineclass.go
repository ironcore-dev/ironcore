// SPDX-FileCopyrightText: 2023 SAP SE or an SAP affiliate company and IronCore contributors
// SPDX-License-Identifier: Apache-2.0

// Code generated by informer-gen. DO NOT EDIT.

package v1alpha1

import (
	"context"
	time "time"

	computev1alpha1 "github.com/ironcore-dev/ironcore/api/compute/v1alpha1"
	internalinterfaces "github.com/ironcore-dev/ironcore/client-go/informers/externalversions/internalinterfaces"
	versioned "github.com/ironcore-dev/ironcore/client-go/ironcore/versioned"
	v1alpha1 "github.com/ironcore-dev/ironcore/client-go/listers/compute/v1alpha1"
	v1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	runtime "k8s.io/apimachinery/pkg/runtime"
	watch "k8s.io/apimachinery/pkg/watch"
	cache "k8s.io/client-go/tools/cache"
)

// MachineClassInformer provides access to a shared informer and lister for
// MachineClasses.
type MachineClassInformer interface {
	Informer() cache.SharedIndexInformer
	Lister() v1alpha1.MachineClassLister
}

type machineClassInformer struct {
	factory          internalinterfaces.SharedInformerFactory
	tweakListOptions internalinterfaces.TweakListOptionsFunc
}

// NewMachineClassInformer constructs a new informer for MachineClass type.
// Always prefer using an informer factory to get a shared informer instead of getting an independent
// one. This reduces memory footprint and number of connections to the server.
func NewMachineClassInformer(client versioned.Interface, resyncPeriod time.Duration, indexers cache.Indexers) cache.SharedIndexInformer {
	return NewFilteredMachineClassInformer(client, resyncPeriod, indexers, nil)
}

// NewFilteredMachineClassInformer constructs a new informer for MachineClass type.
// Always prefer using an informer factory to get a shared informer instead of getting an independent
// one. This reduces memory footprint and number of connections to the server.
func NewFilteredMachineClassInformer(client versioned.Interface, resyncPeriod time.Duration, indexers cache.Indexers, tweakListOptions internalinterfaces.TweakListOptionsFunc) cache.SharedIndexInformer {
	return cache.NewSharedIndexInformer(
		&cache.ListWatch{
			ListFunc: func(options v1.ListOptions) (runtime.Object, error) {
				if tweakListOptions != nil {
					tweakListOptions(&options)
				}
				return client.ComputeV1alpha1().MachineClasses().List(context.TODO(), options)
			},
			WatchFunc: func(options v1.ListOptions) (watch.Interface, error) {
				if tweakListOptions != nil {
					tweakListOptions(&options)
				}
				return client.ComputeV1alpha1().MachineClasses().Watch(context.TODO(), options)
			},
		},
		&computev1alpha1.MachineClass{},
		resyncPeriod,
		indexers,
	)
}

func (f *machineClassInformer) defaultInformer(client versioned.Interface, resyncPeriod time.Duration) cache.SharedIndexInformer {
	return NewFilteredMachineClassInformer(client, resyncPeriod, cache.Indexers{cache.NamespaceIndex: cache.MetaNamespaceIndexFunc}, f.tweakListOptions)
}

func (f *machineClassInformer) Informer() cache.SharedIndexInformer {
	return f.factory.InformerFor(&computev1alpha1.MachineClass{}, f.defaultInformer)
}

func (f *machineClassInformer) Lister() v1alpha1.MachineClassLister {
	return v1alpha1.NewMachineClassLister(f.Informer().GetIndexer())
}