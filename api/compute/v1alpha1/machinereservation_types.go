// SPDX-FileCopyrightText: 2023 SAP SE or an SAP affiliate company and IronCore contributors
// SPDX-License-Identifier: Apache-2.0

package v1alpha1

import (
	corev1alpha1 "github.com/ironcore-dev/ironcore/api/core/v1alpha1"
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

// ReservationSpec defines the desired state of Reservation
type ReservationSpec struct {
	Pools     []corev1.LocalObjectReference `json:"pools"`
	Resources corev1alpha1.ResourceList     `json:"resources,omitempty"`
}

// ReservationStatus defines the observed state of Reservation
type ReservationStatus struct {
	Pools      []ReservationPoolStatus `json:"pools,omitempty"`
	Conditions []ReservationCondition  `json:"conditions,omitempty"`
}

// ReservationState is the state of a Reservation.
// +enum
type ReservationState string

const (
	// ReservationStatePending means the Reservation is being reconciled.
	ReservationStatePending ReservationState = "Pending"
	// ReservationStateAccepted means the pool accepted the reservation and reserved the requested resources.
	ReservationStateAccepted ReservationState = "Accepted"
	// ReservationStateRejected means the pool rejected the reservation.
	ReservationStateRejected ReservationState = "Rejected"
)

// ReservationConditionType is a type a ReservationCondition can have.
type ReservationConditionType string

// ReservationCondition is one of the conditions of a volume.
type ReservationCondition struct {
	// Type is the type of the condition.
	Type ReservationConditionType `json:"type"`
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

type ReservationPoolStatus struct {
	Name  string           `json:"ref,omitempty"`
	State ReservationState `json:"state,omitempty"`
}

// +genclient
// +k8s:deepcopy-gen:interfaces=k8s.io/apimachinery/pkg/runtime.Object

// Reservation is the Schema for the machines API
type Reservation struct {
	metav1.TypeMeta   `json:",inline"`
	metav1.ObjectMeta `json:"metadata,omitempty"`

	Spec   ReservationSpec   `json:"spec,omitempty"`
	Status ReservationStatus `json:"status,omitempty"`
}

// +k8s:deepcopy-gen:interfaces=k8s.io/apimachinery/pkg/runtime.Object

// ReservationList contains a list of Reservation
type ReservationList struct {
	metav1.TypeMeta `json:",inline"`
	metav1.ListMeta `json:"metadata,omitempty"`
	Items           []Reservation `json:"items"`
}
