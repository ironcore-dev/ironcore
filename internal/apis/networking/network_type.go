// SPDX-FileCopyrightText: 2023 SAP SE or an SAP affiliate company and IronCore contributors
// SPDX-License-Identifier: Apache-2.0

package networking

import (
	commonv1alpha1 "github.com/ironcore-dev/ironcore/api/common/v1alpha1"
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
	// Prefixes is a list of prefixes that we want only to be exposed
	// to the peered network, if no prefixes are specified no filtering will be done.
	Prefixes []PeeringPrefix
}

// PeeringPrefixes defines prefixes to be exposed to the peered network
type PeeringPrefix struct {
	// Name is the semantical name of the peering prefixes
	Name string
	// CIDR to be exposed to the peered network
	Prefix *commonv1alpha1.IPPrefix
	// PrefixRef is the reference to the prefix to be exposed to peered network
	// An empty namespace indicates that the prefix resides in the same namespace as the source network.
	PrefixRef PeeringPrefixRef
}

type PeeringPrefixRef struct {
	// Namespace is the namespace of the referenced entity. If empty,
	// the same namespace as the referring resource is implied.
	Namespace string
	// Name is the name of the referenced entity.
	Name string
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

// NetworkPeeringState is the state a NetworkPeering can be in.
type NetworkPeeringState string

const (
	// NetworkPeeringStatePending signals that the network peering is not applied.
	NetworkPeeringStatePending NetworkPeeringState = "Pending"
	// NetworkPeeringStateReady signals that the network peering is ready.
	NetworkPeeringStateReady NetworkPeeringState = "Ready"
	// NetworkPeeringStateError signals that the network peering is in error state.
	NetworkPeeringStateError NetworkPeeringState = "Error"
)

// NetworkPeeringStatus is the status of a network peering.
type NetworkPeeringStatus struct {
	// Name is the name of the network peering.
	Name string
	// State represents the network peering state
	State NetworkPeeringState
	// Prefixes contains the prefixes exposed to the peered network
	Prefixes []PeeringPrefixStatus
}

// PeeringPrefixStatus lists prefixes exposed to peered network
type PeeringPrefixStatus struct {
	// Name is the name of the peering prefix
	Name string
	// CIDR exposed to the peered network
	Prefix *commonv1alpha1.IPPrefix
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
