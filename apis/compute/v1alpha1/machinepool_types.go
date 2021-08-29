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

package v1alpha1

import (
	common "github.com/onmetal/onmetal-api/apis/common/v1alpha1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

// MachinePoolSpec defines the desired state of MachinePool
type MachinePoolSpec struct {
	// Region defines the region where this machine pool is available
	Region string `json:"region,omitempty"`
	// Privacy indicates the privacy scope of the machine pool
	Privacy string `json:"privacy"`
	// Capacity defines the quantity of this machine pool per availability zone
	Capacity []AvailabilityZoneQuantity `json:"capacity"`
}

const (
	PrivacyShared     = "shared"
	PrivacyHypervisor = "hypervisor"
	PrivacyCluster    = "cluster"
)

// AvailabilityZoneQuantity defines the quantity of available MachineClasses in a given AZ
type AvailabilityZoneQuantity struct {
	// AvailabilityZone is the name of the availability zone
	AvailabilityZone string `json:"availabilityZone"`
	// Classes defines a list of machine classes and their corresponding quantities
	Classes []MachineClassQuantity `json:"classes"`
}

// MachineClassQuantity defines the quantity of a given MachineClass
type MachineClassQuantity struct {
	// Name is the name of the machine class quantity
	Name string `json:"name"`
	// Quantity is an absolut number of the available machine class
	// +kubebuilder:validation:Minimum:=0
	Quantity int `json:"quantity"`
}

// MachinePoolStatus defines the observed state of MachinePool
type MachinePoolStatus struct {
	common.StateFields `json:",inline"`
	Used               AvailabilityZoneQuantity `json:"used"`
}

const (
	MachinePoolStateReady   = "Ready"
	MachinePoolStatePending = "Pending"
	MachinePoolStateError   = "Error"
	MachinePoolStateOffline = "Offline"
)

//+kubebuilder:object:root=true
//+kubebuilder:subresource:status
//+kubebuilder:printcolumn:name="State",type=string,JSONPath=`.status.state`
//+kubebuilder:printcolumn:name="Age",type=date,JSONPath=`.metadata.creationTimestamp`

// MachinePool is the Schema for the machinepools API
type MachinePool struct {
	metav1.TypeMeta   `json:",inline"`
	metav1.ObjectMeta `json:"metadata,omitempty"`

	Spec   MachinePoolSpec   `json:"spec,omitempty"`
	Status MachinePoolStatus `json:"status,omitempty"`
}

//+kubebuilder:object:root=true

// MachinePoolList contains a list of MachinePool
type MachinePoolList struct {
	metav1.TypeMeta `json:",inline"`
	metav1.ListMeta `json:"metadata,omitempty"`
	Items           []MachinePool `json:"items"`
}

func init() {
	SchemeBuilder.Register(&MachinePool{}, &MachinePoolList{})
}
