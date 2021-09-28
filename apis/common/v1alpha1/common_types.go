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
	corev1 "k8s.io/api/core/v1"
)

// ConfigMapKeySelector is a reference to a specific 'key' within a ConfigMap resource.
// In some instances, `key` is a required field.
type ConfigMapKeySelector struct {
	// The name of the ConfigMap resource being referred to.
	corev1.LocalObjectReference `json:",inline"`
	// The key of the entry in the ConfigMap resource's `data` field to be used.
	// Some instances of this field may be defaulted, in others it may be
	// required.
	// +optional
	Key string `json:"key,omitempty"`
}

// SecretKeySelector is a reference to a specific 'key' within a Secret resource.
// In some instances, `key` is a required field.
type SecretKeySelector struct {
	// The name of the Secret resource being referred to.
	corev1.LocalObjectReference `json:",inline"`
	// The key of the entry in the Secret resource's `data` field to be used.
	// Some instances of this field may be defaulted, in others it may be
	// required.
	// +optional
	Key string `json:"key,omitempty"`
}

// IPAddr is an IP address.
// TODO: create marshal/unmarshal functions
type IPAddr string

// CIDR represents a network CIDR.
// TODO: create marshal/unmarshal functions
type CIDR string
