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
	State      BucketPoolState
	Conditions []BucketPoolCondition
	// AvailableBucketClasses list the references of any supported BucketClass of this pool
	AvailableBucketClasses []corev1.LocalObjectReference
	// Available list the available capacity of a BucketPool.
	Available corev1.ResourceList
	// Used indicates how much capacity has been used in a BucketPool.
	Used corev1.ResourceList
}

type BucketPoolState string

const (
	BucketPoolStateAvailable   BucketPoolState = "Available"
	BucketPoolStatePending     BucketPoolState = "Pending"
	BucketPoolStateUnavailable BucketPoolState = "Unavailable"
)

// BucketPoolConditionType is a type a BucketPoolCondition can have.
type BucketPoolConditionType string

// BucketPoolCondition is one of the conditions of a bucket.
type BucketPoolCondition struct {
	// Type is the type of the condition.
	Type BucketPoolConditionType
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
