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

	ori "github.com/onmetal/onmetal-api/ori/apis/compute/v1alpha1"
	"sigs.k8s.io/controller-runtime/pkg/client"
)

func (s *Server) CreateVolume(ctx context.Context, req *ori.CreateVolumeRequest) (res *ori.CreateVolumeResponse, retErr error) {
	log := s.loggerFrom(ctx)

	cleaner, cleanup := s.setupCleaner(ctx, log, &retErr)
	defer cleanup()

	machine, err := s.getMachine(ctx, req.MachineId)
	if err != nil {
		return nil, err
	}

	onmetalMachineVolume, onmetalVolumeConfig, err := s.getOnmetalVolumeData(req.MachineId, machine.Metadata, req.Config)
	if err != nil {
		return nil, err
	}

	if onmetalVolumeConfig != nil {
		if err := s.createOnmetalVolume(ctx, log, cleaner, *onmetalVolumeConfig); err != nil {
			return nil, err
		}
	}

	onmetalMachine, err := s.getOnmetalMachine(ctx, req.MachineId)
	if err != nil {
		return nil, err
	}

	baseOnmetalMachine := onmetalMachine.DeepCopy()
	onmetalMachine.Spec.Volumes = append(onmetalMachine.Spec.Volumes, *onmetalMachineVolume)
	if err := s.client.Patch(ctx, onmetalMachine, client.MergeFrom(baseOnmetalMachine)); err != nil {
		return nil, fmt.Errorf("error patching onmetal machine volumes: %w", err)
	}

	return &ori.CreateVolumeResponse{}, nil
}
