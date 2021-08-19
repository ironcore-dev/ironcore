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
	rbacv1 "k8s.io/api/rbac/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime/schema"
)

var AccountGK = schema.GroupKind{
	Group: GroupVersion.Group,
	Kind:  "Account",
}

// AccountSpec defines the desired state of Account
type AccountSpec struct {
	// CreatedBy is a subject representing a user name, an email address, or any other identifier of a user
	// who created the account.
	CreatedBy *rbacv1.Subject `json:"createdBy,omitempty"`
	// Description is a human-readable description of what the account is used for.
	Description string `json:"description,omitempty"`
	// Owner is a subject representing a user name, an email address, or any other identifier of a user owning
	// the account.
	Owner *rbacv1.Subject `json:"owner,omitempty"`
	// Purpose is a human-readable explanation of the account's purpose.
	Purpose string `json:"purpose,omitempty"`
}

// AccountStatus defines the observed state of Account
type AccountStatus struct {
	common.StateFields `json:",inline"`
	// Namespace references the namespace of the account
	Namespace string `json:"namespace,omitempty"`
}

const (
	// AccountKey is the Account lookup key
	AccountKey = "account"
	// AccountStatePending indicates that the Account reconciliation is pending.
	AccountStatePending = "Pending"
	// AccountStateReady indicates that the Account reconciliation was successful.
	AccountStateReady = "Ready"
	// AccountStateFailed indicates that the Account reconciliation failed.
	AccountStateFailed = "Failed"
	// AccountStateTerminating indicates that the Account is in termination process.
	AccountStateTerminating = "Terminating"
	// AccountStateInitial indicates that the Account is in an initial state
	AccountStateInitial = "Initial"
)

//+kubebuilder:object:root=true
//+kubebuilder:subresource:status
//+kubebuilder:resource:scope=Cluster
//+kubebuilder:printcolumn:name="Namespace",type=string,JSONPath=`.status.namespace`
//+kubebuilder:printcolumn:name="State",type=string,JSONPath=`.status.state`
//+kubebuilder:printcolumn:name="Age",type=date,JSONPath=`.metadata.creationTimestamp`

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
