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
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/types"
)

// +genclient
// +k8s:deepcopy-gen:interfaces=k8s.io/apimachinery/pkg/runtime.Object

// VirtualIPRouting is the Schema for the virtualiproutings API
type VirtualIPRouting struct {
	metav1.TypeMeta   `json:",inline"`
	metav1.ObjectMeta `json:"metadata,omitempty"`

	// Subsets are the subsets that make up a VirtualIPRouting.
	Subsets []VirtualIPRoutingSubset `json:"subsets,omitempty"`
}

// LocalUIDReference is a reference to another entity including its UID.
type LocalUIDReference struct {
	// Name is the name of the referenced entity.
	Name string `json:"name"`
	// UID is the UID of the referenced entity.
	UID types.UID `json:"uid"`
}

// VirtualIPRoutingSubset is one of the targets of a VirtualIPRouting.
type VirtualIPRoutingSubset struct {
	// NetworkRef is the network all targets are in.
	NetworkRef LocalUIDReference `json:"networkRef"`
	// Targets are the targets of the virtual IP.
	Targets []VirtualIPRoutingSubsetTarget `json:"targets"`
}

type VirtualIPRoutingSubsetTarget struct {
	LocalUIDReference `json:",inline"`
	// IP is the target ip to route to.
	IP commonv1alpha1.IP `json:"ip"`
}

// +k8s:deepcopy-gen:interfaces=k8s.io/apimachinery/pkg/runtime.Object

// VirtualIPRoutingList contains a list of VirtualIPRouting
type VirtualIPRoutingList struct {
	metav1.TypeMeta `json:",inline"`
	metav1.ListMeta `json:"metadata,omitempty"`
	Items           []VirtualIPRouting `json:"items"`
}
