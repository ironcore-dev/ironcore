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
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/types"
)

// +genclient
// +k8s:deepcopy-gen:interfaces=k8s.io/apimachinery/pkg/runtime.Object

// NATGatewayRouting is the Schema for the natgatewayroutings API
type NATGatewayRouting struct {
	metav1.TypeMeta
	metav1.ObjectMeta

	// NetworkRef is the network the NAT gateway is assigned to.
	NetworkRef commonv1alpha1.LocalUIDReference

	// Destinations are the destinations for an NATGateway.
	Destinations []NATGatewayDestination
}

type NATGatewayDestination struct {
	// Name is the name of the referenced entity.
	Name string
	// UID is the UID of the referenced entity.
	UID types.UID
	// IPs are the nat gateway ips used.
	IPs []NATGatewayDestinationIP
}

type NATGatewayDestinationIP struct {
	// IP is the ip used for the NAT gateway.
	IP commonv1alpha1.IP
	// Port is the first port used by the destination.
	Port int32
	// EndPort is the last port used by the destination.
	EndPort int32
}

// +k8s:deepcopy-gen:interfaces=k8s.io/apimachinery/pkg/runtime.Object

// NATGatewayRoutingList contains a list of NATGatewayRouting
type NATGatewayRoutingList struct {
	metav1.TypeMeta
	metav1.ListMeta
	Items []NATGatewayRouting
}
