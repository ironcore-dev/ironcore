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

	commonv1alpha1 "github.com/onmetal/onmetal-api/api/common/v1alpha1"
	networkingv1alpha1 "github.com/onmetal/onmetal-api/api/networking/v1alpha1"
	ori "github.com/onmetal/onmetal-api/ori/apis/machine/v1alpha1"
)

func (s *Server) UpdateNetworkPeerings(ctx context.Context, req *ori.UpdateNetworkPeeringsRequest) (*ori.UpdateNetworkPeeringsResponse, error) {
	handle := req.Handle
	log := s.loggerFrom(ctx, "Handle", handle, "Peerings", req.Peerings)

	nw := &networkingv1alpha1.Network{
		Spec: networkingv1alpha1.NetworkSpec{
			Handle: handle,
		},
	}

	peerings := []networkingv1alpha1.NetworkPeeringStatus{}
	//TODO fix duplicate code
	for _, peeringSpec := range req.Peerings {

		peering := networkingv1alpha1.NetworkPeeringStatus{
			Name:          peeringSpec.Name,
			NetworkHandle: peeringSpec.Handle,
			Phase:         networkingv1alpha1.NetworkPeeringPhase(peeringSpec.Phase),
		}

		prefixes := []commonv1alpha1.IPPrefix{}
		if len(peeringSpec.Prefixes) > 0 {
			for _, prefix := range peeringSpec.Prefixes {
				prefixes = append(prefixes, commonv1alpha1.MustParseIPPrefix(prefix))
			}
		}

		if len(prefixes) > 0 {
			peering.Prefixes = &prefixes
		}
		peerings = append(peerings, peering)
	}

	nw.Status.Peerings = peerings
	log.Info("Updating", "Network", nw)
	_, networkTransaction, err := s.networks.BeginCreate(ctx, nw)
	if err != nil {
		return nil, fmt.Errorf("error getting network: %w", err)
	}
	if err := networkTransaction.Commit(nw); err != nil {
		return nil, fmt.Errorf("error update network: %w", err)
	}

	return &ori.UpdateNetworkPeeringsResponse{}, nil
}
