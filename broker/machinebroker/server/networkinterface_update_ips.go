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

	corev1alpha1 "github.com/onmetal/onmetal-api/api/core/v1alpha1"
	networkingv1alpha1 "github.com/onmetal/onmetal-api/api/networking/v1alpha1"
	ori "github.com/onmetal/onmetal-api/ori/apis/machine/v1alpha1"
	"sigs.k8s.io/controller-runtime/pkg/client"
)

func (s *Server) setOnmetalNetworkInterfaceIPs(
	ctx context.Context,
	onmetalNetworkInterface *networkingv1alpha1.NetworkInterface,
	ips []corev1alpha1.IP,
) error {
	baseOnmetalNetworkInterface := onmetalNetworkInterface.DeepCopy()
	onmetalNetworkInterface.Spec.IPs = s.onmetalIPsToOnmetalIPSources(ips)
	if err := s.cluster.Client().Patch(ctx, onmetalNetworkInterface, client.MergeFrom(baseOnmetalNetworkInterface)); err != nil {
		return fmt.Errorf("error setting ips: %w", err)
	}
	return nil
}

func (s *Server) UpdateNetworkInterfaceIPs(ctx context.Context, req *ori.UpdateNetworkInterfaceIPsRequest) (*ori.UpdateNetworkInterfaceIPsResponse, error) {
	networkInterfaceID := req.NetworkInterfaceId
	log := s.loggerFrom(ctx, "NetworkInterfaceID", networkInterfaceID)

	ips, err := s.parseIPs(req.Ips)
	if err != nil {
		return nil, fmt.Errorf("error parsing ips: %w", err)
	}

	log.V(1).Info("Getting onmetal network interface")
	onmetalNetworkInterface, err := s.getAggregateOnmetalNetworkInterface(ctx, networkInterfaceID)
	if err != nil {
		return nil, err
	}

	if err := s.setOnmetalNetworkInterfaceIPs(ctx, onmetalNetworkInterface.NetworkInterface, ips); err != nil {
		return nil, fmt.Errorf("error setting onmetal network interface ips: %w", err)
	}

	return &ori.UpdateNetworkInterfaceIPsResponse{}, nil
}
