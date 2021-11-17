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
	"fmt"

	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"

	common "github.com/onmetal/onmetal-api/apis/common/v1alpha1"
)

type GatewayMode string

const (
	// GatewayFinalizer is the finalizer used by the Gateway controller to
	// cleanup the resources owned by the Gateway when a Gateway is being deleted.
	GatewayFinalizer = "gateway.network.onmetal.de"
	// SNATMode is stateless NAT / 1-1 NAT (network address translation).
	SNATMode GatewayMode = "SNAT"
	// NATMode is regular NAT (network address translation).
	NATMode GatewayMode = "NAT"
	// TransparentMode makes the gateway behave transparently.
	TransparentMode GatewayMode = "Transparent"
)

// GatewaySpec defines the desired state of Gateway
type GatewaySpec struct {
	Mode        GatewayMode  `json:"mode"`
	FilterRules []FilterRule `json:"filterRules,omitempty"`
	// Uplink is a Target to route traffic to.
	Uplink Target `json:"uplink"`

	// Subnet is the reference to the Subnet where the Gateway resides
	Subnet corev1.LocalObjectReference `json:"subnet"`
}

type FilterRule struct {
	SecurityGroup corev1.LocalObjectReference `json:"securityGroup,omitempty"`
}

// GatewayStatus defines the observed state of Gateway
type GatewayStatus struct {
	State      GatewayState       `json:"state,omitempty"`
	Conditions []GatewayCondition `json:"conditions,omitempty"`
	IPs        []common.IPAddr    `json:"ips,omitempty"`
}

type GatewayState string

// GatewayConditionType is a type a GatewayCondition can have.
type GatewayConditionType string

// GatewayCondition is one of the conditions of a volume.
type GatewayCondition struct {
	// Type is the type of the condition.
	Type GatewayConditionType `json:"type"`
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
//+kubebuilder:resource:shortName=gw
//+kubebuilder:printcolumn:name="Mode",type=string,JSONPath=`.spec.mode`
//+kubebuilder:printcolumn:name="Uplink",type=string,JSONPath=`.spec.uplink`,priority=100
//+kubebuilder:printcolumn:name="IPs",type=string,JSONPath=`.status.ips`,priority=100
//+kubebuilder:printcolumn:name="FilterRules",type=string,JSONPath=`.spec.filterRules`,priority=100
//+kubebuilder:printcolumn:name="State",type=string,JSONPath=`.status.state`
//+kubebuilder:printcolumn:name="Age",type=date,JSONPath=`.metadata.creationTimestamp`

// Gateway is the Schema for the gateways API
type Gateway struct {
	metav1.TypeMeta   `json:",inline"`
	metav1.ObjectMeta `json:"metadata,omitempty"`

	Spec   GatewaySpec   `json:"spec,omitempty"`
	Status GatewayStatus `json:"status,omitempty"`
}

// IsBeingDeleted returns if the instance is being deleted
func (gw *Gateway) IsBeingDeleted() bool {
	return !gw.DeletionTimestamp.IsZero()
}

// IPAMRangeName returns the name of the corresponding IPAMRange
func (gw *Gateway) IPAMRangeName() string {
	return fmt.Sprintf("gateway-%s", gw.Name)
}

//+kubebuilder:object:root=true

// GatewayList contains a list of Gateway
type GatewayList struct {
	metav1.TypeMeta `json:",inline"`
	metav1.ListMeta `json:"metadata,omitempty"`
	Items           []Gateway `json:"items"`
}

func init() {
	SchemeBuilder.Register(&Gateway{}, &GatewayList{})
}
