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
	commonv1alpha1 "github.com/onmetal/onmetal-api/apis/common/v1alpha1"
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

// ConsoleSpec defines the desired state of Console
type ConsoleSpec struct {
	// Type is the ConsoleType to use.
	Type ConsoleType `json:"type"`
	// MachineRef references the machine to open a console to.
	MachineRef corev1.LocalObjectReference `json:"machineRef"`
	// LighthouseClientConfig is the ConsoleClientConfig for the machine to connect to a common lighthouse.
	LighthouseClientConfig *ConsoleClientConfig `json:"lighthouseClientConfig,omitempty"`
}

type ConsoleClientConfig struct {
	// Service is the service to connect to.
	Service *ServiceReference `json:"service,omitempty"`
	// URL is the url to connect to.
	URL *string `json:"url,omitempty"`
	// CABundle is a PEM encoded CA bundle which will be used to validate the endpoint's server certificate.
	// If unspecified, system trust roots on the machine will be used.
	CABundle []byte `json:"caBundle,omitempty"`
	// KeySecret is the key that will be looked up for a client key.
	KeySecret *commonv1alpha1.SecretKeySelector `json:"keySecret,omitempty"`
}

const DefaultKeySecretKey = "client.cert"

type ServiceReference struct {
	// Name of the referenced service.
	Name string `json:"name"`
	// `path` is an optional URL path which will be sent in any request to
	// this service.
	// +optional
	Path *string `json:"path,omitempty"`

	// Port on the service hosting the console.
	// Defaults to 443 for backward compatibility.
	// `port` should be a valid port number (1-65535, inclusive).
	Port *int32 `json:"port,omitempty"`
}

// ConsoleType represents the type of console.
// +kubebuilder:validation:Enum=Service;Lighthouse
type ConsoleType string

const (
	ConsoleTypeService    ConsoleType = "Service"
	ConsoleTypeLightHouse ConsoleType = "Lighthouse"
)

type ConsoleState string

const (
	ConsoleStatePending ConsoleState = "Pending"
	ConsoleStateReady   ConsoleState = "Ready"
	ConsoleStateError   ConsoleState = "Error"
)

// ConsoleStatus defines the observed state of Console
type ConsoleStatus struct {
	State               ConsoleState         `json:"state,omitempty"`
	ServiceClientConfig *ConsoleClientConfig `json:"serviceClientConfig,omitempty"`
}

//+kubebuilder:object:root=true
//+kubebuilder:subresource:status

// Console is the Schema for the consoles API
type Console struct {
	metav1.TypeMeta   `json:",inline"`
	metav1.ObjectMeta `json:"metadata,omitempty"`

	Spec   ConsoleSpec   `json:"spec,omitempty"`
	Status ConsoleStatus `json:"status,omitempty"`
}

//+kubebuilder:object:root=true

// ConsoleList contains a list of Console
type ConsoleList struct {
	metav1.TypeMeta `json:",inline"`
	metav1.ListMeta `json:"metadata,omitempty"`
	Items           []Console `json:"items"`
}

func init() {
	SchemeBuilder.Register(&Console{}, &ConsoleList{})
}
