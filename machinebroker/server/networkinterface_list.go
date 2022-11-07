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
	"errors"
	"fmt"

	computev1alpha1 "github.com/onmetal/onmetal-api/apis/compute/v1alpha1"
	networkingv1alpha1 "github.com/onmetal/onmetal-api/apis/networking/v1alpha1"
	machinebrokerv1alpha1 "github.com/onmetal/onmetal-api/machinebroker/api/v1alpha1"
	"github.com/onmetal/onmetal-api/machinebroker/apiutils"
	ori "github.com/onmetal/onmetal-api/ori/apis/runtime/v1alpha1"
	"github.com/onmetal/onmetal-api/utils/slices"
	"sigs.k8s.io/controller-runtime/pkg/client"
)

type onmetalNetworkingNetworkInterfaceFilter struct {
	machineID string
}

func (s *Server) listOnmetalNetworkingNetworkInterfaces(ctx context.Context, filter *onmetalNetworkingNetworkInterfaceFilter) ([]networkingv1alpha1.NetworkInterface, error) {
	labels := map[string]string{}
	if filter != nil {
		if filter.machineID != "" {
			labels[machinebrokerv1alpha1.MachineIDLabel] = filter.machineID
		}
	}

	onmetalNetworkingNetworkInterfaceList := &networkingv1alpha1.NetworkInterfaceList{}
	if err := s.client.List(ctx, onmetalNetworkingNetworkInterfaceList,
		client.InNamespace(s.namespace),
		client.MatchingLabels(labels),
	); err != nil {
		return nil, fmt.Errorf("error listing onmetal networking network interfaces: %w", err)
	}

	return onmetalNetworkingNetworkInterfaceList.Items, nil
}

func (s *Server) getMachineNetworkInterfaces(ctx context.Context, machineID string) ([]*ori.NetworkInterface, error) {
	onmetalMachine, err := s.getOnmetalMachine(ctx, machineID)
	if err != nil {
		return nil, err
	}

	onmetalNetworkingNetworkInterfaces, err := s.listOnmetalNetworkingNetworkInterfaces(ctx, &onmetalNetworkingNetworkInterfaceFilter{machineID: machineID})
	if err != nil {
		return nil, err
	}

	onmetalNetworkingNetworkInterfaceByName := slices.ToMap(onmetalNetworkingNetworkInterfaces, func(onmetalNetworkingNetworkInterface networkingv1alpha1.NetworkInterface) string {
		return onmetalNetworkingNetworkInterface.Name
	})

	return s.getOnmetalMachineNetworkInterfaces(ctx, onmetalMachine, onmetalNetworkingNetworkInterfaceByName)
}

func (s *Server) getOnmetalMachineNetworkInterfaces(
	_ context.Context,
	onmetalMachine *computev1alpha1.Machine,
	onmetalNetworkingNetworkInterfaceByName map[string]networkingv1alpha1.NetworkInterface,
) ([]*ori.NetworkInterface, error) {
	machineMetadata, err := apiutils.GetMetadataAnnotation(onmetalMachine)
	if err != nil {
		return nil, err
	}

	var networkInterfaces []*ori.NetworkInterface
	for _, onmetalNetworkInterface := range onmetalMachine.Spec.NetworkInterfaces {
		var onmetalNetworkingNetworkInterface *networkingv1alpha1.NetworkInterface
		if onmetalNetworkingNetworkInterfaceName := computev1alpha1.MachineNetworkInterfaceName(onmetalMachine.Name, onmetalNetworkInterface); onmetalNetworkingNetworkInterfaceName != "" {
			onmetalNetworkingNetworkInterface = &networkingv1alpha1.NetworkInterface{}
			var ok bool
			*onmetalNetworkingNetworkInterface, ok = onmetalNetworkingNetworkInterfaceByName[onmetalNetworkingNetworkInterfaceName]
			if !ok {
				return nil, fmt.Errorf("onmetal networking network interface %s not found", onmetalNetworkingNetworkInterfaceName)
			}
		}

		networkInterface, err := OnmetalNetworkInterfaceToNetworkInterface(onmetalMachine.Name, machineMetadata, &onmetalNetworkInterface, onmetalNetworkingNetworkInterface)
		if err != nil {
			return nil, err
		}

		networkInterfaces = append(networkInterfaces, networkInterface)
	}
	return networkInterfaces, nil
}

func (s *Server) listMachineNetworkInterfaces(ctx context.Context) ([]*ori.NetworkInterface, error) {
	onmetalMachines, err := s.listOnmetalMachines(ctx)
	if err != nil {
		return nil, err
	}

	onmetalNetworkingNetworkInterfaces, err := s.listOnmetalNetworkingNetworkInterfaces(ctx, nil)
	if err != nil {
		return nil, err
	}

	onmetalNetworkingNetworkInterfaceByName := slices.ToMap(onmetalNetworkingNetworkInterfaces, func(onmetalNetworkingNetworkInterface networkingv1alpha1.NetworkInterface) string {
		return onmetalNetworkingNetworkInterface.Name
	})

	var networkInterfaces []*ori.NetworkInterface
	for _, onmetalMachine := range onmetalMachines {
		machineNetworkInterfaces, err := s.getOnmetalMachineNetworkInterfaces(ctx, &onmetalMachine, onmetalNetworkingNetworkInterfaceByName)
		if err != nil {
			return nil, err
		}

		networkInterfaces = append(networkInterfaces, machineNetworkInterfaces...)
	}

	return networkInterfaces, nil
}

func (s *Server) ListNetworkInterfaces(ctx context.Context, req *ori.ListNetworkInterfacesRequest) (*ori.ListNetworkInterfacesResponse, error) {
	if filter := req.Filter; filter != nil && filter.MachineId != "" {
		networkInterfaces, err := s.getMachineNetworkInterfaces(ctx, filter.MachineId)
		if err != nil {
			if !errors.Is(err, ErrMachineNotFound) {
				return nil, err
			}
			return &ori.ListNetworkInterfacesResponse{
				NetworkInterfaces: []*ori.NetworkInterface{},
			}, nil
		}
		return &ori.ListNetworkInterfacesResponse{
			NetworkInterfaces: networkInterfaces,
		}, nil
	}

	networkInterfaces, err := s.listMachineNetworkInterfaces(ctx)
	if err != nil {
		return nil, err
	}
	return &ori.ListNetworkInterfacesResponse{
		NetworkInterfaces: networkInterfaces,
	}, nil
}
