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

// SSHPublicKeySpec defines the desired state of SSHPublicKey
type SSHPublicKeySpec struct {
	SSHPublicKey   string      `json:"sshPublicKey"`
	Description    string      `json:"description,omitempty"`
	ExpirationDate metav1.Time `json:"expirationDate,omitempty"`
}

// SSHPublicKeyStatus defines the observed state of SSHPublicKey
type SSHPublicKeyStatus struct {
	common.StateFields `json:",inline"`
	FingerPrint        string `json:"fingerPrint,omitempty"`
	KeyLength          uint   `json:"keyLength,omitempty"`
	Algorithm          string `json:"algorithm,omitempty"`
	// PublicKey is the PEM encoded public key
	PublicKey string `json:"publicKey,omitempty"`
}

//+kubebuilder:object:root=true
//+kubebuilder:subresource:status
//+kubebuilder:printcolumn:name="Expiration",type=date,JSONPath=`.spec.expirationDate`
//+kubebuilder:printcolumn:name="Algorithm",type=string,JSONPath=`.status.algorithm`
//+kubebuilder:printcolumn:name="KeyLength",type=integer,JSONPath=`.status.keyLength`
//+kubebuilder:printcolumn:name="StateFields",type=string,JSONPath=`.status.state`
//+kubebuilder:printcolumn:name="Age",type=string,JSONPath=`.metadata.CreationTimestamp`

// SSHPublicKey is the Schema for the sshpublickeys API
type SSHPublicKey struct {
	metav1.TypeMeta   `json:",inline"`
	metav1.ObjectMeta `json:"metadata,omitempty"`

	Spec   SSHPublicKeySpec   `json:"spec,omitempty"`
	Status SSHPublicKeyStatus `json:"status,omitempty"`
}

//+kubebuilder:object:root=true

// SSHPublicKeyList contains a list of SSHPublicKey
type SSHPublicKeyList struct {
	metav1.TypeMeta `json:",inline"`
	metav1.ListMeta `json:"metadata,omitempty"`
	Items           []SSHPublicKey `json:"items"`
}

func init() {
	SchemeBuilder.Register(&SSHPublicKey{}, &SSHPublicKeyList{})
}
