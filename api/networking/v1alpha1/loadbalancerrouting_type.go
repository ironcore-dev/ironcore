/*
 * Copyright (c) 2022 by the IronCore authors.
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
	commonv1alpha1 "github.com/ironcore-dev/ironcore/api/common/v1alpha1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/types"
)

// +genclient
// +k8s:deepcopy-gen:interfaces=k8s.io/apimachinery/pkg/runtime.Object

// LoadBalancerRouting is the Schema for the loadbalancerroutings API
type LoadBalancerRouting struct {
	metav1.TypeMeta   `json:",inline"`
	metav1.ObjectMeta `json:"metadata,omitempty"`

	// NetworkRef is the network the load balancer is assigned to.
	NetworkRef commonv1alpha1.LocalUIDReference `json:"networkRef"`

	// Destinations are the destinations for an LoadBalancer.
	Destinations []LoadBalancerDestination `json:"destinations"`
}

// LoadBalancerDestination is the destination of the load balancer.
type LoadBalancerDestination struct {
	// IP is the target IP.
	IP commonv1alpha1.IP `json:"ip"`
	// TargetRef is the target providing the destination.
	TargetRef *LoadBalancerTargetRef `json:"targetRef,omitempty"`
}

// LoadBalancerTargetRef is a load balancer target.
type LoadBalancerTargetRef struct {
	// UID is the UID of the target.
	UID types.UID `json:"uid"`
	// Name is the name of the target.
	Name string `json:"name"`
	// ProviderID is the provider internal id of the target.
	ProviderID string `json:"providerID"`
}

// +k8s:deepcopy-gen:interfaces=k8s.io/apimachinery/pkg/runtime.Object

// LoadBalancerRoutingList contains a list of LoadBalancerRouting
type LoadBalancerRoutingList struct {
	metav1.TypeMeta `json:",inline"`
	metav1.ListMeta `json:"metadata,omitempty"`
	Items           []LoadBalancerRouting `json:"items"`
}
