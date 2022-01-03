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
)

// IPSpec defines the desired state of IP
type IPSpec struct {
	// PrefixRef references the parent to allocate the IP from.
	PrefixRef *PrefixReference `json:"prefixRef,omitempty"`
	// PrefixSelector is the LabelSelector to use for determining the parent for this IP.
	PrefixSelector *PrefixSelector `json:"prefixSelector,omitempty"`
	// IP is the ip to allocate.
	//+nullable
	IP commonv1alpha1.IP `json:"ip,omitempty"`
}

// IPStatus defines the observed state of IP
type IPStatus struct {
	Conditions []IPCondition `json:"conditions,omitempty"`
}

type IPConditionType string

const (
	IPReady IPConditionType = "Ready"
)

const (
	IPReadyReasonPending   = "Pending"
	IPReadyReasonAllocated = "Allocated"
)

type IPCondition struct {
	Type               IPConditionType        `json:"type"`
	Status             corev1.ConditionStatus `json:"status"`
	Reason             string                 `json:"reason,omitempty"`
	Message            string                 `json:"message,omitempty"`
	LastTransitionTime metav1.Time            `json:"lastTransitionTime,omitempty"`
}

//+kubebuilder:object:root=true
//+kubebuilder:subresource:status
//+kubebuilder:printcolumn:name="IP",type=string,JSONPath=`.spec.ip`
//+kubebuilder:printcolumn:name="Prefix",type=string,JSONPath=`.spec.prefixRef.name`
//+kubebuilder:printcolumn:name="State",type=string,JSONPath=`.status.conditions[?(@.type == "Ready")].reason`
//+kubebuilder:printcolumn:name="Age",type=date,JSONPath=".metadata.creationTimestamp"

// IP is the Schema for the ips API
type IP struct {
	metav1.TypeMeta   `json:",inline"`
	metav1.ObjectMeta `json:"metadata,omitempty"`

	Spec   IPSpec   `json:"spec,omitempty"`
	Status IPStatus `json:"status,omitempty"`
}

//+kubebuilder:object:root=true

// IPList contains a list of IP
type IPList struct {
	metav1.TypeMeta `json:",inline"`
	metav1.ListMeta `json:"metadata,omitempty"`
	Items           []IP `json:"items"`
}

func init() {
	SchemeBuilder.Register(&IP{}, &IPList{})
}