// SPDX-FileCopyrightText: 2023 SAP SE or an SAP affiliate company and IronCore contributors
// SPDX-License-Identifier: Apache-2.0

// Code generated by applyconfiguration-gen. DO NOT EDIT.

package v1alpha1

import (
	commonv1alpha1 "github.com/ironcore-dev/ironcore/api/common/v1alpha1"
	networkingv1alpha1 "github.com/ironcore-dev/ironcore/api/networking/v1alpha1"
	v1 "k8s.io/api/core/v1"
)

// VirtualIPSpecApplyConfiguration represents a declarative configuration of the VirtualIPSpec type for use
// with apply.
type VirtualIPSpecApplyConfiguration struct {
	Type      *networkingv1alpha1.VirtualIPType `json:"type,omitempty"`
	IPFamily  *v1.IPFamily                      `json:"ipFamily,omitempty"`
	TargetRef *commonv1alpha1.LocalUIDReference `json:"targetRef,omitempty"`
}

// VirtualIPSpecApplyConfiguration constructs a declarative configuration of the VirtualIPSpec type for use with
// apply.
func VirtualIPSpec() *VirtualIPSpecApplyConfiguration {
	return &VirtualIPSpecApplyConfiguration{}
}

// WithType sets the Type field in the declarative configuration to the given value
// and returns the receiver, so that objects can be built by chaining "With" function invocations.
// If called multiple times, the Type field is set to the value of the last call.
func (b *VirtualIPSpecApplyConfiguration) WithType(value networkingv1alpha1.VirtualIPType) *VirtualIPSpecApplyConfiguration {
	b.Type = &value
	return b
}

// WithIPFamily sets the IPFamily field in the declarative configuration to the given value
// and returns the receiver, so that objects can be built by chaining "With" function invocations.
// If called multiple times, the IPFamily field is set to the value of the last call.
func (b *VirtualIPSpecApplyConfiguration) WithIPFamily(value v1.IPFamily) *VirtualIPSpecApplyConfiguration {
	b.IPFamily = &value
	return b
}

// WithTargetRef sets the TargetRef field in the declarative configuration to the given value
// and returns the receiver, so that objects can be built by chaining "With" function invocations.
// If called multiple times, the TargetRef field is set to the value of the last call.
func (b *VirtualIPSpecApplyConfiguration) WithTargetRef(value commonv1alpha1.LocalUIDReference) *VirtualIPSpecApplyConfiguration {
	b.TargetRef = &value
	return b
}
