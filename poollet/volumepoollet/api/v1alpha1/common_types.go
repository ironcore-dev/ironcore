// SPDX-FileCopyrightText: 2023 SAP SE or an SAP affiliate company and IronCore contributors
// SPDX-License-Identifier: Apache-2.0

package v1alpha1

const (
	VolumeUIDLabel       = "volumepoollet.ironcore.dev/volume-uid"
	VolumeNamespaceLabel = "volumepoollet.ironcore.dev/volume-namespace"
	VolumeNameLabel      = "volumepoollet.ironcore.dev/volume-name"

	VolumeSnapshotUIDLabel       = "volumepoollet.ironcore.dev/volume-snapshot-uid"
	VolumeSnapshotNamespaceLabel = "volumepoollet.ironcore.dev/volume-snapshot-namespace"
	VolumeSnapshotNameLabel      = "volumepoollet.ironcore.dev/volume-snapshot-name"

	FieldOwner      = "volumepoollet.ironcore.dev/field-owner"
	VolumeFinalizer = "volumepoollet.ironcore.dev/volume"

	VolumeSnapshotFinalizer = "volumepoollet.ironcore.dev/volume-snapshot"

	VolumeDownwardAPIPrefix = "downward-api.volumepoollet.ironcore.dev/"

	VolumeSnapshotDownwardAPIPrefix = "downward-api.volumepoollet.ironcore.dev/volume-snapshot-"
)
