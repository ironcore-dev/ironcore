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

package storage

import (
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"

	commonv1alpha1 "github.com/onmetal/onmetal-api/api/common/v1alpha1"
)

// BucketPoolSpec defines the desired state of BucketPool
type BucketPoolSpec struct {
	// ProviderID identifies the BucketPool on provider side.
	ProviderID string
	// Taints of the BucketPool. Only Buckets who tolerate all the taints
	// will land in the BucketPool.
	Taints []commonv1alpha1.Taint
}

// BucketPoolStatus defines the observed state of BucketPool
type BucketPoolStatus struct {
	// State represents the infrastructure state of a BucketPool.
	State BucketPoolState
	// AvailableBucketClasses list the references of any supported BucketClass of this pool
	AvailableBucketClasses []corev1.LocalObjectReference
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
	metav1.TypeMeta
	metav1.ObjectMeta

	Spec   BucketPoolSpec
	Status BucketPoolStatus
}

// +k8s:deepcopy-gen:interfaces=k8s.io/apimachinery/pkg/runtime.Object

// BucketPoolList contains a list of BucketPool
type BucketPoolList struct {
	metav1.TypeMeta
	metav1.ListMeta
	Items []BucketPool
}
