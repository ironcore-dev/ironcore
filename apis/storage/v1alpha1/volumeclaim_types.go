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
	v1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

// VolumeClaimSpec defines the desired state of VolumeClaim
type VolumeClaimSpec struct {
	// VolumeRef is the reference to the Volume used by the VolumeClaim
	VolumeRef v1.LocalObjectReference `json:"volumeRef,omitempty"`
	// Selector is a label query over volumes to consider for binding.
	Selector metav1.LabelSelector `json:"selector,omitempty"`
}

// VolumeClaimStatus defines the observed state of VolumeClaim
type VolumeClaimStatus struct {
	// VolumeClaimPhase represents the state a VolumeClaim can be in.
	Phase VolumeClaimPhase `json:"phase,omitempty"`
}

// VolumeClaimPhase represents the state a VolumeClaim can be in.
type VolumeClaimPhase string

const (
	// VolumeClaimPhasePending is used for a VolumeClaim which is not yet bound.
	VolumeClaimPhasePending VolumeClaimPhase = "Pending"
	// VolumeClaimPhaseBound is used for a VolumeClaim which is bound to a Volume.
	VolumeClaimPhaseBound VolumeClaimPhase = "Bound"
	// VolumeClaimPhaseLost is used for a VolumeClaim that lost its underlying Volume. The claim was bound to a
	// Volume and this volume does not exist any longer and all data on it was lost.
	VolumeClaimPhaseLost VolumeClaimPhase = "Lost"
)

//+kubebuilder:object:root=true
//+kubebuilder:subresource:status
//+kubebuilder:printcolumn:name="ReferencedVolume",type=string,JSONPath=`.spec.volumeRef.name`
//+kubebuilder:printcolumn:name="Phase",type=string,JSONPath=`.status.phase`
//+kubebuilder:printcolumn:name="Age",type=date,JSONPath=`.metadata.creationTimestamp`

// VolumeClaim is the Schema for the volumeclaims API
type VolumeClaim struct {
	metav1.TypeMeta   `json:",inline"`
	metav1.ObjectMeta `json:"metadata,omitempty"`

	Spec   VolumeClaimSpec   `json:"spec,omitempty"`
	Status VolumeClaimStatus `json:"status,omitempty"`
}

//+kubebuilder:object:root=true

// VolumeClaimList contains a list of VolumeClaim
type VolumeClaimList struct {
	metav1.TypeMeta `json:",inline"`
	metav1.ListMeta `json:"metadata,omitempty"`
	Items           []VolumeClaim `json:"items"`
}

func init() {
	SchemeBuilder.Register(&VolumeClaim{}, &VolumeClaimList{})
}
