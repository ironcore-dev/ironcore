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

// NetworkInterfaceSpec defines the desired state of NetworkInterface
type NetworkInterfaceSpec struct {
	// NetworkRef is the Network this NetworkInterface is connected to
	NetworkRef corev1.LocalObjectReference `json:"networkRef"`
	// MachineRef is the Machine this NetworkInterface is used by
	MachineRef *commonv1alpha1.LocalUIDReference `json:"machineRef,omitempty"`
	// IPFamilies defines which IPFamilies this NetworkInterface is supporting
	IPFamilies []corev1.IPFamily `json:"ipFamilies"`
	// IPs is the list of provided IPs or EphemeralIPs which should be assigned to
	// this NetworkInterface
	IPs []IPSource `json:"ips"`
	// VirtualIP specifies the virtual ip that should be assigned to this NetworkInterface.
	VirtualIP *VirtualIPSource `json:"virtualIP,omitempty"`
}

// IPSource is the definition of how to obtain an IP.
type IPSource struct {
	// Value specifies an IP by using an IP literal.
	Value *commonv1alpha1.IP `json:"value,omitempty"`
	// Ephemeral specifies an IP by creating an ephemeral Prefix to allocate the IP with.
	Ephemeral *EphemeralPrefixSource `json:"ephemeral,omitempty"`
}

// VirtualIPSource is the definition of how to obtain a VirtualIP.
type VirtualIPSource struct {
	// VirtualIPRef references a VirtualIP to use.
	VirtualIPRef *corev1.LocalObjectReference `json:"virtualIPRef,omitempty"`
	// Ephemeral instructs to create an ephemeral (i.e. coupled to the lifetime of the surrounding object)
	// VirtualIP.
	Ephemeral *EphemeralVirtualIPSource `json:"ephemeral,omitempty"`
}

// NetworkInterfaceStatus defines the observed state of NetworkInterface
type NetworkInterfaceStatus struct {
	// State is the NetworkInterfaceState of the NetworkInterface.
	State NetworkInterfaceState `json:"state,omitempty"`
	// LastStateTransitionTime is the last time the State transitioned from one value to another.
	LastStateTransitionTime *metav1.Time `json:"lastStateTransitionTime,omitempty"`

	// NetworkHandle is the handle of the network the network interface is part of.
	NetworkHandle string `json:"networkHandle,omitempty"`
	// IPs represent the effective IP addresses of the NetworkInterface
	IPs []commonv1alpha1.IP `json:"ips,omitempty"`
	// VirtualIP is any virtual ip assigned to the NetworkInterface.
	VirtualIP *commonv1alpha1.IP `json:"virtualIP,omitempty"`

	// Phase is the NetworkInterfacePhase of the NetworkInterface.
	Phase NetworkInterfacePhase `json:"phase,omitempty"`
	// MachinePoolRef is the machine pool the network interface is currently on, if any.
	MachinePoolRef *corev1.LocalObjectReference `json:"machinePoolRef,omitempty"`
	// LastPhaseTransitionTime is the last time the Phase transitioned from one value to another.
	LastPhaseTransitionTime *metav1.Time `json:"lastPhaseTransitionTime,omitempty"`
}

// NetworkInterfaceState is the onmetal-api state of a NetworkInterface.
type NetworkInterfaceState string

const (
	// NetworkInterfaceStatePending is used for any NetworkInterface that is pending.
	NetworkInterfaceStatePending NetworkInterfaceState = "Pending"
	// NetworkInterfaceStateAvailable is used for any NetworkInterface where all properties are valid.
	NetworkInterfaceStateAvailable NetworkInterfaceState = "Available"
	// NetworkInterfaceStateError is used for any NetworkInterface where any property has an error.
	NetworkInterfaceStateError NetworkInterfaceState = "Error"
)

// NetworkInterfacePhase is the binding phase of a NetworkInterface.
type NetworkInterfacePhase string

const (
	// NetworkInterfacePhaseUnbound is used for any NetworkInterface that is not bound.
	NetworkInterfacePhaseUnbound NetworkInterfacePhase = "Unbound"
	// NetworkInterfacePhasePending is used for any NetworkInterface that is currently awaiting binding.
	NetworkInterfacePhasePending NetworkInterfacePhase = "Pending"
	// NetworkInterfacePhaseBound is used for any NetworkInterface that is properly bound.
	NetworkInterfacePhaseBound NetworkInterfacePhase = "Bound"
)

// +genclient
// +k8s:deepcopy-gen:interfaces=k8s.io/apimachinery/pkg/runtime.Object

// NetworkInterface is the Schema for the networkinterfaces API
type NetworkInterface struct {
	metav1.TypeMeta   `json:",inline"`
	metav1.ObjectMeta `json:"metadata,omitempty"`

	Spec   NetworkInterfaceSpec   `json:"spec,omitempty"`
	Status NetworkInterfaceStatus `json:"status,omitempty"`
}

// +k8s:deepcopy-gen:interfaces=k8s.io/apimachinery/pkg/runtime.Object

// NetworkInterfaceList contains a list of NetworkInterface
type NetworkInterfaceList struct {
	metav1.TypeMeta `json:",inline"`
	metav1.ListMeta `json:"metadata,omitempty"`
	Items           []NetworkInterface `json:"items"`
}

// NetworkInterfaceTemplateSpec is the specification of a NetworkInterface template.
type NetworkInterfaceTemplateSpec struct {
	metav1.ObjectMeta `json:"metadata,omitempty"`
	Spec              NetworkInterfaceSpec `json:"spec,omitempty"`
}
