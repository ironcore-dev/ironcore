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

import metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"

//+kubebuilder:object:generate=true

// StateFields defines the observed state of an object
type StateFields struct {
	// State indicates the state of a resource
	//+optional
	State string `json:"state,omitempty"`
	// Message contains a message for the corresponding state
	//+optional
	Message string `json:"message,omitempty"`
	// Conditions represents the status for individual operators
	//+optional
	Conditions []metav1.Condition `json:"conditions,omitempty"`
}

const (
	StateAvailable = "Available"
	StateReady     = "Ready"
	StateUp        = "Up"
	StateError     = "Error"
	StateInvalid   = "Invalid"
	StateBusy      = "Busy"
	StatePending   = "Pending"
)
