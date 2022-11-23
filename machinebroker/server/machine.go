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
	machinebrokerv1alpha1 "github.com/onmetal/onmetal-api/machinebroker/api/v1alpha1"
	"github.com/onmetal/onmetal-api/machinebroker/apiutils"
	ori "github.com/onmetal/onmetal-api/ori/apis/machine/v1alpha1"
	"github.com/onmetal/onmetal-api/utils/slices"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

func (s *Server) convertOnmetalMachine(machine *computev1alpha1.Machine) (*ori.Machine, error) {
	id := machine.Labels[machinebrokerv1alpha1.MachineIDLabel]

	metadata, err := apiutils.GetMetadataAnnotation(machine)
	if err != nil {
		return nil, err
	}

	labels, err := apiutils.GetLabelsAnnotation(machine)
	if err != nil {
		return nil, err
	}

	annotations, err := apiutils.GetAnnotationsAnnotation(machine)
	if err != nil {
		return nil, err
	}

	var deletedAt int64
	if deletionTimestamp := machine.DeletionTimestamp; !deletionTimestamp.IsZero() {
		deletedAt = deletionTimestamp.UnixNano()
	}

	return &ori.Machine{
		Id:          id,
		Metadata:    metadata,
		State:       s.convertOnmetalMachineState(machine.Status.State),
		CreatedAt:   machine.CreationTimestamp.UnixNano(),
		DeletedAt:   deletedAt,
		Annotations: annotations,
		Labels:      labels,
	}, nil
}

func (s *Server) getOrListOnmetalMachinesByID(ctx context.Context, machineID string) (map[string]computev1alpha1.Machine, error) {
	var machinesByID map[string]computev1alpha1.Machine
	if machineID != "" {
		machine, err := s.getOnmetalMachine(ctx, machineID)
		if err != nil {
			if status.Code(err) != codes.NotFound {
				return nil, fmt.Errorf("error getting machine %s: %w", machineID, err)
			}
			return nil, nil
		}

		machinesByID = map[string]computev1alpha1.Machine{
			machine.Name: *machine,
		}
	} else {
		machines, err := s.listOnmetalMachines(ctx)
		if err != nil {
			return nil, fmt.Errorf("error listing machines: %w", err)
		}

		machinesByID = slices.ToMap(machines, func(machine computev1alpha1.Machine) string {
			return machine.Name
		})
	}
	return machinesByID, nil
}
