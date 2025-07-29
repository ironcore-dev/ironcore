// SPDX-FileCopyrightText: 2023 SAP SE or an SAP affiliate company and IronCore contributors
// SPDX-License-Identifier: Apache-2.0

package storage

import (
	corev1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/api/resource"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

// VolumeSnapshotResource is a constant for the name of the VolumeSnapshot resource.
const VolumeSnapshotResource = "VolumeSnapshot"

// VolumeSnapshotSpec defines the desired state of VolumeSnapshot
type VolumeSnapshotSpec struct {
	// VolumeRef indicates which Volume to refer for VolumeSnapshot
	VolumeRef *corev1.LocalObjectReference
}

// VolumeSnapshotStatus defines the observed state of VolumeSnapshot
type VolumeSnapshotStatus struct {
	// State represents the storage provider state of VolumeSnapshot
	State VolumeSnapshotState
	// restoreSize is the size of storage required to restore from VolumeSnapshot
	RestoreSize *resource.Quantity
}

// VolumeSnapshotState is the state of a VolumeSnapshot
type VolumeSnapshotState string

const (
	// VolumeSnapshotStatePending means the VolumeSnapshot resource has been created, but the snapshot has not yet been initiated
	VolumeSnapshotStatePending VolumeSnapshotState = "Pending"
	// VolumeSnapshotStateReady means the VolumeSnapshot has been successfully created and is ready to use
	VolumeSnapshotStateReady VolumeSnapshotState = "Ready"
	// VolumeSnapshotStateFailed means the VolumeSnapshot creation has failed
	VolumeSnapshotStateFailed VolumeSnapshotState = "Failed"
)

// +k8s:deepcopy-gen:interfaces=k8s.io/apimachinery/pkg/runtime.Object
// +genclient

// VolumeSnapshot is the Schema for the VolumeSnapshots API
type VolumeSnapshot struct {
	metav1.TypeMeta
	metav1.ObjectMeta

	Spec   VolumeSnapshotSpec
	Status VolumeSnapshotStatus
}

// +k8s:deepcopy-gen:interfaces=k8s.io/apimachinery/pkg/runtime.Object

// VolumeSnapshotList contains a list of VolumeSnapshot
type VolumeSnapshotList struct {
	metav1.TypeMeta
	metav1.ListMeta
	Items []VolumeSnapshot
}
