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
	"github.com/onmetal/onmetal-api/broker/common/utils"
	machinebrokerv1alpha1 "github.com/onmetal/onmetal-api/broker/machinebroker/api/v1alpha1"
	ori "github.com/onmetal/onmetal-api/ori/apis/machine/v1alpha1"
	utilslices "github.com/onmetal/onmetal-api/utils/slices"
	"golang.org/x/exp/slices"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
	apierrors "k8s.io/apimachinery/pkg/api/errors"
	"k8s.io/apimachinery/pkg/labels"
)

func (s *Server) buildIDToLoadBalancersMap(loadBalancers []machinebrokerv1alpha1.LoadBalancer) map[string][]machinebrokerv1alpha1.LoadBalancer {
	res := make(map[string][]machinebrokerv1alpha1.LoadBalancer)
	for _, loadBalancer := range loadBalancers {
		for _, destination := range loadBalancer.Destinations {
			res[destination] = append(res[destination], loadBalancer)
		}
	}
	return res
}

func (s *Server) buildIDToNATGatewaysMap(natGateways []machinebrokerv1alpha1.NATGateway) map[string][]machinebrokerv1alpha1.NATGateway {
	res := make(map[string][]machinebrokerv1alpha1.NATGateway)
	for _, natGateway := range natGateways {
		for _, destination := range natGateway.Destinations {
			res[destination.ID] = append(res[destination.ID], natGateway)
		}
	}
	return res
}

func (s *Server) listAggregateOnmetalNetworkInterfaces(ctx context.Context) ([]AggregateOnmetalNetworkInterface, error) {
	log := s.loggerFrom(ctx)

	onmetalNetworkInterfaceList := &networkingv1alpha1.NetworkInterfaceList{}
	if err := s.listManagedAndCreated(ctx, onmetalNetworkInterfaceList); err != nil {
		return nil, fmt.Errorf("error listing onmetal network interfaces: %w", err)
	}

	onmetalNetworkList := &networkingv1alpha1.NetworkList{}
	if err := s.listManagedAndCreated(ctx, onmetalNetworkList); err != nil {
		return nil, fmt.Errorf("error listing onmetal networks: %w", err)
	}

	onmetalVirtualIPList := &networkingv1alpha1.VirtualIPList{}
	if err := s.listWithPurpose(ctx, onmetalVirtualIPList, machinebrokerv1alpha1.NetworkInterfacePurpose); err != nil {
		return nil, fmt.Errorf("error listing onmetal virtual ips: %w", err)
	}

	onmetalLoadBalancers, err := s.loadBalancers.List(ctx)
	if err != nil {
		return nil, fmt.Errorf("error listing onmetal load balancers: %w", err)
	}

	onmetalNATGateways, err := s.natGateways.List(ctx)
	if err != nil {
		return nil, fmt.Errorf("error listing onmetal nat gateways: %w", err)
	}

	idToLoadBalancers := s.buildIDToLoadBalancersMap(onmetalLoadBalancers)
	idToNATGateways := s.buildIDToNATGatewaysMap(onmetalNATGateways)
	getNetwork := utils.ObjectSliceToByNameGetter(networkingv1alpha1.Resource("networks"), onmetalNetworkList.Items)
	getVirtualIP := utils.ObjectSliceToByNameGetter(networkingv1alpha1.Resource("virtualips"), onmetalVirtualIPList.Items)

	var res []AggregateOnmetalNetworkInterface
	for i := range onmetalNetworkInterfaceList.Items {
		onmetalNetworkInterface := &onmetalNetworkInterfaceList.Items[i]
		networkInterface, err := s.aggregateOnmetalNetworkInterface(
			onmetalNetworkInterface,
			getNetwork,
			getVirtualIP,
			func() ([]machinebrokerv1alpha1.LoadBalancer, error) {
				return idToLoadBalancers[onmetalNetworkInterface.Name], nil
			},
			func() ([]machinebrokerv1alpha1.NATGateway, error) {
				return idToNATGateways[onmetalNetworkInterface.Name], nil
			},
		)
		if err != nil {
			log.Error(err, fmt.Sprintf("error assembling onmetal network interface %s", onmetalNetworkInterface.Name))
		}

		if networkInterface != nil {
			res = append(res, *networkInterface)
		}
	}
	return res, nil
}

