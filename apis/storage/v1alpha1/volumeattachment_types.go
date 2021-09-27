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
)

// VolumeAttachmentSpec defines the desired state of VolumeAttachment
type VolumeAttachmentSpec struct {
	// Volume is a reference of the volume object which should be attached
	Volume corev1.LocalObjectReference `json:"volume,omitempty"`
	// Machine is a reference of the machine object which the volume should be attached to
	Machine corev1.LocalObjectReference `json:"machine"`
}

// VolumeAttachmentStatus defines the observed state of VolumeAttachment
type VolumeAttachmentStatus struct {
	// State reports a VolumeAttachmentState a VolumeAttachment is in.
	State VolumeAttachmentState `json:"state,omitempty"`
	// Conditions reports the conditions a VolumeAttachment may have.
	Conditions []VolumeAttachmentCondition `json:"conditions,omitempty"`
}

// VolumeAttachmentConditionType is a type a VolumeAttachmentCondition can have.
type VolumeAttachmentConditionType string

// VolumeAttachmentCondition is one of the conditions of a volume.
type VolumeAttachmentCondition struct {
	// Type is the type of the condition.
	Type VolumeAttachmentConditionType `json:"type"`
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

// VolumeAttachmentState is a state a VolumeAttachment can be int.
type VolumeAttachmentState string

const (
	// VolumeAttachmentStateAttached reports that the VolumeAttachment has been attached.
	VolumeAttachmentStateAttached = "Attached"
	// VolumeAttachmentStateInvalid reports that the VolumeAttachment is invalid.
	VolumeAttachmentStateInvalid = "Invalid"
	// VolumeAttachmentStateError reports that the VolumeAttachment is in an error state.
	VolumeAttachmentStateError = "Error"
)

//+kubebuilder:object:root=true
//+kubebuilder:subresource:status
//+kubebuilder:printcolumn:name="Volume",type=string,JSONPath=`.spec.volume.name`
//+kubebuilder:printcolumn:name="Machine",type=string,JSONPath=`.spec.machine.name`
//+kubebuilder:printcolumn:name="Device",type=string,JSONPath=`.spec.device`
//+kubebuilder:printcolumn:name="Image",type=string,JSONPath=`.spec.source.image`
//+kubebuilder:printcolumn:name="Snapshot",type=string,JSONPath=`.spec.source.snapshot`
//+kubebuilder:printcolumn:name="State",type=string,JSONPath=`.status.state`
//+kubebuilder:printcolumn:name="Age",type=date,JSONPath=`.metadata.creationTimestamp`

// VolumeAttachment is the Schema for the volumeattachments API
type VolumeAttachment struct {
	metav1.TypeMeta   `json:",inline"`
	metav1.ObjectMeta `json:"metadata,omitempty"`

	Spec   VolumeAttachmentSpec   `json:"spec,omitempty"`
	Status VolumeAttachmentStatus `json:"status,omitempty"`
}

//+kubebuilder:object:root=true

// VolumeAttachmentList contains a list of VolumeAttachment
type VolumeAttachmentList struct {
	metav1.TypeMeta `json:",inline"`
	metav1.ListMeta `json:"metadata,omitempty"`
	Items           []VolumeAttachment `json:"items"`
}

func init() {
	SchemeBuilder.Register(&VolumeAttachment{}, &VolumeAttachmentList{})
}
