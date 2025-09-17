// SPDX-FileCopyrightText: 2023 SAP SE or an SAP affiliate company and IronCore contributors
// SPDX-License-Identifier: Apache-2.0

package compute

import (
	"github.com/ironcore-dev/ironcore/internal/apis/core"
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

// ReservationSpec defines the desired state of Reservation
type ReservationSpec struct {
	Pools     []corev1.LocalObjectReference
	Resources core.ResourceList
}

// ReservationStatus defines the observed state of Reservation
type ReservationStatus struct {
	Pools      []ReservationPoolStatus
	Conditions []ReservationCondition
}

// ReservationState is the state of a Reservation.
// +enum
type ReservationState string

const (
	// ReservationStatePending means the Reservation is being reconciled.
	ReservationStatePending MachineState = "Pending"
	// ReservationStateAccepted means the pool accepted the reservation and reserved the requested resources.
	ReservationStateAccepted MachineState = "Accepted"
	// ReservationStateRejected means the pool rejected the reservation.
	ReservationStateRejected MachineState = "Rejected"
)

type ReservationPoolStatus struct {
	Name  string
	State ReservationState
}

// ReservationConditionType is a type a ReservationCondition can have.
type ReservationConditionType string

// ReservationCondition is one of the conditions of a volume.
type ReservationCondition struct {
	// Type is the type of the condition.
	Type ReservationConditionType
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

// +genclient
// +k8s:deepcopy-gen:interfaces=k8s.io/apimachinery/pkg/runtime.Object

// Reservation is the Schema for the machines API
type Reservation struct {
	metav1.TypeMeta
	metav1.ObjectMeta

	Spec   ReservationSpec
	Status ReservationStatus
}

// +k8s:deepcopy-gen:interfaces=k8s.io/apimachinery/pkg/runtime.Object

// ReservationList contains a list of Reservation
type ReservationList struct {
	metav1.TypeMeta
	metav1.ListMeta
	Items []Reservation
}
