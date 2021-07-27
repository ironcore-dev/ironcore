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

// SubnetIPSpec defines the desired state of SubnetIP
type SubnetIPSpec struct {
	// IP is the used IP in a given subnet
	IP string `json:"ip"`
	// Subnet is the subnet of the IP
	Subnet common.ScopeReference `json:"subnet"`
	// Target is the resource the IP is assigned to
	Target common.KindReference `json:"target"`
}

// SubnetIPStatus defines the observed state of SubnetIP
type SubnetIPStatus struct {
	common.StateFields `json:",inline"`
}

const (
	SubnetIPStateAssigned   = "Assigned"
	SubnetIPStateUnassigned = "Unassigned"
	SubnetIPStateInvalid    = "Invalid"
)

//+kubebuilder:object:root=true
//+kubebuilder:subresource:status
//+kubebuilder:printcolumn:name="IP",type=string,JSONPath=`.spec.ip`
//+kubebuilder:printcolumn:name="Subnet",type=string,JSONPath=`.spec.subnet.name`
//+kubebuilder:printcolumn:name="SubnetScope",type=string,JSONPath=`.spec.subnet.scope`
//+kubebuilder:printcolumn:name="Target",type=string,JSONPath=`.spec.target.name`
//+kubebuilder:printcolumn:name="TargetKind",type=string,JSONPath=`.spec.target.kind`
//+kubebuilder:printcolumn:name="StateFields",type=string,JSONPath=`.status.state`
//+kubebuilder:printcolumn:name="Age",type=date,JSONPath=`.metadata.creationTimestamp`

// SubnetIP is the Schema for the subnetips API
type SubnetIP struct {
	metav1.TypeMeta   `json:",inline"`
	metav1.ObjectMeta `json:"metadata,omitempty"`

	Spec   SubnetIPSpec   `json:"spec,omitempty"`
	Status SubnetIPStatus `json:"status,omitempty"`
}

//+kubebuilder:object:root=true

// SubnetIPList contains a list of SubnetIP
type SubnetIPList struct {
	metav1.TypeMeta `json:",inline"`
	metav1.ListMeta `json:"metadata,omitempty"`
	Items           []SubnetIP `json:"items"`
}

func init() {
	SchemeBuilder.Register(&SubnetIP{}, &SubnetIPList{})
}
