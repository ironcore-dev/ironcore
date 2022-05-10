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

// VolumeClaimGK is a helper to easily access the GroupKind information of an VolumeClaim
var VolumeClaimGK = schema.GroupKind{
	Group: SchemeGroupVersion.Group,
	Kind:  "VolumeClaimRef",
}

// VolumeClaimSpec defines the desired state of VolumeClaim
type VolumeClaimSpec struct {
	// VolumeRef is the reference to the Volume used by the VolumeClaim
	VolumeRef *corev1.LocalObjectReference `json:"volumeRef,omitempty"`
	// Selector is a label query over volumes to consider for binding.
	Selector *metav1.LabelSelector `json:"selector,omitempty"`
	// Resources are the requested Volume resources.
	Resources corev1.ResourceList `json:"resources"`
	// VolumeClassRef references the VolumeClass used by the Volume.
	VolumeClassRef corev1.LocalObjectReference `json:"volumeClassRef"`
}

// VolumeClaimStatus defines the observed state of VolumeClaim
type VolumeClaimStatus struct {
	// Phase represents the state a VolumeClaim can be in.
	Phase VolumeClaimPhase `json:"phase,omitempty"`
}

// VolumeClaimPhase represents the state a VolumeClaim can be in.
type VolumeClaimPhase string

const (
	// VolumeClaimPending is used for a VolumeClaim which is not yet bound.
	VolumeClaimPending VolumeClaimPhase = "Pending"
	// VolumeClaimBound is used for a VolumeClaim which is bound to a Volume.
	VolumeClaimBound VolumeClaimPhase = "Bound"
	// VolumeClaimLost is used for a VolumeClaim that lost its underlying Volume. The claim was bound to a
	// Volume and this volume does not exist any longer and all data on it was lost.
	VolumeClaimLost VolumeClaimPhase = "Lost"
)

// +k8s:deepcopy-gen:interfaces=k8s.io/apimachinery/pkg/runtime.Object
// +genclient

// VolumeClaim is the Schema for the volumeclaims API
type VolumeClaim struct {
	metav1.TypeMeta   `json:",inline"`
	metav1.ObjectMeta `json:"metadata,omitempty"`

	Spec   VolumeClaimSpec   `json:"spec,omitempty"`
	Status VolumeClaimStatus `json:"status,omitempty"`
}

// +k8s:deepcopy-gen:interfaces=k8s.io/apimachinery/pkg/runtime.Object

// VolumeClaimList contains a list of VolumeClaim
type VolumeClaimList struct {
	metav1.TypeMeta `json:",inline"`
	metav1.ListMeta `json:"metadata,omitempty"`
	Items           []VolumeClaim `json:"items"`
}
