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
	ori "github.com/onmetal/onmetal-api/ori/apis/machine/v1alpha1"
	"golang.org/x/exp/slices"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
	"sigs.k8s.io/controller-runtime/pkg/client"
)

func (s *Server) addOnmetalVolumeAttachment(ctx context.Context, onmetalMachine *computev1alpha1.Machine, volume computev1alpha1.Volume) error {
	baseOnmetalMachine := onmetalMachine.DeepCopy()
	onmetalMachine.Spec.Volumes = append(onmetalMachine.Spec.Volumes, volume)
	if err := s.cluster.Client().Patch(ctx, onmetalMachine, client.StrategicMergeFrom(baseOnmetalMachine)); err != nil {
		return fmt.Errorf("error patching onmetal machine volumes: %w", err)
	}
	return nil
}

func (s *Server) CreateVolumeAttachment(ctx context.Context, req *ori.CreateVolumeAttachmentRequest) (*ori.CreateVolumeAttachmentResponse, error) {
	machineID := req.MachineId
	name := req.Volume.Name
	log := s.loggerFrom(ctx, "MachineID", machineID, "Name", name)

	log.V(1).Info("Getting aggregated onmetal machine")
	aggOnmetalMachine, err := s.getAggregateOnmetalMachine(ctx, machineID)
	if err != nil {
		return nil, err
	}

	machine, err := s.convertAggregateOnmetalMachine(aggOnmetalMachine)
	if err != nil {
		return nil, err
	}

	idx := slices.IndexFunc(
		machine.Spec.Volumes,
		func(volume *ori.VolumeAttachment) bool { return volume.Name == name },
	)
	if idx >= 0 {
		return nil, status.Errorf(codes.AlreadyExists, "machine %s already has volume attachment %s", machineID, name)
	}

	onmetalVolumeAttachment, err := s.prepareOnmetalVolumeAttachment(req.Volume)
	if err != nil {
		return nil, fmt.Errorf("error preparing onmetal machine volume: %w", err)
	}

	log.V(1).Info("Adding onmetal machine volume")
	if err := s.addOnmetalVolumeAttachment(ctx, aggOnmetalMachine.Machine, onmetalVolumeAttachment); err != nil {
		return nil, fmt.Errorf("error adding onmetal machine volume: %w", err)
	}

	return &ori.CreateVolumeAttachmentResponse{}, nil
}
