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
	machinebrokerv1alpha1 "github.com/onmetal/onmetal-api/machinebroker/api/v1alpha1"
	ori "github.com/onmetal/onmetal-api/ori/apis/machine/v1alpha1"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
	"sigs.k8s.io/controller-runtime/pkg/client"
)

type networkInterfaceFilter struct {
	machineID string
	name      string
}

func (s *Server) listOnmetalNetworkInterfaces(ctx context.Context, filter networkInterfaceFilter) ([]networkingv1alpha1.NetworkInterface, error) {
	opts := []client.ListOption{
		client.InNamespace(s.namespace),
	}

	if filter.machineID != "" || filter.name != "" {
		labels := make(map[string]string)
		if filter.machineID != "" {
			labels[machinebrokerv1alpha1.MachineIDLabel] = filter.machineID
		}
		if filter.name != "" {
			labels[machinebrokerv1alpha1.NetworkInterfaceNameLabel] = filter.name
		}

		opts = append(opts, client.MatchingLabels(labels))
	}

	onmetalNetworkingNetworkInterfaceList := &networkingv1alpha1.NetworkInterfaceList{}
	if err := s.client.List(ctx, onmetalNetworkingNetworkInterfaceList, opts...); err != nil {
		return nil, fmt.Errorf("error listing onmetal networking network interfaces: %w", err)
	}

	return onmetalNetworkingNetworkInterfaceList.Items, nil
}

func (s *Server) listNetworkInterfaces(ctx context.Context, filter networkInterfaceFilter) ([]*ori.NetworkInterface, error) {
	onmetalNetworkInterfaces, err := s.listOnmetalNetworkInterfaces(ctx, filter)
	if err != nil {
		return nil, err
	}

	res := make([]*ori.NetworkInterface, len(onmetalNetworkInterfaces))
	for i, onmetalNetworkInterface := range onmetalNetworkInterfaces {
		networkInterface, err := s.convertOnmetalNetworkInterface(&onmetalNetworkInterface)
		if err != nil {
			return nil, err
		}

		res[i] = networkInterface
	}
	return res, nil
}

func (s *Server) ListNetworkInterfaces(ctx context.Context, req *ori.ListNetworkInterfacesRequest) (*ori.ListNetworkInterfacesResponse, error) {
	if filter := req.Filter; filter != nil && filter.MachineId != "" && filter.Name != "" {
		networkInterface, err := s.getNetworkInterface(ctx, filter.MachineId, filter.Name)
		if err != nil {
			if status.Code(err) != codes.NotFound {
				return nil, err
			}
			return &ori.ListNetworkInterfacesResponse{
				NetworkInterfaces: []*ori.NetworkInterface{},
			}, nil
		}
		return &ori.ListNetworkInterfacesResponse{
			NetworkInterfaces: []*ori.NetworkInterface{networkInterface},
		}, nil
	}

	networkInterfaces, err := s.listNetworkInterfaces(ctx, networkInterfaceFilter{
		machineID: req.GetFilter().GetMachineId(),
		name:      req.GetFilter().GetName(),
	})
	if err != nil {
		return nil, err
	}
	return &ori.ListNetworkInterfacesResponse{
		NetworkInterfaces: networkInterfaces,
	}, nil
}
