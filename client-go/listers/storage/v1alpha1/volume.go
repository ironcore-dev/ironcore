// SPDX-FileCopyrightText: 2023 SAP SE or an SAP affiliate company and IronCore contributors
// SPDX-License-Identifier: Apache-2.0

// Code generated by lister-gen. DO NOT EDIT.

package v1alpha1

import (
	storagev1alpha1 "github.com/ironcore-dev/ironcore/api/storage/v1alpha1"
	labels "k8s.io/apimachinery/pkg/labels"
	listers "k8s.io/client-go/listers"
	cache "k8s.io/client-go/tools/cache"
)

// VolumeLister helps list Volumes.
// All objects returned here must be treated as read-only.
type VolumeLister interface {
	// List lists all Volumes in the indexer.
	// Objects returned here must be treated as read-only.
	List(selector labels.Selector) (ret []*storagev1alpha1.Volume, err error)
	// Volumes returns an object that can list and get Volumes.
	Volumes(namespace string) VolumeNamespaceLister
	VolumeListerExpansion
}

// volumeLister implements the VolumeLister interface.
type volumeLister struct {
	listers.ResourceIndexer[*storagev1alpha1.Volume]
}

// NewVolumeLister returns a new VolumeLister.
func NewVolumeLister(indexer cache.Indexer) VolumeLister {
	return &volumeLister{listers.New[*storagev1alpha1.Volume](indexer, storagev1alpha1.Resource("volume"))}
}

// Volumes returns an object that can list and get Volumes.
func (s *volumeLister) Volumes(namespace string) VolumeNamespaceLister {
	return volumeNamespaceLister{listers.NewNamespaced[*storagev1alpha1.Volume](s.ResourceIndexer, namespace)}
}

// VolumeNamespaceLister helps list and get Volumes.
// All objects returned here must be treated as read-only.
type VolumeNamespaceLister interface {
	// List lists all Volumes in the indexer for a given namespace.
	// Objects returned here must be treated as read-only.
	List(selector labels.Selector) (ret []*storagev1alpha1.Volume, err error)
	// Get retrieves the Volume from the indexer for a given namespace and name.
	// Objects returned here must be treated as read-only.
	Get(name string) (*storagev1alpha1.Volume, error)
	VolumeNamespaceListerExpansion
}

// volumeNamespaceLister implements the VolumeNamespaceLister
// interface.
type volumeNamespaceLister struct {
	listers.ResourceIndexer[*storagev1alpha1.Volume]
}
