// SPDX-FileCopyrightText: 2023 SAP SE or an SAP affiliate company and IronCore contributors
// SPDX-License-Identifier: Apache-2.0

package v1alpha1

import (
	commonv1alpha1 "github.com/ironcore-dev/ironcore/api/common/v1alpha1"
	corev1alpha1 "github.com/ironcore-dev/ironcore/api/core/v1alpha1"
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

// VolumeEncryption represents information to encrypt a volume.
type VolumeEncryption struct {
	// SecretRef references the Secret containing the encryption key to encrypt a Volume.
	// This secret is created by user with encryptionKey as Key and base64 encoded 256-bit encryption key as Value.
	SecretRef corev1.LocalObjectReference `json:"secretRef"`
}

// VolumeSpec defines the desired state of Volume
type VolumeSpec struct {
	// VolumeClassRef is the VolumeClass of a volume
	// If empty, an external controller has to provision the volume.
	VolumeClassRef *corev1.LocalObjectReference `json:"volumeClassRef,omitempty"`
	// VolumePoolSelector selects a suitable VolumePoolRef by the given labels.
	VolumePoolSelector map[string]string `json:"volumePoolSelector,omitempty"`
	// VolumePoolRef indicates which VolumePool to use for a volume.
	// If unset, the scheduler will figure out a suitable VolumePoolRef.
	VolumePoolRef *corev1.LocalObjectReference `json:"volumePoolRef,omitempty"`
	// ClaimRef is the reference to the claiming entity of the Volume.
	ClaimRef *commonv1alpha1.LocalUIDReference `json:"claimRef,omitempty"`
	// Resources is a description of the volume's resources and capacity.
	Resources corev1alpha1.ResourceList `json:"resources,omitempty"`
	// Image is an optional image to bootstrap the volume with.
	// To be deprecated
	Image string `json:"image,omitempty"`
	// ImagePullSecretRef is an optional secret for pulling the image of a volume.
	ImagePullSecretRef *corev1.LocalObjectReference `json:"imagePullSecretRef,omitempty"`
	// Unclaimable marks the volume as unclaimable.
	Unclaimable bool `json:"unclaimable,omitempty"`
	// Tolerations define tolerations the Volume has. Only any VolumePool whose taints
	// covered by Tolerations will be considered to host the Volume.
	Tolerations []commonv1alpha1.Toleration `json:"tolerations,omitempty"`
	// Encryption is an optional field which provides attributes to encrypt Volume.
	Encryption *VolumeEncryption `json:"encryption,omitempty"`
	// VolumeDataSource is the source where the storage for the Volume resides at.
	VolumeDataSource `json:",inline"`
}

// VolumeDataSource specifies the source to use for a Volume.
type VolumeDataSource struct {
	// VolumeSnapshotRef instructs to use the specified VolumeSnapshot as the data source.
	VolumeSnapshotRef *corev1.LocalObjectReference `json:"volumeSnapshotRef,omitempty"`
	// OSImage is an optional os image to bootstrap the volume.
	OSImage *string `json:"osImage,omitempty"`
}

// VolumeAccess represents information on how to access a volume.
type VolumeAccess struct {
	// SecretRef references the Secret containing the access credentials to consume a Volume.
	SecretRef *corev1.LocalObjectReference `json:"secretRef,omitempty"`
	// Driver is the name of the drive to use for this volume. Required.
	Driver string `json:"driver"`
	// Handle is the unique handle of the volume.
	Handle string `json:"handle"`
	// VolumeAttributes are attributes of the volume to use.
	VolumeAttributes map[string]string `json:"volumeAttributes,omitempty"`
}

// VolumeStatus defines the observed state of Volume
type VolumeStatus struct {
	// VolumeID is the provider-specific volume ID in the format 'TYPE://VOLUME_ID'.
	VolumeID string `json:"volumeID,omitempty"`
	// State represents the infrastructure state of a Volume.
	State VolumeState `json:"state,omitempty"`
	// LastStateTransitionTime is the last time the State transitioned between values.
	LastStateTransitionTime *metav1.Time `json:"lastStateTransitionTime,omitempty"`

	// Access specifies how to access a Volume.
	// This is set by the volume provider when the volume is provisioned.
	Access *VolumeAccess `json:"access,omitempty"`

	// Conditions are the conditions of a volume.
	Conditions []VolumeCondition `json:"conditions,omitempty"`

	// Resources is a effective volume's resources.
	Resources corev1alpha1.ResourceList `json:"resources,omitempty"`
}

// VolumeConditionType is a type a VolumeCondition can have.
type VolumeConditionType string

// VolumeCondition is one of the conditions of a volume.
type VolumeCondition struct {
	// Type is the type of the condition.
	Type VolumeConditionType `json:"type"`
	// Status is the status of the condition.
	Status corev1.ConditionStatus `json:"status"`
	// Reason is a machine-readable indication of why the condition is in a certain state.
	Reason string `json:"reason,omitempty"`
	// Message is a human-readable explanation of why the condition has a certain reason / state.
	Message string `json:"message,omitempty"`
	// LastTransitionTime is the last time the status of a condition has transitioned from one state to another.
	LastTransitionTime metav1.Time `json:"lastTransitionTime,omitempty"`
}

// VolumeState represents the infrastructure state of a Volume.
type VolumeState string

const (
	// VolumeStatePending reports whether a Volume is about to be ready.
	VolumeStatePending VolumeState = "Pending"
	// VolumeStateAvailable reports whether a Volume is available to be used.
	VolumeStateAvailable VolumeState = "Available"
	// VolumeStateError reports that a Volume is in an error state.
	VolumeStateError VolumeState = "Error"
)

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

// VolumeTemplateSpec is the specification of a Volume template.
type VolumeTemplateSpec struct {
	metav1.ObjectMeta `json:"metadata,omitempty"`
	Spec              VolumeSpec `json:"spec,omitempty"`
}
