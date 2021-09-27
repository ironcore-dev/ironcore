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
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

// RoutingDomainSpec defines the desired state of RoutingDomain
// Subnets associated with a RoutingDomain are routed implicitly and don't
// need explicit routing instructions.
type RoutingDomainSpec struct {
	// Routes is a list of routing instructions
	Routes []Route `json:"routes"`
}

// Route describes a single route definition
type Route struct {
	// SubnetRef is a reference to Subnet
	SubnetRef corev1.LocalObjectReference `json:"subnetRef,omitempty"`
	// CIDR is the matching CIDR of a Route
	CIDR string `json:"cidr,omitempty"`
	// Target is the target object of a Route
	Target Target `json:"target,omitempty"`
}

type MachineRouteTarget struct {
	corev1.LocalObjectReference `json:",inline"`
	Interface                   string `json:"interface"`
}

// RoutingDomainStatus defines the observed state of RoutingDomain
type RoutingDomainStatus struct {
	State      RoutingDomainState       `json:"state,omitempty"`
	Conditions []RoutingDomainCondition `json:"conditions,omitempty"`
}

type RoutingDomainState string

// RoutingDomainConditionType is a type a RoutingDomainCondition can have.
type RoutingDomainConditionType string

// RoutingDomainCondition is one of the conditions of a volume.
type RoutingDomainCondition struct {
	// Type is the type of the condition.
	Type RoutingDomainConditionType `json:"type"`
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
//+kubebuilder:resource:shortName=rd
//+kubebuilder:printcolumn:name="Routes",type=string,JSONPath=`.spec.routes`,priority=100
//+kubebuilder:printcolumn:name="State",type=string,JSONPath=`.status.state`
//+kubebuilder:printcolumn:name="Age",type=date,JSONPath=`.metadata.creationTimestamp`

// RoutingDomain is the Schema for the RoutingDomain API
type RoutingDomain struct {
	metav1.TypeMeta   `json:",inline"`
	metav1.ObjectMeta `json:"metadata,omitempty"`

	Spec   RoutingDomainSpec   `json:"spec,omitempty"`
	Status RoutingDomainStatus `json:"status,omitempty"`
}

//+kubebuilder:object:root=true

// RoutingDomainList contains a list of RoutingDomain
type RoutingDomainList struct {
	metav1.TypeMeta `json:",inline"`
	metav1.ListMeta `json:"metadata,omitempty"`
	Items           []RoutingDomain `json:"items"`
}

func init() {
	SchemeBuilder.Register(&RoutingDomain{}, &RoutingDomainList{})
}
