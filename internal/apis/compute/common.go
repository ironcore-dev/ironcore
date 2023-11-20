// SPDX-FileCopyrightText: 2023 SAP SE or an SAP affiliate company and IronCore contributors
// SPDX-License-Identifier: Apache-2.0

package compute

import (
	"github.com/ironcore-dev/ironcore/internal/apis/networking"
	"github.com/ironcore-dev/ironcore/internal/apis/storage"
)

const (
	MachineMachinePoolRefNameField  = "spec.machinePoolRef.name"
	MachineMachineClassRefNameField = "spec.machineClassRef.name"

	// MachinePoolsGroup is the system rbac group all machine pools are in.
	MachinePoolsGroup = "compute.ironcore.dev:system:machinepools"

	// MachinePoolUserNamePrefix is the prefix all machine pool users should have.
	MachinePoolUserNamePrefix = "compute.ironcore.dev:system:machinepool:"
)

// MachinePoolCommonName constructs the common name for a certificate of a machine pool user.
func MachinePoolCommonName(name string) string {
	return MachinePoolUserNamePrefix + name
}

// EphemeralNetworkInterfaceSource is a definition for an ephemeral (i.e. coupled to the lifetime of the surrounding
// object) networking.NetworkInterface.
type EphemeralNetworkInterfaceSource struct {
	// NetworkInterfaceTemplate is the template definition of the networking.NetworkInterface.
	NetworkInterfaceTemplate *networking.NetworkInterfaceTemplateSpec
}

// EphemeralVolumeSource is a definition for an ephemeral (i.e. coupled to the lifetime of the surrounding object)
// storage.Volume.
type EphemeralVolumeSource struct {
	// VolumeTemplate is the template definition of the storage.Volume.
	VolumeTemplate *storage.VolumeTemplateSpec
}
