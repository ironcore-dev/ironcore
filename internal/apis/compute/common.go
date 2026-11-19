// SPDX-FileCopyrightText: SAP SE or an SAP affiliate company and IronCore contributors
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

	// PoolLifecycleControllersGroup is the system rbac group the pool-lifecycle-controller is in.
	PoolLifecycleControllersGroup = "compute.ironcore.dev:system:pool-lifecycle-controllers"

	// PoolLifecycleControllerCommonName is the fixed common name of the pool-lifecycle-controller cert.
	PoolLifecycleControllerCommonName = "compute.ironcore.dev:system:pool-lifecycle-controller"

	// APIServerMachinePoolletClientsGroup is the system rbac group the ironcore apiserver's
	// machinepoollet client identity is in. Membership is authorized by the poollet (via a
	// delegating SubjectAccessReview back to the ironcore apiserver) for machinepools/proxy.
	APIServerMachinePoolletClientsGroup = "compute.ironcore.dev:system:apiserver-machinepoollet-clients"

	// APIServerMachinePoolletClientCommonName is the fixed common name of the ironcore
	// apiserver's machinepoollet client cert (used when it dials a poollet).
	APIServerMachinePoolletClientCommonName = "compute.ironcore.dev:system:apiserver-machinepoollet-client"
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
