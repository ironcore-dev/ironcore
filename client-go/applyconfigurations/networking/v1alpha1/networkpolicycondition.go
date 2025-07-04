// SPDX-FileCopyrightText: 2025 SAP SE or an SAP affiliate company and IronCore contributors
// SPDX-License-Identifier: Apache-2.0

// Code generated by applyconfiguration-gen. DO NOT EDIT.

package v1alpha1

import (
	networkingv1alpha1 "github.com/ironcore-dev/ironcore/api/networking/v1alpha1"
	v1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

// NetworkPolicyConditionApplyConfiguration represents a declarative configuration of the NetworkPolicyCondition type for use
// with apply.
type NetworkPolicyConditionApplyConfiguration struct {
	Type               *networkingv1alpha1.NetworkPolicyConditionType `json:"type,omitempty"`
	Status             *v1.ConditionStatus                            `json:"status,omitempty"`
	Reason             *string                                        `json:"reason,omitempty"`
	Message            *string                                        `json:"message,omitempty"`
	ObservedGeneration *int64                                         `json:"observedGeneration,omitempty"`
	LastTransitionTime *metav1.Time                                   `json:"lastTransitionTime,omitempty"`
}

// NetworkPolicyConditionApplyConfiguration constructs a declarative configuration of the NetworkPolicyCondition type for use with
// apply.
func NetworkPolicyCondition() *NetworkPolicyConditionApplyConfiguration {
	return &NetworkPolicyConditionApplyConfiguration{}
}

// WithType sets the Type field in the declarative configuration to the given value
// and returns the receiver, so that objects can be built by chaining "With" function invocations.
// If called multiple times, the Type field is set to the value of the last call.
func (b *NetworkPolicyConditionApplyConfiguration) WithType(value networkingv1alpha1.NetworkPolicyConditionType) *NetworkPolicyConditionApplyConfiguration {
	b.Type = &value
	return b
}

// WithStatus sets the Status field in the declarative configuration to the given value
// and returns the receiver, so that objects can be built by chaining "With" function invocations.
// If called multiple times, the Status field is set to the value of the last call.
func (b *NetworkPolicyConditionApplyConfiguration) WithStatus(value v1.ConditionStatus) *NetworkPolicyConditionApplyConfiguration {
	b.Status = &value
	return b
}

// WithReason sets the Reason field in the declarative configuration to the given value
// and returns the receiver, so that objects can be built by chaining "With" function invocations.
// If called multiple times, the Reason field is set to the value of the last call.
func (b *NetworkPolicyConditionApplyConfiguration) WithReason(value string) *NetworkPolicyConditionApplyConfiguration {
	b.Reason = &value
	return b
}

// WithMessage sets the Message field in the declarative configuration to the given value
// and returns the receiver, so that objects can be built by chaining "With" function invocations.
// If called multiple times, the Message field is set to the value of the last call.
func (b *NetworkPolicyConditionApplyConfiguration) WithMessage(value string) *NetworkPolicyConditionApplyConfiguration {
	b.Message = &value
	return b
}

// WithObservedGeneration sets the ObservedGeneration field in the declarative configuration to the given value
// and returns the receiver, so that objects can be built by chaining "With" function invocations.
// If called multiple times, the ObservedGeneration field is set to the value of the last call.
func (b *NetworkPolicyConditionApplyConfiguration) WithObservedGeneration(value int64) *NetworkPolicyConditionApplyConfiguration {
	b.ObservedGeneration = &value
	return b
}

// WithLastTransitionTime sets the LastTransitionTime field in the declarative configuration to the given value
// and returns the receiver, so that objects can be built by chaining "With" function invocations.
// If called multiple times, the LastTransitionTime field is set to the value of the last call.
func (b *NetworkPolicyConditionApplyConfiguration) WithLastTransitionTime(value metav1.Time) *NetworkPolicyConditionApplyConfiguration {
	b.LastTransitionTime = &value
	return b
}
