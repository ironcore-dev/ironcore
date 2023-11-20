// SPDX-FileCopyrightText: 2023 SAP SE or an SAP affiliate company and IronCore contributors
// SPDX-License-Identifier: Apache-2.0

package networking

import (
	commonv1alpha1 "github.com/ironcore-dev/ironcore/api/common/v1alpha1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/types"
)

// +genclient
// +k8s:deepcopy-gen:interfaces=k8s.io/apimachinery/pkg/runtime.Object

// LoadBalancerRouting is the Schema for the loadbalancerroutings API
type LoadBalancerRouting struct {
	metav1.TypeMeta
	metav1.ObjectMeta

	// NetworkRef is the network the load balancer is assigned to.
	NetworkRef commonv1alpha1.LocalUIDReference

	// Destinations are the destinations for an LoadBalancer.
	Destinations []LoadBalancerDestination
}

// LoadBalancerDestination is the destination of the load balancer.
type LoadBalancerDestination struct {
	// IP is the target IP.
	IP commonv1alpha1.IP
	// TargetRef is the target providing the destination.
	TargetRef *LoadBalancerTargetRef
}

// LoadBalancerTargetRef is a load balancer target.
type LoadBalancerTargetRef struct {
	// UID is the UID of the target.
	UID types.UID
	// Name is the name of the target.
	Name string
	// ProviderID is the provider internal id of the target.
	ProviderID string
}

// +k8s:deepcopy-gen:interfaces=k8s.io/apimachinery/pkg/runtime.Object

// LoadBalancerRoutingList contains a list of LoadBalancerRouting
type LoadBalancerRoutingList struct {
	metav1.TypeMeta
	metav1.ListMeta
	Items []LoadBalancerRouting
}
