// SPDX-FileCopyrightText: 2023 SAP SE or an SAP affiliate company and IronCore contributors
// SPDX-License-Identifier: Apache-2.0

package v1alpha1

import (
	commonv1alpha1 "github.com/ironcore-dev/ironcore/api/common/v1alpha1"
	corev1alpha1 "github.com/ironcore-dev/ironcore/api/core/v1alpha1"
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

// NetworkPolicySpec defines the desired state of NetworkPolicy.
type NetworkPolicySpec struct {
	// NetworkRef is the network to regulate using this policy.
	NetworkRef corev1.LocalObjectReference `json:"networkRef"`
	// NetworkInterfaceSelector selects the network interfaces that are subject to this policy.
	NetworkInterfaceSelector metav1.LabelSelector `json:"networkInterfaceSelector"`
	// Ingress specifies rules for ingress traffic.
	Ingress []NetworkPolicyIngressRule `json:"ingress,omitempty"`
	// Egress specifies rules for egress traffic.
	Egress []NetworkPolicyEgressRule `json:"egress,omitempty"`
	// PolicyTypes specifies the types of policies this network policy contains.
	PolicyTypes []PolicyType `json:"policyTypes,omitempty"`
}

// NetworkPolicyPort describes a port to allow traffic on
type NetworkPolicyPort struct {
	// Protocol (TCP, UDP, or SCTP) which traffic must match. If not specified, this
	// field defaults to TCP.
	Protocol *corev1.Protocol `json:"protocol,omitempty"`

	// The port on the given protocol. If this field is not provided, this matches
	// all port names and numbers.
	// If present, only traffic on the specified protocol AND port will be matched.
	Port int32 `json:"port,omitempty"`

	// EndPort indicates that the range of ports from Port to EndPort, inclusive,
	// should be allowed by the policy. This field cannot be defined if the port field
	// is not defined. The endPort must be equal or greater than port.
	EndPort *int32 `json:"endPort,omitempty" protobuf:"bytes,3,opt,name=endPort"`
}

// IPBlock specifies an ip block with optional exceptions.
type IPBlock struct {
	// CIDR is a string representing the ip block.
	CIDR commonv1alpha1.IPPrefix `json:"cidr"`
	// Except is a slice of CIDRs that should not be included within the specified CIDR.
	// Values will be rejected if they are outside CIDR.
	Except []commonv1alpha1.IPPrefix `json:"except,omitempty"`
}

// NetworkPolicyPeer describes a peer to allow traffic to / from.
type NetworkPolicyPeer struct {
	// ObjectSelector selects peers with the given kind matching the label selector.
	// Exclusive with other peer specifiers.
	ObjectSelector *corev1alpha1.ObjectSelector `json:"objectSelector,omitempty"`
	// IPBlock specifies the ip block from or to which network traffic may come.
	IPBlock *IPBlock `json:"ipBlock,omitempty"`
}

// NetworkPolicyIngressRule describes a rule to regulate ingress traffic with.
type NetworkPolicyIngressRule struct {
	// Ports specifies the list of ports which should be made accessible for
	// this rule. Each item in this list is combined using a logical OR. Empty matches all ports.
	// As soon as a single item is present, only these ports are allowed.
	Ports []NetworkPolicyPort `json:"ports,omitempty"`
	// From specifies the list of sources which should be able to send traffic to the
	// selected network interfaces. Fields are combined using a logical OR. Empty matches all sources.
	// As soon as a single item is present, only these peers are allowed.
	From []NetworkPolicyPeer `json:"from,omitempty"`
}

// NetworkPolicyEgressRule describes a rule to regulate egress traffic with.
type NetworkPolicyEgressRule struct {
	// Ports specifies the list of destination ports that can be called with
	// this rule. Each item in this list is combined using a logical OR. Empty matches all ports.
	// As soon as a single item is present, only these ports are allowed.
	Ports []NetworkPolicyPort `json:"ports,omitempty"`
	// To specifies the list of destinations which the selected network interfaces should be
	// able to send traffic to. Fields are combined using a logical OR. Empty matches all destinations.
	// As soon as a single item is present, only these peers are allowed.
	To []NetworkPolicyPeer `json:"to,omitempty"`
}

// PolicyType is a type of policy.
type PolicyType string

const (
	// PolicyTypeIngress is a policy that describes ingress traffic.
	PolicyTypeIngress PolicyType = "Ingress"
	// PolicyTypeEgress is a policy that describes egress traffic.
	PolicyTypeEgress PolicyType = "Egress"
)

// NetworkPolicyStatus defines the observed state of NetworkPolicy.
type NetworkPolicyStatus struct {
	// Conditions are various conditions of the NetworkPolicy.
	Conditions []NetworkPolicyCondition `json:"conditions,omitempty"`
}

// NetworkPolicyConditionType is a type a NetworkPolicyCondition can have.
type NetworkPolicyConditionType string

// NetworkPolicyCondition is one of the conditions of a network policy.
type NetworkPolicyCondition struct {
	// Type is the type of the condition.
	Type NetworkPolicyConditionType `json:"type"`
	// Status is the status of the condition.
	Status corev1.ConditionStatus `json:"status"`
	// Reason is a machine-readable indication of why the condition is in a certain state.
	Reason string `json:"reason"`
	// Message is a human-readable explanation of why the condition has a certain reason / state.
	Message string `json:"message"`
	// ObservedGeneration represents the .metadata.generation that the condition was set based upon.
	ObservedGeneration int64 `json:"observedGeneration,omitempty"`
	// LastTransitionTime is the last time the status of a condition has transitioned from one state to another.
	LastTransitionTime metav1.Time `json:"lastTransitionTime,omitempty"`
}

// +genclient
// +k8s:deepcopy-gen:interfaces=k8s.io/apimachinery/pkg/runtime.Object

// NetworkPolicy is the Schema for the networkpolicies API
type NetworkPolicy struct {
	metav1.TypeMeta   `json:",inline"`
	metav1.ObjectMeta `json:"metadata,omitempty"`

	Spec   NetworkPolicySpec   `json:"spec,omitempty"`
	Status NetworkPolicyStatus `json:"status,omitempty"`
}

// +k8s:deepcopy-gen:interfaces=k8s.io/apimachinery/pkg/runtime.Object

// NetworkPolicyList contains a list of NetworkPolicy.
type NetworkPolicyList struct {
	metav1.TypeMeta `json:",inline"`
	metav1.ListMeta `json:"metadata,omitempty"`
	Items           []NetworkPolicy `json:"items"`
}
