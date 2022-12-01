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

	"github.com/gogo/protobuf/proto"
	commonv1alpha1 "github.com/onmetal/onmetal-api/api/common/v1alpha1"
	ori "github.com/onmetal/onmetal-api/ori/apis/machine/v1alpha1"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

func (s *Server) UpdateNetworkInterfaceVirtualIP(ctx context.Context, req *ori.UpdateNetworkInterfaceVirtualIPRequest) (res *ori.UpdateNetworkInterfaceVirtualIPResponse, retErr error) {
	networkInterfaceID := req.NetworkInterfaceId
	log := s.loggerFrom(ctx, "NetworkInterfaceID", networkInterfaceID)

	log.V(1).Info("Getting onmetal network interface")
	onmetalNetworkInterface, err := s.getAggregateOnmetalNetworkInterface(ctx, networkInterfaceID)
	if err != nil {
		return nil, err
	}

	actualNetworkInterface, err := s.convertAggregateOnmetalNetworkInterface(onmetalNetworkInterface)
	if err != nil {
		return nil, err
	}

	actualVirtualIP := actualNetworkInterface.Spec.VirtualIp
	desiredVirtualIP := req.VirtualIp
	if proto.Equal(desiredVirtualIP, actualVirtualIP) {
		return &ori.UpdateNetworkInterfaceVirtualIPResponse{}, nil
	}

	if actualVirtualIP == nil {
		return nil, status.Errorf(codes.NotFound, "network interface %s does not have a virtual ip", networkInterfaceID)
	}

	ip, err := commonv1alpha1.ParseIP(desiredVirtualIP.Ip)
	if err != nil {
		return nil, fmt.Errorf("error parsing ip: %w", err)
	}

	log.V(1).Info("Setting virtual ip status ip")
	if err := s.setOnmetalVirtualIPStatusIP(ctx, onmetalNetworkInterface.VirtualIP, &ip); err != nil {
		return nil, fmt.Errorf("error setting virtual ip status ip: %w", err)
	}

	return &ori.UpdateNetworkInterfaceVirtualIPResponse{}, nil
}
