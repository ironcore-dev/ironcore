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

// ImageSpec defines the desired state of Image
type ImageSpec struct {
	Type           string            `json:"type,omitempty"`
	Maturity       string            `json:"maturity"`
	ExpirationTime metav1.Time       `json:"expirationTime,omitempty"`
	OS             string            `json:"os"`
	Version        string            `json:"version"`
	Source         []SourceAttribute `json:"source"`
}

// SourceAttribute describes the source components of an Image
type SourceAttribute struct {
	Name      string `json:"name"`
	ImageName string `json:"imageName,omitempty"`
	URL       string `json:"url,omitempty"`
	CmdLine   string `json:"cmdLine,omitempty"`
	Hash      *Hash  `json:"hash,omitempty"`
}

// Hash describes a hash value and it's corresponding algorithm
type Hash struct {
	Algorithm string `json:"algorithm"`
	Value     string `json:"value"`
}

// ImageStatus defines the observed state of Image
type ImageStatus struct {
	common.StateFields `json:",inline"`
	Hashes             []HashStatus  `json:"hashes,omitempty"`
	Regions            []RegionState `json:"regions,omitempty"`
}

type RegionState struct {
	Name               string `json:"name"`
	common.StateFields `json:",inline"`
}

type HashStatus struct {
	Name string `json:"name"`
	Hash string `json:"hash"`
}

const (
	ImageStateValid    = "Valid"
	ImageStateInvalid  = "Invalid"
	RegionStateReady   = "Ready"
	RegionStatePending = "Pending"
	RegionStateError   = "Error"
)

//+kubebuilder:object:root=true
//+kubebuilder:subresource:status
//+kubebuilder:printcolumn:name="OS",type=string,JSONPath=`.spec.os`
//+kubebuilder:printcolumn:name="Version",type=string,JSONPath=`.spec.version`
//+kubebuilder:printcolumn:name="Maturity",type=string,JSONPath=`.spec.maturity`
//+kubebuilder:printcolumn:name="Expiration",type=date,JSONPath=`.spec.expirationTime`
//+kubebuilder:printcolumn:name="Type",type=string,JSONPath=`.spec.type`,priority=100
//+kubebuilder:printcolumn:name="StateFields",type=string,JSONPath=`.status.state`
//+kubebuilder:printcolumn:name="Age",type=string,JSONPath=`.metadata.CreationTimestamp`

// Image is the Schema for the images API
type Image struct {
	metav1.TypeMeta   `json:",inline"`
	metav1.ObjectMeta `json:"metadata,omitempty"`

	Spec   ImageSpec   `json:"spec,omitempty"`
	Status ImageStatus `json:"status,omitempty"`
}

//+kubebuilder:object:root=true

// ImageList contains a list of Image
type ImageList struct {
	metav1.TypeMeta `json:",inline"`
	metav1.ListMeta `json:"metadata,omitempty"`
	Items           []Image `json:"items"`
}

func init() {
	SchemeBuilder.Register(&Image{}, &ImageList{})
}
