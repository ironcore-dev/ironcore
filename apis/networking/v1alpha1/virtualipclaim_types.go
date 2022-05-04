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

package v1alpha1

import (
	commonv1alpha1 "github.com/onmetal/onmetal-api/apis/common/v1alpha1"
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

// VirtualIPClaimSpec defines the desired state of VirtualIPClaim
type VirtualIPClaimSpec struct {
	// Type is the type of VirtualIP.
	Type VirtualIPType `json:"type"`
	// IPFamily is the ip family of the VirtualIP.
	IPFamily corev1.IPFamily `json:"ipFamily"`

	// VirtualIPRef references the virtual ip to claim.
	VirtualIPRef *corev1.LocalObjectReference `json:"virtualIPRef,omitempty"`
}

// VirtualIPClaimStatus defines the observed state of VirtualIPClaim
type VirtualIPClaimStatus struct {
	// IP is the allocated IP, if any.
	IP *commonv1alpha1.IP `json:"ip,omitempty"`
}

// +genclient
// +k8s:deepcopy-gen:interfaces=k8s.io/apimachinery/pkg/runtime.Object

// VirtualIPClaim is the Schema for the virtualipclaims API
type VirtualIPClaim struct {
	metav1.TypeMeta   `json:",inline"`
	metav1.ObjectMeta `json:"metadata,omitempty"`

	Spec   VirtualIPClaimSpec   `json:"spec,omitempty"`
	Status VirtualIPClaimStatus `json:"status,omitempty"`
}

// +k8s:deepcopy-gen:interfaces=k8s.io/apimachinery/pkg/runtime.Object

// VirtualIPClaimList contains a list of VirtualIPClaim
type VirtualIPClaimList struct {
	metav1.TypeMeta `json:",inline"`
	metav1.ListMeta `json:"metadata,omitempty"`
	Items           []VirtualIPClaim `json:"items"`
}

type VirtualIPClaimTemplateSpec struct {
	metav1.ObjectMeta `json:"metadata,omitempty"`
	Spec              VirtualIPClaimSpec `json:"spec,omitempty"`
}
