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

// ReservedIPSpec defines the desired state of ReservedIP
type ReservedIPSpec struct {
	// Subnet references the subnet where an IP address should be reserved
	Subnet common.ScopeReference `json:"subnet"`
	// IP specifies an IP address which should be reserved. Must be in the CIDR of the
	// associated Subnet
	IP common.IPAddr `json:"ip,omitempty"`
	// Assignment indicates to which resource this IP address should be assigned
	Assignment Assignment `json:"assignment,omitempty"`
}

// Assignment defines the Assignment type to which an IP address can be bind
type Assignment struct {
	// Kind is the object kind of the assigned resource
	Kind string `json:"kind"`
	// APIGroup is the API group of the assigned resource
	APIGroup              string `json:"apigroup"`
	common.ScopeReference `json:",inline"`
}

// ReservedIPStatus defines the observed state of ReservedIP
type ReservedIPStatus struct {
	// IP indicates the effective reserved IP address
	IP                 common.IPAddr `json:"ip,omitempty"`
	common.StateFields `json:",inline"`
}

//+kubebuilder:object:root=true
//+kubebuilder:subresource:status
//+kubebuilder:printcolumn:name="IP",type=string,JSONPath=`.status.ip`
//+kubebuilder:printcolumn:name="StateFields",type=string,JSONPath=`.status.state`
//+kubebuilder:printcolumn:name="Age",type=string,JSONPath=`.metadata.CreationTimestamp`

// ReservedIP is the Schema for the reservedips API
type ReservedIP struct {
	metav1.TypeMeta   `json:",inline"`
	metav1.ObjectMeta `json:"metadata,omitempty"`

	Spec   ReservedIPSpec   `json:"spec,omitempty"`
	Status ReservedIPStatus `json:"status,omitempty"`
}

//+kubebuilder:object:root=true

// ReservedIPList contains a list of ReservedIP
type ReservedIPList struct {
	metav1.TypeMeta `json:",inline"`
	metav1.ListMeta `json:"metadata,omitempty"`
	Items           []ReservedIP `json:"items"`
}

func init() {
	SchemeBuilder.Register(&ReservedIP{}, &ReservedIPList{})
}
