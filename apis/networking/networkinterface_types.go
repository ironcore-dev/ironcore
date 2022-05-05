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
	commonv1alpha1 "github.com/onmetal/onmetal-api/apis/common/v1alpha1"
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

// NetworkInterfaceSpec defines the desired state of NetworkInterface
type NetworkInterfaceSpec struct {
	// NetworkRef is the Network this NetworkInterface is connected to
	NetworkRef corev1.LocalObjectReference
	// MachineRef is the Machine this NetworkInterface is used by
	MachineRef *corev1.LocalObjectReference
	// IPFamilies defines which IPFamilies this NetworkInterface is supporting
	IPFamilies []corev1.IPFamily
	// IPs is the list of provided IPs or EphemeralIPs which should be assigned to
	// this NetworkInterface
	IPs []IPSource
	// VirtualIP specifies the virtual ip that should be assigned to this NetworkInterface.
	VirtualIP *VirtualIPSource
}

type IPSource struct {
	Value           *commonv1alpha1.IP
	EphemeralPrefix *EphemeralPrefixSource
}

type VirtualIPSource struct {
	VirtualIPClaimRef *corev1.LocalObjectReference
	Ephemeral         *EphemeralVirtualIPSource
}

// NetworkInterfaceStatus defines the observed state of NetworkInterface
type NetworkInterfaceStatus struct {
	// IPs represent the effective IP addresses of the NetworkInterface
	IPs []commonv1alpha1.IP
	// VirtualIP is any virtual ip assigned to the NetworkInterface.
	VirtualIP *commonv1alpha1.IP
}

// +genclient
// +k8s:deepcopy-gen:interfaces=k8s.io/apimachinery/pkg/runtime.Object

// NetworkInterface is the Schema for the networkinterfaces API
type NetworkInterface struct {
	metav1.TypeMeta
	metav1.ObjectMeta

	Spec   NetworkInterfaceSpec
	Status NetworkInterfaceStatus
}

// +k8s:deepcopy-gen:interfaces=k8s.io/apimachinery/pkg/runtime.Object

// NetworkInterfaceList contains a list of NetworkInterface
type NetworkInterfaceList struct {
	metav1.TypeMeta
	metav1.ListMeta
	Items []NetworkInterface
}
