// SPDX-FileCopyrightText: 2023 SAP SE or an SAP affiliate company and IronCore contributors
// SPDX-License-Identifier: Apache-2.0

package v1alpha1

import (
	corev1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/api/resource"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

// VolumeSnapshotSpec defines the desired state of VolumeSnapshot
type VolumeSnapshotSpec struct {
	// VolumeRef indicates which Volume to refer for VolumeSnapshot
	VolumeRef corev1.LocalObjectReference `json:"volumeRef"`
}

// VolumeSnapshotStatus defines the observed state of VolumeSnapshot
type VolumeSnapshotStatus struct {
	// State represents the storage provider state of VolumeSnapshot
	State VolumeSnapshotState `json:"state,omitempty"`
	// RestoreSize is the size of storage required to restore from VolumeSnapshot
	RestoreSize *resource.Quantity `json:"restoreSize,omitempty"`
}

// VolumeSnapshotState is the state of a VolumeSnapshot
type VolumeSnapshotState string

const (
	// VolumeSnapshotStateReady means the VolumeSnapshot has been successfully created and is ready to use
	VolumeSnapshotStateReady VolumeSnapshotState = "Ready"
	// VolumeSnapshotStatePending means the VolumeSnapshot resource has been created, but the snapshot has not yet been initiated
	VolumeSnapshotStatePending VolumeSnapshotState = "Pending"
	// VolumeSnapshotStateProcessing means the VolumeSnapshot is being processed by the storage provider
	VolumeSnapshotStateProcessing VolumeSnapshotState = "Processing"
	// VolumeSnapshotStateFailed means the VolumeSnapshot creation has failed
	VolumeSnapshotStateFailed VolumeSnapshotState = "Failed"
	// VolumeSnapshotStateDeleting means the VolumeSnapshot is being deleted
	VolumeSnapshotStateDeleting VolumeSnapshotState = "Deleting"
)

// +k8s:deepcopy-gen:interfaces=k8s.io/apimachinery/pkg/runtime.Object
// +genclient

// VolumeSnapshot is the Schema for the VolumeSnapshots API
type VolumeSnapshot struct {
	metav1.TypeMeta   `json:",inline"`
	metav1.ObjectMeta `json:"metadata,omitempty"`

	Spec   VolumeSnapshotSpec   `json:"spec,omitempty"`
	Status VolumeSnapshotStatus `json:"status,omitempty"`
}

// +k8s:deepcopy-gen:interfaces=k8s.io/apimachinery/pkg/runtime.Object

// VolumeSnapshotList contains a list of VolumeSnapshot
type VolumeSnapshotList struct {
	metav1.TypeMeta `json:",inline"`
	metav1.ListMeta `json:"metadata,omitempty"`
	Items           []VolumeSnapshot `json:"items"`
}
