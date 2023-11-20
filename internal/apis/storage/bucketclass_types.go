// SPDX-FileCopyrightText: 2023 SAP SE or an SAP affiliate company and IronCore contributors
// SPDX-License-Identifier: Apache-2.0

package storage

import (
	"github.com/ironcore-dev/ironcore/internal/apis/core"
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
	Capabilities core.ResourceList
}

// +k8s:deepcopy-gen:interfaces=k8s.io/apimachinery/pkg/runtime.Object

// BucketClassList contains a list of BucketClass
type BucketClassList struct {
	metav1.TypeMeta
	metav1.ListMeta
	Items []BucketClass
}
