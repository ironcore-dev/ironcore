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

// ClusterPrefixAllocationSpec defines the desired state of ClusterPrefixAllocation
type ClusterPrefixAllocationSpec struct {
	PrefixRef                      *corev1.LocalObjectReference `json:"prefixRef,omitempty"`
	PrefixSelector                 *metav1.LabelSelector        `json:"prefixSelector,omitempty"`
	ClusterPrefixAllocationRequest `json:",inline"`
}

type ClusterPrefixAllocationRequest struct {
	//+nullable
	Prefix       commonv1alpha1.IPPrefix `json:"prefix,omitempty"`
	PrefixLength int8                    `json:"prefixLength,omitempty"`
}

// ClusterPrefixAllocationStatus defines the observed state of ClusterPrefixAllocation
type ClusterPrefixAllocationStatus struct {
	ClusterPrefixAllocationResult `json:",inline,omitempty"`
	Conditions                    []ClusterPrefixAllocationCondition `json:"conditions,omitempty"`
}

type ClusterPrefixAllocationResult struct {
	//+nullable
	Prefix commonv1alpha1.IPPrefix `json:"prefix,omitempty"`
}

type ClusterPrefixAllocationConditionType string

const (
	ClusterPrefixAllocationReady ClusterPrefixAllocationConditionType = "Ready"
)

const (
	ClusterPrefixAllocationReadyReasonFailed    = "Failed"
	ClusterPrefixAllocationReadyReasonSucceeded = "Succeeded"
	ClusterPrefixAllocationReadyReasonPending   = "Pending"
)

type ClusterPrefixAllocationCondition struct {
	Type               ClusterPrefixAllocationConditionType `json:"type"`
	Status             corev1.ConditionStatus               `json:"status"`
	Reason             string                               `json:"reason,omitempty"`
	Message            string                               `json:"message,omitempty"`
	LastTransitionTime metav1.Time                          `json:"lastTransitionTime,omitempty"`
}

//+kubebuilder:object:root=true
//+kubebuilder:resource:scope=Cluster
//+kubebuilder:subresource:status
//+kubebuilder:printcolumn:name="RPrefix",type=string,JSONPath=`.spec.prefix`
//+kubebuilder:printcolumn:name="RPLen",type=string,JSONPath=`.spec.prefixLength`
//+kubebuilder:printcolumn:name="Prefix",type=string,JSONPath=`.status.prefix`
//+kubebuilder:printcolumn:name="State",type=string,JSONPath=`.status.conditions[?(@.type == "Ready")].reason`
//+kubebuilder:printcolumn:name="Age",type=date,JSONPath=".metadata.creationTimestamp"

// ClusterPrefixAllocation is the Schema for the clusterprefixallocations API
type ClusterPrefixAllocation struct {
	metav1.TypeMeta   `json:",inline"`
	metav1.ObjectMeta `json:"metadata,omitempty"`

	Spec   ClusterPrefixAllocationSpec   `json:"spec,omitempty"`
	Status ClusterPrefixAllocationStatus `json:"status,omitempty"`
}

//+kubebuilder:object:root=true

// ClusterPrefixAllocationList contains a list of ClusterPrefixAllocation
type ClusterPrefixAllocationList struct {
	metav1.TypeMeta `json:",inline"`
	metav1.ListMeta `json:"metadata,omitempty"`
	Items           []ClusterPrefixAllocation `json:"items"`
}

func init() {
	SchemeBuilder.Register(&ClusterPrefixAllocation{}, &ClusterPrefixAllocationList{})
}
