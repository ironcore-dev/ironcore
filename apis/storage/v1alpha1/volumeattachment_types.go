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

// VolumeAttachmentSpec defines the desired state of VolumeAttachment
type VolumeAttachmentSpec struct {
	Volume  common.ScopeReference `json:"volume,omitempty"`
	Machine common.ScopeReference `json:"machine"`
	Device  string                `json:"device,omitempty"`
	Source  VolumeSource          `json:"source,omitempty"`
}

type VolumeSource struct {
	Image    string `json:"image,omitempty"`
	Snapshot string `json:"snapshot,omitempty"`
}

// VolumeAttachmentStatus defines the observed state of VolumeAttachment
type VolumeAttachmentStatus struct {
	common.StateFields `json:",inline"`
	Device             string `json:"device,omitempty"`
}

const (
	VolumeAttachmentStateAttachde = "Attached"
	VolumeAttachmentStateInvalid  = "Invalid"
	VolumeAttachmentStateError    = "Error"
)

//+kubebuilder:object:root=true
//+kubebuilder:subresource:status
//+kubebuilder:printcolumn:name="Volume",type=string,JSONPath=`.spec.volume.name`
//+kubebuilder:printcolumn:name="Machine",type=string,JSONPath=`.spec.machine.name`
//+kubebuilder:printcolumn:name="Device",type=string,JSONPath=`.spec.device`
//+kubebuilder:printcolumn:name="Image",type=string,JSONPath=`.spec.source.image`
//+kubebuilder:printcolumn:name="Snapshot",type=string,JSONPath=`.spec.source.snapshot`
//+kubebuilder:printcolumn:name="StateFields",type=string,JSONPath=`.status.state`
//+kubebuilder:printcolumn:name="Age",type=string,JSONPath=`.metadata.CreationTimestamp`

// VolumeAttachment is the Schema for the volumeattachments API
type VolumeAttachment struct {
	metav1.TypeMeta   `json:",inline"`
	metav1.ObjectMeta `json:"metadata,omitempty"`

	Spec   VolumeAttachmentSpec   `json:"spec,omitempty"`
	Status VolumeAttachmentStatus `json:"status,omitempty"`
}

//+kubebuilder:object:root=true

// VolumeAttachmentList contains a list of VolumeAttachment
type VolumeAttachmentList struct {
	metav1.TypeMeta `json:",inline"`
	metav1.ListMeta `json:"metadata,omitempty"`
	Items           []VolumeAttachment `json:"items"`
}

func init() {
	SchemeBuilder.Register(&VolumeAttachment{}, &VolumeAttachmentList{})
}
