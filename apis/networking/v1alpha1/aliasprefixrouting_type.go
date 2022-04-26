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

// AliasPrefixRoutingSubset is one of the targets of a AliasPrefixRouting
type AliasPrefixRoutingSubset struct {
	// TargetRef is the targeted entity
	TargetRef commonv1alpha1.LocalUIDReference `json:"targetRef"`
}

// +genclient
// +k8s:deepcopy-gen:interfaces=k8s.io/apimachinery/pkg/runtime.Object

// AliasPrefixRouting is the Schema for the aliasprefixrouting API
type AliasPrefixRouting struct {
	metav1.TypeMeta   `json:",inline"`
	metav1.ObjectMeta `json:"metadata,omitempty"`

	// NetworkRef is the Network this AliasPrefixRouting should belong to
	NetworkRef corev1.LocalObjectReference `json:"networkRef"`
	// Subsets are the subsets that make up an AliasPrefixRouting
	Subsets []AliasPrefixRoutingSubset `json:"subsets,omitempty"`
}

// +k8s:deepcopy-gen:interfaces=k8s.io/apimachinery/pkg/runtime.Object

// AliasPrefixRoutingList contains a list of AliasPrefixRouting
type AliasPrefixRoutingList struct {
	metav1.TypeMeta `json:",inline"`
	metav1.ListMeta `json:"metadata,omitempty"`
	Items           []AliasPrefixRouting `json:"items"`
}
