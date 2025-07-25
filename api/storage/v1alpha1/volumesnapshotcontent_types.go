// SPDX-FileCopyrightText: 2023 SAP SE or an SAP affiliate company and IronCore contributors
// SPDX-License-Identifier: Apache-2.0

package v1alpha1

import (
	commonv1alpha1 "github.com/ironcore-dev/ironcore/api/common/v1alpha1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

// VolumeSnapshotContentSpec defines the desired state of VolumeSnapshotContent
type VolumeSnapshotContentSpec struct {
	// Source defines the VolumeSnapshot handle
	Source VolumeSnapshotContentSource `json:"source,omitempty"`
	// VolumeSnapshotRef is the reference to the VolumeSnapshot that this content belongs to
	// An empty namespace indicates that the target VolumeSnapshot resides in the same namespace as the source
	VolumeSnapshotRef commonv1alpha1.UIDReference `json:"volumeSnapshotRef,omitempty"`
}

// VolumeSnapshotContentSource contains VolumeSnapshotHandle of the snapshot
type VolumeSnapshotContentSource struct {
	// VolumeSnapshotHandle is a unique identifier for the snapshot in the storage provider
	VolumeSnapshotHandle string `json:"snapshotHandle"`
}

// +k8s:deepcopy-gen:interfaces=k8s.io/apimachinery/pkg/runtime.Object
// +genclient

// VolumeSnapshotContent is the Schema for the volumepools API
type VolumeSnapshotContent struct {
	metav1.TypeMeta   `json:",inline"`
	metav1.ObjectMeta `json:"metadata,omitempty"`

	Spec VolumeSnapshotContentSpec `json:"spec,omitempty"`
}

// +k8s:deepcopy-gen:interfaces=k8s.io/apimachinery/pkg/runtime.Object

// VolumeSnapshotContentList contains a list of VolumeSnapshotContent
type VolumeSnapshotContentList struct {
	metav1.TypeMeta `json:",inline"`
	metav1.ListMeta `json:"metadata,omitempty"`
	Items           []VolumeSnapshotContent `json:"items"`
}
