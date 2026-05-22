// SPDX-FileCopyrightText: 2023 SAP SE or an SAP affiliate company and IronCore contributors
// SPDX-License-Identifier: Apache-2.0

package v1alpha1

import (
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"

	commonv1alpha1 "github.com/ironcore-dev/ironcore/api/common/v1alpha1"
)

// BucketPoolSpec defines the desired state of BucketPool
type BucketPoolSpec struct {
	// ProviderID identifies the BucketPool on provider side.
	ProviderID string `json:"providerID"`
	// Taints of the BucketPool. Only Buckets who tolerate all the taints
	// will land in the BucketPool.
	Taints []commonv1alpha1.Taint `json:"taints,omitempty"`
}

// BucketPoolStatus defines the observed state of BucketPool
type BucketPoolStatus struct {
	// State represents the infrastructure state of a BucketPool.
	State      BucketPoolState       `json:"state,omitempty"`
	Conditions []BucketPoolCondition `json:"conditions,omitempty"`
	// AvailableBucketClasses list the references of any supported BucketClass of this pool
	AvailableBucketClasses []corev1.LocalObjectReference `json:"availableBucketClasses,omitempty"`
}

// BucketPoolConditionType is a type a BucketPoolCondition can have.
type BucketPoolConditionType string

const (
	// BucketPoolReady means the bucket pool is healthy and ready to accept buckets.
	BucketPoolReady BucketPoolConditionType = "Ready"
)

// BucketPoolCondition is one of the conditions of a BucketPool.
type BucketPoolCondition struct {
	// Type is the type of the condition.
	Type BucketPoolConditionType `json:"type"`
	// Status is the status of the condition.
	Status corev1.ConditionStatus `json:"status"`
	// Reason is a machine-readable indication of why the condition is in a certain state.
	Reason string `json:"reason"`
	// Message is a human-readable explanation of why the condition has a certain reason / state.
	Message string `json:"message"`
	// ObservedGeneration represents the .metadata.generation that the condition was set based upon.
	ObservedGeneration int64 `json:"observedGeneration,omitempty"`
	// LastUpdateTime is the last time this condition was updated.
	LastUpdateTime metav1.Time `json:"lastUpdateTime,omitempty"`
	// LastTransitionTime is the last time the status of a condition has transitioned from one state to another.
	LastTransitionTime metav1.Time `json:"lastTransitionTime,omitempty"`
}

type BucketPoolState string

const (
	BucketPoolStateAvailable   BucketPoolState = "Available"
	BucketPoolStatePending     BucketPoolState = "Pending"
	BucketPoolStateUnavailable BucketPoolState = "Unavailable"
)

// +k8s:deepcopy-gen:interfaces=k8s.io/apimachinery/pkg/runtime.Object
// +genclient
// +genclient:nonNamespaced

// BucketPool is the Schema for the bucketpools API
type BucketPool struct {
	metav1.TypeMeta   `json:",inline"`
	metav1.ObjectMeta `json:"metadata,omitempty"`

	Spec   BucketPoolSpec   `json:"spec,omitempty"`
	Status BucketPoolStatus `json:"status,omitempty"`
}

// +k8s:deepcopy-gen:interfaces=k8s.io/apimachinery/pkg/runtime.Object

// BucketPoolList contains a list of BucketPool
type BucketPoolList struct {
	metav1.TypeMeta `json:",inline"`
	metav1.ListMeta `json:"metadata,omitempty"`
	Items           []BucketPool `json:"items"`
}
