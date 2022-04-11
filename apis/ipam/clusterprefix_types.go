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

const (
	ClusterPrefixKind = "ClusterPrefix"
)

// ClusterPrefixSpec defines the desired state of ClusterPrefix
type ClusterPrefixSpec struct {
	// ParentRef references the parent to allocate the Prefix from.
	// If ParentRef and ParentSelector is empty, the Prefix is considered a root prefix and thus
	// allocated by itself.
	ParentRef *corev1.LocalObjectReference
	// ParentSelector is the LabelSelector to use for determining the parent for this Prefix.
	ParentSelector *metav1.LabelSelector
	// PrefixSpace is the space the ClusterPrefix manages.
	PrefixSpace
}

// ClusterPrefixStatus defines the observed state of ClusterPrefix
type ClusterPrefixStatus struct {
	// Conditions is a list of conditions of a ClusterPrefix.
	Conditions []ClusterPrefixCondition
	// Available is a list of available prefixes.
	Available []commonv1alpha1.IPPrefix
	// Reserved is a list of reserved prefixes.
	Reserved []commonv1alpha1.IPPrefix
}

type ClusterPrefixConditionType string

const (
	ClusterPrefixReady ClusterPrefixConditionType = "Ready"
)

const (
	ClusterPrefixReadyReasonPending   = "Pending"
	ClusterPrefixReadyReasonAllocated = "Allocated"
)

type ClusterPrefixCondition struct {
	Type               ClusterPrefixConditionType
	Status             corev1.ConditionStatus
	Reason             string
	Message            string
	LastTransitionTime metav1.Time
}

// +k8s:deepcopy-gen:interfaces=k8s.io/apimachinery/pkg/runtime.Object
// +genclient
// +genclient:nonNamespaced

// ClusterPrefix is the Schema for the clusterprefixes API
type ClusterPrefix struct {
	metav1.TypeMeta
	metav1.ObjectMeta

	Spec   ClusterPrefixSpec
	Status ClusterPrefixStatus
}

// +k8s:deepcopy-gen:interfaces=k8s.io/apimachinery/pkg/runtime.Object

// ClusterPrefixList contains a list of ClusterPrefix
type ClusterPrefixList struct {
	metav1.TypeMeta
	metav1.ListMeta
	Items []ClusterPrefix
}
