// SPDX-FileCopyrightText: 2023 SAP SE or an SAP affiliate company and IronCore contributors
// SPDX-License-Identifier: Apache-2.0

// Code generated by applyconfiguration-gen. DO NOT EDIT.

package v1alpha1

// NetworkPeeringStatusApplyConfiguration represents an declarative configuration of the NetworkPeeringStatus type for use
// with apply.
type NetworkPeeringStatusApplyConfiguration struct {
	Name *string `json:"name,omitempty"`
}

// NetworkPeeringStatusApplyConfiguration constructs an declarative configuration of the NetworkPeeringStatus type for use with
// apply.
func NetworkPeeringStatus() *NetworkPeeringStatusApplyConfiguration {
	return &NetworkPeeringStatusApplyConfiguration{}
}

// WithName sets the Name field in the declarative configuration to the given value
// and returns the receiver, so that objects can be built by chaining "With" function invocations.
// If called multiple times, the Name field is set to the value of the last call.
func (b *NetworkPeeringStatusApplyConfiguration) WithName(value string) *NetworkPeeringStatusApplyConfiguration {
	b.Name = &value
	return b
}
