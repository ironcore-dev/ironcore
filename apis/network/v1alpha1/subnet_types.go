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

// SubnetSpec defines the desired state of Subnet
type SubnetSpec struct {
	// Locations defines in which regions and availability zone this subnet should be available
	Locations []common.RegionAvailability `json:"locations,omitempty"`
	// RouteTable is the reference to the routing table this SubNet should be associated with
	RouteTable common.ScopeReference `json:"routeTable"`
	// Ranges defines the size of the subnet
	// +kubebuilder:validation:MinItems:=1
	Ranges []RangeType `json:"ranges"`
}

// RangeType defines the range/size of a subnet
type RangeType struct {
	// IPAM is a reference to the an range block of a subnet
	IPAM common.ScopeReference `json:"ipam,omitempty"`
	// Size defines the size of a subnet e.g. "/12"
	Size string `json:"size,omitempty"`
	// CIDR is the CIDR block
	CIDR common.Cidr `json:"cidr,omitempty"`
	// BlockedRanges specifies which part of the subnet should be used for static IP assignment
	// e.g. 0/14 means the first /14 subnet is blocked in the allocated /12 subnet
	BlockedRanges []string `json:"blockedRanges,omitempty"`
}

// SubnetStatus defines the observed state of Subnet
type SubnetStatus struct {
	common.StateFields `json:",inline"`
	// CIDRs is a list of CIDR status
	CIDRs []CIDRStatus `json:"cidrs,omitempty"`
}

// CIDRStatus is the status of a CIDR
type CIDRStatus struct {
	// CIDR defines the cidr
	CIDR common.Cidr `json:"cidr"`
	// BlockedRanges is a list of blocked cidr ranges
	BlockedRanges []common.Cidr `json:"blockedRanges"`
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
//+kubebuilder:printcolumn:name="Age",type=date,JSONPath=`.metadata.creationTimestamp`

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
