// SPDX-FileCopyrightText: 2023 SAP SE or an SAP affiliate company and IronCore contributors
// SPDX-License-Identifier: Apache-2.0

package networking

import (
	commonv1alpha1 "github.com/ironcore-dev/ironcore/api/common/v1alpha1"
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

// LoadBalancerType is a type of LoadBalancer.
type LoadBalancerType string

const (
	// LoadBalancerTypePublic is a LoadBalancer that allocates and routes a stable public IP.
	LoadBalancerTypePublic LoadBalancerType = "Public"
	// LoadBalancerTypeInternal is a LoadBalancer that allocates and routes network-internal, stable IPs.
	LoadBalancerTypeInternal LoadBalancerType = "Internal"
)

// LoadBalancerSpec defines the desired state of LoadBalancer
type LoadBalancerSpec struct {
	// Type is the type of LoadBalancer.
	Type LoadBalancerType
	// IPFamilies are the ip families the load balancer should have.
	IPFamilies []corev1.IPFamily
	// IPs are the ips to use. Can only be used when Type is LoadBalancerTypeInternal.
	IPs []IPSource
	// NetworkRef is the Network this LoadBalancer should belong to.
	NetworkRef corev1.LocalObjectReference
	// NetworkInterfaceSelector defines the NetworkInterfaces
	// for which this LoadBalancer should be applied
	NetworkInterfaceSelector *metav1.LabelSelector
	// Ports are the ports the load balancer should allow.
	Ports []LoadBalancerPort
}

type LoadBalancerPort struct {
	// Protocol is the protocol the load balancer should allow.
	// If not specified, defaults to TCP.
	Protocol *corev1.Protocol
	// Port is the port to allow.
	Port int32
	// EndPort marks the end of the port range to allow.
	// If unspecified, only a single port, Port, will be allowed.
	EndPort *int32
}

// LoadBalancerStatus defines the observed state of LoadBalancer
type LoadBalancerStatus struct {
	// IPs are the IPs allocated for the load balancer.
	IPs []commonv1alpha1.IP
}

// +genclient
// +k8s:deepcopy-gen:interfaces=k8s.io/apimachinery/pkg/runtime.Object

// LoadBalancer is the Schema for the LoadBalancer API
type LoadBalancer struct {
	metav1.TypeMeta
	metav1.ObjectMeta

	Spec   LoadBalancerSpec
	Status LoadBalancerStatus
}

// +k8s:deepcopy-gen:interfaces=k8s.io/apimachinery/pkg/runtime.Object

// LoadBalancerList contains a list of LoadBalancer
type LoadBalancerList struct {
	metav1.TypeMeta
	metav1.ListMeta
	Items []LoadBalancer
}
