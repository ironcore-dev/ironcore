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
	commonv1alpha1 "github.com/onmetal/onmetal-api/apis/common/v1alpha1"
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime/schema"
)

var SubnetGK = schema.GroupKind{
	Group: GroupVersion.Group,
	Kind:  "Subnet",
}

// SubnetSpec defines the desired state of Subnet
type SubnetSpec struct {
	// Parent is a reference to a public parent Subnet.
	Parent corev1.LocalObjectReference `json:"parent,omitempty"`
	// MachinePools defines in which pools this subnet should be available
	MachinePools []corev1.LocalObjectReference `json:"machinePools,omitempty"`
	// RoutingDomain is the reference to the routing domain this SubNet should be associated with
	RoutingDomain corev1.LocalObjectReference `json:"routingDomain"`
	// Ranges defines the size of the subnet
	// +kubebuilder:validation:MinItems:=1
	Ranges []RangeType `json:"ranges,omitempty"`
}

// RangeType defines the range/size of a subnet
type RangeType struct {
	// Size defines the size of a subnet e.g. "/12"
	Size string `json:"size,omitempty"`
	// CIDR is the CIDR block
	CIDR commonv1alpha1.CIDR `json:"cidr,omitempty"`
	// BlockedRanges specifies which part of the subnet should be used for static IP assignment
	// e.g. 0/14 means the first /14 subnet is blocked in the allocated /12 subnet
	BlockedRanges []string `json:"blockedRanges,omitempty"`
}

// CIDRStatus is the status of a CIDR
type CIDRStatus struct {
	// CIDR defines the cidr
	CIDR    commonv1alpha1.CIDR          `json:"cidr,omitempty"`
	State   CIDRState                    `json:"state"`
	Request *Request                     `json:"request,omitempty"`
	User    *corev1.LocalObjectReference `json:"user,omitempty"`
}

type CIDRState string

const (
	CIDRFree    CIDRState = "Free"
	CIDRUsed    CIDRState = "Used"
	CIDRFailed  CIDRState = "Failed"
	CIDRPending CIDRState = "Pending"
)

// SubnetStatus defines the observed state of Subnet
type SubnetStatus struct {
	State      SubnetState       `json:"state,omitempty"`
	Conditions []SubnetCondition `json:"conditions,omitempty"`
	// CIDRs is a list of CIDR status
	CIDRs []CIDRStatus `json:"cidrs,omitempty"`
}

type SubnetState string

const (
	SubnetStateInitial SubnetState = "Initial"
	SubnetStateUp      SubnetState = "Up"
	SubnetStateDown    SubnetState = "Down"
	SubnetStateInvalid SubnetState = "Invalid"
)

// SubnetConditionType is a type a SubnetCondition can have.
type SubnetConditionType string

// SubnetCondition is one of the conditions of a volume.
type SubnetCondition struct {
	// Type is the type of the condition.
	Type SubnetConditionType `json:"type"`
	// Status is the status of the condition.
	Status corev1.ConditionStatus `json:"status"`
	// Reason is a machine-readable indication of why the condition is in a certain state.
	Reason string `json:"reason"`
	// Message is a human-readable explanation of why the condition has a certain reason / state.
	Message string `json:"message"`
	// ObservedGeneration represents the .metadata.generation that the condition was set based upon.
	ObservedGeneration int64 `json:"observedGeneration,omitempty"`
	// LastUpdateTime is the last time a condition has been updated.
	LastUpdateTime metav1.Time `json:"lastUpdateTime,omitempty"`
	// LastTransitionTime is the last time the status of a condition has transitioned from one state to another.
	LastTransitionTime metav1.Time `json:"lastTransitionTime,omitempty"`
}

//+kubebuilder:object:root=true
//+kubebuilder:subresource:status
//+kubebuilder:resource:shortName=sn
//+kubebuilder:printcolumn:name="RoutingDomain",type=string,JSONPath=`.spec.routingDomain.name`
//+kubebuilder:printcolumn:name="Ranges",type=string,JSONPath=`.spec.ranges`,priority=100
//+kubebuilder:printcolumn:name="State",type=string,JSONPath=`.status.state`
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
