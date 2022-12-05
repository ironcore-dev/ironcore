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
	machinebrokerv1alpha1 "github.com/onmetal/onmetal-api/broker/machinebroker/api/v1alpha1"
	ori "github.com/onmetal/onmetal-api/ori/apis/machine/v1alpha1"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
	apierrors "k8s.io/apimachinery/pkg/api/errors"
	"k8s.io/apimachinery/pkg/labels"
)

func (s *Server) listAggregateOnmetalNetworkInterfaces(ctx context.Context) ([]AggregateOnmetalNetworkInterface, error) {
	onmetalNetworkInterfaceList := &networkingv1alpha1.NetworkInterfaceList{}
	if err := s.listManagedAndCreated(ctx, onmetalNetworkInterfaceList); err != nil {
		return nil, fmt.Errorf("error listing onmetal network interfaces: %w", err)
	}

	onmetalNetworkList := &networkingv1alpha1.NetworkList{}
	if err := s.listWithPurpose(ctx, onmetalNetworkList, machinebrokerv1alpha1.NetworkInterfacePurpose); err != nil {
		return nil, fmt.Errorf("error listing onmetal networks: %w", err)
	}

	onmetalVirtualIPList := &networkingv1alpha1.VirtualIPList{}
	if err := s.listWithPurpose(ctx, onmetalVirtualIPList, machinebrokerv1alpha1.NetworkInterfacePurpose); err != nil {
		return nil, fmt.Errorf("error listing onmetal virtual ips: %w", err)
	}

	onmetalNetworkByName := objectStructsToObjectPtrByNameMap(onmetalNetworkList.Items)
	onmetalVirtualIPByName := objectStructsToObjectPtrByNameMap(onmetalVirtualIPList.Items)

	getNetwork := objectByNameMapGetter(networkingv1alpha1.Resource("networks"), onmetalNetworkByName)
	getVirtualIP := objectByNameMapGetter(networkingv1alpha1.Resource("virtualips"), onmetalVirtualIPByName)

	var res []AggregateOnmetalNetworkInterface
	for i := range onmetalNetworkInterfaceList.Items {
		onmetalNetworkInterface := &onmetalNetworkInterfaceList.Items[i]
		networkInterface, err := s.aggregateOnmetalNetworkInterface(onmetalNetworkInterface, getNetwork, getVirtualIP)
		if err != nil {
			return nil, fmt.Errorf("error assembling onmetal network interface %s: %w", onmetalNetworkInterface.Name, err)
		}

		res = append(res, *networkInterface)
	}
	return res, nil
}

func (s *Server) aggregateOnmetalNetworkInterface(
	onmetalNetworkInterface *networkingv1alpha1.NetworkInterface,
	getNetwork func(name string) (*networkingv1alpha1.Network, error),
	getVirtualIP func(name string) (*networkingv1alpha1.VirtualIP, error),
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

	return &AggregateOnmetalNetworkInterface{
		NetworkInterface: onmetalNetworkInterface,
		Network:          network,
		VirtualIP:        virtualIP,
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
		clientGetter[networkingv1alpha1.Network](ctx, s.client, s.namespace),
		clientGetter[networkingv1alpha1.VirtualIP](ctx, s.client, s.namespace),
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
