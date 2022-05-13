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

// ConsoleSpec defines the desired state of Console
type ConsoleSpec struct {
	// MachineRef references the machine to open a console to.
	MachineRef corev1.LocalObjectReference `json:"machineRef"`
}

type ConsoleClientConfig struct {
	// Service is the service to connect to.
	Service ServiceReference `json:"service"`
}

// ServiceReference is a reference to a Service in the same namespace as the referent.
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

// ConsoleState is a state a Console can be in.
//+enum
type ConsoleState string

const (
	ConsoleStatePending ConsoleState = "Pending"
	ConsoleStateReady   ConsoleState = "Ready"
	ConsoleStateError   ConsoleState = "Error"
)

// ConsoleStatus defines the observed state of Console
type ConsoleStatus struct {
	// State is the state of a Console.
	State ConsoleState `json:"state,omitempty"`
	// ClientConfig is the client configuration to connect to a console.
	// Only usable if the ConsoleStatus.State is ConsoleStateReady.
	ClientConfig *ConsoleClientConfig `json:"clientConfig,omitempty"`
}

// +k8s:deepcopy-gen:interfaces=k8s.io/apimachinery/pkg/runtime.Object
// +genclient

// Console is the Schema for the consoles API
type Console struct {
	metav1.TypeMeta   `json:",inline"`
	metav1.ObjectMeta `json:"metadata,omitempty"`

	Spec   ConsoleSpec   `json:"spec,omitempty"`
	Status ConsoleStatus `json:"status,omitempty"`
}

// +k8s:deepcopy-gen:interfaces=k8s.io/apimachinery/pkg/runtime.Object

// ConsoleList contains a list of Console
type ConsoleList struct {
	metav1.TypeMeta `json:",inline"`
	metav1.ListMeta `json:"metadata,omitempty"`
	Items           []Console `json:"items"`
}
