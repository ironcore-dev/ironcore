/*
 * Copyright (c) 2022 by the OnMetal authors.
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

// NetworkPeeringStatus is the status of a network peering.
type NetworkPeeringStatus struct {
	// Name is the name of the network peering.
	Name string `json:"name"`
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
