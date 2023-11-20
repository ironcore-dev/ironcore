// SPDX-FileCopyrightText: 2023 SAP SE or an SAP affiliate company and IronCore contributors
// SPDX-License-Identifier: Apache-2.0

package server

import (
	"context"
	"fmt"

	iri "github.com/ironcore-dev/ironcore/iri/apis/machine/v1alpha1"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
	apierrors "k8s.io/apimachinery/pkg/api/errors"
)

func (s *Server) DeleteMachine(ctx context.Context, req *iri.DeleteMachineRequest) (*iri.DeleteMachineResponse, error) {
	machineID := req.MachineId
	log := s.loggerFrom(ctx, "MachineID", machineID)

	ironcoreMachine, err := s.getAggregateIronCoreMachine(ctx, machineID)
	if err != nil {
		return nil, err
	}

	log.V(1).Info("Deleting ironcore machine")
	if err := s.cluster.Client().Delete(ctx, ironcoreMachine.Machine); err != nil {
		if !apierrors.IsNotFound(err) {
			return nil, fmt.Errorf("error deleting ironcore machine: %w", err)
		}
		return nil, status.Errorf(codes.NotFound, "machine %s not found", machineID)
	}

	return &iri.DeleteMachineResponse{}, nil
}
