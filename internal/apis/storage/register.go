// SPDX-FileCopyrightText: 2023 SAP SE or an SAP affiliate company and IronCore contributors
// SPDX-License-Identifier: Apache-2.0

// Package storage contains API Schema definitions for the storage internal API group
// +groupName=storage.ironcore.dev
package storage

import (
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apimachinery/pkg/runtime/schema"
)

var (
	// SchemeGroupVersion is group version used to register these objects
	SchemeGroupVersion = schema.GroupVersion{Group: "storage.ironcore.dev", Version: runtime.APIVersionInternal}

	// SchemeBuilder is used to add go types to the GroupVersionKind scheme
	SchemeBuilder = runtime.NewSchemeBuilder(addKnownTypes)

	// AddToScheme adds the types in this group-version to the given scheme.
	AddToScheme = SchemeBuilder.AddToScheme
)

func Resource(name string) schema.GroupResource {
	return SchemeGroupVersion.WithResource(name).GroupResource()
}

func Kind(name string) schema.GroupKind {
	return SchemeGroupVersion.WithKind(name).GroupKind()
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
	return nil
}
