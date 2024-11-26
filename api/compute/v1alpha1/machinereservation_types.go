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
	Resources corev1alpha1.ResourceList     `json:"capabilities,omitempty"`
}

// ReservationStatus defines the observed state of Reservation
type ReservationStatus struct {
	Pools []ReservationPoolStatus `json:"pools,omitempty"`
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
