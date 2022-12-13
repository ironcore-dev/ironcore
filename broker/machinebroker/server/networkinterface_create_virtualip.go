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

	networkingv1alpha1 "github.com/onmetal/onmetal-api/api/networking/v1alpha1"
	"github.com/onmetal/onmetal-api/broker/machinebroker/apiutils"
	ori "github.com/onmetal/onmetal-api/ori/apis/machine/v1alpha1"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
	corev1 "k8s.io/api/core/v1"
)

func (s *Server) CreateNetworkInterfaceVirtualIP(ctx context.Context, req *ori.CreateNetworkInterfaceVirtualIPRequest) (res *ori.CreateNetworkInterfaceVirtualIPResponse, retErr error) {
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
	if actualVirtualIP != nil {
		return nil, status.Errorf(codes.AlreadyExists, "network interface %s already has a virtual ip", networkInterfaceID)
	}

	newOnmetalVirtualIP, err := s.prepareOnmetalVirtualIP(desiredVirtualIP)
	if err != nil {
		return nil, err
	}

	c, cleanup := s.setupCleaner(ctx, log, &retErr)
	defer cleanup()

	log.V(1).Info("Creating new onmetal virtual ip")
	if err := s.createOnmetalVirtualIP(ctx, log, c, newOnmetalVirtualIP); err != nil {
		return nil, fmt.Errorf("error creating new onmetal virtual ip: %w", err)
	}

	log.V(1).Info("Patching new onmetal virtual ip as controlled by network interface")
	if err := apiutils.PatchControlledBy(ctx, s.cluster.Client(), onmetalNetworkInterface.NetworkInterface, newOnmetalVirtualIP); err != nil {
		return nil, fmt.Errorf("error patching new onmetal virtual ip as controlled by network interface: %w", err)
	}

	log.V(1).Info("Setting new onmetal virtual ip as source")
	if err := s.setOnmetalNetworkInterfaceVirtualIPSource(ctx, onmetalNetworkInterface.NetworkInterface,
		&networkingv1alpha1.VirtualIPSource{VirtualIPRef: &corev1.LocalObjectReference{Name: newOnmetalVirtualIP.Name}},
	); err != nil {
		return nil, fmt.Errorf("error setting new onmetal virtual ip as source: %w", err)
	}

	return &ori.CreateNetworkInterfaceVirtualIPResponse{}, nil
}
