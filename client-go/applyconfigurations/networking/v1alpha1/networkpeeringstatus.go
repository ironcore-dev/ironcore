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
// Code generated by applyconfiguration-gen. DO NOT EDIT.

package v1alpha1

import (
	v1alpha1 "github.com/onmetal/onmetal-api/api/networking/v1alpha1"
	v1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

// NetworkPeeringStatusApplyConfiguration represents an declarative configuration of the NetworkPeeringStatus type for use
// with apply.
type NetworkPeeringStatusApplyConfiguration struct {
	Name                    *string                       `json:"name,omitempty"`
	NetworkHandle           *string                       `json:"networkHandle,omitempty"`
	Phase                   *v1alpha1.NetworkPeeringPhase `json:"phase,omitempty"`
	LastPhaseTransitionTime *v1.Time                      `json:"lastPhaseTransitionTime,omitempty"`
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

// WithNetworkHandle sets the NetworkHandle field in the declarative configuration to the given value
// and returns the receiver, so that objects can be built by chaining "With" function invocations.
// If called multiple times, the NetworkHandle field is set to the value of the last call.
func (b *NetworkPeeringStatusApplyConfiguration) WithNetworkHandle(value string) *NetworkPeeringStatusApplyConfiguration {
	b.NetworkHandle = &value
	return b
}

// WithPhase sets the Phase field in the declarative configuration to the given value
// and returns the receiver, so that objects can be built by chaining "With" function invocations.
// If called multiple times, the Phase field is set to the value of the last call.
func (b *NetworkPeeringStatusApplyConfiguration) WithPhase(value v1alpha1.NetworkPeeringPhase) *NetworkPeeringStatusApplyConfiguration {
	b.Phase = &value
	return b
}

// WithLastPhaseTransitionTime sets the LastPhaseTransitionTime field in the declarative configuration to the given value
// and returns the receiver, so that objects can be built by chaining "With" function invocations.
// If called multiple times, the LastPhaseTransitionTime field is set to the value of the last call.
func (b *NetworkPeeringStatusApplyConfiguration) WithLastPhaseTransitionTime(value v1.Time) *NetworkPeeringStatusApplyConfiguration {
	b.LastPhaseTransitionTime = &value
	return b
}