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

// RouteTableSpec defines the desired state of RouteTable
type RouteTableSpec struct {
	Routes []Route `json:"routes"`
}

// Route
type Route struct {
	SubnetRef common.ScopeReference `json:"subnetRef,omitempty"`
	CIDR      string                `json:"cidr,omitempty"`
	Target    common.ScopeReference `json:"target,omitempty"`
}

// RouteTableStatus defines the observed state of RouteTable
type RouteTableStatus struct {
	common.StateFields `json:",inline"`
}

//+kubebuilder:object:root=true
//+kubebuilder:subresource:status
//+kubebuilder:printcolumn:name="StateFields",type=string,JSONPath=`.status.state`
//+kubebuilder:printcolumn:name="Age",type=date,JSONPath=`.metadata.creationTimestamp`

// RouteTable is the Schema for the routetables API
type RouteTable struct {
	metav1.TypeMeta   `json:",inline"`
	metav1.ObjectMeta `json:"metadata,omitempty"`

	Spec   RouteTableSpec   `json:"spec,omitempty"`
	Status RouteTableStatus `json:"status,omitempty"`
}

//+kubebuilder:object:root=true

// RouteTableList contains a list of RouteTable
type RouteTableList struct {
	metav1.TypeMeta `json:",inline"`
	metav1.ListMeta `json:"metadata,omitempty"`
	Items           []RouteTable `json:"items"`
}

func init() {
	SchemeBuilder.Register(&RouteTable{}, &RouteTableList{})
}
