// SPDX-FileCopyrightText: 2023 SAP SE or an SAP affiliate company and IronCore contributors
// SPDX-License-Identifier: Apache-2.0

package v1alpha1

import (
	commonv1alpha1 "github.com/ironcore-dev/ironcore/api/common/v1alpha1"
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

// BucketSpec defines the desired state of Bucket
type BucketSpec struct {
	// BucketClassRef is the BucketClass of a bucket
	// If empty, an external controller has to provision the bucket.
	BucketClassRef *corev1.LocalObjectReference `json:"bucketClassRef,omitempty"`
	// BucketPoolSelector selects a suitable BucketPoolRef by the given labels.
	BucketPoolSelector map[string]string `json:"bucketPoolSelector,omitempty"`
	// BucketPoolRef indicates which BucketPool to use for a bucket.
	// If unset, the scheduler will figure out a suitable BucketPoolRef.
	BucketPoolRef *corev1.LocalObjectReference `json:"bucketPoolRef,omitempty"`
	// Tolerations define tolerations the Bucket has. Only any BucketPool whose taints
	// covered by Tolerations will be considered to host the Bucket.
	Tolerations []commonv1alpha1.Toleration `json:"tolerations,omitempty"`
}

// BucketAccess represents information on how to access a bucket.
type BucketAccess struct {
	// SecretRef references the Secret containing the access credentials to consume a Bucket.
	SecretRef *corev1.LocalObjectReference `json:"secretRef,omitempty"`
	// Endpoint defines address of the Bucket REST-API.
	Endpoint string `json:"endpoint"`
}

// BucketStatus defines the observed state of Bucket
type BucketStatus struct {
	// State represents the infrastructure state of a Bucket.
	State BucketState `json:"state,omitempty"`
	// LastStateTransitionTime is the last time the State transitioned between values.
	LastStateTransitionTime *metav1.Time `json:"lastStateTransitionTime,omitempty"`

	// Access specifies how to access a Bucket.
	// This is set by the bucket provider when the bucket is provisioned.
	Access *BucketAccess `json:"access,omitempty"`

	// Conditions are the conditions of a bucket.
	Conditions []BucketCondition `json:"conditions,omitempty"`
}

// BucketConditionType is a type a BucketCondition can have.
type BucketConditionType string

// BucketCondition is one of the conditions of a bucket.
type BucketCondition struct {
	// Type is the type of the condition.
	Type BucketConditionType `json:"type"`
	// Status is the status of the condition.
	Status corev1.ConditionStatus `json:"status"`
	// Reason is a machine-readable indication of why the condition is in a certain state.
	Reason string `json:"reason,omitempty"`
	// Message is a human-readable explanation of why the condition has a certain reason / state.
	Message string `json:"message,omitempty"`
	// ObservedGeneration represents the .metadata.generation that the condition was set based upon.
	ObservedGeneration int64 `json:"observedGeneration,omitempty"`
	// LastTransitionTime is the last time the status of a condition has transitioned from one state to another.
	LastTransitionTime metav1.Time `json:"lastTransitionTime,omitempty"`
}

// BucketState represents the infrastructure state of a Bucket.
type BucketState string

const (
	// BucketStatePending reports whether a Bucket is about to be ready.
	BucketStatePending BucketState = "Pending"
	// BucketStateAvailable reports whether a Bucket is available to be used.
	BucketStateAvailable BucketState = "Available"
	// BucketStateError reports that a Bucket is in an error state.
	BucketStateError BucketState = "Error"
)

// +k8s:deepcopy-gen:interfaces=k8s.io/apimachinery/pkg/runtime.Object
// +genclient

// Bucket is the Schema for the buckets API
type Bucket struct {
	metav1.TypeMeta   `json:",inline"`
	metav1.ObjectMeta `json:"metadata,omitempty"`

	Spec   BucketSpec   `json:"spec,omitempty"`
	Status BucketStatus `json:"status,omitempty"`
}

// +k8s:deepcopy-gen:interfaces=k8s.io/apimachinery/pkg/runtime.Object

// BucketList contains a list of Bucket
type BucketList struct {
	metav1.TypeMeta `json:",inline"`
	metav1.ListMeta `json:"metadata,omitempty"`
	Items           []Bucket `json:"items"`
}

// BucketTemplateSpec is the specification of a Bucket template.
type BucketTemplateSpec struct {
	metav1.ObjectMeta `json:"metadata,omitempty"`
	Spec              BucketSpec `json:"spec,omitempty"`
}
