// Copyright 2022 OnMetal authors
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
//      http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.

package networking

import (
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

// VirtualIPClaimSpec defines the desired state of VirtualIPClaim
type VirtualIPClaimSpec struct {
	// Type is the type of VirtualIP.
	Type VirtualIPType
	// IPFamily is the ip family of the VirtualIP.
	IPFamily corev1.IPFamily

	// Selector is the selector for a VirtualIP.
	Selector *metav1.LabelSelector `json:"selector,omitempty"`

	// VirtualIPRef references the virtual ip to claim.
	VirtualIPRef *corev1.LocalObjectReference
}

// VirtualIPClaimStatus defines the observed state of VirtualIPClaim
type VirtualIPClaimStatus struct {
	Phase VirtualIPClaimPhase
}

// VirtualIPClaimPhase represents the state a VirtualIPClaim can be in.
type VirtualIPClaimPhase string

const (
	// VirtualIPClaimPhasePending is used for a VirtualIPClaim which is not yet bound.
	VirtualIPClaimPhasePending VirtualIPClaimPhase = "Pending"
	// VirtualIPClaimPhaseBound is used for a VirtualIPClaim which is bound to a VirtualIP.
	VirtualIPClaimPhaseBound VirtualIPClaimPhase = "Bound"
	// VirtualIPClaimPhaseLost is used for a VirtualIPClaim that lost its underlying VirtualIP.
	VirtualIPClaimPhaseLost VirtualIPClaimPhase = "Lost"
)

// +genclient
// +k8s:deepcopy-gen:interfaces=k8s.io/apimachinery/pkg/runtime.Object

// VirtualIPClaim is the Schema for the virtualipclaims API
type VirtualIPClaim struct {
	metav1.TypeMeta
	metav1.ObjectMeta

	Spec   VirtualIPClaimSpec
	Status VirtualIPClaimStatus
}

// +k8s:deepcopy-gen:interfaces=k8s.io/apimachinery/pkg/runtime.Object

// VirtualIPClaimList contains a list of VirtualIPClaim
type VirtualIPClaimList struct {
	metav1.TypeMeta
	metav1.ListMeta
	Items []VirtualIPClaim
}

type VirtualIPClaimTemplateSpec struct {
	metav1.ObjectMeta
	Spec VirtualIPClaimSpec
}
