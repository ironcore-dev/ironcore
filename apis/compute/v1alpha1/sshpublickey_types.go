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
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

// DefaultSSHPublicKeyDataKey is the default data key for the common.ConfigMapKeySelector in an SSHPublicKeySpec.
const DefaultSSHPublicKeyDataKey = "id_rsa.pub"

// SSHPublicKeySpec defines the desired state of SSHPublicKey
type SSHPublicKeySpec struct {
	// ConfigMapRef is the reference to a key in a ConfigMap resource
	// containing the public key data. If key is not specified, it defaults to 'id_rsa.pub'.
	ConfigMapRef common.ConfigMapKeySelector `json:"configMapRef"`
}

// SSHPublicKeyStatus defines the observed state of SSHPublicKey
type SSHPublicKeyStatus struct {
	// Conditions are a list of elements indicating the status of various properties of an SSHPublicKey.
	Conditions []SSHPublicKeyCondition `json:"conditions,omitempty"`
	// Algorithm is the algorithm used for the public key.
	Algorithm string `json:"algorithm,omitempty"`
	// Fingerprint is the fingerprint of the ssh public key.
	Fingerprint string `json:"fingerprint,omitempty"`
	// KeyLength is the byte length of the ssh key.
	// +kubebuilder:validation:Minimum:=0
	KeyLength int `json:"keyLength,omitempty"`
}

// SSHPublicKeyConditionType is a possible type of SSHPublicKeyCondition.
type SSHPublicKeyConditionType string

const (
	SSHPublicKeyAvailable SSHPublicKeyConditionType = "Available"
)

// SSHPublicKeyCondition is a condition of an SSHPublicKey.
type SSHPublicKeyCondition struct {
	// Type is the SSHPublicKeyConditionType of the condition.
	Type SSHPublicKeyConditionType `json:"type"`
	// Status is the corev1.ConditionStatus of the condition.
	Status corev1.ConditionStatus `json:"status"`
	// Reason is a machine-readable indication of why the condition is in a certain state.
	Reason string `json:"reason"`
	// Message is a human-readable explanation of why the condition has a certain reason / state.
	Message string `json:"message"`
	// ObservedGeneration represents the .metadata.generation that the condition was set based upon.
	ObservedGeneration int64 `json:"observedGeneration,omitempty"`
	// LastUpdateTime is the last time a condition has been updated.
	LastUpdateTime metav1.Time `json:"lastUpdateTime,omitempty"`
	// LastTransitionTime is the last time the status of a condition has transitioned from one state to another.
	LastTransitionTime metav1.Time `json:"lastTransitionTime,omitempty"`
}

//+kubebuilder:object:root=true
//+kubebuilder:subresource:status
//+kubebuilder:printcolumn:name="Expiration",type=date,JSONPath=`.spec.expirationDate`
//+kubebuilder:printcolumn:name="Algorithm",type=string,JSONPath=`.status.algorithm`
//+kubebuilder:printcolumn:name="KeyLength",type=integer,JSONPath=`.status.keyLength`
//+kubebuilder:printcolumn:name="State",type=string,JSONPath=`.status.state`
//+kubebuilder:printcolumn:name="Age",type=date,JSONPath=`.metadata.creationTimestamp`

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
