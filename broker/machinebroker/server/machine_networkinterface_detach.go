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
	ori "github.com/onmetal/onmetal-api/ori/apis/machine/v1alpha1"
	"golang.org/x/exp/slices"
	"google.golang.org/grpc/codes"
	grpcstatus "google.golang.org/grpc/status"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"sigs.k8s.io/controller-runtime/pkg/client"
)

func (s *Server) DetachNetworkInterface(
	ctx context.Context,
	req *ori.DetachNetworkInterfaceRequest,
) (*ori.DetachNetworkInterfaceResponse, error) {
	machineID := req.MachineId
	nicName := req.Name
	log := s.loggerFrom(ctx, "MachineID", machineID, "NetworkInterfaceName", nicName)

	log.V(1).Info("Getting onmetal machine")
	onmetalMachine, err := s.getOnmetalMachine(ctx, machineID)
	if err != nil {
		return nil, err
	}

	idx := onmetalMachineNetworkInterfaceIndex(onmetalMachine, nicName)
	if idx < 0 {
		return nil, grpcstatus.Errorf(codes.NotFound, "machine %s network interface %s not found", machineID, nicName)
	}

	onmetalMachineNic := onmetalMachine.Spec.NetworkInterfaces[idx]

	log.V(1).Info("Patching onmetal machine network interfaces")
	baseOnmetalMachine := onmetalMachine.DeepCopy()
	onmetalMachine.Spec.NetworkInterfaces = slices.Delete(onmetalMachine.Spec.NetworkInterfaces, idx, idx+1)
	if err := s.cluster.Client().Patch(ctx, onmetalMachine, client.StrategicMergeFrom(baseOnmetalMachine)); err != nil {
		return nil, fmt.Errorf("error patching onmetal machine network interfaces: %w", err)
	}

	onmetalMachineNicRef := onmetalMachineNic.NetworkInterfaceRef
	if onmetalMachineNicRef == nil {
		return nil, fmt.Errorf("onmetal machine %s network interface %s does not have a reference", machineID, nicName)
	}
	onmetalNicName := onmetalMachineNicRef.Name
	log = log.WithValues("OnmetalNetworkInterfaceName", onmetalNicName)

	log.V(1).Info("Deleting onmetal network interface")
	onmetalNic := &networkingv1alpha1.NetworkInterface{
		ObjectMeta: metav1.ObjectMeta{
			Namespace: s.cluster.Namespace(),
			Name:      onmetalNicName,
		},
	}
	if err := s.cluster.Client().Delete(ctx, onmetalNic); client.IgnoreNotFound(err) != nil {
		return nil, fmt.Errorf("error deleting onmetal network interface %s: %w", onmetalNicName, err)
	}

	log.V(1).Info("Detached onmetal network interface")
	return &ori.DetachNetworkInterfaceResponse{}, nil
}
