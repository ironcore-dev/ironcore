// SPDX-FileCopyrightText: 2023 SAP SE or an SAP affiliate company and IronCore contributors
// SPDX-License-Identifier: Apache-2.0

package networking

import (
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/types"
)

// NetworkSpec defines the desired state of Network
type NetworkSpec struct {
	// ProviderID is the provider-internal ID of the network.
	ProviderID string
	// Peerings are the network peerings with this network.
	// +optional
	// +patchMergeKey=name
	// +patchStrategy=merge,retainKeys
	Peerings []NetworkPeering

	// PeeringClaimRefs are the peering claim references of other networks.
	// +optional
	// +patchMergeKey=name
	// +patchStrategy=merge,retainKeys
	PeeringClaimRefs []NetworkPeeringClaimRef
}

type NetworkPeeringClaimRef struct {
	// Namespace is the namespace of the referenced entity. If empty,
	// the same namespace as the referring resource is implied.
	Namespace string
	// Name is the name of the referenced entity.
	Name string
	// UID is the UID of the referenced entity.
	UID types.UID
}

// NetworkPeeringNetworkRef is a reference to a network to peer with.
type NetworkPeeringNetworkRef struct {
	// Namespace is the namespace of the referenced entity. If empty,
	// the same namespace as the referring resource is implied.
	Namespace string
	// Name is the name of the referenced entity.
	Name string
}

// NetworkPeering defines a network peering with another network.
type NetworkPeering struct {
	// Name is the semantical name of the network peering.
	Name string
	// NetworkRef is the reference to the network to peer with.
	// An empty namespace indicates that the target network resides in the same namespace as the source network.
	NetworkRef NetworkPeeringNetworkRef
}

// NetworkStatus defines the observed state of Network
type NetworkStatus struct {
	// State is the state of the machine.
	State NetworkState
	// Peerings contains the states of the network peerings for the network.
	// +optional
	// +patchMergeKey=name
	// +patchStrategy=merge,retainKeys
	Peerings []NetworkPeeringStatus
}

// NetworkState is the state of a network.
// +enum
type NetworkState string

// NetworkPeeringStatus is the status of a network peering.
type NetworkPeeringStatus struct {
	// Name is the name of the network peering.
	Name string
}

const (
	// NetworkStatePending means the network is being provisioned.
	NetworkStatePending NetworkState = "Pending"
	// NetworkStateAvailable means the network is ready to use.
	NetworkStateAvailable NetworkState = "Available"
	// NetworkStateError means the network is in an error state.
	NetworkStateError NetworkState = "Error"
)

// +genclient
// +k8s:deepcopy-gen:interfaces=k8s.io/apimachinery/pkg/runtime.Object

// Network is the Schema for the network API
type Network struct {
	metav1.TypeMeta
	metav1.ObjectMeta

	Spec   NetworkSpec
	Status NetworkStatus
}

// +k8s:deepcopy-gen:interfaces=k8s.io/apimachinery/pkg/runtime.Object

// NetworkList contains a list of Network
type NetworkList struct {
	metav1.TypeMeta
	metav1.ListMeta
	Items []Network
}
