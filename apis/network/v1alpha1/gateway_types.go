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

const (
	SNATMode = "SNAT"
	// Stateless NAT/1-1 NAT
	NATMode         = "NAT"
	TransparentMode = "Transparent"
)

// GatewaySpec defines the desired state of Gateway
type GatewaySpec struct {
	Mode string `json:"mode"`
	// Regions is a list of regions where this Gateway should be available
	Regions     []string     `json:"regions,omitempty"`
	FilterRules []FilterRule `json:"filterRules,omitempty"`
	// Uplink is either a ReservedIP or a Subnet
	Uplink common.ScopedKindReference `json:"uplink"`
}

type FilterRule struct {
	SecurityGroup common.ScopedReference `json:"securityGroup,omitempty"`
}

// GatewayStatus defines the observed state of Gateway
type GatewayStatus struct {
	common.StateFields `json:",inline"`
	IPs                []common.IPAddr `json:"ips,omitempty"`
}

//+kubebuilder:object:root=true
//+kubebuilder:subresource:status
//+kubebuilder:resource:shortName=gw
//+kubebuilder:printcolumn:name="Mode",type=string,JSONPath=`.spec.mode`
//+kubebuilder:printcolumn:name="Regions",type=string,JSONPath=`.spec.regions`
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
