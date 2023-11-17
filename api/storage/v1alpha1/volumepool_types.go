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

package v1alpha1

import (
	corev1alpha1 "github.com/ironcore-dev/ironcore/api/core/v1alpha1"
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"

	commonv1alpha1 "github.com/ironcore-dev/ironcore/api/common/v1alpha1"
)

// VolumePoolSpec defines the desired state of VolumePool
type VolumePoolSpec struct {
	// ProviderID identifies the VolumePool on provider side.
	ProviderID string `json:"providerID"`
	// Taints of the VolumePool. Only Volumes who tolerate all the taints
	// will land in the VolumePool.
	Taints []commonv1alpha1.Taint `json:"taints,omitempty"`
}

// VolumePoolStatus defines the observed state of VolumePool
type VolumePoolStatus struct {
	State      VolumePoolState       `json:"state,omitempty"`
	Conditions []VolumePoolCondition `json:"conditions,omitempty"`
	// AvailableVolumeClasses list the references of any supported VolumeClass of this pool
	AvailableVolumeClasses []corev1.LocalObjectReference `json:"availableVolumeClasses,omitempty"`
	// Capacity represents the total resources of a machine pool.
	Capacity corev1alpha1.ResourceList `json:"capacity,omitempty"`
	// Allocatable represents the resources of a machine pool that are available for scheduling.
	Allocatable corev1alpha1.ResourceList `json:"allocatable,omitempty"`
}

type VolumePoolState string

const (
	VolumePoolStateAvailable   VolumePoolState = "Available"
	VolumePoolStatePending     VolumePoolState = "Pending"
	VolumePoolStateUnavailable VolumePoolState = "Unavailable"
)

// VolumePoolConditionType is a type a VolumePoolCondition can have.
type VolumePoolConditionType string

// VolumePoolCondition is one of the conditions of a volume.
type VolumePoolCondition struct {
	// Type is the type of the condition.
	Type VolumePoolConditionType `json:"type"`
	// Status is the status of the condition.
	Status corev1.ConditionStatus `json:"status"`
	// Reason is a machine-readable indication of why the condition is in a certain state.
	Reason string `json:"reason"`
	// Message is a human-readable explanation of why the condition has a certain reason / state.
	Message string `json:"message"`
	// ObservedGeneration represents the .metadata.generation that the condition was set based upon.
	ObservedGeneration int64 `json:"observedGeneration,omitempty"`
	// LastTransitionTime is the last time the status of a condition has transitioned from one state to another.
	LastTransitionTime metav1.Time `json:"lastTransitionTime,omitempty"`
}

// +k8s:deepcopy-gen:interfaces=k8s.io/apimachinery/pkg/runtime.Object
// +genclient
// +genclient:nonNamespaced

// VolumePool is the Schema for the volumepools API
type VolumePool struct {
	metav1.TypeMeta   `json:",inline"`
	metav1.ObjectMeta `json:"metadata,omitempty"`

	Spec   VolumePoolSpec   `json:"spec,omitempty"`
	Status VolumePoolStatus `json:"status,omitempty"`
}

// +k8s:deepcopy-gen:interfaces=k8s.io/apimachinery/pkg/runtime.Object

// VolumePoolList contains a list of VolumePool
type VolumePoolList struct {
	metav1.TypeMeta `json:",inline"`
	metav1.ListMeta `json:"metadata,omitempty"`
	Items           []VolumePool `json:"items"`
}
