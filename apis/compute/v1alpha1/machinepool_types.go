/*
Copyright 2021.

Licensed under the Apache License, Version 2.0 (the "License");
you may not use this file except in compliance with the License.
You may obtain a copy of the License at

    http://www.apache.org/licenses/LICENSE-2.0

Unless required by applicable law or agreed to in writing, software
distributed under the License is distributed on an "AS IS" BASIS,
WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
See the License for the specific language governing permissions and
limitations under the License.
*/

package v1alpha1

import (
	common "github.com/onmetal/onmetal-api/apis/common/v1alpha1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

// MachinePoolSpec defines the desired state of MachinePool
type MachinePoolSpec struct {
	Region   string                     `json:"region,omitempty"`
	Privacy  PrivacyType                `json:"privacy"`
	Capacity []AvailabilityZoneQuantity `json:"capacity"`
}

// PrivacyType describes the isolation requirements of a MachinePool
type PrivacyType string

const (
	PrivacyShared     = "shared"
	PrivacyHypervisor = "hypervisor"
	PrivacyCluster    = "cluster"
)

// AvailabilityZoneQuantity defines the quantity of available MachineClasses in a given AZ
type AvailabilityZoneQuantity struct {
	AvailabilityZone string                 `json:"availabilityZone"`
	Classes          []MachineClassQuantity `json:"classes"`
}

// MachineClassQuantity defines the quantity of a given MachineClass
type MachineClassQuantity struct {
	Name     string `json:"name"`
	Quantity uint   `json:"quantity"`
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
//+kubebuilder:printcolumn:name="StateFields",type=string,JSONPath=`.status.state`
//+kubebuilder:printcolumn:name="Age",type=string,JSONPath=`.metadata.CreationTimestamp`

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
