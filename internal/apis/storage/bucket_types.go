/*
 * Copyright (c) 2021 by the IronCore authors.
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
	commonv1alpha1 "github.com/ironcore-dev/ironcore/api/common/v1alpha1"
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

// BucketSpec defines the desired state of Bucket
type BucketSpec struct {
	// BucketClassRef is the BucketClass of a bucket
	// If empty, an external controller has to provision the bucket.
	BucketClassRef *corev1.LocalObjectReference
	// BucketPoolSelector selects a suitable BucketPoolRef by the given labels.
	BucketPoolSelector map[string]string
	// BucketPoolRef indicates which BucketPool to use for a bucket.
	// If unset, the scheduler will figure out a suitable BucketPoolRef.
	BucketPoolRef *corev1.LocalObjectReference
	// Tolerations define tolerations the Bucket has. Only any BucketPool whose taints
	// covered by Tolerations will be considered to host the Bucket.
	Tolerations []commonv1alpha1.Toleration
}

// BucketAccess represents information on how to access a bucket.
type BucketAccess struct {
	// SecretRef references the Secret containing the access credentials to consume a Bucket.
	SecretRef *corev1.LocalObjectReference
	// Endpoint defines address of the Bucket REST-API.
	Endpoint string
}

// BucketStatus defines the observed state of Bucket
type BucketStatus struct {
	// State represents the infrastructure state of a Bucket.
	State BucketState
	// LastStateTransitionTime is the last time the State transitioned between values.
	LastStateTransitionTime *metav1.Time

	// Access specifies how to access a Bucket.
	// This is set by the bucket provider when the bucket is provisioned.
	Access *BucketAccess

	// Conditions are the conditions of a bucket.
	Conditions []BucketCondition
}

// BucketConditionType is a type a BucketCondition can have.
type BucketConditionType string

// BucketCondition is one of the conditions of a bucket.
type BucketCondition struct {
	// Type is the type of the condition.
	Type BucketConditionType
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
	metav1.TypeMeta
	metav1.ObjectMeta

	Spec   BucketSpec
	Status BucketStatus
}

// +k8s:deepcopy-gen:interfaces=k8s.io/apimachinery/pkg/runtime.Object

// BucketList contains a list of Bucket
type BucketList struct {
	metav1.TypeMeta
	metav1.ListMeta
	Items []Bucket
}

// BucketTemplateSpec is the specification of a Bucket template.
type BucketTemplateSpec struct {
	metav1.ObjectMeta
	Spec BucketSpec
}
