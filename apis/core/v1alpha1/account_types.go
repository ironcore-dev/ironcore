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
	rbacv1 "k8s.io/api/rbac/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

// AccountSpec defines the desired state of Account
type AccountSpec struct {
	// CreatedBy is a subject representing a user name, an email address, or any other identifier of a user
	// who created the account.
	CreatedBy *rbacv1.Subject `json:"createdBy,omitempty"`
	// Description is a human-readable description of what the account is used for.
	Description *string `json:"description,omitempty"`
	// Owner is a subject representing a user name, an email address, or any other identifier of a user owning
	// the account.
	Owner *rbacv1.Subject `json:"owner,omitempty"`
	// Purpose is a human-readable explanation of the account's purpose.
	Purpose *string `json:"purpose,omitempty"`
}

// AccountStatus defines the observed state of Account
type AccountStatus struct {
	// AccountState represents the state of the account
	State AccountState `json:"state,omitempty"`
	// Namespace references the namespace of the account
	Namespace string `json:"namespace,omitempty"`
}

// AccountState is a label for the condition of a account at the current time.
type AccountState string

const (
	// AccountPending indicates that the account reconciliation is pending.
	AccountPending AccountState = "Pending"
	// AccountReady indicates that the account reconciliation was successful.
	AccountReady AccountState = "Ready"
	// AccountFailed indicates that the account reconciliation failed.
	AccountFailed AccountState = "Failed"
	// AccountTerminating indicates that the account is in termination process.
	AccountTerminating AccountState = "Terminating"
)

//+kubebuilder:object:root=true
//+kubebuilder:subresource:status
//+kubebuilder:resource:scope=Cluster

// Account is the Schema for the accounts API
type Account struct {
	metav1.TypeMeta   `json:",inline"`
	metav1.ObjectMeta `json:"metadata,omitempty"`

	Spec   AccountSpec   `json:"spec,omitempty"`
	Status AccountStatus `json:"status,omitempty"`
}

//+kubebuilder:object:root=true

// AccountList contains a list of Account
type AccountList struct {
	metav1.TypeMeta `json:",inline"`
	metav1.ListMeta `json:"metadata,omitempty"`
	Items           []Account `json:"items"`
}

func init() {
	SchemeBuilder.Register(&Account{}, &AccountList{})
}
