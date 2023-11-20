// Copyright 2022 IronCore authors
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

	networkingv1alpha1 "github.com/ironcore-dev/ironcore/api/networking/v1alpha1"
	iri "github.com/ironcore-dev/ironcore/iri/apis/machine/v1alpha1"
	"golang.org/x/exp/slices"
	"google.golang.org/grpc/codes"
	grpcstatus "google.golang.org/grpc/status"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"sigs.k8s.io/controller-runtime/pkg/client"
)

func (s *Server) DetachNetworkInterface(
	ctx context.Context,
	req *iri.DetachNetworkInterfaceRequest,
) (*iri.DetachNetworkInterfaceResponse, error) {
	machineID := req.MachineId
	nicName := req.Name
	log := s.loggerFrom(ctx, "MachineID", machineID, "NetworkInterfaceName", nicName)

	log.V(1).Info("Getting ironcore machine")
	ironcoreMachine, err := s.getIronCoreMachine(ctx, machineID)
	if err != nil {
		return nil, err
	}

	idx := ironcoreMachineNetworkInterfaceIndex(ironcoreMachine, nicName)
	if idx < 0 {
		return nil, grpcstatus.Errorf(codes.NotFound, "machine %s network interface %s not found", machineID, nicName)
	}

	ironcoreMachineNic := ironcoreMachine.Spec.NetworkInterfaces[idx]

	log.V(1).Info("Patching ironcore machine network interfaces")
	baseIronCoreMachine := ironcoreMachine.DeepCopy()
	ironcoreMachine.Spec.NetworkInterfaces = slices.Delete(ironcoreMachine.Spec.NetworkInterfaces, idx, idx+1)
	if err := s.cluster.Client().Patch(ctx, ironcoreMachine, client.StrategicMergeFrom(baseIronCoreMachine)); err != nil {
		return nil, fmt.Errorf("error patching ironcore machine network interfaces: %w", err)
	}

	ironcoreMachineNicRef := ironcoreMachineNic.NetworkInterfaceRef
	if ironcoreMachineNicRef == nil {
		return nil, fmt.Errorf("ironcore machine %s network interface %s does not have a reference", machineID, nicName)
	}
	ironcoreNicName := ironcoreMachineNicRef.Name
	log = log.WithValues("IronCoreNetworkInterfaceName", ironcoreNicName)

	log.V(1).Info("Deleting ironcore network interface")
	ironcoreNic := &networkingv1alpha1.NetworkInterface{
		ObjectMeta: metav1.ObjectMeta{
			Namespace: s.cluster.Namespace(),
			Name:      ironcoreNicName,
		},
	}
	if err := s.cluster.Client().Delete(ctx, ironcoreNic); client.IgnoreNotFound(err) != nil {
		return nil, fmt.Errorf("error deleting ironcore network interface %s: %w", ironcoreNicName, err)
	}

	log.V(1).Info("Detached ironcore network interface")
	return &iri.DetachNetworkInterfaceResponse{}, nil
}
