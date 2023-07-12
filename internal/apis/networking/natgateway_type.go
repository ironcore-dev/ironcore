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
	commonv1alpha1 "github.com/onmetal/onmetal-api/api/common/v1alpha1"
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

const (
	// DefaultPortsPerNetworkInterface is the default number of ports per network interface.
	DefaultPortsPerNetworkInterface int32 = 2048
)

// NATGatewayType is a type of NATGateway.
type NATGatewayType string

const (
	// NATGatewayTypePublic is a NATGateway that allocates and routes a stable public IP.
	NATGatewayTypePublic NATGatewayType = "Public"
)

// NATGatewaySpec defines the desired state of NATGateway
type NATGatewaySpec struct {
	// Type is the type of NATGateway.
	Type NATGatewayType
	// IPFamily is the ip family the NAT gateway should have.
	IPFamily corev1.IPFamily
	// NetworkRef is the Network this NATGateway should belong to.
	NetworkRef corev1.LocalObjectReference
	// PortsPerNetworkInterface defines the number of concurrent connections per target network interface.
	// Has to be a power of 2. If empty, 2048 (DefaultPortsPerNetworkInterface) is the default.
	PortsPerNetworkInterface *int32
}

// NATGatewayStatus defines the observed state of NATGateway
type NATGatewayStatus struct {
	// IPs are the IPs allocated for the NAT gateway.
	IPs []commonv1alpha1.IP
}

// +genclient
// +k8s:deepcopy-gen:interfaces=k8s.io/apimachinery/pkg/runtime.Object

// NATGateway is the Schema for the NATGateway API
type NATGateway struct {
	metav1.TypeMeta
	metav1.ObjectMeta

	Spec   NATGatewaySpec
	Status NATGatewayStatus
}

// +k8s:deepcopy-gen:interfaces=k8s.io/apimachinery/pkg/runtime.Object

// NATGatewayList contains a list of NATGateway
type NATGatewayList struct {
	metav1.TypeMeta
	metav1.ListMeta
	Items []NATGateway
}
