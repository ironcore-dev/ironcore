// SPDX-FileCopyrightText: 2023 SAP SE or an SAP affiliate company and IronCore contributors
// SPDX-License-Identifier: Apache-2.0

// Package v1alpha1 contains API Schema definitions for the storage v1alpha1 API group
// +groupName=storage.ironcore.dev
package v1alpha1

import (
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apimachinery/pkg/runtime/schema"
)

var (
	// SchemeGroupVersion is group version used to register these objects
	SchemeGroupVersion = schema.GroupVersion{Group: "storage.ironcore.dev", Version: "v1alpha1"}

	// SchemeBuilder is used to add go types to the GroupVersionKind scheme
	SchemeBuilder = runtime.NewSchemeBuilder(addKnownTypes)

	// AddToScheme adds the types in this group-version to the given scheme.
	AddToScheme = SchemeBuilder.AddToScheme
)

func Resource(name string) schema.GroupResource {
	return schema.GroupResource{
		Group:    SchemeGroupVersion.Group,
		Resource: name,
	}
}

func addKnownTypes(scheme *runtime.Scheme) error {
	scheme.AddKnownTypes(SchemeGroupVersion,
		&VolumeClass{},
		&VolumeClassList{},
		&VolumePool{},
		&VolumePoolList{},
		&Volume{},
		&VolumeList{},
		&VolumeSnapshot{},
		&VolumeSnapshotList{},
		&BucketClass{},
		&BucketClassList{},
		&BucketPool{},
		&BucketPoolList{},
		&Bucket{},
		&BucketList{},
	)
	metav1.AddToGroupVersion(scheme, SchemeGroupVersion)
	return nil
}
