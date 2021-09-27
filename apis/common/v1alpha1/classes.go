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

package v1alpha1

import corev1 "k8s.io/api/core/v1"

//+kubebuilder:object:generate=true

type Availability []RegionAvailability

//+kubebuilder:object:generate=true

// RegionAvailability defines a region with its availability zones
type RegionAvailability struct {
	// Region is the name of the region
	Region string `json:"region"`
	// Zones is a list of zones in this region
	//+optional
	Zones []ZoneAvailability `json:"availabilityZones,omitempty"`
}

//+kubebuilder:object:generate=true

// Location describes the location of a resource
type Location struct {
	// Region defines the region of a resource
	Region string `json:"region"`
	// AvailabilityZone is the availability zone of a resource
	//+optional
	AvailabilityZone string `json:"availabilityZone,omitempty"`
}

//+kubebuilder:object:generate=true

// ZoneAvailability defines the name of a zone
type ZoneAvailability struct {
	// Name is the name of the availability zone
	Name string `json:"name"`
}

// ConfigMapKeySelector is a reference to a specific 'key' within a ConfigMap resource.
// In some instances, `key` is a required field.
type ConfigMapKeySelector struct {
	// The name of the ConfigMap resource being referred to.
	corev1.LocalObjectReference `json:",inline"`
	// The key of the entry in the ConfigMap resource's `data` field to be used.
	// Some instances of this field may be defaulted, in others it may be
	// required.
	// +optional
	Key string `json:"key,omitempty"`
}

// TODO: create marshal/unmarshal functions
type IPAddr string

// TODO: create marshal/unmarshal functions
type Cidr string
