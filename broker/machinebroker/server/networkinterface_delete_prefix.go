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
	ori "github.com/onmetal/onmetal-api/ori/apis/machine/v1alpha1"
	"golang.org/x/exp/slices"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

func (s *Server) DeleteNetworkInterfacePrefix(ctx context.Context, req *ori.DeleteNetworkInterfacePrefixRequest) (res *ori.DeleteNetworkInterfacePrefixResponse, retErr error) {
	networkInterfaceID := req.NetworkInterfaceId
	log := s.loggerFrom(ctx, "NetworkInterfaceID", networkInterfaceID)

	prefix, err := commonv1alpha1.ParseIPPrefix(req.Prefix)
	if err != nil {
		return nil, fmt.Errorf("error parsing prefix: %w", err)
	}

	log.V(1).Info("Getting onmetal network interface")
	onmetalNetworkInterface, err := s.getAggregateOnmetalNetworkInterface(ctx, networkInterfaceID)
	if err != nil {
		return nil, err
	}

	actualNetworkInterface, err := s.convertAggregateOnmetalNetworkInterface(onmetalNetworkInterface)
	if err != nil {
		return nil, err
	}

	if !slices.Contains(actualNetworkInterface.Spec.Prefixes, req.Prefix) {
		return nil, status.Errorf(codes.NotFound, "network interface %s prefix %s not found", networkInterfaceID, req.Prefix)
	}

	if err := s.aliasPrefixes.Delete(ctx, onmetalNetworkInterface.Network.Spec.Handle, prefix, onmetalNetworkInterface.NetworkInterface); err != nil {
		return nil, fmt.Errorf("error creating prefix: %w", err)
	}

	return &ori.DeleteNetworkInterfacePrefixResponse{}, nil
}
