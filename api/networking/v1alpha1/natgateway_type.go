// SPDX-FileCopyrightText: 2023 SAP SE or an SAP affiliate company and IronCore contributors
// SPDX-License-Identifier: Apache-2.0

package v1alpha1

import (
	commonv1alpha1 "github.com/ironcore-dev/ironcore/api/common/v1alpha1"
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
	Type NATGatewayType `json:"type"`
	// IPFamily is the ip family the NAT gateway should have.
	IPFamily corev1.IPFamily `json:"ipFamily"`
	// NetworkRef is the Network this NATGateway should belong to.
	NetworkRef corev1.LocalObjectReference `json:"networkRef"`
	// PortsPerNetworkInterface defines the number of concurrent connections per target network interface.
	// Has to be a power of 2. If empty, 2048 (DefaultPortsPerNetworkInterface) is the default.
	PortsPerNetworkInterface *int32 `json:"portsPerNetworkInterface,omitempty"`
}

// NATGatewayStatus defines the observed state of NATGateway
type NATGatewayStatus struct {
	// IPs are the IPs allocated for the NAT gateway.
	IPs []commonv1alpha1.IP `json:"ips,omitempty"`
}

// +genclient
// +k8s:deepcopy-gen:interfaces=k8s.io/apimachinery/pkg/runtime.Object

// NATGateway is the Schema for the NATGateway API
type NATGateway struct {
	metav1.TypeMeta   `json:",inline"`
	metav1.ObjectMeta `json:"metadata,omitempty"`

	Spec   NATGatewaySpec   `json:"spec,omitempty"`
	Status NATGatewayStatus `json:"status,omitempty"`
}

// +k8s:deepcopy-gen:interfaces=k8s.io/apimachinery/pkg/runtime.Object

// NATGatewayList contains a list of NATGateway
type NATGatewayList struct {
	metav1.TypeMeta `json:",inline"`
	metav1.ListMeta `json:"metadata,omitempty"`
	Items           []NATGateway `json:"items"`
}
