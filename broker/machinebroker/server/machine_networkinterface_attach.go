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
	commonv1alpha1 "github.com/onmetal/onmetal-api/api/common/v1alpha1"
	computev1alpha1 "github.com/onmetal/onmetal-api/api/compute/v1alpha1"
	networkingv1alpha1 "github.com/onmetal/onmetal-api/api/networking/v1alpha1"
	"github.com/onmetal/onmetal-api/broker/common/cleaner"
	machinebrokerv1alpha1 "github.com/onmetal/onmetal-api/broker/machinebroker/api/v1alpha1"
	"github.com/onmetal/onmetal-api/broker/machinebroker/apiutils"
	ori "github.com/onmetal/onmetal-api/ori/apis/machine/v1alpha1"
	"google.golang.org/grpc/codes"
	grpcstatus "google.golang.org/grpc/status"
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"sigs.k8s.io/controller-runtime/pkg/client"
)

type OnmetalNetworkInterfaceConfig struct {
	Name       string
	NetworkID  string
	IPs        []commonv1alpha1.IP
	Attributes map[string]string
}

func (s *Server) getOnmetalNetworkInterfaceConfig(nic *ori.NetworkInterface) (*OnmetalNetworkInterfaceConfig, error) {
	ips, err := s.parseIPs(nic.Ips)
	if err != nil {
		return nil, err
	}

	return &OnmetalNetworkInterfaceConfig{
		Name:       nic.Name,
		NetworkID:  nic.NetworkId,
		IPs:        ips,
		Attributes: nic.Attributes,
	}, nil
}

func (s *Server) createOnmetalNetworkInterface(
	ctx context.Context,
	log logr.Logger,
	c *cleaner.Cleaner,
	optOnmetalMachine client.Object,
	cfg *OnmetalNetworkInterfaceConfig,
) (onmetalMachineNic *computev1alpha1.NetworkInterface, aggOnmetalNic *AggregateOnmetalNetworkInterface, retErr error) {
	log.V(1).Info("Getting network for handle")
	onmetalNetwork, err := s.networks.GetNetwork(ctx, cfg.NetworkID)
	if err != nil {
		return nil, nil, fmt.Errorf("error getting network: %w", err)
	}

	onmetalNic := &networkingv1alpha1.NetworkInterface{
		ObjectMeta: metav1.ObjectMeta{
			Namespace: s.cluster.Namespace(),
			Name:      s.cluster.IDGen().Generate(),
			Annotations: map[string]string{
				commonv1alpha1.ManagedByAnnotation: machinebrokerv1alpha1.MachineBrokerManager,
			},
			Labels: map[string]string{
				machinebrokerv1alpha1.ManagerLabel: machinebrokerv1alpha1.MachineBrokerManager,
			},
			OwnerReferences: s.optionalOwnerReferences(onmetalMachineGVK, optOnmetalMachine),
		},
		Spec: networkingv1alpha1.NetworkInterfaceSpec{
			NetworkRef: corev1.LocalObjectReference{Name: onmetalNetwork.Name},
			MachineRef: s.optionalLocalUIDReference(optOnmetalMachine),
			IPFamilies: s.getOnmetalIPsIPFamilies(cfg.IPs),
			IPs:        s.onmetalIPsToOnmetalIPSources(cfg.IPs),
			Attributes: cfg.Attributes,
		},
	}
	log.V(1).Info("Creating onmetal network interface")
	if err := s.cluster.Client().Create(ctx, onmetalNic); err != nil {
		return nil, nil, fmt.Errorf("error creating onmetal network interface: %w", err)
	}
	c.Add(cleaner.CleanupObject(s.cluster.Client(), onmetalNic))

	log.V(1).Info("Patching onmetal network to be owned by onmetal network interface")
	if err := apiutils.PatchOwnedBy(ctx, s.cluster.Client(), onmetalNic, onmetalNetwork); err != nil {
		return nil, nil, fmt.Errorf("error patching network to be owned by network interface: %w", err)
	}

	return &computev1alpha1.NetworkInterface{
			Name: cfg.Name,
			NetworkInterfaceSource: computev1alpha1.NetworkInterfaceSource{
				NetworkInterfaceRef: &corev1.LocalObjectReference{Name: onmetalNic.Name},
			},
		}, &AggregateOnmetalNetworkInterface{
			Network:          onmetalNetwork,
			NetworkInterface: onmetalNic,
		}, nil
}

func (s *Server) attachOnmetalNetworkInterface(
	ctx context.Context,
	log logr.Logger,
	onmetalMachine *computev1alpha1.Machine,
	onmetalMachineNic *computev1alpha1.NetworkInterface,
) error {
	baseMachine := onmetalMachine.DeepCopy()
	onmetalMachine.Spec.NetworkInterfaces = append(onmetalMachine.Spec.NetworkInterfaces, *onmetalMachineNic)
	if err := s.cluster.Client().Patch(ctx, onmetalMachine, client.StrategicMergeFrom(baseMachine)); err != nil {
		return fmt.Errorf("error patching onmetal machine network interfaces: %w", err)
	}
	return nil
}

func (s *Server) AttachNetworkInterface(ctx context.Context, req *ori.AttachNetworkInterfaceRequest) (res *ori.AttachNetworkInterfaceResponse, retErr error) {
	machineID := req.MachineId
	networkInterfaceName := req.NetworkInterface.Name
	log := s.loggerFrom(ctx, "MachineID", machineID, "NetworkInterfaceName", networkInterfaceName)

	log.V(1).Info("Getting aggregate onmetal machine")
	onmetalMachine, err := s.getOnmetalMachine(ctx, req.MachineId)
	if err != nil {
		return nil, err
	}

	idx := onmetalMachineNetworkInterfaceIndex(onmetalMachine, networkInterfaceName)
	if idx >= 0 {
		return nil, grpcstatus.Errorf(codes.AlreadyExists, "machine %s network interface %s already exists", req.MachineId, req.NetworkInterface.Name)
	}

	cfg, err := s.getOnmetalNetworkInterfaceConfig(req.NetworkInterface)
	if err != nil {
		return nil, fmt.Errorf("error getting onmetal network interface config: %w", err)
	}

	c, cleanup := s.setupCleaner(ctx, log, &retErr)
	defer cleanup()

	onmetalMachineNic, _, err := s.createOnmetalNetworkInterface(ctx, log, c, onmetalMachine, cfg)
	if err != nil {
		return nil, err
	}

	if err := s.attachOnmetalNetworkInterface(ctx, log, onmetalMachine, onmetalMachineNic); err != nil {
		return nil, fmt.Errorf("error creating onmetal network interface: %w", err)
	}

	return &ori.AttachNetworkInterfaceResponse{}, nil
}
