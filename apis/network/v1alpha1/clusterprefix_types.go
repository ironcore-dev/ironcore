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

const (
	ClusterPrefixKind = "ClusterPrefix"
)

// ClusterPrefixSpec defines the desired state of ClusterPrefix
type ClusterPrefixSpec struct {
	// ParentRef references the parent to allocate the Prefix from.
	// If ParentRef and ParentSelector is empty, the Prefix is considered a root prefix and thus
	// allocated by itself.
	ParentRef *corev1.LocalObjectReference `json:"parentRef,omitempty"`
	// ParentSelector is the LabelSelector to use for determining the parent for this Prefix.
	ParentSelector *metav1.LabelSelector `json:"parentSelector,omitempty"`
	// PrefixSpace is the space the ClusterPrefix manages.
	PrefixSpace `json:",inline"`
}

// ClusterPrefixStatus defines the observed state of ClusterPrefix
type ClusterPrefixStatus struct {
	// Conditions is a list of conditions of a ClusterPrefix.
	Conditions []ClusterPrefixCondition `json:"conditions,omitempty"`
	// Available is a list of available prefixes.
	Available []commonv1alpha1.IPPrefix `json:"available,omitempty"`
	// Reserved is a list of reserved prefixes.
	Reserved []commonv1alpha1.IPPrefix `json:"reserved,omitempty"`
}

type ClusterPrefixConditionType string

const (
	ClusterPrefixReady ClusterPrefixConditionType = "Ready"
)

const (
	ClusterPrefixReadyReasonPending   = "Pending"
	ClusterPrefixReadyReasonAllocated = "Allocated"
)

type ClusterPrefixCondition struct {
	Type               ClusterPrefixConditionType `json:"type"`
	Status             corev1.ConditionStatus     `json:"status"`
	Reason             string                     `json:"reason,omitempty"`
	Message            string                     `json:"message,omitempty"`
	LastTransitionTime metav1.Time                `json:"lastTransitionTime,omitempty"`
}

//+kubebuilder:object:root=true
//+kubebuilder:resource:scope=Cluster
//+kubebuilder:subresource:status
//+kubebuilder:printcolumn:name="Prefix",type=string,JSONPath=`.spec.prefix`
//+kubebuilder:printcolumn:name="Parent",type=string,JSONPath=`.spec.parentRef.name`
//+kubebuilder:printcolumn:name="State",type=string,JSONPath=`.status.conditions[?(@.type == "Ready")].reason`
//+kubebuilder:printcolumn:name="Age",type=date,JSONPath=".metadata.creationTimestamp"

// ClusterPrefix is the Schema for the clusterprefixes API
type ClusterPrefix struct {
	metav1.TypeMeta   `json:",inline"`
	metav1.ObjectMeta `json:"metadata,omitempty"`

	Spec   ClusterPrefixSpec   `json:"spec,omitempty"`
	Status ClusterPrefixStatus `json:"status,omitempty"`
}

//+kubebuilder:object:root=true

// ClusterPrefixList contains a list of ClusterPrefix
type ClusterPrefixList struct {
	metav1.TypeMeta `json:",inline"`
	metav1.ListMeta `json:"metadata,omitempty"`
	Items           []ClusterPrefix `json:"items"`
}

func init() {
	SchemeBuilder.Register(&ClusterPrefix{}, &ClusterPrefixList{})
}
