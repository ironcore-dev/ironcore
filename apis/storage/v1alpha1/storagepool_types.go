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

package v1alpha1

import (
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"

	commonv1alpha1 "github.com/onmetal/onmetal-api/apis/common/v1alpha1"
)

// StoragePoolSpec defines the desired state of StoragePool
type StoragePoolSpec struct {
	// ProviderID identifies the StoragePool on provider side.
	ProviderID string `json:"providerID"`
	// Taints define a list of taints to be added for StoragePool.
	Taints []commonv1alpha1.Taint `json:"taints,omitempty"`
}

// StoragePoolStatus defines the observed state of StoragePool
type StoragePoolStatus struct {
	State      StoragePoolState       `json:"state,omitempty"`
	Conditions []StoragePoolCondition `json:"conditions,omitempty"`
	// AvailableStorageClasses list the references of supported StorageClasses of this pool
	AvailableStorageClasses []corev1.LocalObjectReference `json:"availableStorageClasses,omitempty"`
	// Available list the available capacity of a storage pool
	Available corev1.ResourceList `json:"available,omitempty"`
	// Used indicates how much capacity has been used in a storage pool
	Used corev1.ResourceList `json:"used,omitempty"`
}

type StoragePoolState string

const (
	StoragePoolStateAvailable    StoragePoolState = "Available"
	StoragePoolStatePending      StoragePoolState = "Pending"
	StoragePoolStateNotAvailable StoragePoolState = "NotAvailable"
)

// StoragePoolConditionType is a type a StoragePoolCondition can have.
type StoragePoolConditionType string

// StoragePoolCondition is one of the conditions of a volume.
type StoragePoolCondition struct {
	// Type is the type of the condition.
	Type StoragePoolConditionType `json:"type"`
	// Status is the status of the condition.
	Status corev1.ConditionStatus `json:"status"`
	// Reason is a machine-readable indication of why the condition is in a certain state.
	Reason string `json:"reason"`
	// Message is a human-readable explanation of why the condition has a certain reason / state.
	Message string `json:"message"`
	// ObservedGeneration represents the .metadata.generation that the condition was set based upon.
	ObservedGeneration int64 `json:"observedGeneration,omitempty"`
	// LastUpdateTime is the last time a condition has been updated.
	LastUpdateTime metav1.Time `json:"lastUpdateTime,omitempty"`
	// LastTransitionTime is the last time the status of a condition has transitioned from one state to another.
	LastTransitionTime metav1.Time `json:"lastTransitionTime,omitempty"`
}

//+kubebuilder:object:root=true
//+kubebuilder:resource:scope=Cluster
//+kubebuilder:subresource:status
//+kubebuilder:printcolumn:name="State",type=string,JSONPath=`.status.state`
//+kubebuilder:printcolumn:name="Age",type=date,JSONPath=`.metadata.creationTimestamp`

// StoragePool is the Schema for the storagepools API
type StoragePool struct {
	metav1.TypeMeta   `json:",inline"`
	metav1.ObjectMeta `json:"metadata,omitempty"`

	Spec   StoragePoolSpec   `json:"spec,omitempty"`
	Status StoragePoolStatus `json:"status,omitempty"`
}

//+kubebuilder:object:root=true

// StoragePoolList contains a list of StoragePool
type StoragePoolList struct {
	metav1.TypeMeta `json:",inline"`
	metav1.ListMeta `json:"metadata,omitempty"`
	Items           []StoragePool `json:"items"`
}

func init() {
	SchemeBuilder.Register(&StoragePool{}, &StoragePoolList{})
}
