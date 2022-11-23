// Copyright 2022 OnMetal authors
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
//      http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.

package server

import (
	"context"
	"fmt"

	computev1alpha1 "github.com/onmetal/onmetal-api/api/compute/v1alpha1"
	storagev1alpha1 "github.com/onmetal/onmetal-api/api/storage/v1alpha1"
	machinebrokerv1alpha1 "github.com/onmetal/onmetal-api/machinebroker/api/v1alpha1"
	"github.com/onmetal/onmetal-api/machinebroker/apiutils"
	ori "github.com/onmetal/onmetal-api/ori/apis/machine/v1alpha1"
	"golang.org/x/exp/slices"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
	apierrors "k8s.io/apimachinery/pkg/api/errors"
	"sigs.k8s.io/controller-runtime/pkg/client"
)

type OnmetalVolume struct {
	Volume    *storagev1alpha1.Volume
	EmptyDisk *OnmetalEmptyDiskVolume
}

type OnmetalEmptyDiskVolume struct {
	MachineID string
	Name      string
	Device    string
	computev1alpha1.EmptyDiskVolumeSource
}

func (o *OnmetalVolume) Name() string {
	switch {
	case o.Volume != nil:
		return o.Volume.Labels[machinebrokerv1alpha1.VolumeNameLabel]
	case o.EmptyDisk != nil:
		return o.EmptyDisk.Name
	default:
		panic("both volume and empty disk are unset")
	}
}

func (o *OnmetalVolume) MachineID() string {
	switch {
	case o.Volume != nil:
		return o.Volume.Labels[machinebrokerv1alpha1.MachineIDLabel]
	case o.EmptyDisk != nil:
		return o.EmptyDisk.MachineID
	default:
		panic("both volume and empty disk are unset")
	}
}

func (s *Server) onmetalEmptyDiskVolume(
	onmetalMachine *computev1alpha1.Machine,
	onmetalMachineVolume computev1alpha1.Volume,
) *OnmetalEmptyDiskVolume {
	return &OnmetalEmptyDiskVolume{
		MachineID:             onmetalMachine.Name,
		Name:                  onmetalMachineVolume.Name,
		Device:                *onmetalMachineVolume.Device,
		EmptyDiskVolumeSource: *onmetalMachineVolume.EmptyDisk,
	}
}

func (s *Server) getOnmetalVolume(ctx context.Context, onmetalMachine *computev1alpha1.Machine, name string) (*OnmetalVolume, error) {
	idx := slices.IndexFunc(onmetalMachine.Spec.Volumes, func(volume computev1alpha1.Volume) bool {
		return volume.Name == name
	})
	if idx == -1 {
		return nil, status.Errorf(codes.NotFound, "machine %s volume %s not found", onmetalMachine.Name, name)
	}

	onmetalMachineVolume := onmetalMachine.Spec.Volumes[idx]
	switch {
	case onmetalMachineVolume.EmptyDisk != nil:
		return &OnmetalVolume{
			EmptyDisk: s.onmetalEmptyDiskVolume(onmetalMachine, onmetalMachineVolume),
		}, nil
	case onmetalMachineVolume.VolumeRef != nil:
		onmetalVolume := &storagev1alpha1.Volume{}
		onmetalVolumeKey := client.ObjectKey{Namespace: s.namespace, Name: s.onmetalVolumeName(onmetalMachine.Name, name)}
		if err := s.client.Get(ctx, onmetalVolumeKey, onmetalVolume); err != nil {
			if !apierrors.IsNotFound(err) {
				return nil, fmt.Errorf("error getting onmetal volume %s: %w", onmetalVolumeKey, err)
			}
			return nil, status.Errorf(codes.NotFound, "machine %s volume %s not found", onmetalMachine.Name, name)
		}
		return &OnmetalVolume{
			Volume: onmetalVolume,
		}, nil
	default:
		return nil, fmt.Errorf("machine %s contains unrecognized volume %#v", onmetalMachine.Name, onmetalMachineVolume)
	}
}

func (s *Server) convertOnmetalVolume(
	machine *computev1alpha1.Machine,
	volume *OnmetalVolume,
) (*ori.Volume, error) {
	volumeName := volume.Name()
	idx := slices.IndexFunc(machine.Status.Volumes,
		func(volume computev1alpha1.VolumeStatus) bool {
			return volume.Name == volumeName
		},
	)
	state := ori.VolumeState_VOLUME_DETACHED
	if idx >= 0 {
		state = s.convertOnmetalVolumeState(machine.Status.Volumes[idx].State)
	}

	machineMetadata, err := apiutils.GetMetadataAnnotation(machine)
	if err != nil {
		return nil, err
	}

	switch {
	case volume.EmptyDisk != nil:
		var sizeLimitBytes uint64
		if sizeLimit := volume.EmptyDisk.SizeLimit; sizeLimit != nil {
			sizeLimitBytes = uint64(sizeLimit.Value())
		}

		emptyDisk := &ori.EmptyDisk{
			SizeLimitBytes: sizeLimitBytes,
		}
		return &ori.Volume{
			MachineId:       machine.Name,
			MachineMetadata: machineMetadata,
			Name:            volume.EmptyDisk.Name,
			Device:          volume.EmptyDisk.Device,
			EmptyDisk:       emptyDisk,
			State:           state,
		}, nil
	case volume.Volume != nil:
		name := volume.Volume.Labels[machinebrokerv1alpha1.VolumeNameLabel]
		device := volume.Volume.Labels[machinebrokerv1alpha1.DeviceLabel]

		onmetalVolumeAccess := volume.Volume.Status.Access
		if onmetalVolumeAccess == nil {
			return nil, fmt.Errorf("onmetal volume %s/%s does not specify access", volume.Volume.Namespace, volume.Volume.Name)
		}

		access := &ori.VolumeAccess{
			Driver: onmetalVolumeAccess.Driver,
			Handle: onmetalVolumeAccess.Handle,
		}

		return &ori.Volume{
			MachineId:       machine.Name,
			MachineMetadata: machineMetadata,
			Name:            name,
			Device:          device,
			Access:          access,
			State:           state,
		}, nil
	default:
		return nil, fmt.Errorf("both onmetal empty disk volume and onmetal volume are empty")
	}
}

func (s *Server) getVolume(ctx context.Context, machineID, name string) (*ori.Volume, error) {
	onmetalMachine, err := s.getOnmetalMachine(ctx, machineID)
	if err != nil {
		return nil, err
	}

	onmetalVolume, err := s.getOnmetalVolume(ctx, onmetalMachine, name)
	if err != nil {
		return nil, err
	}

	return s.convertOnmetalVolume(onmetalMachine, onmetalVolume)
}
