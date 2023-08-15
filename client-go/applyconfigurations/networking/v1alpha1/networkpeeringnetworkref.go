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

// NetworkPeeringNetworkRefApplyConfiguration represents an declarative configuration of the NetworkPeeringNetworkRef type for use
// with apply.
type NetworkPeeringNetworkRefApplyConfiguration struct {
	Namespace *string `json:"namespace,omitempty"`
	Name      *string `json:"name,omitempty"`
}

// NetworkPeeringNetworkRefApplyConfiguration constructs an declarative configuration of the NetworkPeeringNetworkRef type for use with
// apply.
func NetworkPeeringNetworkRef() *NetworkPeeringNetworkRefApplyConfiguration {
	return &NetworkPeeringNetworkRefApplyConfiguration{}
}

// WithNamespace sets the Namespace field in the declarative configuration to the given value
// and returns the receiver, so that objects can be built by chaining "With" function invocations.
// If called multiple times, the Namespace field is set to the value of the last call.
func (b *NetworkPeeringNetworkRefApplyConfiguration) WithNamespace(value string) *NetworkPeeringNetworkRefApplyConfiguration {
	b.Namespace = &value
	return b
}

// WithName sets the Name field in the declarative configuration to the given value
// and returns the receiver, so that objects can be built by chaining "With" function invocations.
// If called multiple times, the Name field is set to the value of the last call.
func (b *NetworkPeeringNetworkRefApplyConfiguration) WithName(value string) *NetworkPeeringNetworkRefApplyConfiguration {
	b.Name = &value
	return b
}
