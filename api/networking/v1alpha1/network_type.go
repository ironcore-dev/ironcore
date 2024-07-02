// SPDX-FileCopyrightText: 2023 SAP SE or an SAP affiliate company and IronCore contributors
// SPDX-License-Identifier: Apache-2.0

package v1alpha1

import (
	commonv1alpha1 "github.com/ironcore-dev/ironcore/api/common/v1alpha1"
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/types"
)

// NetworkSpec defines the desired state of Network
type NetworkSpec struct {
	// ProviderID is the provider-internal ID of the network.
	ProviderID string `json:"providerID,omitempty"`
	// Peerings are the network peerings with this network.
	// +optional
	// +patchMergeKey=name
	// +patchStrategy=merge,retainKeys
	Peerings []NetworkPeering `json:"peerings,omitempty" patchStrategy:"merge,retainKeys" patchMergeKey:"name"`

	// PeeringClaimRefs are the peering claim references of other networks.
	// +optional
	// +patchMergeKey=name
	// +patchStrategy=merge,retainKeys
	PeeringClaimRefs []NetworkPeeringClaimRef `json:"incomingPeerings,omitempty" patchStrategy:"merge,retainKeys" patchMergeKey:"name"`
}

type NetworkPeeringClaimRef struct {
	// Namespace is the namespace of the referenced entity. If empty,
	// the same namespace as the referring resource is implied.
	Namespace string `json:"namespace,omitempty"`
	// Name is the name of the referenced entity.
	Name string `json:"name"`
	// UID is the UID of the referenced entity.
	UID types.UID `json:"uid,omitempty"`
}

// NetworkPeeringNetworkRef is a reference to a network to peer with.
type NetworkPeeringNetworkRef struct {
	// Namespace is the namespace of the referenced entity. If empty,
	// the same namespace as the referring resource is implied.
	Namespace string `json:"namespace,omitempty"`
	// Name is the name of the referenced entity.
	Name string `json:"name"`
}

// NetworkPeering defines a network peering with another network.
type NetworkPeering struct {
	// Name is the semantical name of the network peering.
	Name string `json:"name"`
	// NetworkRef is the reference to the network to peer with.
	// An empty namespace indicates that the target network resides in the same namespace as the source network.
	NetworkRef NetworkPeeringNetworkRef `json:"networkRef"`
	// Prefixes is a list of prefixes that we want only to be exposed
	// to the peered network, if no prefixes are specified no filtering will be done.
	Prefixes []PeeringPrefix `json:"prefixes,omitempty"`
}

// PeeringPrefixes defines prefixes to be exposed to the peered network
type PeeringPrefix struct {
	// Name is the semantical name of the peering prefixes
	Name string `json:"name"`
	// CIDR to be exposed to the peered network
	Prefix *commonv1alpha1.IPPrefix `json:"prefix,omitempty"`
	// PrefixRef is the reference to the prefix to be exposed to peered network
	// An empty namespace indicates that the prefix resides in the same namespace as the source network.
	PrefixRef corev1.LocalObjectReference `json:"prefixRef,omitempty"`
}

// NetworkStatus defines the observed state of Network
type NetworkStatus struct {
	// State is the state of the machine.
	State NetworkState `json:"state,omitempty"`
	// Peerings contains the states of the network peerings for the network.
	// +optional
	// +patchMergeKey=name
	// +patchStrategy=merge,retainKeys
	Peerings []NetworkPeeringStatus `json:"peerings,omitempty" patchStrategy:"merge,retainKeys" patchMergeKey:"name"`
}

// NetworkState is the state of a network.
// +enum
type NetworkState string

// NetworkPeeringState is the state a NetworkPeering can be in
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
	Name string `json:"name"`
	// State represents the network peering state
	State NetworkPeeringState `json:"state,omitempty"`
	// Prefixes contains the prefixes exposed to the peered network
	Prefixes []PeeringPrefixStatus `json:"prefixes,omitempty"`
}

// PeeringPrefixStatus lists prefixes exposed to peered network
type PeeringPrefixStatus struct {
	// Name is the name of the peering prefix
	Name string `json:"name"`
	// CIDR exposed to the peered network
	Prefix *commonv1alpha1.IPPrefix `json:"prefix,omitempty"`
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
	metav1.TypeMeta   `json:",inline"`
	metav1.ObjectMeta `json:"metadata,omitempty"`

	Spec   NetworkSpec   `json:"spec,omitempty"`
	Status NetworkStatus `json:"status,omitempty"`
}

// +k8s:deepcopy-gen:interfaces=k8s.io/apimachinery/pkg/runtime.Object

// NetworkList contains a list of Network
type NetworkList struct {
	metav1.TypeMeta `json:",inline"`
	metav1.ListMeta `json:"metadata,omitempty"`
	Items           []Network `json:"items"`
}
