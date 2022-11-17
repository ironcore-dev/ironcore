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

const (
	// ResourceTPS defines max throughput per second. (e.g. 1Gi)
	ResourceTPS corev1.ResourceName = "tps"
	// ResourceIOPS defines max IOPS in input/output operations per second.
	ResourceIOPS corev1.ResourceName = "iops"
)

var (
	// VolumeClassFinalizer is the finalizer for VolumeClass.
	VolumeClassFinalizer = SchemeGroupVersion.Group + "/volumeclass"
)

// +k8s:deepcopy-gen:interfaces=k8s.io/apimachinery/pkg/runtime.Object
// +genclient
// +genclient:nonNamespaced
// +genclient:noStatus

// VolumeClass is the Schema for the volumeclasses API
type VolumeClass struct {
	metav1.TypeMeta   `json:",inline"`
	metav1.ObjectMeta `json:"metadata,omitempty"`
	// Capabilities describes the capabilities of a VolumeClass.
	Capabilities corev1.ResourceList `json:"capabilities,omitempty"`
}

// +k8s:deepcopy-gen:interfaces=k8s.io/apimachinery/pkg/runtime.Object

// VolumeClassList contains a list of VolumeClass
type VolumeClassList struct {
	metav1.TypeMeta `json:",inline"`
	metav1.ListMeta `json:"metadata,omitempty"`
	Items           []VolumeClass `json:"items"`
}
