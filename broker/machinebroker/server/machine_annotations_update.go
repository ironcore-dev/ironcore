// SPDX-FileCopyrightText: 2023 SAP SE or an SAP affiliate company and IronCore contributors
// SPDX-License-Identifier: Apache-2.0

package server

import (
	"context"
	"fmt"

	"github.com/ironcore-dev/ironcore/broker/machinebroker/apiutils"
	iri "github.com/ironcore-dev/ironcore/iri/apis/machine/v1alpha1"
	"sigs.k8s.io/controller-runtime/pkg/client"
)

func (s *Server) UpdateMachineAnnotations(ctx context.Context, req *iri.UpdateMachineAnnotationsRequest) (*iri.UpdateMachineAnnotationsResponse, error) {
	machineID := req.MachineId
	log := s.loggerFrom(ctx, "MachineID", machineID)

	log.V(1).Info("Getting ironcore machine")
	aggIronCoreMachine, err := s.getAggregateIronCoreMachine(ctx, machineID)
	if err != nil {
		return nil, err
	}

	base := aggIronCoreMachine.Machine.DeepCopy()
	if err := apiutils.SetAnnotationsAnnotation(aggIronCoreMachine.Machine, req.Annotations); err != nil {
		return nil, err
	}

	log.V(1).Info("Patching ironcore machine annotations")
	if err := s.cluster.Client().Patch(ctx, aggIronCoreMachine.Machine, client.MergeFrom(base)); err != nil {
		return nil, fmt.Errorf("error patching ironcore machine annotations: %w", err)
	}

	return &iri.UpdateMachineAnnotationsResponse{}, nil
}
