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
	v1alpha1 "github.com/onmetal/onmetal-api/api/common/v1alpha1"
	computev1alpha1 "github.com/onmetal/onmetal-api/api/compute/v1alpha1"
	v1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

// NetworkInterfaceStatusApplyConfiguration represents an declarative configuration of the NetworkInterfaceStatus type for use
// with apply.
type NetworkInterfaceStatusApplyConfiguration struct {
	Name                    *string                                `json:"name,omitempty"`
	NetworkHandle           *string                                `json:"networkHandle,omitempty"`
	IPs                     []v1alpha1.IP                          `json:"ips,omitempty"`
	VirtualIP               *v1alpha1.IP                           `json:"virtualIP,omitempty"`
	State                   *computev1alpha1.NetworkInterfaceState `json:"state,omitempty"`
	LastStateTransitionTime *v1.Time                               `json:"lastStateTransitionTime,omitempty"`
	Phase                   *computev1alpha1.NetworkInterfacePhase `json:"phase,omitempty"`
	LastPhaseTransitionTime *v1.Time                               `json:"lastPhaseTransitionTime,omitempty"`
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

// WithNetworkHandle sets the NetworkHandle field in the declarative configuration to the given value
// and returns the receiver, so that objects can be built by chaining "With" function invocations.
// If called multiple times, the NetworkHandle field is set to the value of the last call.
func (b *NetworkInterfaceStatusApplyConfiguration) WithNetworkHandle(value string) *NetworkInterfaceStatusApplyConfiguration {
	b.NetworkHandle = &value
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

// WithLastStateTransitionTime sets the LastStateTransitionTime field in the declarative configuration to the given value
// and returns the receiver, so that objects can be built by chaining "With" function invocations.
// If called multiple times, the LastStateTransitionTime field is set to the value of the last call.
func (b *NetworkInterfaceStatusApplyConfiguration) WithLastStateTransitionTime(value v1.Time) *NetworkInterfaceStatusApplyConfiguration {
	b.LastStateTransitionTime = &value
	return b
}

// WithPhase sets the Phase field in the declarative configuration to the given value
// and returns the receiver, so that objects can be built by chaining "With" function invocations.
// If called multiple times, the Phase field is set to the value of the last call.
func (b *NetworkInterfaceStatusApplyConfiguration) WithPhase(value computev1alpha1.NetworkInterfacePhase) *NetworkInterfaceStatusApplyConfiguration {
	b.Phase = &value
	return b
}

// WithLastPhaseTransitionTime sets the LastPhaseTransitionTime field in the declarative configuration to the given value
// and returns the receiver, so that objects can be built by chaining "With" function invocations.
// If called multiple times, the LastPhaseTransitionTime field is set to the value of the last call.
func (b *NetworkInterfaceStatusApplyConfiguration) WithLastPhaseTransitionTime(value v1.Time) *NetworkInterfaceStatusApplyConfiguration {
	b.LastPhaseTransitionTime = &value
	return b
}
