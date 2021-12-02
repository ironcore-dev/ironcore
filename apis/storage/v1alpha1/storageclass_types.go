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
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

var (
	// StorageClassFinalizer
	StorageClassFinalizer = GroupVersion.Group + "/storageclass"
)

// StorageClassSpec defines the desired state of StorageClass
type StorageClassSpec struct {
	// Capabilities describes the capabilities of a storage class
	Capabilities corev1.ResourceList `json:"capabilities,omitempty"`
}

// StorageClassStatus defines the observed state of StorageClass
type StorageClassStatus struct {
}

//+kubebuilder:object:root=true
//+kubebuilder:resource:scope=Cluster
//+kubebuilder:subresource:status
//+kubebuilder:printcolumn:name="Age",type=date,JSONPath=`.metadata.creationTimestamp`

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
