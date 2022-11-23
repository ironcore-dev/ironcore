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
	networkingv1alpha1 "github.com/onmetal/onmetal-api/api/networking/v1alpha1"
	machinebrokerv1alpha1 "github.com/onmetal/onmetal-api/machinebroker/api/v1alpha1"
	"github.com/onmetal/onmetal-api/machinebroker/apiutils"
	ori "github.com/onmetal/onmetal-api/ori/apis/machine/v1alpha1"
	"golang.org/x/exp/slices"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
	apierrors "k8s.io/apimachinery/pkg/api/errors"
	"sigs.k8s.io/controller-runtime/pkg/client"
)

func (s *Server) getOnmetalNetworkInterface(ctx context.Context, onmetalMachine *computev1alpha1.Machine, name string) (*networkingv1alpha1.NetworkInterface, error) {
	onmetalNetworkInterface := &networkingv1alpha1.NetworkInterface{}
	onmetalNetworkInterfaceKey := client.ObjectKey{Namespace: s.namespace, Name: s.onmetalNetworkInterfaceName(onmetalMachine.Name, name)}
	if err := s.client.Get(ctx, onmetalNetworkInterfaceKey, onmetalNetworkInterface); err != nil {
		if !apierrors.IsNotFound(err) {
			return nil, fmt.Errorf("error getting machine %s network interface %s: %w", onmetalMachine.Name, name, err)
		}
		return nil, status.Errorf(codes.NotFound, "machine %s network interface %s not found", onmetalMachine.Name, name)
	}
	return onmetalNetworkInterface, nil
}

func (s *Server) convertOnmetalNetworkInterface(
	machine *computev1alpha1.Machine,
	networkInterface *networkingv1alpha1.NetworkInterface,
) (*ori.NetworkInterface, error) {
	machineID := networkInterface.Labels[machinebrokerv1alpha1.MachineIDLabel]

	metadata, err := apiutils.GetMetadataAnnotation(networkInterface)
	if err != nil {
		return nil, err
	}

	name := networkInterface.Labels[machinebrokerv1alpha1.NetworkInterfaceNameLabel]

	ips := make([]string, len(networkInterface.Status.IPs))
	for i, ip := range networkInterface.Status.IPs {
		ips[i] = ip.String()
	}

	var virtualIPConfig *ori.VirtualIPConfig
	if onmetalVirtualIP := networkInterface.Status.VirtualIP; onmetalVirtualIP != nil {
		virtualIPConfig = &ori.VirtualIPConfig{
			Ip: onmetalVirtualIP.String(),
		}
	}

	idx := slices.IndexFunc(machine.Status.NetworkInterfaces,
		func(networkInterface computev1alpha1.NetworkInterfaceStatus) bool {
			return networkInterface.Name == name
		},
	)
	state := ori.NetworkInterfaceState_NETWORK_INTERFACE_DETACHED
	if idx != -1 {
		state = s.convertOnmetalNetworkInterfaceState(machine.Status.NetworkInterfaces[idx].State)
	}

	return &ori.NetworkInterface{
		MachineId:       machineID,
		MachineMetadata: metadata,
		Name:            name,
		Network:         &ori.NetworkConfig{Handle: networkInterface.Status.NetworkHandle},
		Ips:             ips,
		VirtualIp:       virtualIPConfig,
		State:           state,
	}, nil
}

func (s *Server) getNetworkInterface(
	ctx context.Context,
	machineID, name string,
) (*ori.NetworkInterface, error) {
	onmetalMachine, err := s.getOnmetalMachine(ctx, machineID)
	if err != nil {
		return nil, err
	}

	onmetalNetworkInterface, err := s.getOnmetalNetworkInterface(ctx, onmetalMachine, name)
	if err != nil {
		return nil, err
	}

	return s.convertOnmetalNetworkInterface(onmetalMachine, onmetalNetworkInterface)
}
