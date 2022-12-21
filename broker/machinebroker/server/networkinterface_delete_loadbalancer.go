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

	machinebrokerv1alpha1 "github.com/onmetal/onmetal-api/broker/machinebroker/api/v1alpha1"
	ori "github.com/onmetal/onmetal-api/ori/apis/machine/v1alpha1"
	"golang.org/x/exp/slices"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

func (s *Server) DeleteNetworkInterfaceLoadBalancerTarget(ctx context.Context, req *ori.DeleteNetworkInterfaceLoadBalancerTargetRequest) (res *ori.DeleteNetworkInterfaceLoadBalancerTargetResponse, retErr error) {
	networkInterfaceID := req.NetworkInterfaceId
	log := s.loggerFrom(ctx, "NetworkInterfaceID", networkInterfaceID)

	loadBalancerTarget, err := s.prepareOnmetalLoadBalancerTarget(req.LoadBalancerTarget)
	if err != nil {
		return nil, fmt.Errorf("error converting load balancer: %w", err)
	}

	log.V(1).Info("Getting onmetal network interface")
	onmetalNetworkInterface, err := s.getAggregateOnmetalNetworkInterface(ctx, networkInterfaceID)
	if err != nil {
		return nil, err
	}

	if !slices.ContainsFunc(onmetalNetworkInterface.LoadBalancerTargets,
		func(tgt machinebrokerv1alpha1.LoadBalancerTarget) bool { return tgt.Key() == loadBalancerTarget.Key() }) {
		return nil, status.Errorf(codes.NotFound, "network interface %s load balancer %s not found", networkInterfaceID, loadBalancerTarget.Key())
	}

	if err := s.loadBalancers.Delete(ctx, onmetalNetworkInterface.Network.Spec.Handle, *loadBalancerTarget, onmetalNetworkInterface.NetworkInterface); err != nil {
		return nil, fmt.Errorf("error deleting load balancer: %w", err)
	}
	return &ori.DeleteNetworkInterfaceLoadBalancerTargetResponse{}, nil
}
