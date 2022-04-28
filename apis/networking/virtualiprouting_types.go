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

package networking

import (
	commonv1alpha1 "github.com/onmetal/onmetal-api/apis/common/v1alpha1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

// +genclient
// +k8s:deepcopy-gen:interfaces=k8s.io/apimachinery/pkg/runtime.Object

// VirtualIPRouting is the Schema for the virtualiproutings API
type VirtualIPRouting struct {
	metav1.TypeMeta
	metav1.ObjectMeta

	// Subsets are the subsets that make up a VirtualIPRouting.
	Subsets []VirtualIPRoutingSubset
}

// VirtualIPRoutingSubset is one of the targets of a VirtualIPRouting.
type VirtualIPRoutingSubset struct {
	// NetworkRef is the network all targets are in.
	NetworkRef commonv1alpha1.LocalUIDReference
	// Targets are the targets of the virtual IP.
	Targets []VirtualIPRoutingSubsetTarget
}

type VirtualIPRoutingSubsetTarget struct {
	commonv1alpha1.LocalUIDReference
	// IP is the target ip to route to.
	IP commonv1alpha1.IP
}

// +k8s:deepcopy-gen:interfaces=k8s.io/apimachinery/pkg/runtime.Object

// VirtualIPRoutingList contains a list of VirtualIPRouting
type VirtualIPRoutingList struct {
	metav1.TypeMeta
	metav1.ListMeta
	Items []VirtualIPRouting
}
