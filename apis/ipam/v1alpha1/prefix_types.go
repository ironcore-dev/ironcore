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
	PrefixKind = "Prefix"
)

type PrefixReference struct {
	// Kind is the kind of prefix to select.
	//+kubebuilder:validation:Enum=Prefix;ClusterPrefix
	//+optional
	Kind string `json:"kind"`
	Name string `json:"name"`
}

type PrefixSelector struct {
	// Kind is the kind of prefix to select.
	//+kubebuilder:validation:Enum=Prefix;ClusterPrefix
	//+optional
	Kind                 string `json:"kind"`
	metav1.LabelSelector `json:",inline"`
}

// PrefixSpec defines the desired state of Prefix
type PrefixSpec struct {
	// ParentRef references the parent to allocate the Prefix from.
	// If ParentRef and ParentSelector is empty, the Prefix is considered a root prefix and thus
	// allocated by itself.
	ParentRef *PrefixReference `json:"parentRef,omitempty"`
	// ParentSelector is the LabelSelector to use for determining the parent for this Prefix.
	ParentSelector *PrefixSelector `json:"parentSelector,omitempty"`
	// PrefixSpace is the definition of the space the prefix manages.
	PrefixSpace `json:",inline"`
}

// PrefixSpace is the space a prefix manages.
type PrefixSpace struct {
	// PrefixLength is the length of prefix to allocate for this Prefix.
	PrefixLength int32 `json:"prefixLength,omitempty"`
	// Prefix is the prefix to allocate for this Prefix.
	//+optional
	//+nullable
	Prefix commonv1alpha1.IPPrefix `json:"prefix,omitempty"`

	// Reservations is a list of IPPrefixes to reserve for this Prefix.
	Reservations []commonv1alpha1.IPPrefix `json:"reservations,omitempty"`
	// ReservationLengths is a list of IPPrefixes to reserve for this Prefix.
	ReservationLengths []int32 `json:"reservationLengths,omitempty"`
}

// PrefixStatus defines the observed state of Prefix
type PrefixStatus struct {
	// Conditions is a list of conditions of a Prefix.
	Conditions []PrefixCondition `json:"conditions,omitempty"`
	// Available is a list of available prefixes.
	Available []commonv1alpha1.IPPrefix `json:"available,omitempty"`
	// Reserved is a list of reserved prefixes.
	Reserved []commonv1alpha1.IPPrefix `json:"reserved,omitempty"`
}

type PrefixConditionType string

const (
	PrefixReady PrefixConditionType = "Ready"
)

const (
	PrefixReadyReasonPending   = "Pending"
	PrefixReadyReasonAllocated = "Allocated"
)

type PrefixCondition struct {
	Type               PrefixConditionType    `json:"type"`
	Status             corev1.ConditionStatus `json:"status"`
	Reason             string                 `json:"reason,omitempty"`
	Message            string                 `json:"message,omitempty"`
	LastTransitionTime metav1.Time            `json:"lastTransitionTime,omitempty"`
}

// +genclient
// +k8s:deepcopy-gen:interfaces=k8s.io/apimachinery/pkg/runtime.Object

// Prefix is the Schema for the prefixes API
type Prefix struct {
	metav1.TypeMeta   `json:",inline"`
	metav1.ObjectMeta `json:"metadata,omitempty"`

	Spec   PrefixSpec   `json:"spec,omitempty"`
	Status PrefixStatus `json:"status,omitempty"`
}

// +k8s:deepcopy-gen:interfaces=k8s.io/apimachinery/pkg/runtime.Object

// PrefixList contains a list of Prefix
type PrefixList struct {
	metav1.TypeMeta `json:",inline"`
	metav1.ListMeta `json:"metadata,omitempty"`
	Items           []Prefix `json:"items"`
}
