// SPDX-FileCopyrightText: 2023 SAP SE or an SAP affiliate company and IronCore contributors
// SPDX-License-Identifier: Apache-2.0

// Code generated by applyconfiguration-gen. DO NOT EDIT.

package v1alpha1

import (
	commonv1alpha1 "github.com/ironcore-dev/ironcore/api/common/v1alpha1"
)

// VirtualIPStatusApplyConfiguration represents a declarative configuration of the VirtualIPStatus type for use
// with apply.
type VirtualIPStatusApplyConfiguration struct {
	IP *commonv1alpha1.IP `json:"ip,omitempty"`
}

// VirtualIPStatusApplyConfiguration constructs a declarative configuration of the VirtualIPStatus type for use with
// apply.
func VirtualIPStatus() *VirtualIPStatusApplyConfiguration {
	return &VirtualIPStatusApplyConfiguration{}
}

// WithIP sets the IP field in the declarative configuration to the given value
// and returns the receiver, so that objects can be built by chaining "With" function invocations.
// If called multiple times, the IP field is set to the value of the last call.
func (b *VirtualIPStatusApplyConfiguration) WithIP(value commonv1alpha1.IP) *VirtualIPStatusApplyConfiguration {
	b.IP = &value
	return b
}
