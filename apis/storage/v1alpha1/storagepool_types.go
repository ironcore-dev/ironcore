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
	"k8s.io/apimachinery/pkg/api/resource"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

// StoragePoolSpec defines the desired state of StoragePool
type StoragePoolSpec struct {
	Privacy      PrivacyType            `json:"privacy"`
	Replication  int                    `json:"replication"`
	Reservations []StorageClassCapacity `json:"reservations,omitempty"`
}

type PrivacyType string

const (
	PrivacyShared  PrivacyType = "shared"
	PrivacyDisk    PrivacyType = "disk"
	PrivacyCluster PrivacyType = "cluster"
)

type StorageClassCapacity struct {
	Name     string            `json:"name"`
	Scope    *string           `json:"scope,omitempty"`
	Capacity resource.Quantity `json:"capacity"`
}

// StoragePoolStatus defines the observed state of StoragePool
type StoragePoolStatus struct {
	State   *string                `json:"state,omitempty"`
	Message *string                `json:"message,omitempty"`
	Used    []StorageClassCapacity `json:"used,omitempty"`
}

//+kubebuilder:object:root=true
//+kubebuilder:subresource:status

// StoragePool is the Schema for the storagepools API
type StoragePool struct {
	metav1.TypeMeta   `json:",inline"`
	metav1.ObjectMeta `json:"metadata,omitempty"`

	Spec   StoragePoolSpec   `json:"spec,omitempty"`
	Status StoragePoolStatus `json:"status,omitempty"`
}

//+kubebuilder:object:root=true

// StoragePoolList contains a list of StoragePool
type StoragePoolList struct {
	metav1.TypeMeta `json:",inline"`
	metav1.ListMeta `json:"metadata,omitempty"`
	Items           []StoragePool `json:"items"`
}

func init() {
	SchemeBuilder.Register(&StoragePool{}, &StoragePoolList{})
}
