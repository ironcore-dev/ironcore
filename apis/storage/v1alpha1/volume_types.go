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

import (
	common "github.com/onmetal/onmetal-api/apis/common/v1alpha1"
	"k8s.io/apimachinery/pkg/api/resource"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

// VolumeSpec defines the desired state of Volume
type VolumeSpec struct {
	// StorageClass is the storage class of a volume
	StorageClass string `json:"storage_class"`
	// StoragePool indicates which storage pool to use for a volume
	StoragePool common.ScopeReference `json:"storagepool"`
	// Size defines the size of the volume
	Size *resource.Quantity `json:"size"`
}

// VolumeStatus defines the observed state of Volume
type VolumeStatus struct {
	common.StateFields `json:",inline"`
}

const (
	VolumeStateAvailable = "Available"
	VolumeStatePending   = "Pending"
	VolumeStateAttached  = "Attached"
	VolumeStateError     = "Error"
)

//+kubebuilder:object:root=true
//+kubebuilder:subresource:status

// Volume is the Schema for the volumes API
type Volume struct {
	metav1.TypeMeta   `json:",inline"`
	metav1.ObjectMeta `json:"metadata,omitempty"`

	Spec   VolumeSpec   `json:"spec,omitempty"`
	Status VolumeStatus `json:"status,omitempty"`
}

//+kubebuilder:object:root=true

// VolumeList contains a list of Volume
type VolumeList struct {
	metav1.TypeMeta `json:",inline"`
	metav1.ListMeta `json:"metadata,omitempty"`
	Items           []Volume `json:"items"`
}

func init() {
	SchemeBuilder.Register(&Volume{}, &VolumeList{})
}
