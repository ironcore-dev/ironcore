// SPDX-FileCopyrightText: 2023 SAP SE or an SAP affiliate company and IronCore contributors
// SPDX-License-Identifier: Apache-2.0

package storage

import (
	commonv1alpha1 "github.com/ironcore-dev/ironcore/api/common/v1alpha1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

// VolumeSnapshotContentSpec defines the desired state of VolumeSnapshotContent
type VolumeSnapshotContentSpec struct {
	// Source defines the VolumeSnapshot handle
	Source *VolumeSnapshotContentSource
	// VolumeSnapshotRef is the reference to the VolumeSnapshot that this content belongs to
	// An empty namespace indicates that the target VolumeSnapshot resides in the same namespace as the source
	VolumeSnapshotRef *commonv1alpha1.UIDReference
}

// VolumeSnapshotContentSource contains VolumeSnapshotHandle of the snapshot
type VolumeSnapshotContentSource struct {
	// VolumeSnapshotHandle is a unique identifier for the snapshot in the storage provider
	VolumeSnapshotHandle string
}

// +k8s:deepcopy-gen:interfaces=k8s.io/apimachinery/pkg/runtime.Object
// +genclient

// VolumeSnapshotContent is the Schema for the volumepools API
type VolumeSnapshotContent struct {
	metav1.TypeMeta
	metav1.ObjectMeta

	Spec VolumeSnapshotContentSpec
}

// +k8s:deepcopy-gen:interfaces=k8s.io/apimachinery/pkg/runtime.Object

// VolumeSnapshotContentList contains a list of VolumeSnapshotContent
type VolumeSnapshotContentList struct {
	metav1.TypeMeta
	metav1.ListMeta
	Items []VolumeSnapshotContent
}
