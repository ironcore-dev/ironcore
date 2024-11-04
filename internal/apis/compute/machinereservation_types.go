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
	Pools []ReservationPoolStatus
}

// ReservationState is the state of a Reservation.
// +enum
type ReservationState string

const (
	// ReservationStatePending means the Reservation is beeing reconciled.
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
