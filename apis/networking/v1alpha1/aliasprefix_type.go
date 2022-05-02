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

package v1alpha1

import (
	commonv1alpha1 "github.com/onmetal/onmetal-api/apis/common/v1alpha1"
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

// AliasPrefixSpec defines the desired state of AliasPrefix
type AliasPrefixSpec struct {
	// NetworkRef is the Network this AliasPrefix should belong to
	NetworkRef corev1.LocalObjectReference `json:"networkRef"`
	// NetworkInterfaceSelector defines the NetworkInterfaces
	// for which this AliasPrefix should be applied
	NetworkInterfaceSelector *metav1.LabelSelector `json:"networkInterfaceSelector,omitempty"`
	// Prefix is the provided Prefix or EphemeralPrefix which
	// should be used by this AliasPrefix
	Prefix PrefixSource `json:"prefix,omitempty"`
}

// PrefixSource is the source of the Prefix definition in an AliasPrefix
type PrefixSource struct {
	// Value is a single IPPrefix value as defined in the AliasPrefix
	Value *commonv1alpha1.IPPrefix `json:"value,omitempty"`
	// EphemeralPrefix defines the Prefix which should be allocated by the AliasPrefix
	EphemeralPrefix *EphemeralPrefixSource `json:"ephemeralPrefix,omitempty"`
}

// AliasPrefixStatus defines the observed state of AliasPrefix
type AliasPrefixStatus struct {
	// Prefix is the Prefix reserved by this AliasPrefix
	Prefix *commonv1alpha1.IPPrefix `json:"prefix,omitempty"`
}

// +genclient
// +k8s:deepcopy-gen:interfaces=k8s.io/apimachinery/pkg/runtime.Object

// AliasPrefix is the Schema for the AliasPrefix API
type AliasPrefix struct {
	metav1.TypeMeta   `json:",inline"`
	metav1.ObjectMeta `json:"metadata,omitempty"`

	Spec   AliasPrefixSpec   `json:"spec,omitempty"`
	Status AliasPrefixStatus `json:"status,omitempty"`
}

// +k8s:deepcopy-gen:interfaces=k8s.io/apimachinery/pkg/runtime.Object

// AliasPrefixList contains a list of AliasPrefix
type AliasPrefixList struct {
	metav1.TypeMeta `json:",inline"`
	metav1.ListMeta `json:"metadata,omitempty"`
	Items           []AliasPrefix `json:"items"`
}
