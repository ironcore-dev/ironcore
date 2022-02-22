// Copyright 2021 OnMetal authors
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

package v1alpha1

import (
	corev1 "k8s.io/api/core/v1"
)

// Target is a target for network traffic.
// It may be either
// * a v1alpha1.Machine
// * a Gateway
// * a ReservedIP
// * a Subnet
// * a raw IP
// * a raw CIDR.
type Target struct {
	Machine    *MachineRouteTarget          `json:"machine,omitempty"`
	Gateway    *corev1.LocalObjectReference `json:"gateway,omitempty"`
	ReservedIP *corev1.LocalObjectReference `json:"reservedIP,omitempty"`
	Subnet     *corev1.LocalObjectReference `json:"subnet,omitempty"`
	IP         string                       `json:"ip,omitempty"`
	CIDR       string                       `json:"cidr,omitempty"`
}
