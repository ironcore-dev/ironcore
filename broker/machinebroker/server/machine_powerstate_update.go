// SPDX-FileCopyrightText: 2023 SAP SE or an SAP affiliate company and IronCore contributors
// SPDX-License-Identifier: Apache-2.0

package server

import (
	"context"
	"fmt"

	iri "github.com/ironcore-dev/ironcore/iri/apis/machine/v1alpha1"
	"sigs.k8s.io/controller-runtime/pkg/client"
)

func (s *Server) UpdateMachinePower(ctx context.Context, req *iri.UpdateMachinePowerRequest) (*iri.UpdateMachinePowerResponse, error) {
	machineID := req.MachineId
	log := s.loggerFrom(ctx, "MachineID", machineID)

	log.V(1).Info("Getting ironcore machine")
	aggIronCoreMachine, err := s.getAggregateIronCoreMachine(ctx, machineID)
	if err != nil {
		return nil, err
	}

	power, err := s.prepareIronCoreMachinePower(req.Power)
	if err != nil {
		return nil, err
	}

	base := aggIronCoreMachine.Machine.DeepCopy()
	aggIronCoreMachine.Machine.Spec.Power = power
	log.V(1).Info("Patching ironcore machine power")
	if err := s.cluster.Client().Patch(ctx, aggIronCoreMachine.Machine, client.MergeFrom(base)); err != nil {
		return nil, fmt.Errorf("error patching ironcore machine power: %w", err)
	}

	return &iri.UpdateMachinePowerResponse{}, nil
}
