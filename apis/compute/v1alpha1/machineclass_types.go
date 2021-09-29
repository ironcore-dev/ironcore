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
	"encoding/json"
	"errors"

	corev1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/api/resource"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

// MachineClassSpec defines the desired state of MachineClass
type MachineClassSpec struct {
	// Description is a short description of size constraint set
	// +kubebuilder:validation:Optional
	Description string `json:"description,omitempty"`
	// Capabilities describes the resources a machine class can provide.
	Capabilities corev1.ResourceList `json:"capabilities,omitempty"`
}

// MachineClassStatus defines the observed state of MachineClass
type MachineClassStatus struct {
	// Availability describes the regions and zones where this MachineClass is available
	Availability Availability     `json:"availability,omitempty"`
	Constraints  []ConstraintSpec `json:"constraints,omitempty"`
}

//+kubebuilder:object:root=true
//+kubebuilder:resource:scope=Cluster
//+kubebuilder:subresource:status
//+kubebuilder:printcolumn:name="Age",type=date,JSONPath=`.metadata.creationTimestamp`

// MachineClass is the Schema for the machineclasses API
type MachineClass struct {
	metav1.TypeMeta   `json:",inline"`
	metav1.ObjectMeta `json:"metadata,omitempty"`

	Spec   MachineClassSpec   `json:"spec,omitempty"`
	Status MachineClassStatus `json:"status,omitempty"`
}

//+kubebuilder:object:root=true

// MachineClassList contains a list of MachineClass
type MachineClassList struct {
	metav1.TypeMeta `json:",inline"`
	metav1.ListMeta `json:"metadata,omitempty"`
	Items           []MachineClass `json:"items"`
}

type Availability []RegionAvailability

//+kubebuilder:object:generate=true

// RegionAvailability defines a region with its availability zones
type RegionAvailability struct {
	// Region is the name of the region
	Region string `json:"region"`
	// Zones is a list of zones in this region
	Zones []ZoneAvailability `json:"availabilityZone"`
}

//+kubebuilder:object:generate=true

// ZoneAvailability defines the name of a zone
type ZoneAvailability struct {
	// Name is the name of the availability zone
	Name string `json:"name"`
}

// Capability describes a single feature of a MachineClass
type Capability struct {
	// Name is the name of the capability
	Name string `json:"name"`
	// Type defines the type of the capability
	Type string `json:"type"`
	// Value is the effective value of the capability
	Value string `json:"value"`
}

// ConstraintValSpec is a wrapper around value for constraint.
// Since it is not possible to set oneOf/anyOf through kubebuilder
// markers, type is set to number here, and patched with kustomize
// See https://github.com/kubernetes-sigs/kubebuilder/issues/301
//
// Known limitation: literal field set to numeric string, e.g. "12"
// will be deserialized into the numeric field on the other side.
// This should be taken in mind while comparing 2 objects.
// +kubebuilder:validation:Type=number
type ConstraintValSpec struct {
	Literal *string            `json:"-"`
	Numeric *resource.Quantity `json:"-"`
}

func (s *ConstraintValSpec) MarshalJSON() ([]byte, error) {
	if s.Literal != nil && s.Numeric != nil {
		return nil, errors.New("unable to marshal JSON since both numeric and literal fields are set")
	}
	if s.Literal != nil {
		return json.Marshal(s.Literal)
	}
	if s.Numeric != nil {
		return json.Marshal(s.Numeric)
	}
	return nil, nil
}

func (s *ConstraintValSpec) UnmarshalJSON(data []byte) error {
	if len(data) == 0 || string(data) == "null" {
		s.Literal = nil
		s.Numeric = nil
		return nil
	}
	q := resource.Quantity{}
	err := q.UnmarshalJSON(data)
	if err == nil {
		s.Numeric = &q
		return nil
	}
	var str string
	err = json.Unmarshal(data, &str)
	if err != nil {
		return err
	}
	s.Literal = &str
	return nil
}

type AggregateType string

type ConstraintSpec struct {
	// Path is a path to the struct field constraint will be applied to
	// +kubebuilder:validation:Optional
	Path string `json:"path,omitempty"`
	// Aggregate defines whether collection values should be aggregated
	// for constraint checks, in case if path defines selector for collection
	// +kubebuilder:validation:Optional
	// +kubebuilder:validation:Enum=avg;sum
	Aggregate AggregateType `json:"agg,omitempty"`
	// Equal contains an exact expected value
	// +kubebuilder:validation:Optional
	Equal *ConstraintValSpec `json:"eq,omitempty"`
	// NotEqual contains an exact not expected value
	// +kubebuilder:validation:Optional
	NotEqual *ConstraintValSpec `json:"neq,omitempty"`
	// LessThan contains an highest expected value, exclusive
	// +kubebuilder:validation:Optional
	LessThan *resource.Quantity `json:"lt,omitempty"`
	// LessThan contains an highest expected value, inclusive
	// +kubebuilder:validation:Optional
	LessThanOrEqual *resource.Quantity `json:"lte,omitempty"`
	// LessThan contains an lowest expected value, exclusive
	// +kubebuilder:validation:Optional
	GreaterThan *resource.Quantity `json:"gt,omitempty"`
	// GreaterThanOrEqual contains an lowest expected value, inclusive
	// +kubebuilder:validation:Optional
	GreaterThanOrEqual *resource.Quantity `json:"gte,omitempty"`
}

func init() {
	SchemeBuilder.Register(&MachineClass{}, &MachineClassList{})
}
