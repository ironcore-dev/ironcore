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

// SubnetSpec defines the desired state of Subnet
type SubnetSpec struct {
	// ParentSubnet is a reference to the parent of a subnet
	ParentSubnet common.ScopeReference `json:"parentsubnet,omitempty"`
	// Locations defines in which regions and availability zone this subnet should be available
	Locations []common.RegionAvailability `json:"locations,omitempty"`
	// Ranges defines the size of the subnet
	// +kubebuilder:validation:MinItems:=1
	Ranges []RangeType `json:"ranges"`
}

// RangeType defines the range/size of a subnet
type RangeType struct {
	// CIDR is the CIDR block
	CIDR common.Cidr `json:"cidr,omitempty"`
	// BlockedRanges specifies which part of the subnet should be used for static IP assignment
	BlockedRanges []common.Cidr `json:"blockedranges,omitempty"`
	// Size defines the size of a subnet e.g. "/12"
	Size string `json:"size,omitempty"`
	// Capacity defines the absolute number of IPs in a given subnet
	// +kubebuilder:validation:Minimum:=0
	Capacity int `json:"capacity,omitempty"`
}

// SubnetStatus defines the observed state of Subnet
type SubnetStatus struct {
	common.StateFields `json:",inline"`
}

const (
	SubnetStateInitial = "Initial"
	SubnetStateUp      = "Up"
	SubnetStateDown    = "Down"
	SubnetStateInvalid = "Invalid"
)

//+kubebuilder:object:root=true
//+kubebuilder:subresource:status
//+kubebuilder:printcolumn:name="StateFields",type=string,JSONPath=`.status.state`
//+kubebuilder:printcolumn:name="Age",type=string,JSONPath=`.metadata.CreationTimestamp`

// Subnet is the Schema for the subnets API
type Subnet struct {
	metav1.TypeMeta   `json:",inline"`
	metav1.ObjectMeta `json:"metadata,omitempty"`

	Spec   SubnetSpec   `json:"spec,omitempty"`
	Status SubnetStatus `json:"status,omitempty"`
}

//+kubebuilder:object:root=true

// SubnetList contains a list of Subnet
type SubnetList struct {
	metav1.TypeMeta `json:",inline"`
	metav1.ListMeta `json:"metadata,omitempty"`
	Items           []Subnet `json:"items"`
}

func init() {
	SchemeBuilder.Register(&Subnet{}, &SubnetList{})
}
