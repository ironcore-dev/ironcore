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

	commonv1alpha1 "github.com/onmetal/onmetal-api/apis/common/v1alpha1"
)

// VolumePoolSpec defines the desired state of VolumePool
type VolumePoolSpec struct {
	// ProviderID identifies the VolumePool on provider side.
	ProviderID string
	// Taints of the VolumePool. Only Volumes who tolerate all the taints
	// will land in the VolumePool.
	Taints []commonv1alpha1.Taint
}

// VolumePoolStatus defines the observed state of VolumePool
type VolumePoolStatus struct {
	State      VolumePoolState
	Conditions []VolumePoolCondition
	// AvailableVolumeClasses list the references of any supported VolumeClass of this pool
	AvailableVolumeClasses []corev1.LocalObjectReference
	// Available list the available capacity of a volume pool
	Available corev1.ResourceList
	// Used indicates how much capacity has been used in a volume pool
	Used corev1.ResourceList
}

type VolumePoolState string

const (
	VolumePoolStateAvailable    VolumePoolState = "Available"
	VolumePoolStatePending      VolumePoolState = "Pending"
	VolumePoolStateNotAvailable VolumePoolState = "NotAvailable"
)

// VolumePoolConditionType is a type a VolumePoolCondition can have.
type VolumePoolConditionType string

// VolumePoolCondition is one of the conditions of a volume.
type VolumePoolCondition struct {
	// Type is the type of the condition.
	Type VolumePoolConditionType
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

// VolumePool is the Schema for the volumepools API
type VolumePool struct {
	metav1.TypeMeta
	metav1.ObjectMeta

	Spec   VolumePoolSpec
	Status VolumePoolStatus
}

// +k8s:deepcopy-gen:interfaces=k8s.io/apimachinery/pkg/runtime.Object

// VolumePoolList contains a list of VolumePool
type VolumePoolList struct {
	metav1.TypeMeta
	metav1.ListMeta
	Items []VolumePool
}
