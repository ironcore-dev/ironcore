// SPDX-FileCopyrightText: 2023 SAP SE or an SAP affiliate company and IronCore contributors
// SPDX-License-Identifier: Apache-2.0

// Code generated by lister-gen. DO NOT EDIT.

package v1alpha1

import (
	computev1alpha1 "github.com/ironcore-dev/ironcore/api/compute/v1alpha1"
	labels "k8s.io/apimachinery/pkg/labels"
	listers "k8s.io/client-go/listers"
	cache "k8s.io/client-go/tools/cache"
)

// MachineClassLister helps list MachineClasses.
// All objects returned here must be treated as read-only.
type MachineClassLister interface {
	// List lists all MachineClasses in the indexer.
	// Objects returned here must be treated as read-only.
	List(selector labels.Selector) (ret []*computev1alpha1.MachineClass, err error)
	// Get retrieves the MachineClass from the index for a given name.
	// Objects returned here must be treated as read-only.
	Get(name string) (*computev1alpha1.MachineClass, error)
	MachineClassListerExpansion
}

// machineClassLister implements the MachineClassLister interface.
type machineClassLister struct {
	listers.ResourceIndexer[*computev1alpha1.MachineClass]
}

// NewMachineClassLister returns a new MachineClassLister.
func NewMachineClassLister(indexer cache.Indexer) MachineClassLister {
	return &machineClassLister{listers.New[*computev1alpha1.MachineClass](indexer, computev1alpha1.Resource("machineclass"))}
}
