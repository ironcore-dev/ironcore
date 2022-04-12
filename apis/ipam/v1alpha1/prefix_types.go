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
	commonv1alpha1 "github.com/onmetal/onmetal-api/apis/common/v1alpha1"
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

const (
	PrefixKind = "Prefix"
)

// PrefixSpec defines the desired state of Prefix
type PrefixSpec struct {
	// IPFamily is the IPFamily of the prefix.
	// If unset but Prefix is set, this can be inferred.
	IPFamily corev1.IPFamily `json:"ipFamily,omitempty"`
	// Prefix is the prefix to allocate for this Prefix.
	Prefix *commonv1alpha1.IPPrefix `json:"prefix,omitempty"`
	// PrefixLength is the length of prefix to allocate for this Prefix.
	PrefixLength int32 `json:"prefixLength,omitempty"`

	// ParentRef references the parent to allocate the Prefix from.
	// If ParentRef and ParentSelector is empty, the Prefix is considered a root prefix and thus
	// allocated by itself.
	ParentRef *corev1.LocalObjectReference `json:"parentRef,omitempty"`
	// ParentSelector is the LabelSelector to use for determining the parent for this Prefix.
	ParentSelector *metav1.LabelSelector `json:"parentSelector,omitempty"`
}

func (s *PrefixSpec) IsRoot() bool {
	return s.ParentRef == nil && s.ParentSelector == nil
}

// PrefixStatus defines the observed state of Prefix
type PrefixStatus struct {
	// Conditions is a list of conditions of a Prefix.
	Conditions []PrefixCondition `json:"conditions,omitempty"`
	// Used is a list of used prefixes.
	Used []commonv1alpha1.IPPrefix `json:"used,omitempty"`
}

type PrefixConditionType string

const (
	PrefixReady PrefixConditionType = "Ready"
)

type PrefixCondition struct {
	Type               PrefixConditionType    `json:"type"`
	Status             corev1.ConditionStatus `json:"status"`
	Reason             string                 `json:"reason,omitempty"`
	Message            string                 `json:"message,omitempty"`
	LastTransitionTime metav1.Time            `json:"lastTransitionTime,omitempty"`
}

// +genclient
// +k8s:deepcopy-gen:interfaces=k8s.io/apimachinery/pkg/runtime.Object

// Prefix is the Schema for the prefixes API
type Prefix struct {
	metav1.TypeMeta   `json:",inline"`
	metav1.ObjectMeta `json:"metadata,omitempty"`

	Spec   PrefixSpec   `json:"spec,omitempty"`
	Status PrefixStatus `json:"status,omitempty"`
}

func (p *Prefix) IsRoot() bool {
	return p.Spec.IsRoot()
}

func (p *Prefix) Readiness() Readiness {
	readiness, _ := GetPrefixConditionsReadinessAndIndex(p.Status.Conditions)
	return readiness
}

// +k8s:deepcopy-gen:interfaces=k8s.io/apimachinery/pkg/runtime.Object

// PrefixList contains a list of Prefix
type PrefixList struct {
	metav1.TypeMeta `json:",inline"`
	metav1.ListMeta `json:"metadata,omitempty"`
	Items           []Prefix `json:"items"`
}
