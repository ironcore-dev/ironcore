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
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

// VirtualIPSpec defines the desired state of VirtualIP
type VirtualIPSpec struct {
	// Type is the type of VirtualIP.
	Type VirtualIPType
	// IPFamily is the ip family of the VirtualIP.
	IPFamily corev1.IPFamily
	// NetworkInterfaceSelector selects any NetworkInterface that should get the VirtualIP routed.
	// If empty, it is assumed that an external process manages the VirtualIPRouting for this VirtualIP.
	NetworkInterfaceSelector *metav1.LabelSelector
}

// VirtualIPType is a type of VirtualIP.
type VirtualIPType string

const (
	// VirtualIPTypePublic is a VirtualIP that allocates and routes a stable public IP.
	VirtualIPTypePublic VirtualIPType = "Public"
)

// VirtualIPStatus defines the observed state of VirtualIP
type VirtualIPStatus struct {
	// IP is the allocated IP, if any.
	IP *commonv1alpha1.IP
}

// +genclient
// +k8s:deepcopy-gen:interfaces=k8s.io/apimachinery/pkg/runtime.Object

// VirtualIP is the Schema for the virtualips API
type VirtualIP struct {
	metav1.TypeMeta
	metav1.ObjectMeta

	Spec   VirtualIPSpec
	Status VirtualIPStatus
}

// +k8s:deepcopy-gen:interfaces=k8s.io/apimachinery/pkg/runtime.Object

// VirtualIPList contains a list of VirtualIP
type VirtualIPList struct {
	metav1.TypeMeta
	metav1.ListMeta
	Items []VirtualIP
}
