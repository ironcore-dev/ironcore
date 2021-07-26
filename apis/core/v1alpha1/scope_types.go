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

// ScopeSpec defines the desired state of Scope
type ScopeSpec struct {
	// Description is a human-readable description of what the scope is used for.
	Description string `json:"description,omitempty"`
	// Region describes the region scope
	Region string `json:"region,omitempty"`
}

// ScopeStatus defines the observed state of Scope
type ScopeStatus struct {
	common.StateFields `json:",inline"`
	// Namespace references the namespace of the scope
	Namespace string `json:"namespace,omitempty"`
	// ParentScope describes the parent scope, if empty the account
	// should be used as the top reference
	ParentScope string `json:"parentScope"`
	// ParentNamespace represents the namespace of the parent scope
	ParentNamespace string `json:"parentNamespace"`
	// Account describes the account this scope belongs to
	Account string `json:"account"`
}

const (
	// ScopePending indicates that the scope reconciliation is pending.
	ScopePending = "Pending"
	// ScopeReady indicates that the scope reconciliation was successful.
	ScopeReady = "Ready"
	// ScopeFailed indicates that the scope reconciliation failed.
	ScopeFailed = "Failed"
	// ScopeTerminating indicates that the scope is in termination process.
	ScopeTerminating = "Terminating"
)

//+kubebuilder:object:root=true
//+kubebuilder:subresource:status
//+kubebuilder:printcolumn:name="Namespace",type=string,JSONPath=`.status.namespace`
//+kubebuilder:printcolumn:name="Account",type=string,JSONPath=`.status.account`
//+kubebuilder:printcolumn:name="ParentScope",type=string,JSONPath=`.status.parentScope`
//+kubebuilder:printcolumn:name="ParentNamespace",type=string,JSONPath=`.status.parentNamespace`
//+kubebuilder:printcolumn:name="StateFields",type=string,JSONPath=`.status.state`
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
