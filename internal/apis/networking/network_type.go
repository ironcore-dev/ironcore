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

package networking

import (
	corev1alpha1 "github.com/onmetal/onmetal-api/api/core/v1alpha1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

// NetworkSpec defines the desired state of Network
type NetworkSpec struct {
	// Handle is the identifier of the network provider.
	Handle string
	// Peerings are the network peerings with this network.
	Peerings []NetworkPeering
}

// NetworkPeering defines a network peering with another network.
type NetworkPeering struct {
	// Name is the semantical name of the network peering.
	Name string
	// NetworkRef is the reference to the network to peer with.
	// If the UID is empty, it will be populated once when the peering is successfully bound.
	// If namespace is empty it is implied that the target network resides in the same network.
	NetworkRef corev1alpha1.UIDReference
}

// NetworkStatus defines the observed state of Network
type NetworkStatus struct {
	// State is the state of the machine.
	State NetworkState
	// Peerings contains the states of the network peerings for the network.
	Peerings []NetworkPeeringStatus
}

// NetworkState is the state of a network.
// +enum
type NetworkState string

// NetworkPeeringPhase is the phase a NetworkPeering can be in.
type NetworkPeeringPhase string

const (
	// NetworkPeeringPhasePending signals that the network peering is not bound.
	NetworkPeeringPhasePending NetworkPeeringPhase = "Pending"
	// NetworkPeeringPhaseBound signals that the network peering is bound.
	NetworkPeeringPhaseBound NetworkPeeringPhase = "Bound"
)

// NetworkPeeringStatus is the status of a network peering.
type NetworkPeeringStatus struct {
	// Name is the name of the network peering.
	Name string
	// NetworkHandle is the handle of the peered network.
	NetworkHandle string
	// Phase represents the binding phase of a network peering.
	Phase NetworkPeeringPhase
	// LastPhaseTransitionTime is the last time the Phase transitioned.
	LastPhaseTransitionTime *metav1.Time
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
