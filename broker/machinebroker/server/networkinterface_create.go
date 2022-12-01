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
	networkingv1alpha1 "github.com/onmetal/onmetal-api/api/networking/v1alpha1"
	onmetalapiannotations "github.com/onmetal/onmetal-api/apiutils/annotations"
	machinebrokerv1alpha1 "github.com/onmetal/onmetal-api/broker/machinebroker/api/v1alpha1"
	"github.com/onmetal/onmetal-api/broker/machinebroker/apiutils"
	"github.com/onmetal/onmetal-api/broker/machinebroker/cleaner"
	ori "github.com/onmetal/onmetal-api/ori/apis/machine/v1alpha1"
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"sigs.k8s.io/controller-runtime/pkg/client"
)

func (s *Server) prepareOnmetalNetwork(networkSpec *ori.NetworkSpec) (*networkingv1alpha1.Network, error) {
	onmetalNetwork := &networkingv1alpha1.Network{
		ObjectMeta: metav1.ObjectMeta{
			Namespace: s.namespace,
			Name:      s.generateID(),
		},
		Spec: networkingv1alpha1.NetworkSpec{
			Handle: networkSpec.Handle,
		},
	}
	onmetalapiannotations.SetExternallyMangedBy(onmetalNetwork, machinebrokerv1alpha1.MachineBrokerManager)
	apiutils.SetPurpose(onmetalNetwork, machinebrokerv1alpha1.NetworkInterfacePurpose)
	return onmetalNetwork, nil
}

func (s *Server) prepareOnmetalVirtualIP(virtualIPSpec *ori.VirtualIPSpec) (*networkingv1alpha1.VirtualIP, error) {
	ip, err := commonv1alpha1.ParseIP(virtualIPSpec.Ip)
	if err != nil {
		return nil, err
	}

	onmetalVirtualIP := &networkingv1alpha1.VirtualIP{
		ObjectMeta: metav1.ObjectMeta{
			Namespace: s.namespace,
			Name:      s.generateID(),
		},
		Spec: networkingv1alpha1.VirtualIPSpec{
			Type:     networkingv1alpha1.VirtualIPTypePublic,
			IPFamily: ip.Family(),
		},
		Status: networkingv1alpha1.VirtualIPStatus{
			IP: &ip,
		},
	}
	onmetalapiannotations.SetExternallyMangedBy(onmetalVirtualIP, machinebrokerv1alpha1.MachineBrokerManager)
	apiutils.SetPurpose(onmetalVirtualIP, machinebrokerv1alpha1.NetworkInterfacePurpose)
	return onmetalVirtualIP, nil
}

func (s *Server) prepareAggregateOnmetalNetworkInterface(networkInterface *ori.NetworkInterface) (*AggregateOnmetalNetworkInterface, error) {
	onmetalNetwork, err := s.prepareOnmetalNetwork(networkInterface.Spec.Network)
	if err != nil {
		return nil, fmt.Errorf("error preparing onmetal network: %w", err)
	}

	var onmetalVirtualIP *networkingv1alpha1.VirtualIP
	if virtualIPSpec := networkInterface.Spec.VirtualIp; virtualIPSpec != nil {
		v, err := s.prepareOnmetalVirtualIP(virtualIPSpec)
		if err != nil {
			return nil, err
		}

		onmetalVirtualIP = v
	}

	ips, err := s.parseIPs(networkInterface.Spec.Ips)
	if err != nil {
		return nil, err
	}

	var virtualIPSource *networkingv1alpha1.VirtualIPSource
	if onmetalVirtualIP != nil {
		virtualIPSource = &networkingv1alpha1.VirtualIPSource{
			VirtualIPRef: &corev1.LocalObjectReference{Name: onmetalVirtualIP.Name},
		}
	}
	onmetalNetworkInterface := &networkingv1alpha1.NetworkInterface{
		ObjectMeta: metav1.ObjectMeta{
			Namespace: s.namespace,
			Name:      s.generateID(),
		},
		Spec: networkingv1alpha1.NetworkInterfaceSpec{
			NetworkRef: corev1.LocalObjectReference{Name: onmetalNetwork.Name},
			IPFamilies: s.getOnmetalIPsIPFamilies(ips),
			IPs:        s.onmetalIPsToOnmetalIPSources(ips),
			VirtualIP:  virtualIPSource,
		},
	}
	if err := apiutils.SetObjectMetadata(onmetalNetworkInterface, networkInterface.Metadata); err != nil {
		return nil, err
	}
	apiutils.SetManagerLabel(onmetalNetworkInterface, machinebrokerv1alpha1.MachineBrokerManager)

	onmetalNetworkInterfaceConfig := &AggregateOnmetalNetworkInterface{
		Network:          onmetalNetwork,
		VirtualIP:        onmetalVirtualIP,
		NetworkInterface: onmetalNetworkInterface,
	}
	return onmetalNetworkInterfaceConfig, nil
}

