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

// SecurityGroupSpec defines the desired state of SecurityGroup
type SecurityGroupSpec struct {
	// Ingress is a list of inbound rules
	Ingress []IngressSecurityGroupRule `json:"ingress,omitempty"`
	// Egress is a list of outbound rules
	Egress []EgressSecurityGroupRule `json:"egress,omitempty"`
}

// IngressSecurityGroupRule is an ingress rule of a security group
type IngressSecurityGroupRule struct {
	SecurityGroupRule `json:",inline"`
	// Source is either the cird or a reference to another security group
	Source IPSetSpec `json:"source,omitempty"`
}

// EgressSecurityGroupRule is an egress rule of a security group
type EgressSecurityGroupRule struct {
	SecurityGroupRule `json:",inline"`
	// Destination is either the cird or a reference to another security group
	Destination IPSetSpec `json:"destination,omitempty"`
}

// SecurityGroupRule is a single access rule
type SecurityGroupRule struct {
	// Name is the name of the SecurityGroupRule
	Name string `json:"name"`
	// SecurityGroupRef is a reference to an existing SecurityGroup
	SecurityGroupRef corev1.LocalObjectReference `json:"securityGroupRef,omitempty"`
	// Action defines the action type of a SecurityGroupRule
	Action SecurityGroupAction `json:"action,omitempty"`
	// Protocol defines the protocol of a SecurityGroupRule
	Protocol string `json:"protocol,omitempty"`
	// PortRange is the port range of the SecurityGroupRule
	PortRange *PortRange `json:"portRange,omitempty"`
}

// IPSetSpec defines either a cidr or a security group reference
type IPSetSpec struct {
	// CIDR block for source/destination
	CIDR common.IPPrefix `json:"cidr,omitempty"`
	// SecurityGroupRef references a security group
	SecurityGroupRef corev1.LocalObjectReference `json:"securityGroupref,omitempty"`
}

// PortRange defines the start and end of a port range
type PortRange struct {
	// StartPort is the start port of the port range
	StartPort int `json:"startPort"`
	// EndPort is the end port of the port range
	EndPort int `json:"endPort,omitempty"`
}

// SecurityGroupStatus defines the observed state of SecurityGroup
type SecurityGroupStatus struct {
	State      SecurityGroupState       `json:"state,omitempty"`
	Conditions []SecurityGroupCondition `json:"conditions,omitempty"`
}

// SecurityGroupConditionType is a type a SecurityGroupCondition can have.
type SecurityGroupConditionType string

// SecurityGroupCondition is one of the conditions of a volume.
type SecurityGroupCondition struct {
	// Type is the type of the condition.
	Type SecurityGroupConditionType `json:"type"`
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

// SecurityGroupAction describes the action of a SecurityGroupRule.
type SecurityGroupAction string

const (
	SecurityGroupActionTypeAllow SecurityGroupAction = "Allow"
	SecurityGroupActionTypeDeny  SecurityGroupAction = "Deny"
)

// SecurityGroupState is the state of a SecurityGroup.
type SecurityGroupState string

const (
	SecurityGroupStateUsed    SecurityGroupState = "Used"
	SecurityGroupStateUnused  SecurityGroupState = "Unused"
	SecurityGroupStateInvalid SecurityGroupState = "Invalid"
)

//+kubebuilder:object:root=true
//+kubebuilder:subresource:status
//+kubebuilder:resource:shortName=sg
//+kubebuilder:printcolumn:name="State",type=string,JSONPath=`.status.state`
//+kubebuilder:printcolumn:name="Age",type=date,JSONPath=`.metadata.creationTimestamp`

// SecurityGroup is the Schema for the securitygroups API
type SecurityGroup struct {
	metav1.TypeMeta   `json:",inline"`
	metav1.ObjectMeta `json:"metadata,omitempty"`

	Spec   SecurityGroupSpec   `json:"spec,omitempty"`
	Status SecurityGroupStatus `json:"status,omitempty"`
}

//+kubebuilder:object:root=true

// SecurityGroupList contains a list of SecurityGroup
type SecurityGroupList struct {
	metav1.TypeMeta `json:",inline"`
	metav1.ListMeta `json:"metadata,omitempty"`
	Items           []SecurityGroup `json:"items"`
}

func init() {
	SchemeBuilder.Register(&SecurityGroup{}, &SecurityGroupList{})
}
