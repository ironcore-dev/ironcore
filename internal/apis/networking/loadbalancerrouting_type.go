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
	commonv1alpha1 "github.com/onmetal/onmetal-api/api/common/v1alpha1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

// +genclient
// +k8s:deepcopy-gen:interfaces=k8s.io/apimachinery/pkg/runtime.Object

// LoadBalancerRouting is the Schema for the loadbalancerroutings API
type LoadBalancerRouting struct {
	metav1.TypeMeta
	metav1.ObjectMeta

	// NetworkRef is the network the load balancer is assigned to.
	NetworkRef commonv1alpha1.LocalUIDReference

	// Destinations are the destinations for an LoadBalancer.
	Destinations []commonv1alpha1.LocalUIDReference
}

// +k8s:deepcopy-gen:interfaces=k8s.io/apimachinery/pkg/runtime.Object

// LoadBalancerRoutingList contains a list of LoadBalancerRouting
type LoadBalancerRoutingList struct {
	metav1.TypeMeta
	metav1.ListMeta
	Items []LoadBalancerRouting
}
