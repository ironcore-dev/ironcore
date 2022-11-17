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
	commonv1alpha1 "github.com/onmetal/onmetal-api/api/common/v1alpha1"
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

// LoadBalancerType is a type of LoadBalancer.
type LoadBalancerType string

const (
	// LoadBalancerTypePublic is a LoadBalancer that allocates and routes a stable public IP.
	LoadBalancerTypePublic LoadBalancerType = "Public"
)

// LoadBalancerSpec defines the desired state of LoadBalancer
type LoadBalancerSpec struct {
	// Type is the type of LoadBalancer.
	Type LoadBalancerType `json:"type"`
	// IPFamilies are the ip families the load balancer should have.
	IPFamilies []corev1.IPFamily `json:"ipFamilies"`
	// NetworkRef is the Network this LoadBalancer should belong to.
	NetworkRef corev1.LocalObjectReference `json:"networkRef"`
	// NetworkInterfaceSelector defines the NetworkInterfaces
	// for which this LoadBalancer should be applied
	NetworkInterfaceSelector *metav1.LabelSelector `json:"networkInterfaceSelector,omitempty"`
	// Ports are the ports the load balancer should allow.
	Ports []LoadBalancerPort `json:"ports,omitempty"`
}

type LoadBalancerPort struct {
	// Protocol is the protocol the load balancer should allow.
	// If not specified, defaults to TCP.
	Protocol *corev1.Protocol `json:"protocol,omitempty"`
	// Port is the port to allow.
	Port int32 `json:"port"`
	// EndPort marks the end of the port range to allow.
	// If unspecified, only a single port, Port, will be allowed.
	EndPort *int32 `json:"endPort,omitempty"`
}

// LoadBalancerStatus defines the observed state of LoadBalancer
type LoadBalancerStatus struct {
	// IPs are the IPs allocated for the load balancer.
	IPs []commonv1alpha1.IP `json:"ips,omitempty"`
}

// +genclient
// +k8s:deepcopy-gen:interfaces=k8s.io/apimachinery/pkg/runtime.Object

// LoadBalancer is the Schema for the LoadBalancer API
type LoadBalancer struct {
	metav1.TypeMeta   `json:",inline"`
	metav1.ObjectMeta `json:"metadata,omitempty"`

	Spec   LoadBalancerSpec   `json:"spec,omitempty"`
	Status LoadBalancerStatus `json:"status,omitempty"`
}

// +k8s:deepcopy-gen:interfaces=k8s.io/apimachinery/pkg/runtime.Object

// LoadBalancerList contains a list of LoadBalancer
type LoadBalancerList struct {
	metav1.TypeMeta `json:",inline"`
	metav1.ListMeta `json:"metadata,omitempty"`
	Items           []LoadBalancer `json:"items"`
}
