// SPDX-FileCopyrightText: 2023 SAP SE or an SAP affiliate company and IronCore contributors
// SPDX-License-Identifier: Apache-2.0

package storage

import (
	commonv1alpha1 "github.com/ironcore-dev/ironcore/api/common/v1alpha1"
	"github.com/ironcore-dev/ironcore/internal/apis/core"
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

// VolumeEncryption represents information to encrypt a volume.
type VolumeEncryption struct {
	// SecretRef references the Secret containing the encryption key to encrypt a Volume.
	// This secret is created by user with encryptionKey as Key and base64 encoded 256-bit encryption key as Value.
	SecretRef corev1.LocalObjectReference
}

// VolumeSpec defines the desired state of Volume
type VolumeSpec struct {
	// VolumeClassRef is the volume class of a volume
	VolumeClassRef *corev1.LocalObjectReference
	// VolumePoolSelector selects a suitable VolumePoolRef by the given labels.
	VolumePoolSelector map[string]string
	// VolumePoolRef indicates which volume pool to use for a volume.
	// If unset, the scheduler will figure out a suitable VolumePoolRef.
	VolumePoolRef *corev1.LocalObjectReference
	// ClaimRef is the reference to the claiming entity of the Volume.
	ClaimRef *commonv1alpha1.LocalUIDReference
	// Resources is a description of the volume's resources and capacity.
	Resources core.ResourceList
	// Image is an optional image to bootstrap the volume with.
	Image string
	// ImagePullSecretRef is an optional secret for pulling the image of a volume.
	ImagePullSecretRef *corev1.LocalObjectReference
	// Unclaimable marks the volume as unclaimable.
	Unclaimable bool
	// Tolerations define tolerations the Volume has. Only a VolumePool whose taints
	// covered by Tolerations will be considered to host the Volume.
	Tolerations []commonv1alpha1.Toleration
	// Encryption is an optional field which provides attributes to encrypt Volume.
	Encryption *VolumeEncryption
}

// VolumeAccess represents information on how to access a volume.
type VolumeAccess struct {
	// SecretRef references the Secret containing the access credentials to consume a Volume.
	SecretRef *corev1.LocalObjectReference
	// Driver is the name of the drive to use for this volume. Required.
	Driver string
	// Handle is the unique handle of the volume.
	Handle string
	// VolumeAttributes are attributes of the volume to use.
	VolumeAttributes map[string]string
}

// VolumeStatus defines the observed state of Volume
type VolumeStatus struct {
	// State represents the infrastructure state of a Volume.
	State VolumeState
	// LastStateTransitionTime is the last time the State transitioned between values.
	LastStateTransitionTime *metav1.Time

	// Access specifies how to access a Volume.
	// This is set by the volume provider when the volume is provisioned.
	Access *VolumeAccess

	// Conditions are the conditions of a volume.
	Conditions []VolumeCondition
}

// VolumeConditionType is a type a VolumeCondition can have.
type VolumeConditionType string

// VolumeCondition is one of the conditions of a volume.
type VolumeCondition struct {
	// Type is the type of the condition.
	Type VolumeConditionType
	// Status is the status of the condition.
	Status corev1.ConditionStatus
	// Reason is a machine-readable indication of why the condition is in a certain state.
	Reason string
	// Message is a human-readable explanation of why the condition has a certain reason / state.
	Message string
	// ObservedGeneration represents the .metadata.generation that the condition was set based upon.
	ObservedGeneration int64
	// LastTransitionTime is the last time the status of a condition has transitioned from one state to another.
	LastTransitionTime metav1.Time
}

// VolumeState represents the infrastructure state of a Volume.
type VolumeState string

const (
	// VolumeStateAvailable reports whether the volume is available to be used.
	VolumeStateAvailable VolumeState = "Available"
	// VolumeStatePending reports whether the volume is about to be ready.
	VolumeStatePending VolumeState = "Pending"
	// VolumeStateError reports that the volume is in an error state.
	VolumeStateError VolumeState = "Error"
)

// +k8s:deepcopy-gen:interfaces=k8s.io/apimachinery/pkg/runtime.Object
// +genclient

// Volume is the Schema for the volumes API
type Volume struct {
	metav1.TypeMeta
	metav1.ObjectMeta

	Spec   VolumeSpec
	Status VolumeStatus
}

// +k8s:deepcopy-gen:interfaces=k8s.io/apimachinery/pkg/runtime.Object

// VolumeList contains a list of Volume
type VolumeList struct {
	metav1.TypeMeta
	metav1.ListMeta
	Items []Volume
}

// VolumeTemplateSpec is the specification of a Volume template.
type VolumeTemplateSpec struct {
	metav1.ObjectMeta
	Spec VolumeSpec
}
