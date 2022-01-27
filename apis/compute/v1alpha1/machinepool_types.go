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

// MachinePoolSpec defines the desired state of MachinePool
type MachinePoolSpec struct {
	// ProviderID identifies the MachinePool on provider side.
	ProviderID string `json:"providerID"`
	// Taints of the MachinePool. Only Machines who tolerate all the taints
	// will land in the MachinePool.
	Taints []commonv1alpha1.Taint `json:"taints,omitempty"`
}

// MachinePoolStatus defines the observed state of MachinePool
type MachinePoolStatus struct {
	State                   MachinePoolState              `json:"state,omitempty"`
	Conditions              []MachinePoolCondition        `json:"conditions,omitempty"`
	AvailableMachineClasses []corev1.LocalObjectReference `json:"availableMachineClasses,omitempty"`
}

// MachinePoolConditionType is a type a MachinePoolCondition can have.
type MachinePoolConditionType string

// MachinePoolCondition is one of the conditions of a volume.
type MachinePoolCondition struct {
	// Type is the type of the condition.
	Type MachinePoolConditionType `json:"type"`
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

// MachinePoolState is a state a MachinePool can be in.
type MachinePoolState string

const (
	MachinePoolStateReady   MachinePoolState = "Ready"
	MachinePoolStatePending MachinePoolState = "Pending"
	MachinePoolStateError   MachinePoolState = "Error"
	MachinePoolStateOffline MachinePoolState = "Offline"
)

//+kubebuilder:object:root=true
//+kubebuilder:resource:scope=Cluster
//+kubebuilder:subresource:status
//+kubebuilder:printcolumn:name="State",type=string,JSONPath=`.status.state`
//+kubebuilder:printcolumn:name="Age",type=date,JSONPath=`.metadata.creationTimestamp`

// MachinePool is the Schema for the machinepools API
type MachinePool struct {
	metav1.TypeMeta   `json:",inline"`
	metav1.ObjectMeta `json:"metadata,omitempty"`

	Spec   MachinePoolSpec   `json:"spec,omitempty"`
	Status MachinePoolStatus `json:"status,omitempty"`
}

//+kubebuilder:object:root=true

// MachinePoolList contains a list of MachinePool
type MachinePoolList struct {
	metav1.TypeMeta `json:",inline"`
	metav1.ListMeta `json:"metadata,omitempty"`
	Items           []MachinePool `json:"items"`
}

func init() {
	SchemeBuilder.Register(&MachinePool{}, &MachinePoolList{})
}
