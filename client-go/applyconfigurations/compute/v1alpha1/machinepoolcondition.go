// SPDX-FileCopyrightText: 2025 SAP SE or an SAP affiliate company and IronCore contributors
// SPDX-License-Identifier: Apache-2.0

// Code generated by applyconfiguration-gen. DO NOT EDIT.

package v1alpha1

import (
	computev1alpha1 "github.com/ironcore-dev/ironcore/api/compute/v1alpha1"
	v1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

// MachinePoolConditionApplyConfiguration represents a declarative configuration of the MachinePoolCondition type for use
// with apply.
type MachinePoolConditionApplyConfiguration struct {
	Type               *computev1alpha1.MachinePoolConditionType `json:"type,omitempty"`
	Status             *v1.ConditionStatus                       `json:"status,omitempty"`
	Reason             *string                                   `json:"reason,omitempty"`
	Message            *string                                   `json:"message,omitempty"`
	ObservedGeneration *int64                                    `json:"observedGeneration,omitempty"`
	LastTransitionTime *metav1.Time                              `json:"lastTransitionTime,omitempty"`
}

// MachinePoolConditionApplyConfiguration constructs a declarative configuration of the MachinePoolCondition type for use with
// apply.
func MachinePoolCondition() *MachinePoolConditionApplyConfiguration {
	return &MachinePoolConditionApplyConfiguration{}
}

// WithType sets the Type field in the declarative configuration to the given value
// and returns the receiver, so that objects can be built by chaining "With" function invocations.
// If called multiple times, the Type field is set to the value of the last call.
func (b *MachinePoolConditionApplyConfiguration) WithType(value computev1alpha1.MachinePoolConditionType) *MachinePoolConditionApplyConfiguration {
	b.Type = &value
	return b
}

// WithStatus sets the Status field in the declarative configuration to the given value
// and returns the receiver, so that objects can be built by chaining "With" function invocations.
// If called multiple times, the Status field is set to the value of the last call.
func (b *MachinePoolConditionApplyConfiguration) WithStatus(value v1.ConditionStatus) *MachinePoolConditionApplyConfiguration {
	b.Status = &value
	return b
}

// WithReason sets the Reason field in the declarative configuration to the given value
// and returns the receiver, so that objects can be built by chaining "With" function invocations.
// If called multiple times, the Reason field is set to the value of the last call.
func (b *MachinePoolConditionApplyConfiguration) WithReason(value string) *MachinePoolConditionApplyConfiguration {
	b.Reason = &value
	return b
}

// WithMessage sets the Message field in the declarative configuration to the given value
// and returns the receiver, so that objects can be built by chaining "With" function invocations.
// If called multiple times, the Message field is set to the value of the last call.
func (b *MachinePoolConditionApplyConfiguration) WithMessage(value string) *MachinePoolConditionApplyConfiguration {
	b.Message = &value
	return b
}

// WithObservedGeneration sets the ObservedGeneration field in the declarative configuration to the given value
// and returns the receiver, so that objects can be built by chaining "With" function invocations.
// If called multiple times, the ObservedGeneration field is set to the value of the last call.
func (b *MachinePoolConditionApplyConfiguration) WithObservedGeneration(value int64) *MachinePoolConditionApplyConfiguration {
	b.ObservedGeneration = &value
	return b
}

// WithLastTransitionTime sets the LastTransitionTime field in the declarative configuration to the given value
// and returns the receiver, so that objects can be built by chaining "With" function invocations.
// If called multiple times, the LastTransitionTime field is set to the value of the last call.
func (b *MachinePoolConditionApplyConfiguration) WithLastTransitionTime(value metav1.Time) *MachinePoolConditionApplyConfiguration {
	b.LastTransitionTime = &value
	return b
}
