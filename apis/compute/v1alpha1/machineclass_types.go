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

// MachineClassSpec defines the desired state of MachineClass
type MachineClassSpec struct {
	// Capabilities describes the features of the MachineClass
	Capabilities []Capability `json:"capabilities"`
}

// Capability describes a single feature of a MachineClass
type Capability struct {
	Name  string `json:"name,omitempty"`
	Type  string `json:"type,omitempty"`
	Value string `json:"value,omitempty"`
}

// MachineClassStatus defines the observed state of MachineClass
type MachineClassStatus struct {
	// Availability describes the regions and zones where this MachineClass is available
	Availability common.Availability `json:"availability,omitempty"`
}

//+kubebuilder:object:root=true
//+kubebuilder:subresource:status
//+kubebuilder:printcolumn:name="Age",type=string,JSONPath=`.metadata.CreationTimestamp`

// MachineClass is the Schema for the machineclasses API
type MachineClass struct {
	metav1.TypeMeta   `json:",inline"`
	metav1.ObjectMeta `json:"metadata,omitempty"`

	Spec   MachineClassSpec   `json:"spec,omitempty"`
	Status MachineClassStatus `json:"status,omitempty"`
}

//+kubebuilder:object:root=true

// MachineClassList contains a list of MachineClass
type MachineClassList struct {
	metav1.TypeMeta `json:",inline"`
	metav1.ListMeta `json:"metadata,omitempty"`
	Items           []MachineClass `json:"items"`
}

func init() {
	SchemeBuilder.Register(&MachineClass{}, &MachineClassList{})
}
