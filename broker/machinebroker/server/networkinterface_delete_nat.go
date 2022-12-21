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
	machinebrokerv1alpha1 "github.com/onmetal/onmetal-api/broker/machinebroker/api/v1alpha1"
	ori "github.com/onmetal/onmetal-api/ori/apis/machine/v1alpha1"
	"golang.org/x/exp/slices"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

func (s *Server) DeleteNetworkInterfaceNAT(ctx context.Context, req *ori.DeleteNetworkInterfaceNATRequest) (res *ori.DeleteNetworkInterfaceNATResponse, retErr error) {
	networkInterfaceID := req.NetworkInterfaceId
	log := s.loggerFrom(ctx, "NetworkInterfaceID", networkInterfaceID)

	natIP, err := commonv1alpha1.ParseIP(req.NatIp)
	if err != nil {
		return nil, fmt.Errorf("error parsing nat ip: %w", err)
	}

	log.V(1).Info("Getting onmetal network interface")
	onmetalNetworkInterface, err := s.getAggregateOnmetalNetworkInterface(ctx, networkInterfaceID)
	if err != nil {
		return nil, err
	}

	if !slices.ContainsFunc(onmetalNetworkInterface.NATGatewayTargets,
		func(tgt machinebrokerv1alpha1.NATGatewayTarget) bool { return tgt.IP == natIP }) {
		return nil, status.Errorf(codes.NotFound, "network interface %s nat %s not found", networkInterfaceID, natIP)
	}

	if err := s.natGateways.Delete(ctx, onmetalNetworkInterface.Network.Spec.Handle, natIP, onmetalNetworkInterface.NetworkInterface); err != nil {
		return nil, fmt.Errorf("error deleting nat: %w", err)
	}
	return &ori.DeleteNetworkInterfaceNATResponse{}, nil
}
