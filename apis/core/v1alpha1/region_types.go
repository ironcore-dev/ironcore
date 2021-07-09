/*
Copyright 2021.

Licensed under the Apache License, Version 2.0 (the "License");
you may not use this file except in compliance with the License.
You may obtain a copy of the License at

    http://www.apache.org/licenses/LICENSE-2.0

Unless required by applicable law or agreed to in writing, software
distributed under the License is distributed on an "AS IS" BASIS,
WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
See the License for the specific language governing permissions and
limitations under the License.
*/

package v1alpha1

import (
	common "github.com/onmetal/onmetal-api/apis/common/v1alpha1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

// RegionSpec defines the desired state of Region
type RegionSpec struct {
	// Location describes the physical location of the region
	Location string `json:"location,omitempty"`
	// AvailabilityZones represents the availability zones in a given region
	AvailabilityZones []string `json:"availabiltyZone"`
}

// RegionStatus defines the observed state of Region
type RegionStatus struct {
	common.StateFields `json:",inline"`
	AvailabilityZones  []AZState `json:"availabilityZones,omitempty"`
}

// ZoneState describes the state of an AvailabilityZone within a region
type AZState struct {
	Name               string `json:"name,omitempty"`
	common.StateFields `json:",inline"`
}

const (
	// ZoneStateActive represents the active state of an AZ
	ZoneStateActive = "Active"
	// ZoneStateOffline represents the offline state of an AZ
	ZoneStateOffline = "Offline"
	// RegionStateActive represents the active state of a region
	RegionStateActive = "Active"
	// RegionStateOffline represents the offline state of a region
	RegionStateOffline = "Offline"
)

//+kubebuilder:object:root=true
//+kubebuilder:subresource:status
//+kubebuilder:resource:scope=Cluster
//+kubebuilder:printcolumn:name="StateFields",type=string,JSONPath=`.status.state`
//+kubebuilder:printcolumn:name="Age",type=string,JSONPath=`.metadata.CreationTimestamp`

// Region is the Schema for the regions API
type Region struct {
	metav1.TypeMeta   `json:",inline"`
	metav1.ObjectMeta `json:"metadata,omitempty"`

	Spec   RegionSpec   `json:"spec,omitempty"`
	Status RegionStatus `json:"status,omitempty"`
}

//+kubebuilder:object:root=true

// RegionList contains a list of Region
type RegionList struct {
	metav1.TypeMeta `json:",inline"`
	metav1.ListMeta `json:"metadata,omitempty"`
	Items           []Region `json:"items"`
}

func init() {
	SchemeBuilder.Register(&Region{}, &RegionList{})
}
