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
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime/schema"
)

// IPAMRangeGK is a helper to easily access the GroupKind information of an IPAMRange
var IPAMRangeGK = schema.GroupKind{
	Group: GroupVersion.Group,
	Kind:  "IPAMRange",
}

// IPAMRangeSpec defines the desired state of IPAMRange
// Either parent and size or a give CIDR must be specified. If parent is specified,
// the effective range of the given size is allocated from the parent IP range. If parent and CIDR
// is defined, the given CIDR must be in the parent range and unused. It will be allocated if possible.
// Otherwise, the status of the object will be set to "Invalid".
type IPAMRangeSpec struct {
	// Parent is the reference of the Parent IPAMRange from which the Cidr or size should be derived
	Parent *corev1.LocalObjectReference `json:"parent,omitempty"`
	// CIDRs is a list of CIDR specs which are defined for this IPAMRange
	CIDRs []string `json:"cidrs,omitempty"`
	// Mode is the mode to request an IPAMRange.
	Mode IPAMRangeMode `json:"mode,omitempty"`
}

// IPAMRangeMode is the mode to request IPAMRanges.
type IPAMRangeMode string

const (
	// ModeRoundRobin requests IPAMRanges in a round-robin fashion, distributing evenly.
	ModeRoundRobin IPAMRangeMode = "RoundRobin"
	// ModeFirstMatch requests IPAMRanges by using the first possible match.
	ModeFirstMatch IPAMRangeMode = "FirstMatch"
)

type IPAMRangeState string

const (
	IPAMRangeAvailable IPAMRangeState = "Available"
	IPAMRangeReady     IPAMRangeState = "Ready"
	IPAMRangeUp        IPAMRangeState = "Up"
	IPAMRangeError     IPAMRangeState = "Error"
	IPAMRangeInvalid   IPAMRangeState = "Invalid"
	IPAMRangeBusy      IPAMRangeState = "Busy"
	IPAMRangePending   IPAMRangeState = "Pending"
)

// IPAMRangeStatus defines the observed state of IPAMRange
type IPAMRangeStatus struct {
	State      IPAMRangeState       `json:"state,omitempty"`
	Message    string               `json:"message,omitempty"`
	Conditions []IPAMRangeCondition `json:"conditions,omitempty"`
	// CIDRs is a list of effective cidrs which belong to this IPAMRange
	CIDRs            []CIDRAllocationStatus `json:"cidrs,omitempty"`
	AllocationState  []string               `json:"allocationState,omitempty"`
	RoundRobinState  []string               `json:"roundRobinState,omitempty"`
	PendingRequest   *IPAMPendingRequest    `json:"pendingRequests,omitempty"`
	PendingDeletions []CIDRAllocationStatus `json:"pendingDeletions,omitempty"`
}

// IPAMRangeConditionType is a type a IPAMRangeCondition can have.
type IPAMRangeConditionType string

// IPAMRangeCondition is one of the conditions of a volume.
type IPAMRangeCondition struct {
	// Type is the type of the condition.
	Type IPAMRangeConditionType `json:"type"`
	// Status is the status of the condition.
	Status corev1.ConditionStatus `json:"status"`
	// Reason is a machine-readable indication of why the condition is in a certain state.
	Reason string `json:"reason"`
	// Message is a human-readable explanation of why the condition has a certain reason / state.
	Message string `json:"message"`
	// ObservedGeneration represents the .metadata.generation that the condition was set based upon.
	ObservedGeneration int64 `json:"observedGeneration,omitempty"`
	// LastUpdateTime is the last time a condition has been updated.
	LastUpdateTime metav1.Time `json:"lastUpdateTime,omitempty"`
	// LastTransitionTime is the last time the status of a condition has transitioned from one state to another.
	LastTransitionTime metav1.Time `json:"lastTransitionTime,omitempty"`
}

type CIDRAllocation struct {
	Request string `json:"request"`
	CIDR    string `json:"cidr"`
}

// CIDRAllocationStatus is the result of a CIDR allocation request
type CIDRAllocationStatus struct {
	CIDRAllocation `json:",inline"`
	Status         AllocationState `json:"status"`
	Message        string          `json:"message,omitempty"`
}

// AllocationState is a state an allocation can be in.
type AllocationState string

const (
	// AllocationStateAllocated reports that the allocation has been made successfully.
	AllocationStateAllocated AllocationState = "Allocated"
	// AllocationStateBusy reports that an allocation is busy.
	AllocationStateBusy AllocationState = "Busy"
	// AllocationStateFailed reports that an allocation has failed.
	AllocationStateFailed AllocationState = "Failed"
)

type IPAMPendingRequest struct {
	Name      string           `json:"name"`
	Namespace string           `json:"namespace"`
	CIDRs     []CIDRAllocation `json:"cidrs,omitempty"`
	Deletions []CIDRAllocation `json:"deletions,omitempty"`
}

//+kubebuilder:object:root=true
//+kubebuilder:subresource:status
//+kubebuilder:resource:shortName=ipr
//+kubebuilder:printcolumn:name="CIDRS",type=string,JSONPath=`.spec.cidrs`
//+kubebuilder:printcolumn:name="EffectiveCIDRs",type=string,JSONPath=`.status.cidrs`
//+kubebuilder:printcolumn:name="Parent",type=string,JSONPath=`.spec.parent.name`
//+kubebuilder:printcolumn:name="State",type=string,JSONPath=`.status.state`
//+kubebuilder:printcolumn:name="Age",type=date,JSONPath=`.metadata.creationTimestamp`

// IPAMRange is the Schema for the ipamranges API
type IPAMRange struct {
	metav1.TypeMeta   `json:",inline"`
	metav1.ObjectMeta `json:"metadata,omitempty"`

	Spec   IPAMRangeSpec   `json:"spec,omitempty"`
	Status IPAMRangeStatus `json:"status,omitempty"`
}

//+kubebuilder:object:root=true

// IPAMRangeList contains a list of IPAMRange
type IPAMRangeList struct {
	metav1.TypeMeta `json:",inline"`
	metav1.ListMeta `json:"metadata,omitempty"`
	Items           []IPAMRange `json:"items"`
}

func init() {
	SchemeBuilder.Register(&IPAMRange{}, &IPAMRangeList{})
}
