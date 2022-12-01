// Copyright 2022 OnMetal authors
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
//      http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.

package endpoints

import (
	"sort"

	computev1alpha1 "github.com/onmetal/onmetal-api/api/compute/v1alpha1"
)

// Notifier is an interface that allows any Listener to be notified upon changes.
type Notifier interface {
	AddListener(listener Listener)
}

// Listener is a listener to a Notifier.
//
// Upon changes, Enqueue is called.
type Listener interface {
	Enqueue()
}

// ListenerFunc is a function that implements Listener.
type ListenerFunc func()

// Enqueue implements Listener.
func (f ListenerFunc) Enqueue() {
	f()
}

// Endpoints allows getting endpoints of a computev1alpha1.MachinePool.
type Endpoints interface {
	// Notifier allows listeners to be notified from endpoint changes.
	Notifier
	// GetEndpoints gets the currently available machine pool endpoints.
	//
	// The method intentionally returns no error, as any error should be reported and handled during the
	// initialization of an Endpoints instance.
	GetEndpoints() (addresses []computev1alpha1.MachinePoolAddress, port int32)
}

// NoEndpoints is a no-op implementation of Endpoints.
type NoEndpoints struct{}

// AddListener implements Notifier.
func (NoEndpoints) AddListener(Listener) {}

// GetEndpoints implements Endpoints.
func (NoEndpoints) GetEndpoints() (addresses []computev1alpha1.MachinePoolAddress, port int32) {
	return nil, 0
}

// MachinePoolAddressSet is a set of computev1alpha1.MachinePoolAddress.
type MachinePoolAddressSet map[computev1alpha1.MachinePoolAddress]struct{}

// Insert inserts the computev1alpha1.MachinePoolAddress into the set.
func (s MachinePoolAddressSet) Insert(items ...computev1alpha1.MachinePoolAddress) {
	for _, item := range items {
		s[item] = struct{}{}
	}
}

// List returns a sorted list of the items.
func (s MachinePoolAddressSet) List() []computev1alpha1.MachinePoolAddress {
	res := make([]computev1alpha1.MachinePoolAddress, 0, len(s))
	for item := range s {
		res = append(res, item)
	}
	sort.Slice(res, func(i, j int) bool { return res[i].Type < res[j].Type || res[i].Address < res[j].Address })
	return res
}

// UnsortedList returns an unsorted list of the items.
func (s MachinePoolAddressSet) UnsortedList() []computev1alpha1.MachinePoolAddress {
	res := make([]computev1alpha1.MachinePoolAddress, 0, len(s))
	for item := range s {
		res = append(res, item)
	}
	return res
}

// NewMachinePoolAddressSet initializes a new MachinePoolAddressSet with the given items.
func NewMachinePoolAddressSet(items ...computev1alpha1.MachinePoolAddress) MachinePoolAddressSet {
	s := make(MachinePoolAddressSet)
	s.Insert(items...)
	return s
}