func (s *Server) aggregateOnmetalNetworkInterface(
	onmetalNetworkInterface *networkingv1alpha1.NetworkInterface,
	getNetwork func(name string) (*networkingv1alpha1.Network, error),
	getVirtualIP func(name string) (*networkingv1alpha1.VirtualIP, error),
	listLoadBalancers func() ([]machinebrokerv1alpha1.LoadBalancer, error),
	listNATGateways func() ([]machinebrokerv1alpha1.NATGateway, error),
) (*AggregateOnmetalNetworkInterface, error) {
	network, err := getNetwork(onmetalNetworkInterface.Spec.NetworkRef.Name)
	if err != nil {
		return nil, fmt.Errorf("error getting network %s: %w", onmetalNetworkInterface.Name, err)
	}

	var virtualIP *networkingv1alpha1.VirtualIP
	if virtualIPSrc := onmetalNetworkInterface.Spec.VirtualIP; virtualIPSrc != nil {
		virtualIPRef := virtualIPSrc.VirtualIPRef
		if virtualIPRef == nil {
			return nil, fmt.Errorf("network interface specifies a non-ref virtual ip source")
		}

		v, err := getVirtualIP(virtualIPRef.Name)
		if err != nil {
			return nil, fmt.Errorf("error getting virtual ip %s: %w", virtualIPRef.Name, err)
		}

		virtualIP = v
	}

	loadBalancers, err := listLoadBalancers()
	if err != nil {
		return nil, fmt.Errorf("error listing load balancers: %w", err)
	}

	var lbTgts []machinebrokerv1alpha1.LoadBalancerTarget
	for _, loadBalancer := range loadBalancers {
		if !slices.Contains(loadBalancer.Destinations, onmetalNetworkInterface.Name) {
			continue
		}

		lbTgts = append(lbTgts, machinebrokerv1alpha1.LoadBalancerTarget{
			LoadBalancerType: loadBalancer.Type,
			IP:               loadBalancer.IP,
			Ports:            loadBalancer.Ports,
		})
	}

	natGateways, err := listNATGateways()
	if err != nil {
		return nil, fmt.Errorf("error listing nat gateways: %w", err)
	}

	var natGatewayTgts []machinebrokerv1alpha1.NATGatewayTarget
	for _, natGateway := range natGateways {
		dst, ok := utilslices.FindFunc(natGateway.Destinations,
			func(dest machinebrokerv1alpha1.NATGatewayDestination) bool {
				return dest.ID == onmetalNetworkInterface.Name
			},
		)
		if !ok {
			continue
		}

		natGatewayTgts = append(natGatewayTgts, machinebrokerv1alpha1.NATGatewayTarget{
			IP:      natGateway.IP,
			Port:    dst.Port,
			EndPort: dst.EndPort,
		})
	}

	return &AggregateOnmetalNetworkInterface{
		NetworkInterface:    onmetalNetworkInterface,
		Network:             network,
		VirtualIP:           virtualIP,
		LoadBalancerTargets: lbTgts,
		NATGatewayTargets:   natGatewayTgts,
	}, nil
}

func (s *Server) getAggregateOnmetalNetworkInterface(ctx context.Context, id string) (*AggregateOnmetalNetworkInterface, error) {
	onmetalNetworkInterface := &networkingv1alpha1.NetworkInterface{}
	if err := s.getManagedAndCreated(ctx, id, onmetalNetworkInterface); err != nil {
		if !apierrors.IsNotFound(err) {
			return nil, fmt.Errorf("error getting onmetal network interface %s: %w", id, err)
		}
		return nil, status.Errorf(codes.NotFound, "network interface %s not found", id)
	}

	return s.aggregateOnmetalNetworkInterface(
		onmetalNetworkInterface,
		utils.ClientObjectGetter[*networkingv1alpha1.Network](ctx, s.cluster.Client(), s.cluster.Namespace()),
		utils.ClientObjectGetter[*networkingv1alpha1.VirtualIP](ctx, s.cluster.Client(), s.cluster.Namespace()),
		func() ([]machinebrokerv1alpha1.LoadBalancer, error) {
			return s.loadBalancers.ListByDependent(ctx, id)
		},
		func() ([]machinebrokerv1alpha1.NATGateway, error) {
			return s.natGateways.ListByDependent(ctx, id)
		},
	)
}

func (s *Server) getNetworkInterface(ctx context.Context, id string) (*ori.NetworkInterface, error) {
	onmetalNetworkInterface, err := s.getAggregateOnmetalNetworkInterface(ctx, id)
	if err != nil {
		return nil, err
	}

	return s.convertAggregateOnmetalNetworkInterface(onmetalNetworkInterface)
}

func (s *Server) listNetworkInterfaces(ctx context.Context) ([]*ori.NetworkInterface, error) {
	onmetalNetworkInterfaces, err := s.listAggregateOnmetalNetworkInterfaces(ctx)
	if err != nil {
		return nil, err
	}

	res := make([]*ori.NetworkInterface, len(onmetalNetworkInterfaces))
	for i := range onmetalNetworkInterfaces {
		onmetalNetworkInterface := &onmetalNetworkInterfaces[i]
		networkInterface, err := s.convertAggregateOnmetalNetworkInterface(onmetalNetworkInterface)
		if err != nil {
			return nil, fmt.Errorf("error converting onmetal network interface %s: %w", onmetalNetworkInterface.NetworkInterface.Name, err)
		}

		res[i] = networkInterface
	}
	return res, nil
}

func (s *Server) filterNetworkInterfaces(networkInterfaces []*ori.NetworkInterface, filter *ori.NetworkInterfaceFilter) []*ori.NetworkInterface {
	if filter == nil {
		return networkInterfaces
	}

	var (
		res []*ori.NetworkInterface
		sel = labels.SelectorFromSet(filter.LabelSelector)
	)
	for _, networkInterface := range networkInterfaces {
		if !sel.Matches(labels.Set(networkInterface.Metadata.Labels)) {
			continue
		}

		res = append(res, networkInterface)
	}
	return res
}

func (s *Server) ListNetworkInterfaces(ctx context.Context, req *ori.ListNetworkInterfacesRequest) (*ori.ListNetworkInterfacesResponse, error) {
	if filter := req.Filter; filter != nil && filter.Id != "" {
		networkInterface, err := s.getNetworkInterface(ctx, filter.Id)
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

	networkInterfaces, err := s.listNetworkInterfaces(ctx)
	if err != nil {
		return nil, err
	}

	networkInterfaces = s.filterNetworkInterfaces(networkInterfaces, req.Filter)

	return &ori.ListNetworkInterfacesResponse{
		NetworkInterfaces: networkInterfaces,
	}, nil
}
