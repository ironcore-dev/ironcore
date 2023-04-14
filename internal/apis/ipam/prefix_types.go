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
	corev1alpha1 "github.com/onmetal/onmetal-api/api/core/v1alpha1"
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

// PrefixSpec defines the desired state of Prefix
type PrefixSpec struct {
	// IPFamily is the IPFamily of the prefix.
	// If unset but Prefix is set, this can be inferred.
	IPFamily corev1.IPFamily
	// Prefix is the prefix to allocate for this Prefix.
	Prefix *corev1alpha1.IPPrefix
	// PrefixLength is the length of prefix to allocate for this Prefix.
	PrefixLength int32

	// ParentRef references the parent to allocate the Prefix from.
	// If ParentRef and ParentSelector is empty, the Prefix is considered a root prefix and thus
	// allocated by itself.
	ParentRef *corev1.LocalObjectReference
	// ParentSelector is the LabelSelector to use for determining the parent for this Prefix.
	ParentSelector *metav1.LabelSelector
}

func (s *PrefixSpec) IsRoot() bool {
	return s.ParentRef == nil && s.ParentSelector == nil
}

// PrefixStatus defines the observed state of Prefix
type PrefixStatus struct {
	// Phase is the PrefixPhase of the Prefix.
	Phase PrefixPhase
	// LastPhaseTransitionTime is the last time the Phase changed values.
	LastPhaseTransitionTime *metav1.Time

	// Used is a list of used prefixes.
	Used []corev1alpha1.IPPrefix
}

// PrefixPhase is a phase a Prefix can be in.
type PrefixPhase string

const (
	// PrefixPhasePending marks a prefix as waiting for allocation.
	PrefixPhasePending PrefixPhase = "Pending"
	// PrefixPhaseAllocated marks a prefix as allocated.
	PrefixPhaseAllocated PrefixPhase = "Allocated"
)

// +k8s:deepcopy-gen:interfaces=k8s.io/apimachinery/pkg/runtime.Object
// +genclient

// Prefix is the Schema for the prefixes API
type Prefix struct {
	metav1.TypeMeta
	metav1.ObjectMeta

	Spec   PrefixSpec
	Status PrefixStatus
}

func (p *Prefix) IsRoot() bool {
	return p.Spec.IsRoot()
}

// +k8s:deepcopy-gen:interfaces=k8s.io/apimachinery/pkg/runtime.Object

// PrefixList contains a list of Prefix
type PrefixList struct {
	metav1.TypeMeta
	metav1.ListMeta
	Items []Prefix
}

type PrefixTemplateSpec struct {
	metav1.ObjectMeta
	Spec PrefixSpec
}
