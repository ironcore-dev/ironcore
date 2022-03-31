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

// PrefixAllocationSpec defines the desired state of PrefixAllocation
type PrefixAllocationSpec struct {
	// PrefixRef references the prefix to allocate from.
	//+optional
	PrefixRef               *PrefixReference `json:"prefixRef,omitempty"`
	PrefixSelector          *PrefixSelector  `json:"prefixSelector,omitempty"`
	PrefixAllocationRequest `json:",inline"`
}

type PrefixAllocationRequest struct {
	// +nullable
	Prefix       commonv1alpha1.IPPrefix `json:"prefix,omitempty"`
	PrefixLength int32                   `json:"prefixLength,omitempty"`
	Range        *commonv1alpha1.IPRange `json:"range,omitempty"`
	RangeLength  int64                   `json:"rangeLength,omitempty"`
}

type PrefixAllocationResult struct {
	// +nullable
	Prefix commonv1alpha1.IPPrefix `json:"prefix,omitempty"`
	Range  *commonv1alpha1.IPRange `json:"range,omitempty"`
}

type PrefixAllocationStatus struct {
	PrefixAllocationResult `json:",inline,omitempty"`
	Conditions             []PrefixAllocationCondition `json:"conditions,omitempty"`
}

type PrefixAllocationConditionType string

const (
	PrefixAllocationReady PrefixAllocationConditionType = "Ready"
)

const (
	PrefixAllocationReadyReasonFailed    = "Failed"
	PrefixAllocationReadyReasonPending   = "Pending"
	PrefixAllocationReadyReasonSucceeded = "Succeeded"
)

type PrefixAllocationCondition struct {
	Type               PrefixAllocationConditionType `json:"type"`
	Status             corev1.ConditionStatus        `json:"status"`
	Reason             string                        `json:"reason,omitempty"`
	Message            string                        `json:"message,omitempty"`
	LastTransitionTime metav1.Time                   `json:"lastTransitionTime,omitempty"`
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
