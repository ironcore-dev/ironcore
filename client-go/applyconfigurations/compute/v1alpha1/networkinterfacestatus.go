// SPDX-FileCopyrightText: 2023 SAP SE or an SAP affiliate company and IronCore contributors
// SPDX-License-Identifier: Apache-2.0

// Code generated by applyconfiguration-gen. DO NOT EDIT.

package v1alpha1

import (
	v1alpha1 "github.com/ironcore-dev/ironcore/api/common/v1alpha1"
	computev1alpha1 "github.com/ironcore-dev/ironcore/api/compute/v1alpha1"
	v1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

// NetworkInterfaceStatusApplyConfiguration represents an declarative configuration of the NetworkInterfaceStatus type for use
// with apply.
type NetworkInterfaceStatusApplyConfiguration struct {
	Name                    *string                                `json:"name,omitempty"`
	Handle                  *string                                `json:"handle,omitempty"`
	IPs                     []v1alpha1.IP                          `json:"ips,omitempty"`
	VirtualIP               *v1alpha1.IP                           `json:"virtualIP,omitempty"`
	State                   *computev1alpha1.NetworkInterfaceState `json:"state,omitempty"`
	NetworkInterfaceRef     *v1.LocalObjectReference               `json:"networkInterfaceRef,omitempty"`
	LastStateTransitionTime *metav1.Time                           `json:"lastStateTransitionTime,omitempty"`
}

// NetworkInterfaceStatusApplyConfiguration constructs an declarative configuration of the NetworkInterfaceStatus type for use with
// apply.
func NetworkInterfaceStatus() *NetworkInterfaceStatusApplyConfiguration {
	return &NetworkInterfaceStatusApplyConfiguration{}
}

// WithName sets the Name field in the declarative configuration to the given value
// and returns the receiver, so that objects can be built by chaining "With" function invocations.
// If called multiple times, the Name field is set to the value of the last call.
func (b *NetworkInterfaceStatusApplyConfiguration) WithName(value string) *NetworkInterfaceStatusApplyConfiguration {
	b.Name = &value
	return b
}

// WithHandle sets the Handle field in the declarative configuration to the given value
// and returns the receiver, so that objects can be built by chaining "With" function invocations.
// If called multiple times, the Handle field is set to the value of the last call.
func (b *NetworkInterfaceStatusApplyConfiguration) WithHandle(value string) *NetworkInterfaceStatusApplyConfiguration {
	b.Handle = &value
	return b
}

// WithIPs adds the given value to the IPs field in the declarative configuration
// and returns the receiver, so that objects can be build by chaining "With" function invocations.
// If called multiple times, values provided by each call will be appended to the IPs field.
func (b *NetworkInterfaceStatusApplyConfiguration) WithIPs(values ...v1alpha1.IP) *NetworkInterfaceStatusApplyConfiguration {
	for i := range values {
		b.IPs = append(b.IPs, values[i])
	}
	return b
}

// WithVirtualIP sets the VirtualIP field in the declarative configuration to the given value
// and returns the receiver, so that objects can be built by chaining "With" function invocations.
// If called multiple times, the VirtualIP field is set to the value of the last call.
func (b *NetworkInterfaceStatusApplyConfiguration) WithVirtualIP(value v1alpha1.IP) *NetworkInterfaceStatusApplyConfiguration {
	b.VirtualIP = &value
	return b
}

// WithState sets the State field in the declarative configuration to the given value
// and returns the receiver, so that objects can be built by chaining "With" function invocations.
// If called multiple times, the State field is set to the value of the last call.
func (b *NetworkInterfaceStatusApplyConfiguration) WithState(value computev1alpha1.NetworkInterfaceState) *NetworkInterfaceStatusApplyConfiguration {
	b.State = &value
	return b
}

// WithNetworkInterfaceRef sets the NetworkInterfaceRef field in the declarative configuration to the given value
// and returns the receiver, so that objects can be built by chaining "With" function invocations.
// If called multiple times, the NetworkInterfaceRef field is set to the value of the last call.
func (b *NetworkInterfaceStatusApplyConfiguration) WithNetworkInterfaceRef(value v1.LocalObjectReference) *NetworkInterfaceStatusApplyConfiguration {
	b.NetworkInterfaceRef = &value
	return b
}

// WithLastStateTransitionTime sets the LastStateTransitionTime field in the declarative configuration to the given value
// and returns the receiver, so that objects can be built by chaining "With" function invocations.
// If called multiple times, the LastStateTransitionTime field is set to the value of the last call.
func (b *NetworkInterfaceStatusApplyConfiguration) WithLastStateTransitionTime(value metav1.Time) *NetworkInterfaceStatusApplyConfiguration {
	b.LastStateTransitionTime = &value
	return b
}
