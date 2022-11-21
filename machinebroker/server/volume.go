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
	ori "github.com/onmetal/onmetal-api/ori/apis/compute/v1alpha1"
	"golang.org/x/exp/slices"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
	apierrors "k8s.io/apimachinery/pkg/api/errors"
	"sigs.k8s.io/controller-runtime/pkg/client"
)

type OnmetalEmptyDiskVolume struct {
	Machine *computev1alpha1.Machine
	Volume  *computev1alpha1.Volume
}

type OnmetalVolume struct {
	Volume    *storagev1alpha1.Volume
	EmptyDisk *OnmetalEmptyDiskVolume
}

func (s *Server) getOnmetalVolume(ctx context.Context, machineID, name string) (*OnmetalVolume, error) {
	onmetalVolume := &storagev1alpha1.Volume{}
	onmetalVolumeKey := client.ObjectKey{Namespace: s.namespace, Name: s.onmetalVolumeName(machineID, name)}
	if err := s.client.Get(ctx, onmetalVolumeKey, onmetalVolume); err != nil {
		if !apierrors.IsNotFound(err) {
			return nil, fmt.Errorf("error getting onmetal volume %s: %w", onmetalVolumeKey, err)
		}
	} else {
		return &OnmetalVolume{
			Volume: onmetalVolume,
		}, nil
	}

	onmetalMachine, err := s.getOnmetalMachine(ctx, machineID)
	if err != nil {
		if status.Code(err) != codes.NotFound {
			return nil, fmt.Errorf("error getting volume %s machine %s: %w", name, machineID, err)
		}
		return nil, status.Errorf(codes.NotFound, "machine %s volume %s not found", machineID, name)
	}

	idx := slices.IndexFunc(onmetalMachine.Spec.Volumes, func(volume computev1alpha1.Volume) bool {
		return volume.Name == name && volume.EmptyDisk != nil
	})
	if idx < 0 {
		return nil, status.Errorf(codes.NotFound, "machine %s volume %s not found", machineID, name)
	}
	volume := onmetalMachine.Spec.Volumes[idx]
	return &OnmetalVolume{
		EmptyDisk: &OnmetalEmptyDiskVolume{
			Machine: onmetalMachine,
			Volume:  &volume,
		},
	}, nil
}

func (s *Server) convertOnmetalVolume(
	onmetalVolume *OnmetalVolume,
) (*ori.Volume, error) {
	switch {
	case onmetalVolume.EmptyDisk != nil:
		machineMetadata, err := apiutils.GetMetadataAnnotation(onmetalVolume.EmptyDisk.Machine)
		if err != nil {
			return nil, err
		}

		var sizeLimitBytes uint64
		if sizeLimit := onmetalVolume.EmptyDisk.Volume.EmptyDisk.SizeLimit; sizeLimit != nil {
			sizeLimitBytes = uint64(sizeLimit.Value())
		}

		emptyDisk := &ori.EmptyDisk{
			SizeLimitBytes: sizeLimitBytes,
		}
		return &ori.Volume{
			MachineId:       onmetalVolume.EmptyDisk.Machine.Name,
			MachineMetadata: machineMetadata,
			Name:            onmetalVolume.EmptyDisk.Volume.Name,
			Device:          *onmetalVolume.EmptyDisk.Volume.Device,
			EmptyDisk:       emptyDisk,
		}, nil
	case onmetalVolume.Volume != nil:
		machineID := onmetalVolume.Volume.Labels[machinebrokerv1alpha1.MachineIDLabel]
		name := onmetalVolume.Volume.Labels[machinebrokerv1alpha1.VolumeNameLabel]
		device := onmetalVolume.Volume.Labels[machinebrokerv1alpha1.DeviceLabel]

		machineMetadata, err := apiutils.GetMetadataAnnotation(onmetalVolume.Volume)
		if err != nil {
			return nil, err
		}

		onmetalVolumeAccess := onmetalVolume.Volume.Status.Access
		if onmetalVolumeAccess == nil {
			return nil, fmt.Errorf("onmetal volume %s/%s does not specify access", onmetalVolume.Volume.Namespace, onmetalVolume.Volume.Name)
		}

		access := &ori.VolumeAccess{
			Driver: onmetalVolumeAccess.Driver,
			Handle: onmetalVolumeAccess.Handle,
		}

		return &ori.Volume{
			MachineId:       machineID,
			MachineMetadata: machineMetadata,
			Name:            name,
			Device:          device,
			Access:          access,
		}, nil
	default:
		return nil, fmt.Errorf("both onmetal empty disk volume and onmetal volume are empty")
	}
}

func (s *Server) getVolume(ctx context.Context, machineID, name string) (*ori.Volume, error) {
	onmetalVolume, err := s.getOnmetalVolume(ctx, machineID, name)
	if err != nil {
		return nil, err
	}

	return s.convertOnmetalVolume(onmetalVolume)
}
