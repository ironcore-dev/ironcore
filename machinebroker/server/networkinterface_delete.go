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

	computev1alpha1 "github.com/onmetal/onmetal-api/apis/compute/v1alpha1"
	networkingv1alpha1 "github.com/onmetal/onmetal-api/apis/networking/v1alpha1"
	machinebrokerv1alpha1 "github.com/onmetal/onmetal-api/machinebroker/api/v1alpha1"
	ori "github.com/onmetal/onmetal-api/ori/apis/runtime/v1alpha1"
	"golang.org/x/exp/slices"
	"sigs.k8s.io/controller-runtime/pkg/client"
)

func (s *Server) DeleteNetworkInterface(ctx context.Context, req *ori.DeleteNetworkInterfaceRequest) (*ori.DeleteNetworkInterfaceResponse, error) {
	log := s.loggerFrom(ctx)

	log.V(1).Info("Getting machine")
	onmetalMachine, err := s.getOnmetalMachine(ctx, req.MachineId)
	if err != nil {
		return nil, err
	}

	idx := slices.IndexFunc(onmetalMachine.Spec.NetworkInterfaces,
		func(networkInterface computev1alpha1.NetworkInterface) bool {
			return networkInterface.Name == req.NetworkInterfaceName
		},
	)
	if idx < 0 {
		log.V(1).Info("Network interface not present in machine")
		return &ori.DeleteNetworkInterfaceResponse{}, nil
	}

	log.V(1).Info("Deleting network interface from machine")
	baseOnmetalMachine := onmetalMachine.DeepCopy()
	onmetalMachine.Spec.NetworkInterfaces = slices.Delete(onmetalMachine.Spec.NetworkInterfaces, idx, idx+1)
	if err := s.client.Patch(ctx, onmetalMachine, client.MergeFrom(baseOnmetalMachine)); err != nil {
		return nil, fmt.Errorf("error deleting network interface from machine: %w", err)
	}

	log.V(1).Info("Deleting network interface")
	if err := s.client.DeleteAllOf(ctx, &networkingv1alpha1.NetworkInterface{},
		client.InNamespace(s.namespace),
		client.MatchingLabels{
			machinebrokerv1alpha1.MachineIDLabel:            req.MachineId,
			machinebrokerv1alpha1.NetworkInterfaceNameLabel: req.NetworkInterfaceName,
		},
	); err != nil {
		return nil, fmt.Errorf("error deleting network interface: %w", err)
	}

	log.V(1).Info("Deleting network")
	if err := s.client.DeleteAllOf(ctx, &networkingv1alpha1.Network{},
		client.InNamespace(s.namespace),
		client.MatchingLabels{
			machinebrokerv1alpha1.MachineIDLabel:            req.MachineId,
			machinebrokerv1alpha1.NetworkInterfaceNameLabel: req.NetworkInterfaceName,
		},
	); err != nil {
		return nil, fmt.Errorf("error deleting network interface: %w", err)
	}

	log.V(1).Info("Deleting virtual ip")
	if err := s.client.DeleteAllOf(ctx, &networkingv1alpha1.VirtualIP{},
		client.InNamespace(s.namespace),
		client.MatchingLabels{
			machinebrokerv1alpha1.MachineIDLabel:            req.MachineId,
			machinebrokerv1alpha1.NetworkInterfaceNameLabel: req.NetworkInterfaceName,
		},
	); err != nil {
		return nil, fmt.Errorf("error deleting virtual ip: %w", err)
	}

	return &ori.DeleteNetworkInterfaceResponse{}, nil
}
