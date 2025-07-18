// SPDX-FileCopyrightText: 2025 SAP SE or an SAP affiliate company and IronCore contributors
// SPDX-License-Identifier: Apache-2.0

// Code generated by informer-gen. DO NOT EDIT.

package externalversions

import (
	fmt "fmt"

	v1alpha1 "github.com/ironcore-dev/ironcore/api/compute/v1alpha1"
	corev1alpha1 "github.com/ironcore-dev/ironcore/api/core/v1alpha1"
	ipamv1alpha1 "github.com/ironcore-dev/ironcore/api/ipam/v1alpha1"
	networkingv1alpha1 "github.com/ironcore-dev/ironcore/api/networking/v1alpha1"
	storagev1alpha1 "github.com/ironcore-dev/ironcore/api/storage/v1alpha1"
	schema "k8s.io/apimachinery/pkg/runtime/schema"
	cache "k8s.io/client-go/tools/cache"
)

// GenericInformer is type of SharedIndexInformer which will locate and delegate to other
// sharedInformers based on type
type GenericInformer interface {
	Informer() cache.SharedIndexInformer
	Lister() cache.GenericLister
}

type genericInformer struct {
	informer cache.SharedIndexInformer
	resource schema.GroupResource
}

// Informer returns the SharedIndexInformer.
func (f *genericInformer) Informer() cache.SharedIndexInformer {
	return f.informer
}

// Lister returns the GenericLister.
func (f *genericInformer) Lister() cache.GenericLister {
	return cache.NewGenericLister(f.Informer().GetIndexer(), f.resource)
}

// ForResource gives generic access to a shared informer of the matching type
// TODO extend this to unknown resources with a client pool
func (f *sharedInformerFactory) ForResource(resource schema.GroupVersionResource) (GenericInformer, error) {
	switch resource {
	// Group=compute.ironcore.dev, Version=v1alpha1
	case v1alpha1.SchemeGroupVersion.WithResource("machines"):
		return &genericInformer{resource: resource.GroupResource(), informer: f.Compute().V1alpha1().Machines().Informer()}, nil
	case v1alpha1.SchemeGroupVersion.WithResource("machineclasses"):
		return &genericInformer{resource: resource.GroupResource(), informer: f.Compute().V1alpha1().MachineClasses().Informer()}, nil
	case v1alpha1.SchemeGroupVersion.WithResource("machinepools"):
		return &genericInformer{resource: resource.GroupResource(), informer: f.Compute().V1alpha1().MachinePools().Informer()}, nil

		// Group=core.ironcore.dev, Version=v1alpha1
	case corev1alpha1.SchemeGroupVersion.WithResource("resourcequotas"):
		return &genericInformer{resource: resource.GroupResource(), informer: f.Core().V1alpha1().ResourceQuotas().Informer()}, nil

		// Group=ipam.ironcore.dev, Version=v1alpha1
	case ipamv1alpha1.SchemeGroupVersion.WithResource("prefixes"):
		return &genericInformer{resource: resource.GroupResource(), informer: f.Ipam().V1alpha1().Prefixes().Informer()}, nil
	case ipamv1alpha1.SchemeGroupVersion.WithResource("prefixallocations"):
		return &genericInformer{resource: resource.GroupResource(), informer: f.Ipam().V1alpha1().PrefixAllocations().Informer()}, nil

		// Group=networking.ironcore.dev, Version=v1alpha1
	case networkingv1alpha1.SchemeGroupVersion.WithResource("loadbalancers"):
		return &genericInformer{resource: resource.GroupResource(), informer: f.Networking().V1alpha1().LoadBalancers().Informer()}, nil
	case networkingv1alpha1.SchemeGroupVersion.WithResource("loadbalancerroutings"):
		return &genericInformer{resource: resource.GroupResource(), informer: f.Networking().V1alpha1().LoadBalancerRoutings().Informer()}, nil
	case networkingv1alpha1.SchemeGroupVersion.WithResource("natgateways"):
		return &genericInformer{resource: resource.GroupResource(), informer: f.Networking().V1alpha1().NATGateways().Informer()}, nil
	case networkingv1alpha1.SchemeGroupVersion.WithResource("networks"):
		return &genericInformer{resource: resource.GroupResource(), informer: f.Networking().V1alpha1().Networks().Informer()}, nil
	case networkingv1alpha1.SchemeGroupVersion.WithResource("networkinterfaces"):
		return &genericInformer{resource: resource.GroupResource(), informer: f.Networking().V1alpha1().NetworkInterfaces().Informer()}, nil
	case networkingv1alpha1.SchemeGroupVersion.WithResource("networkpolicies"):
		return &genericInformer{resource: resource.GroupResource(), informer: f.Networking().V1alpha1().NetworkPolicies().Informer()}, nil
	case networkingv1alpha1.SchemeGroupVersion.WithResource("virtualips"):
		return &genericInformer{resource: resource.GroupResource(), informer: f.Networking().V1alpha1().VirtualIPs().Informer()}, nil

		// Group=storage.ironcore.dev, Version=v1alpha1
	case storagev1alpha1.SchemeGroupVersion.WithResource("buckets"):
		return &genericInformer{resource: resource.GroupResource(), informer: f.Storage().V1alpha1().Buckets().Informer()}, nil
	case storagev1alpha1.SchemeGroupVersion.WithResource("bucketclasses"):
		return &genericInformer{resource: resource.GroupResource(), informer: f.Storage().V1alpha1().BucketClasses().Informer()}, nil
	case storagev1alpha1.SchemeGroupVersion.WithResource("bucketpools"):
		return &genericInformer{resource: resource.GroupResource(), informer: f.Storage().V1alpha1().BucketPools().Informer()}, nil
	case storagev1alpha1.SchemeGroupVersion.WithResource("volumes"):
		return &genericInformer{resource: resource.GroupResource(), informer: f.Storage().V1alpha1().Volumes().Informer()}, nil
	case storagev1alpha1.SchemeGroupVersion.WithResource("volumeclasses"):
		return &genericInformer{resource: resource.GroupResource(), informer: f.Storage().V1alpha1().VolumeClasses().Informer()}, nil
	case storagev1alpha1.SchemeGroupVersion.WithResource("volumepools"):
		return &genericInformer{resource: resource.GroupResource(), informer: f.Storage().V1alpha1().VolumePools().Informer()}, nil

	}

	return nil, fmt.Errorf("no informer found for %v", resource)
}
