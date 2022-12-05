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

func (s *Server) deleteOnmetalVolumeAttachment(ctx context.Context, onmetalMachine *computev1alpha1.Machine, idx int) error {
	baseOnmetalMachine := onmetalMachine.DeepCopy()
	onmetalMachine.Spec.Volumes = slices.Delete(onmetalMachine.Spec.Volumes, idx, idx+1)
	if err := s.client.Patch(ctx, onmetalMachine, client.StrategicMergeFrom(baseOnmetalMachine)); err != nil {
		return fmt.Errorf("error patching onmetal machine volumes: %w", err)
	}
	return nil
}

func (s *Server) DeleteVolumeAttachment(ctx context.Context, req *ori.DeleteVolumeAttachmentRequest) (*ori.DeleteVolumeAttachmentResponse, error) {
	machineID := req.MachineId
	name := req.Name
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
	if idx < 0 {
		return nil, status.Errorf(codes.NotFound, "machine %s does have volume attachment %s", machineID, name)
	}

	log.V(1).Info("Deleting onmetal machine volume")
	if err := s.deleteOnmetalVolumeAttachment(ctx, aggOnmetalMachine.Machine, idx); err != nil {
		return nil, fmt.Errorf("error adding onmetal machine volume: %w", err)
	}

	return &ori.DeleteVolumeAttachmentResponse{}, nil
}
