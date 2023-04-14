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
	corev1alpha1 "github.com/onmetal/onmetal-api/api/core/v1alpha1"
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

// PrefixAllocationSpec defines the desired state of PrefixAllocation
type PrefixAllocationSpec struct {
	// IPFamily is the IPFamily of the prefix.
	// If unset but Prefix is set, this can be inferred.
	IPFamily corev1.IPFamily `json:"ipFamily,omitempty"`
	// Prefix is the prefix to allocate for this Prefix.
	Prefix *corev1alpha1.IPPrefix `json:"prefix,omitempty"`
	// PrefixLength is the length of prefix to allocate for this Prefix.
	PrefixLength int32 `json:"prefixLength,omitempty"`

	// PrefixRef references the prefix to allocate from.
	PrefixRef *corev1.LocalObjectReference `json:"prefixRef,omitempty"`
	// PrefixSelector selects the prefix to allocate from.
	PrefixSelector *metav1.LabelSelector `json:"prefixSelector,omitempty"`
}

// PrefixAllocationPhase is a phase a PrefixAllocation can be in.
type PrefixAllocationPhase string

func (p PrefixAllocationPhase) IsTerminal() bool {
	switch p {
	case PrefixAllocationPhaseFailed, PrefixAllocationPhaseAllocated:
		return true
	default:
		return false
	}
}

const (
	// PrefixAllocationPhasePending marks a PrefixAllocation as waiting for allocation.
	PrefixAllocationPhasePending PrefixAllocationPhase = "Pending"
	// PrefixAllocationPhaseAllocated marks a PrefixAllocation as allocated by a Prefix.
	PrefixAllocationPhaseAllocated PrefixAllocationPhase = "Allocated"
	// PrefixAllocationPhaseFailed marks a PrefixAllocation as failed.
	PrefixAllocationPhaseFailed PrefixAllocationPhase = "Failed"
)

// PrefixAllocationStatus is the status of a PrefixAllocation.
type PrefixAllocationStatus struct {
	// Prefix is the allocated prefix, if any
	Prefix *corev1alpha1.IPPrefix `json:"prefix,omitempty"`

	// Phase is the phase of the PrefixAllocation.
	Phase PrefixAllocationPhase `json:"phase,omitempty"`
	// LastPhaseTransitionTime is the last time the Phase changed values.
	LastPhaseTransitionTime *metav1.Time `json:"lastPhaseTransitionTime,omitempty"`
}

// +genclient
// +k8s:deepcopy-gen:interfaces=k8s.io/apimachinery/pkg/runtime.Object

// PrefixAllocation is the Schema for the prefixallocations API
type PrefixAllocation struct {
	metav1.TypeMeta   `json:",inline"`
	metav1.ObjectMeta `json:"metadata,omitempty"`

	Spec   PrefixAllocationSpec   `json:"spec,omitempty"`
	Status PrefixAllocationStatus `json:"status,omitempty"`
}

// +k8s:deepcopy-gen:interfaces=k8s.io/apimachinery/pkg/runtime.Object

// PrefixAllocationList contains a list of PrefixAllocation
type PrefixAllocationList struct {
	metav1.TypeMeta `json:",inline"`
	metav1.ListMeta `json:"metadata,omitempty"`
	Items           []PrefixAllocation `json:"items"`
}
