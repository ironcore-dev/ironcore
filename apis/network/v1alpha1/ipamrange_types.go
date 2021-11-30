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
	// Parent is the reference of the Parent IPAMRange from which the CIDR or size should be derived
	Parent *corev1.LocalObjectReference `json:"parent,omitempty"`
	// Items is a list of CIDR specs which are defined for this IPAMRange
	Items []IPAMRangeItem `json:"itmes,omitempty"`
}

type IPAMRangeItem struct {
	// Size requests a CIDR of the given size.
	Size int32                   `json:"size,omitempty"`
	CIDR *commonv1alpha1.CIDR    `json:"cidr,omitempty"`
	IPs  *commonv1alpha1.IPRange `json:"ips,omitempty"`
	// IPCount requests a consecutive IP range of the given size.
	IPCount int32 `json:"ipCount,omitempty"`
}

// IPAMRangeStatus defines the observed state of IPAMRange
type IPAMRangeStatus struct {
	ObservedGeneration int64                       `json:"observedGeneration,omitempty"`
	Allocations        []IPAMRangeAllocationStatus `json:"allocations,omitempty"`
}

// IPAMRangeAllocationStatus is the status of an IPAMRange allocation.
type IPAMRangeAllocationStatus struct {
	// CIDR defines the cidr
	CIDR    *commonv1alpha1.CIDR         `json:"cidr,omitempty"`
	IPs     *commonv1alpha1.IPRange      `json:"ips,omitempty"`
	State   IPAMRangeAllocationState     `json:"state"`
	Request *IPAMRangeItem               `json:"request,omitempty"`
	User    *corev1.LocalObjectReference `json:"user,omitempty"`
}

// IPAMRangeAllocationState is a state an IPAMRangeAllocationStatus can have.
type IPAMRangeAllocationState string

const (
	// IPAMRangeAllocationFree reports that an allocation is available for use (by children or direct consumers).
	IPAMRangeAllocationFree IPAMRangeAllocationState = "Free"
	// IPAMRangeAllocationUsed reports that an allocation is used by a child.
	IPAMRangeAllocationUsed IPAMRangeAllocationState = "Used"
	// IPAMRangeAllocationFailed reports that an allocation could not be made successfully.
	IPAMRangeAllocationFailed IPAMRangeAllocationState = "Failed"
)

//+kubebuilder:object:root=true
//+kubebuilder:subresource:status
//+kubebuilder:resource:shortName=ipr
//+kubebuilder:printcolumn:name="Parent",type=string,JSONPath=`.spec.parent.name`
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
