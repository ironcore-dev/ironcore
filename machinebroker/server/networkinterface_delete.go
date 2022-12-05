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

	ori "github.com/onmetal/onmetal-api/ori/apis/machine/v1alpha1"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
	apierrors "k8s.io/apimachinery/pkg/api/errors"
)

func (s *Server) DeleteNetworkInterface(ctx context.Context, req *ori.DeleteNetworkInterfaceRequest) (*ori.DeleteNetworkInterfaceResponse, error) {
	networkInterfaceID := req.NetworkInterfaceId
	log := s.loggerFrom(ctx, "NetworkInterfaceID", networkInterfaceID)

	log.V(1).Info("Getting aggregate onmetal network interface")
	aggOnmetalNetworkInterface, err := s.getAggregateOnmetalNetworkInterface(ctx, networkInterfaceID)
	if err != nil {
		return nil, err
	}

	log.V(1).Info("Deleting onmetal network interface")
	if err := s.client.Delete(ctx, aggOnmetalNetworkInterface.NetworkInterface); err != nil {
		if !apierrors.IsNotFound(err) {
			return nil, fmt.Errorf("error deleting onmetal network interface: %w", err)
		}
		return nil, status.Errorf(codes.NotFound, "network interface %s not found", networkInterfaceID)
	}

	return &ori.DeleteNetworkInterfaceResponse{}, nil
}
