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
	machinebrokerv1alpha1 "github.com/onmetal/onmetal-api/broker/machinebroker/api/v1alpha1"
	"github.com/onmetal/onmetal-api/broker/machinebroker/apiutils"
	ori "github.com/onmetal/onmetal-api/ori/apis/machine/v1alpha1"
	"sigs.k8s.io/controller-runtime/pkg/client"
)

type AggregateOnmetalNetworkInterface struct {
	NetworkInterface    *networkingv1alpha1.NetworkInterface
	Network             *networkingv1alpha1.Network
	VirtualIP           *networkingv1alpha1.VirtualIP
	Prefixes            []corev1alpha1.IPPrefix
	LoadBalancerTargets []machinebrokerv1alpha1.LoadBalancerTarget
	NATGatewayTargets   []machinebrokerv1alpha1.NATGatewayTarget
}

func (s *Server) convertAggregateOnmetalNetworkInterface(networkInterface *AggregateOnmetalNetworkInterface) (*ori.NetworkInterface, error) {
	metadata, err := apiutils.GetObjectMetadata(networkInterface.NetworkInterface)
	if err != nil {
		return nil, err
	}

	ips, err := s.convertOnmetalIPSourcesToIPs(networkInterface.NetworkInterface.Spec.IPs)
	if err != nil {
		return nil, err
	}

	var virtualIPSpec *ori.VirtualIPSpec
	if networkInterface.VirtualIP != nil {
		virtualIPSpec = &ori.VirtualIPSpec{
			Ip: networkInterface.VirtualIP.Status.IP.String(),
		}
	}

	loadBalancerTargets, err := s.convertOnmetalLoadBalancerTargets(networkInterface.LoadBalancerTargets)
	if err != nil {
		return nil, err
	}

	nats, err := s.convertOnmetalNATGatewayTargets(networkInterface.NATGatewayTargets)
	if err != nil {
		return nil, err
	}

	return &ori.NetworkInterface{
		Metadata: metadata,
		Spec: &ori.NetworkInterfaceSpec{
			Network: &ori.NetworkSpec{
				Handle: networkInterface.Network.Spec.Handle,
			},
			Ips:                 ips,
			VirtualIp:           virtualIPSpec,
			Prefixes:            s.convertOnmetalPrefixes(networkInterface.Prefixes),
			LoadBalancerTargets: loadBalancerTargets,
			Nats:                nats,
		},
	}, nil
}

func (s *Server) setOnmetalNetworkInterfaceVirtualIPSource(
	ctx context.Context,
	onmetalNetworkInterface *networkingv1alpha1.NetworkInterface,
	virtualIPSrc *networkingv1alpha1.VirtualIPSource,
) error {
	baseOnmetalNetworkInterface := onmetalNetworkInterface.DeepCopy()
	onmetalNetworkInterface.Spec.VirtualIP = virtualIPSrc
	if err := s.cluster.Client().Patch(ctx, onmetalNetworkInterface, client.MergeFrom(baseOnmetalNetworkInterface)); err != nil {
		return fmt.Errorf("error setting virtual ip source: %w", err)
	}
	return nil
}
