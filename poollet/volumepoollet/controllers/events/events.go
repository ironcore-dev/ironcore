// SPDX-FileCopyrightText: SAP SE or an SAP affiliate company and IronCore contributors
// SPDX-License-Identifier: Apache-2.0

package events

// Event reasons.
const (
	VolumeClassNotReady            = "VolumeClassNotReady"
	VolumeEncryptionSecretNotReady = "VolumeEncryptionSecretNotReady"
	VolumeSnapshotNotFound         = "VolumeSnapshotNotFound"
	VolumeSnapshotNotReady         = "VolumeSnapshotNotReady"
	SourceVolumeNotFound           = "SourceVolumeNotFound"
	SourceVolumeNotAvailable       = "SourceVolumeNotAvailable"
)

// Event actions.
const (
	ResolvingVolumeClass            = "ResolvingVolumeClass"
	ResolvingVolumeSnapshot         = "ResolvingVolumeSnapshot"
	ResolvingVolumeEncryptionSecret = "ResolvingVolumeEncryptionSecret"
	ResolvingSourceVolume           = "ResolvingSourceVolume"
	PreparingVolumeEncryption       = "PreparingVolumeEncryption"
)
