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
	State BucketPoolState `json:"state,omitempty"`
	// AvailableBucketClasses list the references of any supported BucketClass of this pool
	AvailableBucketClasses []corev1.LocalObjectReference `json:"availableBucketClasses,omitempty"`
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
