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

// VolumeSpec defines the desired state of Volume
type VolumeSpec struct {
	// StorageClass is the storage class of a volume
	StorageClass corev1.LocalObjectReference `json:"storageClass"`
	// StoragePoolSelector selects a suitable StoragePool by the given labels.
	StoragePoolSelector map[string]string `json:"storagePoolSelector,omitempty"`
	// StoragePool indicates which storage pool to use for a volume.
	// If unset, the scheduler will figure out a suitable StoragePool.
	StoragePool corev1.LocalObjectReference `json:"storagePool"`
	// Resources is a description of the volume's resources and capacity.
	Resources corev1.ResourceList `json:"resources,omitempty"`
}

// VolumeStatus defines the observed state of Volume
type VolumeStatus struct {
	State      VolumeState       `json:"state,omitempty"`
	Conditions []VolumeCondition `json:"conditions,omitempty"`
}

// VolumeState is a possible state a volume can be in.
type VolumeState string

const (
	// VolumeStateAvailable reports whether the volume is available to be used.
	VolumeStateAvailable VolumeState = "Available"
	// VolumeStatePending reports whether the volume is about to be ready.
	VolumeStatePending VolumeState = "Pending"
	// VolumeStateAttached reports that the volume is attached and in-use.
	VolumeStateAttached VolumeState = "Attached"
	// VolumeStateError reports that the volume is in an error state.
	VolumeStateError VolumeState = "Error"
)

// VolumeConditionType is a type a VolumeCondition can have.
type VolumeConditionType string

// VolumeCondition is one of the conditions of a volume.
type VolumeCondition struct {
	// Type is the type of the condition.
	Type VolumeConditionType `json:"type"`
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

//+kubebuilder:object:root=true
//+kubebuilder:subresource:status
//+kubebuilder:printcolumn:name="State",type=string,JSONPath=`.status.state`
//+kubebuilder:printcolumn:name="Age",type=date,JSONPath=`.metadata.creationTimestamp`

// Volume is the Schema for the volumes API
type Volume struct {
	metav1.TypeMeta   `json:",inline"`
	metav1.ObjectMeta `json:"metadata,omitempty"`

	Spec   VolumeSpec   `json:"spec,omitempty"`
	Status VolumeStatus `json:"status,omitempty"`
}

//+kubebuilder:object:root=true

// VolumeList contains a list of Volume
type VolumeList struct {
	metav1.TypeMeta `json:",inline"`
	metav1.ListMeta `json:"metadata,omitempty"`
	Items           []Volume `json:"items"`
}

func init() {
	SchemeBuilder.Register(&Volume{}, &VolumeList{})
}
