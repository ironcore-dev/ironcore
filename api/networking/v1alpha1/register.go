/*
 * Copyright (c) 2021 by the OnMetal authors.
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

// Package v1alpha1 contains API Schema definitions for the networking v1alpha1 API group
// +groupName=networking.api.onmetal.de
package v1alpha1

import (
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apimachinery/pkg/runtime/schema"
)

var (
	// SchemeGroupVersion is group version used to register these objects
	SchemeGroupVersion = schema.GroupVersion{Group: "networking.api.onmetal.de", Version: "v1alpha1"}

	// SchemeBuilder is used to add go types to the GroupVersionKind scheme
	SchemeBuilder = runtime.NewSchemeBuilder(addKnownTypes)

	// AddToScheme adds the types in this group-version to the given scheme.
	AddToScheme = SchemeBuilder.AddToScheme
)

func Resource(name string) schema.GroupResource {
	return schema.GroupResource{
		Group:    SchemeGroupVersion.Group,
		Resource: name,
	}
}

func addKnownTypes(scheme *runtime.Scheme) error {
	scheme.AddKnownTypes(SchemeGroupVersion,
		&Network{},
		&NetworkList{},
		&NetworkInterface{},
		&NetworkInterfaceList{},
		&VirtualIP{},
		&VirtualIPList{},
		&AliasPrefix{},
		&AliasPrefixList{},
		&AliasPrefixRouting{},
		&AliasPrefixRoutingList{},
		&LoadBalancer{},
		&LoadBalancerList{},
		&LoadBalancerRouting{},
		&LoadBalancerRoutingList{},
		&NATGateway{},
		&NATGatewayList{},
		&NATGatewayRouting{},
		&NATGatewayRoutingList{},
	)
	metav1.AddToGroupVersion(scheme, SchemeGroupVersion)
	return nil
}
