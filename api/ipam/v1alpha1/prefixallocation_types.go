// SPDX-FileCopyrightText: 2023 SAP SE or an SAP affiliate company and IronCore contributors
// SPDX-License-Identifier: Apache-2.0

package v1alpha1

import (
	commonv1alpha1 "github.com/ironcore-dev/ironcore/api/common/v1alpha1"
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

// PrefixAllocationSpec defines the desired state of PrefixAllocation
type PrefixAllocationSpec struct {
	// IPFamily is the IPFamily of the prefix.
	// If unset but Prefix is set, this can be inferred.
	IPFamily corev1.IPFamily `json:"ipFamily,omitempty"`
	// Prefix is the prefix to allocate for this Prefix.
	Prefix *commonv1alpha1.IPPrefix `json:"prefix,omitempty"`
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
	Prefix *commonv1alpha1.IPPrefix `json:"prefix,omitempty"`

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
