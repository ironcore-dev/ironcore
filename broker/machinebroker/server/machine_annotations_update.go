// Copyright 2022 IronCore authors
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
