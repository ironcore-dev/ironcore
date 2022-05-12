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

package compute

import (
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"

	commonv1alpha1 "github.com/onmetal/onmetal-api/apis/common/v1alpha1"
)

// MachinePoolSpec defines the desired state of MachinePool
type MachinePoolSpec struct {
	// ProviderID identifies the MachinePool on provider side.
	ProviderID string
	// Taints of the MachinePool. Only Machines who tolerate all the taints
	// will land in the MachinePool.
	Taints []commonv1alpha1.Taint
}

// MachinePoolStatus defines the observed state of MachinePool
type MachinePoolStatus struct {
	State                   MachinePoolState
	Conditions              []MachinePoolCondition
	AvailableMachineClasses []corev1.LocalObjectReference
}

// MachinePoolConditionType is a type a MachinePoolCondition can have.
type MachinePoolConditionType string

// MachinePoolCondition is one of the conditions of a volume.
type MachinePoolCondition struct {
	// Type is the type of the condition.
	Type MachinePoolConditionType
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

// MachinePoolState is a state a MachinePool can be in.
//+enum
type MachinePoolState string

const (
	// MachinePoolStateReady marks a MachinePool as ready for accepting a Machine.
	MachinePoolStateReady MachinePoolState = "Ready"
	// MachinePoolStatePending marks a MachinePool as pending readiness.
	MachinePoolStatePending MachinePoolState = "Pending"
	// MachinePoolStateError marks a MachinePool in an error state.
	MachinePoolStateError MachinePoolState = "Error"
	// MachinePoolStateOffline marks a MachinePool as offline.
	MachinePoolStateOffline MachinePoolState = "Offline"
)

// +k8s:deepcopy-gen:interfaces=k8s.io/apimachinery/pkg/runtime.Object
// +genclient
// +genclient:nonNamespaced

// MachinePool is the Schema for the machinepools API
type MachinePool struct {
	metav1.TypeMeta
	metav1.ObjectMeta

	Spec   MachinePoolSpec
	Status MachinePoolStatus
}

// +k8s:deepcopy-gen:interfaces=k8s.io/apimachinery/pkg/runtime.Object

// MachinePoolList contains a list of MachinePool
type MachinePoolList struct {
	metav1.TypeMeta
	metav1.ListMeta
	Items []MachinePool
}
