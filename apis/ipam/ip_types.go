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

// IPSpec defines the desired state of IP
type IPSpec struct {
	// PrefixRef references the parent to allocate the IP from.
	PrefixRef *PrefixReference
	// PrefixSelector is the LabelSelector to use for determining the parent for this IP.
	PrefixSelector *PrefixSelector
	// IP is the ip to allocate.
	//+nullable
	IP commonv1alpha1.IP
}

// IPStatus defines the observed state of IP
type IPStatus struct {
	Conditions []IPCondition
}

type IPConditionType string

const (
	IPReady IPConditionType = "Ready"
)

const (
	IPReadyReasonPending   = "Pending"
	IPReadyReasonAllocated = "Allocated"
)

type IPCondition struct {
	Type               IPConditionType
	Status             corev1.ConditionStatus
	Reason             string
	Message            string
	LastTransitionTime metav1.Time
}

// +k8s:deepcopy-gen:interfaces=k8s.io/apimachinery/pkg/runtime.Object
// +genclient

// IP is the Schema for the ips API
type IP struct {
	metav1.TypeMeta
	metav1.ObjectMeta

	Spec   IPSpec
	Status IPStatus
}

// +k8s:deepcopy-gen:interfaces=k8s.io/apimachinery/pkg/runtime.Object

// IPList contains a list of IP
type IPList struct {
	metav1.TypeMeta
	metav1.ListMeta
	Items []IP
}
