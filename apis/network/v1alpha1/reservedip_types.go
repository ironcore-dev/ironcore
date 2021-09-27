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
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

// ReservedIPSpec defines the desired state of ReservedIP
type ReservedIPSpec struct {
	// Subnet references the subnet where an IP address should be reserved
	Subnet corev1.LocalObjectReference `json:"subnet"`
	// IP specifies an IP address which should be reserved. Must be in the CIDR of the
	// associated Subnet
	IP common.IPAddr `json:"ip,omitempty"`
	// Assignment indicates to which resource this IP address should be assigned
	Assignment ReservedIPAssignment `json:"assignment,omitempty"`
}

// ReservedIPAssignment contains information that points to the resource being used.
type ReservedIPAssignment struct {
	// APIGroup is the group for the resource being referenced
	APIGroup string `json:"apiGroup" protobuf:"bytes,1,opt,name=apiGroup"`
	// Kind is the type of resource being referenced
	Kind string `json:"kind" protobuf:"bytes,2,opt,name=kind"`
	// Name is the name of resource being referenced
	Name string `json:"name" protobuf:"bytes,3,opt,name=name"`
}

// ReservedIPStatus defines the observed state of ReservedIP
type ReservedIPStatus struct {
	// IP indicates the effective reserved IP address
	IP                 common.IPAddr `json:"ip,omitempty"`
	common.StateFields `json:",inline"`
	Bound              *ReservedIPBound `json:"bound,omitempty"`
}

// ReservedIPBound describes the binding state of a ReservedIP
type ReservedIPBound struct {
	Mode       string               `json:"mode"`
	Assignment ReservedIPAssignment `json:"assignment,omitempty"`
}

const (
	// BoundModeFloating defines a ReservedIP which is dynamically assigned
	// as additional DNAT-ed IP for the target resource.
	BoundModeFloating = "Floating"
	// BoundModeStatic defines a ReservedIP which is directly assigned to an interface
	// of the target resource. This means the target is directly connected to the Subnet
	// of the reserved IP.
	BoundModeStatic = "Static"
)

//+kubebuilder:object:root=true
//+kubebuilder:subresource:status
//+kubebuilder:resource:shortName=rip
//+kubebuilder:printcolumn:name="IP",type=string,JSONPath=`.status.ip`
//+kubebuilder:printcolumn:name="Mode",type=string,JSONPath=`.status.bound.mode`
//+kubebuilder:printcolumn:name="Assignment",type=string,JSONPath=`.status.bound.assignment`,priority=100
//+kubebuilder:printcolumn:name="State",type=string,JSONPath=`.status.state`
//+kubebuilder:printcolumn:name="Age",type=date,JSONPath=`.metadata.creationTimestamp`

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
