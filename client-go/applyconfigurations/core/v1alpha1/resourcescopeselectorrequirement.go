// SPDX-FileCopyrightText: 2025 SAP SE or an SAP affiliate company and IronCore contributors
// SPDX-License-Identifier: Apache-2.0

// Code generated by applyconfiguration-gen. DO NOT EDIT.

package v1alpha1

import (
	corev1alpha1 "github.com/ironcore-dev/ironcore/api/core/v1alpha1"
)

// ResourceScopeSelectorRequirementApplyConfiguration represents a declarative configuration of the ResourceScopeSelectorRequirement type for use
// with apply.
type ResourceScopeSelectorRequirementApplyConfiguration struct {
	ScopeName *corev1alpha1.ResourceScope                 `json:"scopeName,omitempty"`
	Operator  *corev1alpha1.ResourceScopeSelectorOperator `json:"operator,omitempty"`
	Values    []string                                    `json:"values,omitempty"`
}

// ResourceScopeSelectorRequirementApplyConfiguration constructs a declarative configuration of the ResourceScopeSelectorRequirement type for use with
// apply.
func ResourceScopeSelectorRequirement() *ResourceScopeSelectorRequirementApplyConfiguration {
	return &ResourceScopeSelectorRequirementApplyConfiguration{}
}

// WithScopeName sets the ScopeName field in the declarative configuration to the given value
// and returns the receiver, so that objects can be built by chaining "With" function invocations.
// If called multiple times, the ScopeName field is set to the value of the last call.
func (b *ResourceScopeSelectorRequirementApplyConfiguration) WithScopeName(value corev1alpha1.ResourceScope) *ResourceScopeSelectorRequirementApplyConfiguration {
	b.ScopeName = &value
	return b
}

// WithOperator sets the Operator field in the declarative configuration to the given value
// and returns the receiver, so that objects can be built by chaining "With" function invocations.
// If called multiple times, the Operator field is set to the value of the last call.
func (b *ResourceScopeSelectorRequirementApplyConfiguration) WithOperator(value corev1alpha1.ResourceScopeSelectorOperator) *ResourceScopeSelectorRequirementApplyConfiguration {
	b.Operator = &value
	return b
}

// WithValues adds the given value to the Values field in the declarative configuration
// and returns the receiver, so that objects can be build by chaining "With" function invocations.
// If called multiple times, values provided by each call will be appended to the Values field.
func (b *ResourceScopeSelectorRequirementApplyConfiguration) WithValues(values ...string) *ResourceScopeSelectorRequirementApplyConfiguration {
	for i := range values {
		b.Values = append(b.Values, values[i])
	}
	return b
}
