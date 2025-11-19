// SPDX-FileCopyrightText: 2023 SAP SE or an SAP affiliate company and IronCore contributors
// SPDX-License-Identifier: Apache-2.0

package compute

import (
	"github.com/ironcore-dev/ironcore/internal/apis/core"
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"

	commonv1alpha1 "github.com/ironcore-dev/ironcore/api/common/v1alpha1"
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
	Addresses               []MachinePoolAddress
	DaemonEndpoints         MachinePoolDaemonEndpoints
	// Capacity represents the total resources of a machine pool.
	Capacity core.ResourceList
	// Allocatable represents the resources of a machine pool that are available for scheduling.
	Allocatable core.ResourceList
}

// MachinePoolDaemonEndpoints lists ports opened by daemons running on the MachinePool.
type MachinePoolDaemonEndpoints struct {
	// Endpoint on which machinepoollet is listening.
	// +optional
	MachinepoolletEndpoint DaemonEndpoint
}

// DaemonEndpoint contains information about a single Daemon endpoint.
type DaemonEndpoint struct {
	// Port number of the given endpoint.
	Port int32
}

type MachinePoolAddressType string

const (
	// MachinePoolHostName identifies a name of the machine pool. Although every machine pool can be assumed
	// to have a MachinePoolAddress of this type, its exact syntax and semantics are not
	// defined, and are not consistent between different clusters.
	MachinePoolHostName MachinePoolAddressType = "Hostname"

	// MachinePoolInternalIP identifies an IP address which may not be visible to hosts outside the cluster.
	// By default, it is assumed that ironcore-apiserver can reach machine pool internal IPs, though it is possible
	// to configure clusters where this is not the case.
	//
	// MachinePoolInternalIP is the default type of machine pool IP, and does not necessarily imply
	// that the IP is ONLY reachable internally. If a machine pool has multiple internal IPs,
	// no specific semantics are assigned to the additional IPs.
	MachinePoolInternalIP MachinePoolAddressType = "InternalIP"

	// MachinePoolExternalIP identifies an IP address which is, in some way, intended to be more usable from outside
	// the cluster than an internal IP, though no specific semantics are defined.
	MachinePoolExternalIP MachinePoolAddressType = "ExternalIP"

	// MachinePoolInternalDNS identifies a DNS name which resolves to an IP address which has
	// the characteristics of a MachinePoolInternalIP. The IP it resolves to may or may not
	// be a listed MachinePoolInternalIP address.
	MachinePoolInternalDNS MachinePoolAddressType = "InternalDNS"

	// MachinePoolExternalDNS identifies a DNS name which resolves to an IP address which has the characteristics
	// of MachinePoolExternalIP. The IP it resolves to may or may not be a listed MachineExternalIP address.
	MachinePoolExternalDNS MachinePoolAddressType = "ExternalDNS"
)

type MachinePoolAddress struct {
	Type    MachinePoolAddressType
	Address string
}

// MachinePoolConditionType is a type a MachinePoolCondition can have.
type MachinePoolConditionType string

// MachinePoolCondition is one of the conditions of a MachinePool.
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
// +enum
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
