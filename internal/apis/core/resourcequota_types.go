// Copyright 2023 IronCore authors
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
//      http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.

package core

import (
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

// ResourceQuotaSpec defines the desired state of ResourceQuotaSpec
type ResourceQuotaSpec struct {
	// Hard is a ResourceList of the strictly enforced amount of resources.
	Hard ResourceList
	// ScopeSelector selects the resources that are subject to this quota.
	// Note: By using certain ScopeSelectors, only certain resources may be tracked.
	ScopeSelector *ResourceScopeSelector
}

// ResourceScopeSelector selects
type ResourceScopeSelector struct {
	// MatchExpressions is a list of ResourceScopeSelectorRequirement to match resources by.
	MatchExpressions []ResourceScopeSelectorRequirement
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

// ResourceScopeSelectorRequirement is a requirement for a resource using a ResourceScope alongside
// a ResourceScopeSelectorOperator with Values (depending on the ResourceScopeSelectorOperator).
type ResourceScopeSelectorRequirement struct {
	// ScopeName is the ResourceScope to make a requirement for.
	ScopeName ResourceScope
	// Operator is the ResourceScopeSelectorOperator to check the ScopeName with in a resource.
	Operator ResourceScopeSelectorOperator
	// Values are the values to compare the Operator with the ScopeName. May be optional.
	Values []string
}

// ResourceQuotaStatus is the status of a ResourceQuota.
type ResourceQuotaStatus struct {
	// Hard are the currently enforced hard resource limits. Hard may be less than used in
	// case the limits were introduced / updated after more than allowed resources were already present.
	Hard ResourceList
	// Used is the amount of currently used resources.
	Used ResourceList
}

// +genclient
// +k8s:deepcopy-gen:interfaces=k8s.io/apimachinery/pkg/runtime.Object

// ResourceQuota is the Schema for the resourcequotas API
type ResourceQuota struct {
	metav1.TypeMeta
	metav1.ObjectMeta

	Spec   ResourceQuotaSpec
	Status ResourceQuotaStatus
}

// +k8s:deepcopy-gen:interfaces=k8s.io/apimachinery/pkg/runtime.Object

// ResourceQuotaList contains a list of ResourceQuota
type ResourceQuotaList struct {
	metav1.TypeMeta
	metav1.ListMeta
	Items []ResourceQuota
}
