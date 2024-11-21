// SPDX-FileCopyrightText: 2023 SAP SE or an SAP affiliate company and IronCore contributors
// SPDX-License-Identifier: Apache-2.0

// Code generated by applyconfiguration-gen. DO NOT EDIT.

package v1alpha1

import (
	commonv1alpha1 "github.com/ironcore-dev/ironcore/api/common/v1alpha1"
	networkingv1alpha1 "github.com/ironcore-dev/ironcore/api/networking/v1alpha1"
	v1 "k8s.io/api/core/v1"
)

// EphemeralVirtualIPSpecApplyConfiguration represents an declarative configuration of the EphemeralVirtualIPSpec type for use
// with apply.
type EphemeralVirtualIPSpecApplyConfiguration struct {
	VirtualIPSpecApplyConfiguration `json:",inline"`
	ReclaimPolicy                   *networkingv1alpha1.ReclaimPolicyType `json:"reclaimPolicy,omitempty"`
}

// EphemeralVirtualIPSpecApplyConfiguration constructs an declarative configuration of the EphemeralVirtualIPSpec type for use with
// apply.
func EphemeralVirtualIPSpec() *EphemeralVirtualIPSpecApplyConfiguration {
	return &EphemeralVirtualIPSpecApplyConfiguration{}
}

// WithType sets the Type field in the declarative configuration to the given value
// and returns the receiver, so that objects can be built by chaining "With" function invocations.
// If called multiple times, the Type field is set to the value of the last call.
func (b *EphemeralVirtualIPSpecApplyConfiguration) WithType(value networkingv1alpha1.VirtualIPType) *EphemeralVirtualIPSpecApplyConfiguration {
	b.Type = &value
	return b
}

// WithIPFamily sets the IPFamily field in the declarative configuration to the given value
// and returns the receiver, so that objects can be built by chaining "With" function invocations.
// If called multiple times, the IPFamily field is set to the value of the last call.
func (b *EphemeralVirtualIPSpecApplyConfiguration) WithIPFamily(value v1.IPFamily) *EphemeralVirtualIPSpecApplyConfiguration {
	b.IPFamily = &value
	return b
}

// WithTargetRef sets the TargetRef field in the declarative configuration to the given value
// and returns the receiver, so that objects can be built by chaining "With" function invocations.
// If called multiple times, the TargetRef field is set to the value of the last call.
func (b *EphemeralVirtualIPSpecApplyConfiguration) WithTargetRef(value commonv1alpha1.LocalUIDReference) *EphemeralVirtualIPSpecApplyConfiguration {
	b.TargetRef = &value
	return b
}

// WithReclaimPolicy sets the ReclaimPolicy field in the declarative configuration to the given value
// and returns the receiver, so that objects can be built by chaining "With" function invocations.
// If called multiple times, the ReclaimPolicy field is set to the value of the last call.
func (b *EphemeralVirtualIPSpecApplyConfiguration) WithReclaimPolicy(value networkingv1alpha1.ReclaimPolicyType) *EphemeralVirtualIPSpecApplyConfiguration {
	b.ReclaimPolicy = &value
	return b
}
