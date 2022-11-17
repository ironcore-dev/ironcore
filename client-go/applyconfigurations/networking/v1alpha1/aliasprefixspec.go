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
	metav1 "github.com/onmetal/onmetal-api/client-go/applyconfigurations/meta/v1"
	v1 "k8s.io/api/core/v1"
)

// AliasPrefixSpecApplyConfiguration represents an declarative configuration of the AliasPrefixSpec type for use
// with apply.
type AliasPrefixSpecApplyConfiguration struct {
	NetworkRef               *v1.LocalObjectReference                `json:"networkRef,omitempty"`
	NetworkInterfaceSelector *metav1.LabelSelectorApplyConfiguration `json:"networkInterfaceSelector,omitempty"`
	Prefix                   *PrefixSourceApplyConfiguration         `json:"prefix,omitempty"`
}

// AliasPrefixSpecApplyConfiguration constructs an declarative configuration of the AliasPrefixSpec type for use with
// apply.
func AliasPrefixSpec() *AliasPrefixSpecApplyConfiguration {
	return &AliasPrefixSpecApplyConfiguration{}
}

// WithNetworkRef sets the NetworkRef field in the declarative configuration to the given value
// and returns the receiver, so that objects can be built by chaining "With" function invocations.
// If called multiple times, the NetworkRef field is set to the value of the last call.
func (b *AliasPrefixSpecApplyConfiguration) WithNetworkRef(value v1.LocalObjectReference) *AliasPrefixSpecApplyConfiguration {
	b.NetworkRef = &value
	return b
}

// WithNetworkInterfaceSelector sets the NetworkInterfaceSelector field in the declarative configuration to the given value
// and returns the receiver, so that objects can be built by chaining "With" function invocations.
// If called multiple times, the NetworkInterfaceSelector field is set to the value of the last call.
func (b *AliasPrefixSpecApplyConfiguration) WithNetworkInterfaceSelector(value *metav1.LabelSelectorApplyConfiguration) *AliasPrefixSpecApplyConfiguration {
	b.NetworkInterfaceSelector = value
	return b
}

// WithPrefix sets the Prefix field in the declarative configuration to the given value
// and returns the receiver, so that objects can be built by chaining "With" function invocations.
// If called multiple times, the Prefix field is set to the value of the last call.
func (b *AliasPrefixSpecApplyConfiguration) WithPrefix(value *PrefixSourceApplyConfiguration) *AliasPrefixSpecApplyConfiguration {
	b.Prefix = value
	return b
}
