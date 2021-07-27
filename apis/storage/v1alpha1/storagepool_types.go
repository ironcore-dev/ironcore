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

// StoragePoolSpec defines the desired state of StoragePool
type StoragePoolSpec struct {
	// Region defines the region of the storage pool
	Region string `json:"region,omitempty"`
	// Privacy defines the privacy scope of the storage pool
	Privacy string `json:"privacy"`
	// Replication indicates the replication factor in the storage pool
	// +kubebuilder:validation:Minimum:=0
	Replication int `json:"replication,omitempty"`
	// Capacity list the available capacity of a storage pool
	Capacity []StorageClassCapacity `json:"capacity,omitempty"`
}

const (
	PrivacyShared  = "shared"
	PrivacyDisk    = "disk"
	PrivacyCluster = "cluster"
)

// StorageClassCapacity defines capacity attribute of a storage class
type StorageClassCapacity struct {
	// Name is the name of a storage class capacity
	Name string `json:"name"`
	// Capacity is the quantity of a capacity
	Capacity *resource.Quantity `json:"capacity,omitempty"`
}

// StoragePoolStatus defines the observed state of StoragePool
type StoragePoolStatus struct {
	common.StateFields `json:",inline"`
	// Used indicates how much capacity has been used in a storage pool
	Used []StorageClassCapacity `json:"used,omitempty"`
}

const (
	StoragePoolStateAvailable    = "Available"
	StoragePoolStatePending      = "Pending"
	StoragePoolStateNotAvailable = "NotAvailable"
)

//+kubebuilder:object:root=true
//+kubebuilder:subresource:status
//+kubebuilder:printcolumn:name="StateFields",type=string,JSONPath=`.status.state`
//+kubebuilder:printcolumn:name="Age",type=date,JSONPath=`.metadata.creationTimestamp`

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
