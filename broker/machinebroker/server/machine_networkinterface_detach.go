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

	computev1alpha1 "github.com/onmetal/onmetal-api/api/compute/v1alpha1"
	ori "github.com/onmetal/onmetal-api/ori/apis/machine/v1alpha1"
	"golang.org/x/exp/slices"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
	"sigs.k8s.io/controller-runtime/pkg/client"
)

func (s *Server) deleteOnmetalNetworkInterfaceAttachment(ctx context.Context, onmetalMachine *computev1alpha1.Machine, networkInterfaceName string) error {
	baseOnmetalMachine := onmetalMachine.DeepCopy()
	idx := slices.IndexFunc(
		onmetalMachine.Spec.NetworkInterfaces,
		func(networkInterface computev1alpha1.NetworkInterface) bool {
			return networkInterface.Name == networkInterfaceName
		},
	)
	onmetalMachine.Spec.NetworkInterfaces = slices.Delete(onmetalMachine.Spec.NetworkInterfaces, idx, idx+1)
	if err := s.cluster.Client().Patch(ctx, onmetalMachine, client.StrategicMergeFrom(baseOnmetalMachine)); err != nil {
		return fmt.Errorf("error patching onmetal machine network interfaces: %w", err)
	}
	return nil
}

func (s *Server) DeleteNetworkInterfaceAttachment(ctx context.Context, req *ori.DeleteNetworkInterfaceAttachmentRequest) (*ori.DeleteNetworkInterfaceAttachmentResponse, error) {
	machineID := req.MachineId
	networkInterfaceName := req.Name
	log := s.loggerFrom(ctx, "MachineID", machineID, "NetworkInterfaceName", networkInterfaceName)

	log.V(1).Info("Getting aggregated onmetal machine")
	aggOnmetalMachine, err := s.getAggregateOnmetalMachine(ctx, machineID)
	if err != nil {
		return nil, err
	}

	machine, err := s.convertAggregateOnmetalMachine(aggOnmetalMachine)
	if err != nil {
		return nil, err
	}

	idx := slices.IndexFunc(
		machine.Spec.NetworkInterfaces,
		func(networkInterface *ori.NetworkInterfaceAttachment) bool {
			return networkInterface.Name == networkInterfaceName
		},
	)
	if idx < 0 {
		return nil, status.Errorf(codes.NotFound, "machine %s does not attach network interface %s", machineID, networkInterfaceName)
	}

	log.V(1).Info("Deleting onmetal machine network interface")
	if err := s.deleteOnmetalNetworkInterfaceAttachment(ctx, aggOnmetalMachine.Machine, networkInterfaceName); err != nil {
		return nil, fmt.Errorf("error adding onmetal machine network interface: %w", err)
	}

	return &ori.DeleteNetworkInterfaceAttachmentResponse{}, nil
}
