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

	"github.com/go-logr/logr"
	ori "github.com/onmetal/onmetal-api/ori/apis/machine/v1alpha1"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
	"sigs.k8s.io/controller-runtime/pkg/client"
)

func (s *Server) DeleteNetworkInterfaceVirtualIP(ctx context.Context, req *ori.DeleteNetworkInterfaceVirtualIPRequest) (*ori.DeleteNetworkInterfaceVirtualIPResponse, error) {
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
	if actualVirtualIP == nil {
		return nil, status.Errorf(codes.NotFound, "network interface %s does not have a virtual ip", networkInterfaceID)
	}

	if err := s.deleteOnmetalVirtualIP(ctx, log, onmetalNetworkInterface); err != nil {
		return nil, err
	}

	return &ori.DeleteNetworkInterfaceVirtualIPResponse{}, nil
}

func (s *Server) deleteOnmetalVirtualIP(ctx context.Context, log logr.Logger, onmetalNetworkInterface *AggregateOnmetalNetworkInterface) error {
	log.V(1).Info("Removing virtual ip source from onmetal network interface")
	if err := s.setOnmetalNetworkInterfaceVirtualIPSource(ctx, onmetalNetworkInterface.NetworkInterface, nil); err != nil {
		return fmt.Errorf("error removing virtual ip source: %w", err)
	}

	log.V(1).Info("Deleting outdated onmetal virtual ip")
	if err := s.client.Delete(ctx, onmetalNetworkInterface.VirtualIP); client.IgnoreNotFound(err) != nil {
		log.Error(err, "Error deleting outdated onmetal network interface virtual ip")
	}
	return nil
}