func (s *Server) createOnmetalNetwork(ctx context.Context, log logr.Logger, cleaner *cleaner.Cleaner, onmetalNetwork *networkingv1alpha1.Network) error {
	log.V(1).Info("Creating onmetal network")
	if err := s.client.Create(ctx, onmetalNetwork); err != nil {
		return fmt.Errorf("error creating onmetal network: %w", err)
	}
	cleaner.Add(func(ctx context.Context) error {
		if err := s.client.Delete(ctx, onmetalNetwork); client.IgnoreNotFound(err) != nil {
			return fmt.Errorf("error deleting onmetal network: %w", err)
		}
		return nil
	})

	baseOnmetalNetwork := onmetalNetwork.DeepCopy()
	onmetalNetwork.Status.State = networkingv1alpha1.NetworkStateAvailable
	if err := s.client.Status().Patch(ctx, onmetalNetwork, client.MergeFrom(baseOnmetalNetwork)); err != nil {
		return fmt.Errorf("error patching onmetal network status state to available: %w", err)
	}

	return nil
}

func (s *Server) setOnmetalVirtualIPStatusIP(ctx context.Context, onmetalVirtualIP *networkingv1alpha1.VirtualIP, ip *commonv1alpha1.IP) error {
	baseOnmetalVirtualIP := onmetalVirtualIP.DeepCopy()
	onmetalVirtualIP.Status.IP = ip
	if err := s.client.Status().Patch(ctx, onmetalVirtualIP, client.MergeFrom(baseOnmetalVirtualIP)); err != nil {
		return fmt.Errorf("error patching onmetal virtual ip status: %w", err)
	}
	return nil
}

func (s *Server) createOnmetalVirtualIP(ctx context.Context, log logr.Logger, c *cleaner.Cleaner, onmetalVirtualIP *networkingv1alpha1.VirtualIP) error {
	ip := *onmetalVirtualIP.Status.IP

	log.V(1).Info("Creating onmetal virtual ip")
	if err := s.client.Create(ctx, onmetalVirtualIP); err != nil {
		return fmt.Errorf("error creating onmetal virtual ip: %w", err)
	}
	c.Add(func(ctx context.Context) error {
		if err := s.client.Delete(ctx, onmetalVirtualIP); client.IgnoreNotFound(err) != nil {
			return fmt.Errorf("error deleting virtual ip: %w", err)
		}
		return nil
	})

	log.V(1).Info("Patching onmetal virtual ip status")
	if err := s.setOnmetalVirtualIPStatusIP(ctx, onmetalVirtualIP, &ip); err != nil {
		return fmt.Errorf("error setting onmetal virtual ip status ip: %w", err)
	}

	return nil
}

func (s *Server) createOnmetalNetworkInterface(ctx context.Context, log logr.Logger, onmetalNetworkInterface *AggregateOnmetalNetworkInterface) (retErr error) {
	c, cleanup := s.setupCleaner(ctx, log, &retErr)
	defer cleanup()

	if err := s.createOnmetalNetwork(ctx, log, c, onmetalNetworkInterface.Network); err != nil {
		return err
	}

	if onmetalVirtualIP := onmetalNetworkInterface.VirtualIP; onmetalVirtualIP != nil {
		if err := s.createOnmetalVirtualIP(ctx, log, c, onmetalVirtualIP); err != nil {
			return err
		}
	}

	log.V(1).Info("Creating onmetal network interface")
	if err := s.client.Create(ctx, onmetalNetworkInterface.NetworkInterface); err != nil {
		return fmt.Errorf("error creating onmetal network interface: %w", err)
	}
	c.Add(func(ctx context.Context) error {
		if err := s.client.Delete(ctx, onmetalNetworkInterface.NetworkInterface); client.IgnoreNotFound(err) != nil {
			return fmt.Errorf("error deleting onmetal network interface: %w", err)
		}
		return nil
	})

	log.V(1).Info("Patching onmetal network to be controlled by onmetal network interface")
	if err := apiutils.PatchControlledBy(ctx, s.client, onmetalNetworkInterface.NetworkInterface, onmetalNetworkInterface.Network); err != nil {
		return fmt.Errorf("error patching onmetal network to be controlled by onmetal network interface: %w", err)
	}

	if onmetalNetworkInterface.VirtualIP != nil {
		log.V(1).Info("Patching onmetal virtual ip to be controlled by onmetal network interface")
		if err := apiutils.PatchControlledBy(ctx, s.client, onmetalNetworkInterface.NetworkInterface, onmetalNetworkInterface.VirtualIP); err != nil {
			return fmt.Errorf("error patching onmetal virtual ip to be controlled by onmetal network interface: %w", err)
		}
	}

	log.V(1).Info("Patching onmetal network interface as created")
	if err := apiutils.PatchCreated(ctx, s.client, onmetalNetworkInterface.NetworkInterface); err != nil {
		return fmt.Errorf("error patching onmetal network interface as created: %w", err)
	}

	return nil
}

func (s *Server) CreateNetworkInterface(ctx context.Context, req *ori.CreateNetworkInterfaceRequest) (res *ori.CreateNetworkInterfaceResponse, retErr error) {
	log := s.loggerFrom(ctx)

	onmetalNetworkInterface, err := s.prepareAggregateOnmetalNetworkInterface(req.NetworkInterface)
	if err != nil {
		return nil, fmt.Errorf("error getting onmetal network interface config: %w", err)
	}

	if err := s.createOnmetalNetworkInterface(ctx, log, onmetalNetworkInterface); err != nil {
		return nil, fmt.Errorf("error creating onmetal network interface: %w", err)
	}

	networkInterface, err := s.convertAggregateOnmetalNetworkInterface(onmetalNetworkInterface)
	if err != nil {
		return nil, fmt.Errorf("error converting onmetal network interface: %w", err)
	}

	return &ori.CreateNetworkInterfaceResponse{
		NetworkInterface: networkInterface,
	}, nil
}
