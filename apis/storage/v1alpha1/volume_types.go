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
	"k8s.io/apimachinery/pkg/types"
)

// VolumeGK is a helper to easily access the GroupKind information of an Volume
var VolumeGK = schema.GroupKind{
	Group: SchemeGroupVersion.Group,
	Kind:  "Volume",
}

// VolumeSpec defines the desired state of Volume
type VolumeSpec struct {
	// StorageClassRef is the storage class of a volume
	StorageClassRef corev1.LocalObjectReference `json:"storageClassRef"`
	// StoragePoolSelector selects a suitable StoragePool by the given labels.
	StoragePoolSelector map[string]string `json:"storagePoolSelector,omitempty"`
	// StoragePool indicates which storage pool to use for a volume.
	// If unset, the scheduler will figure out a suitable StoragePool.
	StoragePool corev1.LocalObjectReference `json:"storagePool"`
	// ClaimRef is the reference to the VolumeClaim used by the Volume.
	ClaimRef ClaimReference `json:"claimRef,omitempty"`
	// Resources is a description of the volume's resources and capacity.
	Resources corev1.ResourceList `json:"resources,omitempty"`
	// Tolerations define tolerations the Volume has. Only StoragePools whose taints
	// covered by Tolerations will be considered to host the Volume.
	Tolerations []commonv1alpha1.Toleration `json:"tolerations,omitempty"`
}

// VolumeAccess represents information on how to access a volume.
type VolumeAccess struct {
	// SecretRef references the Secret containing the access credentials to consume a Volume.
	SecretRef corev1.LocalObjectReference `json:"secretRef,omitempty"`
	// Driver is the name of the drive to use for this volume. Required.
	Driver string `json:"driver"`
	// VolumeAttributes are attributes of the volume to use.
	VolumeAttributes map[string]string `json:"volumeAttributes,omitempty"`
}

// ClaimReference points to a referenced VolumeClaim.
type ClaimReference struct {
	// Name is the name of the referenced VolumeClaim.
	Name string `json:"name"`
	// UID is the UID of the referenced VolumeClaim.
	UID types.UID `json:"uid"`
}

// VolumeStatus defines the observed state of Volume
type VolumeStatus struct {
	// State represents the infrastructure state of a Volume.
	State VolumeState `json:"state,omitempty"`
	// Phase represents the VolumeClaim binding phase of a Volume.
	Phase VolumePhase `json:"phase,omitempty"`
	// Conditions represents different status aspects of a Volume.
	Conditions []VolumeCondition `json:"conditions,omitempty"`
	// Access specifies how to access a Volume.
	// This is set by the volume provider when the volume is provisioned.
	Access *VolumeAccess `json:"access,omitempty"`
}

// VolumePhase represents the VolumeClaim binding phase of a Volume
// +kubebuilder:validation:Enum=Pending;Available;Bound;Failed
type VolumePhase string

const (
	// VolumePending is used for Volumes that are not available.
	VolumePending VolumePhase = "Pending"
	// VolumeAvailable is used for Volumes that are not yet bound
	// Available volumes are held by the binder and matched to VolumeClaims.
	VolumeAvailable VolumePhase = "Available"
	// VolumeBound is used for Volumes that are bound.
	VolumeBound VolumePhase = "Bound"
	// VolumeFailed is used for Volumes that failed to be correctly freed from a VolumeClaim.
	VolumeFailed VolumePhase = "Failed"
)

// VolumeState is a possible state a volume can be in.
type VolumeState string

const (
	// VolumeStateAvailable reports whether the volume is available to be used.
	VolumeStateAvailable VolumeState = "Available"
	// VolumeStatePending reports whether the volume is about to be ready.
	VolumeStatePending VolumeState = "Pending"
	// VolumeStateError reports that the volume is in an error state.
	VolumeStateError VolumeState = "Error"
)

// VolumeConditionType is a type a VolumeCondition can have.
type VolumeConditionType string

const (
	// VolumeSynced represents the condition of a volume being synced with its backing resources
	VolumeSynced VolumeConditionType = "Synced"
)

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

// +k8s:deepcopy-gen:interfaces=k8s.io/apimachinery/pkg/runtime.Object
// +genclient

// Volume is the Schema for the volumes API
type Volume struct {
	metav1.TypeMeta   `json:",inline"`
	metav1.ObjectMeta `json:"metadata,omitempty"`

	Spec   VolumeSpec   `json:"spec,omitempty"`
	Status VolumeStatus `json:"status,omitempty"`
}

// +k8s:deepcopy-gen:interfaces=k8s.io/apimachinery/pkg/runtime.Object

// VolumeList contains a list of Volume
type VolumeList struct {
	metav1.TypeMeta `json:",inline"`
	metav1.ListMeta `json:"metadata,omitempty"`
	Items           []Volume `json:"items"`
}
