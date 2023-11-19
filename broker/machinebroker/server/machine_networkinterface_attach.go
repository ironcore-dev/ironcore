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

	"github.com/go-logr/logr"
	commonv1alpha1 "github.com/ironcore-dev/ironcore/api/common/v1alpha1"
	computev1alpha1 "github.com/ironcore-dev/ironcore/api/compute/v1alpha1"
	networkingv1alpha1 "github.com/ironcore-dev/ironcore/api/networking/v1alpha1"
	"github.com/ironcore-dev/ironcore/broker/common/cleaner"
	machinebrokerv1alpha1 "github.com/ironcore-dev/ironcore/broker/machinebroker/api/v1alpha1"
	"github.com/ironcore-dev/ironcore/broker/machinebroker/apiutils"
	iri "github.com/ironcore-dev/ironcore/iri/apis/machine/v1alpha1"
	"google.golang.org/grpc/codes"
	grpcstatus "google.golang.org/grpc/status"
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"sigs.k8s.io/controller-runtime/pkg/client"
)

type IronCoreNetworkInterfaceConfig struct {
	Name       string
	NetworkID  string
	IPs        []commonv1alpha1.IP
	Attributes map[string]string
}

func (s *Server) getIronCoreNetworkInterfaceConfig(nic *iri.NetworkInterface) (*IronCoreNetworkInterfaceConfig, error) {
	ips, err := s.parseIPs(nic.Ips)
	if err != nil {
		return nil, err
	}

	return &IronCoreNetworkInterfaceConfig{
		Name:       nic.Name,
		NetworkID:  nic.NetworkId,
		IPs:        ips,
		Attributes: nic.Attributes,
	}, nil
}

func (s *Server) createIronCoreNetworkInterface(
	ctx context.Context,
	log logr.Logger,
	c *cleaner.Cleaner,
	optIronCoreMachine client.Object,
	cfg *IronCoreNetworkInterfaceConfig,
) (ironcoreMachineNic *computev1alpha1.NetworkInterface, aggIronCoreNic *AggregateIronCoreNetworkInterface, retErr error) {
	log.V(1).Info("Getting network for handle")
	ironcoreNetwork, err := s.networks.GetNetwork(ctx, cfg.NetworkID)
	if err != nil {
		return nil, nil, fmt.Errorf("error getting network: %w", err)
	}

	ironcoreNic := &networkingv1alpha1.NetworkInterface{
		ObjectMeta: metav1.ObjectMeta{
			Namespace: s.cluster.Namespace(),
			Name:      s.cluster.IDGen().Generate(),
			Annotations: map[string]string{
				commonv1alpha1.ManagedByAnnotation: machinebrokerv1alpha1.MachineBrokerManager,
			},
			Labels: map[string]string{
				machinebrokerv1alpha1.ManagerLabel: machinebrokerv1alpha1.MachineBrokerManager,
			},
			OwnerReferences: s.optionalOwnerReferences(ironcoreMachineGVK, optIronCoreMachine),
		},
		Spec: networkingv1alpha1.NetworkInterfaceSpec{
			NetworkRef: corev1.LocalObjectReference{Name: ironcoreNetwork.Name},
			MachineRef: s.optionalLocalUIDReference(optIronCoreMachine),
			IPFamilies: s.getIronCoreIPsIPFamilies(cfg.IPs),
			IPs:        s.ironcoreIPsToIronCoreIPSources(cfg.IPs),
			Attributes: cfg.Attributes,
		},
	}
	log.V(1).Info("Creating ironcore network interface")
	if err := s.cluster.Client().Create(ctx, ironcoreNic); err != nil {
		return nil, nil, fmt.Errorf("error creating ironcore network interface: %w", err)
	}
	c.Add(cleaner.CleanupObject(s.cluster.Client(), ironcoreNic))

	log.V(1).Info("Patching ironcore network to be owned by ironcore network interface")
	if err := apiutils.PatchOwnedBy(ctx, s.cluster.Client(), ironcoreNic, ironcoreNetwork); err != nil {
		return nil, nil, fmt.Errorf("error patching network to be owned by network interface: %w", err)
	}

	return &computev1alpha1.NetworkInterface{
			Name: cfg.Name,
			NetworkInterfaceSource: computev1alpha1.NetworkInterfaceSource{
				NetworkInterfaceRef: &corev1.LocalObjectReference{Name: ironcoreNic.Name},
			},
		}, &AggregateIronCoreNetworkInterface{
			Network:          ironcoreNetwork,
			NetworkInterface: ironcoreNic,
		}, nil
}

func (s *Server) attachIronCoreNetworkInterface(
	ctx context.Context,
	log logr.Logger,
	ironcoreMachine *computev1alpha1.Machine,
	ironcoreMachineNic *computev1alpha1.NetworkInterface,
) error {
	baseMachine := ironcoreMachine.DeepCopy()
	ironcoreMachine.Spec.NetworkInterfaces = append(ironcoreMachine.Spec.NetworkInterfaces, *ironcoreMachineNic)
	if err := s.cluster.Client().Patch(ctx, ironcoreMachine, client.StrategicMergeFrom(baseMachine)); err != nil {
		return fmt.Errorf("error patching ironcore machine network interfaces: %w", err)
	}
	return nil
}

func (s *Server) AttachNetworkInterface(ctx context.Context, req *iri.AttachNetworkInterfaceRequest) (res *iri.AttachNetworkInterfaceResponse, retErr error) {
	machineID := req.MachineId
	networkInterfaceName := req.NetworkInterface.Name
	log := s.loggerFrom(ctx, "MachineID", machineID, "NetworkInterfaceName", networkInterfaceName)

	log.V(1).Info("Getting aggregate ironcore machine")
	ironcoreMachine, err := s.getIronCoreMachine(ctx, req.MachineId)
	if err != nil {
		return nil, err
	}

	idx := ironcoreMachineNetworkInterfaceIndex(ironcoreMachine, networkInterfaceName)
	if idx >= 0 {
		return nil, grpcstatus.Errorf(codes.AlreadyExists, "machine %s network interface %s already exists", req.MachineId, req.NetworkInterface.Name)
	}

	cfg, err := s.getIronCoreNetworkInterfaceConfig(req.NetworkInterface)
	if err != nil {
		return nil, fmt.Errorf("error getting ironcore network interface config: %w", err)
	}

	c, cleanup := s.setupCleaner(ctx, log, &retErr)
	defer cleanup()

	ironcoreMachineNic, _, err := s.createIronCoreNetworkInterface(ctx, log, c, ironcoreMachine, cfg)
	if err != nil {
		return nil, err
	}

	if err := s.attachIronCoreNetworkInterface(ctx, log, ironcoreMachine, ironcoreMachineNic); err != nil {
		return nil, fmt.Errorf("error creating ironcore network interface: %w", err)
	}

	return &iri.AttachNetworkInterfaceResponse{}, nil
}
