// SPDX-FileCopyrightText: 2023 SAP SE or an SAP affiliate company and IronCore contributors
// SPDX-License-Identifier: Apache-2.0

package machinevolumedevices

import (
	"context"
	"fmt"
	"io"

	"github.com/ironcore-dev/ironcore/internal/admission/plugin/machinevolumedevices/device"
	"github.com/ironcore-dev/ironcore/internal/apis/compute"
	apierrors "k8s.io/apimachinery/pkg/api/errors"
	"k8s.io/apiserver/pkg/admission"
)

// PluginName indicates name of admission plugin.
const PluginName = "MachineVolumeDevices"

// Register registers a plugin
func Register(plugins *admission.Plugins) {
	plugins.Register(PluginName, func(config io.Reader) (admission.Interface, error) {
		return NewMachineVolumeDevices(), nil
	})
}

type MachineVolumeDevices struct {
	*admission.Handler
}

func NewMachineVolumeDevices() *MachineVolumeDevices {
	return &MachineVolumeDevices{
		Handler: admission.NewHandler(admission.Create, admission.Update),
	}
}

func (d *MachineVolumeDevices) Admit(ctx context.Context, a admission.Attributes, o admission.ObjectInterfaces) error {
	if shouldIgnore(a) {
		return nil
	}

	machine, ok := a.GetObject().(*compute.Machine)
	if !ok {
		return apierrors.NewBadRequest("Resource was marked with kind Machine but was unable to be converted")
	}

	namer, err := deviceNamerFromMachineVolumes(machine)
	if err != nil {
		return apierrors.NewBadRequest("Machine has conflicting volume device names")
	}

	for i := range machine.Spec.Volumes {
		volume := &machine.Spec.Volumes[i]
		if volume.Device != "" {
			continue
		}

		newDevice, err := namer.Generate(device.IronCorePrefix) // TODO: We should have a better way for a device prefix.
		if err != nil {
			return apierrors.NewBadRequest("No device names left for machine")
		}

		volume.Device = newDevice
	}

	return nil
}

func shouldIgnore(a admission.Attributes) bool {
	if a.GetKind().GroupKind() != compute.Kind("Machine") {
		return true
	}

	machine, ok := a.GetObject().(*compute.Machine)
	if !ok {
		return true
	}

	return !machineHasAnyVolumeWithoutDevice(machine)
}

func machineHasAnyVolumeWithoutDevice(machine *compute.Machine) bool {
	for _, volume := range machine.Spec.Volumes {
		if volume.Device == "" {
			return true
		}
	}
	return false
}

func deviceNamerFromMachineVolumes(machine *compute.Machine) (*device.Namer, error) {
	namer := device.NewNamer()
	for _, volume := range machine.Spec.Volumes {
		if dev := volume.Device; dev != "" {
			if err := namer.Observe(dev); err != nil {
				return nil, fmt.Errorf("error observing device %s: %w", dev, err)
			}
		}
	}
	return namer, nil
}
