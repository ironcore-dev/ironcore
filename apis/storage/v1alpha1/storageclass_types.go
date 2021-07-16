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
	"k8s.io/apimachinery/pkg/util/intstr"
)

// StorageClassSpec defines the desired state of StorageClass
type StorageClassSpec struct {
	// Capabilities describes the capabilities of a storage class
	Capabilities []StorageClassCapability `json:"capabilities"`
	// Description is a human readable description of a storage class
	Description string `json:"description,omitempty"`
}

// StorageClassCapability describes one attribute of the StorageClass
type StorageClassCapability struct {
	// Name is the name of a capability
	Name string `json:"name"`
	// Value is the value of a capability
	Value intstr.IntOrString `json:"value"`
}

// StorageClassStatus defines the observed state of StorageClass
type StorageClassStatus struct {
	// Availability describes the regions and zones where this MachineClass is available
	Availability common.Availability `json:"availability,omitempty"`
}

//+kubebuilder:object:root=true
//+kubebuilder:subresource:status
//+kubebuilder:printcolumn:name="Age",type=string,JSONPath=`.metadata.CreationTimestamp`

// StorageClass is the Schema for the storageclasses API
type StorageClass struct {
	metav1.TypeMeta   `json:",inline"`
	metav1.ObjectMeta `json:"metadata,omitempty"`

	Spec   StorageClassSpec   `json:"spec,omitempty"`
	Status StorageClassStatus `json:"status,omitempty"`
}

//+kubebuilder:object:root=true

// StorageClassList contains a list of StorageClass
type StorageClassList struct {
	metav1.TypeMeta `json:",inline"`
	metav1.ListMeta `json:"metadata,omitempty"`
	Items           []StorageClass `json:"items"`
}

func init() {
	SchemeBuilder.Register(&StorageClass{}, &StorageClassList{})
}
