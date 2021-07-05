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
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

// ScopeSpec defines the desired state of Scope
type ScopeSpec struct {
	// ParentScope describes the parent scope, if empty the account
	// should be used as the top reference
	ParentScope *string `json:"parentScope,omitempty"`
	// Description is a human-readable description of what the scope is used for.
	Description *string `json:"description,omitempty"`
}

// ScopeStatus defines the observed state of Scope
type ScopeStatus struct {
	// State represents the state of the scope
	State ScopeState `json:"state,omitempty"`
	// Namespace references the namespace of the scope
	Namespace string `json:"namespace,omitempty"`
}

// ScopeState is a label for the condition of a scope at the current time.
type ScopeState string

const (
	// ScopePending indicates that the scope reconciliation is pending.
	ScopePending ScopeState = "Pending"
	// ScopeReady indicates that the scope reconciliation was successful.
	ScopeReady ScopeState = "Ready"
	// ScopeFailed indicates that the scope reconciliation failed.
	ScopeFailed ScopeState = "Failed"
	// ScopeTerminating indicates that the scope is in termination process.
	ScopeTerminating ScopeState = "Terminating"
)

//+kubebuilder:object:root=true
//+kubebuilder:subresource:status
//+kubebuilder:printcolumn:name="Namespace",type=string,JSONPath=`.status.namespace`
//+kubebuilder:printcolumn:name="ParentScope",type=string,JSONPath=`.spec.parentScope`
//+kubebuilder:printcolumn:name="State",type=string,JSONPath=`.status.state`
//+kubebuilder:printcolumn:name="Age",type="date",JSONPath=".metadata.creationTimestamp"

// Scope is the Schema for the scopes API
type Scope struct {
	metav1.TypeMeta   `json:",inline"`
	metav1.ObjectMeta `json:"metadata,omitempty"`

	Spec   ScopeSpec   `json:"spec,omitempty"`
	Status ScopeStatus `json:"status,omitempty"`
}

//+kubebuilder:object:root=true

// ScopeList contains a list of Scope
type ScopeList struct {
	metav1.TypeMeta `json:",inline"`
	metav1.ListMeta `json:"metadata,omitempty"`
	Items           []Scope `json:"items"`
}

func init() {
	SchemeBuilder.Register(&Scope{}, &ScopeList{})
}
