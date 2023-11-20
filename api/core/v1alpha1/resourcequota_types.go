// SPDX-FileCopyrightText: 2023 SAP SE or an SAP affiliate company and IronCore contributors
// SPDX-License-Identifier: Apache-2.0

package v1alpha1

import (
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

// ResourceQuotaSpec defines the desired state of ResourceQuotaSpec
type ResourceQuotaSpec struct {
	// Hard is a ResourceList of the strictly enforced amount of resources.
	Hard ResourceList `json:"hard,omitempty"`
	// ScopeSelector selects the resources that are subject to this quota.
	// Note: By using certain ScopeSelectors, only certain resources may be tracked.
	ScopeSelector *ResourceScopeSelector `json:"scopeSelector,omitempty"`
}

// ResourceScopeSelector selects
type ResourceScopeSelector struct {
	// MatchExpressions is a list of ResourceScopeSelectorRequirement to match resources by.
	MatchExpressions []ResourceScopeSelectorRequirement `json:"matchExpressions,omitempty"`
}

// ResourceScope is a scope of a resource.
type ResourceScope string

const (
	// ResourceScopeMachineClass refers to the machine class of a resource.
	ResourceScopeMachineClass ResourceScope = "MachineClass"
	// ResourceScopeVolumeClass refers to the volume class of a resource.
	ResourceScopeVolumeClass ResourceScope = "VolumeClass"
	// ResourceScopeBucketClass refers to the bucket class of a resource.
	ResourceScopeBucketClass ResourceScope = "BucketClass"
)

// ResourceScopeSelectorOperator is an operator to compare a ResourceScope with values.
type ResourceScopeSelectorOperator string

const (
	ResourceScopeSelectorOperatorIn           ResourceScopeSelectorOperator = "In"
	ResourceScopeSelectorOperatorNotIn        ResourceScopeSelectorOperator = "NotIn"
	ResourceScopeSelectorOperatorExists       ResourceScopeSelectorOperator = "Exists"
	ResourceScopeSelectorOperatorDoesNotExist ResourceScopeSelectorOperator = "DoesNotExist"
)

// ResourceScopeSelectorRequirement is a requirement for a resource using a ResourceScope alongside
// a ResourceScopeSelectorOperator with Values (depending on the ResourceScopeSelectorOperator).
type ResourceScopeSelectorRequirement struct {
	// ScopeName is the ResourceScope to make a requirement for.
	ScopeName ResourceScope `json:"scopeName"`
	// Operator is the ResourceScopeSelectorOperator to check the ScopeName with in a resource.
	Operator ResourceScopeSelectorOperator `json:"operator"`
	// Values are the values to compare the Operator with the ScopeName. May be optional.
	Values []string `json:"values,omitempty"`
}

// ResourceQuotaStatus is the status of a ResourceQuota.
type ResourceQuotaStatus struct {
	// Hard are the currently enforced hard resource limits. Hard may be less than used in
	// case the limits were introduced / updated after more than allowed resources were already present.
	Hard ResourceList `json:"hard,omitempty"`
	// Used is the amount of currently used resources.
	Used ResourceList `json:"used,omitempty"`
}

// +genclient
// +k8s:deepcopy-gen:interfaces=k8s.io/apimachinery/pkg/runtime.Object

// ResourceQuota is the Schema for the resourcequotas API
type ResourceQuota struct {
	metav1.TypeMeta   `json:",inline"`
	metav1.ObjectMeta `json:"metadata,omitempty"`

	Spec   ResourceQuotaSpec   `json:"spec,omitempty"`
	Status ResourceQuotaStatus `json:"status,omitempty"`
}

// +k8s:deepcopy-gen:interfaces=k8s.io/apimachinery/pkg/runtime.Object

// ResourceQuotaList contains a list of ResourceQuota
type ResourceQuotaList struct {
	metav1.TypeMeta `json:",inline"`
	metav1.ListMeta `json:"metadata,omitempty"`
	Items           []ResourceQuota `json:"items"`
}
