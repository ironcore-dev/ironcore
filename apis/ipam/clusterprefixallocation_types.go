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

package ipam

import (
	commonv1alpha1 "github.com/onmetal/onmetal-api/apis/common/v1alpha1"
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

// ClusterPrefixAllocationSpec defines the desired state of ClusterPrefixAllocation
type ClusterPrefixAllocationSpec struct {
	PrefixRef      *corev1.LocalObjectReference
	PrefixSelector *metav1.LabelSelector
	ClusterPrefixAllocationRequest
}

type ClusterPrefixAllocationRequest struct {
	//+nullable
	Prefix       commonv1alpha1.IPPrefix
	PrefixLength int32
}

// ClusterPrefixAllocationStatus defines the observed state of ClusterPrefixAllocation
type ClusterPrefixAllocationStatus struct {
	ClusterPrefixAllocationResult
	Conditions []ClusterPrefixAllocationCondition
}

type ClusterPrefixAllocationResult struct {
	//+nullable
	Prefix commonv1alpha1.IPPrefix
}

type ClusterPrefixAllocationConditionType string

const (
	ClusterPrefixAllocationReady ClusterPrefixAllocationConditionType = "Ready"
)

const (
	ClusterPrefixAllocationReadyReasonFailed    = "Failed"
	ClusterPrefixAllocationReadyReasonSucceeded = "Succeeded"
	ClusterPrefixAllocationReadyReasonPending   = "Pending"
)

type ClusterPrefixAllocationCondition struct {
	Type               ClusterPrefixAllocationConditionType
	Status             corev1.ConditionStatus
	Reason             string
	Message            string
	LastTransitionTime metav1.Time
}

// +k8s:deepcopy-gen:interfaces=k8s.io/apimachinery/pkg/runtime.Object
// +genclient
// +genclient:nonNamespaced
// +genclient:noStatus

// ClusterPrefixAllocation is the Schema for the clusterprefixallocations API
type ClusterPrefixAllocation struct {
	metav1.TypeMeta
	metav1.ObjectMeta

	Spec   ClusterPrefixAllocationSpec
	Status ClusterPrefixAllocationStatus
}

// +k8s:deepcopy-gen:interfaces=k8s.io/apimachinery/pkg/runtime.Object

// ClusterPrefixAllocationList contains a list of ClusterPrefixAllocation
type ClusterPrefixAllocationList struct {
	metav1.TypeMeta
	metav1.ListMeta
	Items []ClusterPrefixAllocation
}
