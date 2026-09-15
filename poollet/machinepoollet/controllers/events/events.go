// SPDX-FileCopyrightText: SAP SE or an SAP affiliate company and IronCore contributors
// SPDX-License-Identifier: Apache-2.0

package events

// Event reasons.
const (
	MachineClassNotReady     = "MachineClassNotReady"
	NetworkInterfaceNotReady = "NetworkInterfaceNotReady"
	VolumeNotReady           = "VolumeNotReady"
	IgnitionNotReady         = "IgnitionNotReady"
)

// Event actions.
const (
	ResolvingMachineClass     = "ResolvingMachineClass"
	ResolvingIgnition         = "ResolvingIgnition"
	AttachingNetworkInterface = "AttachingNetworkInterface"
	AttachingVolume           = "AttachingVolume"
	CreatingEphemeralVolume   = "CreatingEphemeralVolume"
)
