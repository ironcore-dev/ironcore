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

package storage

import (
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

var (
	// BucketClassFinalizer is the finalizer for BucketClass.
	BucketClassFinalizer = SchemeGroupVersion.Group + "/bucketclass"
)

// +k8s:deepcopy-gen:interfaces=k8s.io/apimachinery/pkg/runtime.Object
// +genclient
// +genclient:nonNamespaced

// BucketClass is the Schema for the bucketclasses API
type BucketClass struct {
	metav1.TypeMeta
	metav1.ObjectMeta
	// Capabilities describes the capabilities of a BucketClass.
	Capabilities corev1.ResourceList
}

// +k8s:deepcopy-gen:interfaces=k8s.io/apimachinery/pkg/runtime.Object

// BucketClassList contains a list of BucketClass
type BucketClassList struct {
	metav1.TypeMeta
	metav1.ListMeta
	Items []BucketClass
}
