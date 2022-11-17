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
	"errors"
	"fmt"

	"github.com/go-logr/logr"
	networkingv1alpha1 "github.com/onmetal/onmetal-api/api/networking/v1alpha1"
	machinebrokerv1alpha1 "github.com/onmetal/onmetal-api/machinebroker/api/v1alpha1"
	ori "github.com/onmetal/onmetal-api/ori/apis/compute/v1alpha1"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
	corev1 "k8s.io/api/core/v1"
	apierrors "k8s.io/apimachinery/pkg/api/errors"
	"k8s.io/apimachinery/pkg/labels"
	"k8s.io/apimachinery/pkg/selection"
	"sigs.k8s.io/controller-runtime/pkg/client"
)

func (s *Server) getOnmetalNetworkInterface(ctx context.Context, machineID, networkInterfaceName string) (*networkingv1alpha1.NetworkInterface, error) {
	onmetalNetworkInterface := &networkingv1alpha1.NetworkInterface{}
	onmetalNetworkInterfaceKey := client.ObjectKey{Namespace: s.namespace, Name: s.onmetalNetworkInterfaceName(machineID, networkInterfaceName)}
	if err := s.client.Get(ctx, onmetalNetworkInterfaceKey, onmetalNetworkInterface); err != nil {
		if !apierrors.IsNotFound(err) {
			return nil, fmt.Errorf("error getting machine %s network interface %s: %w", machineID, networkInterfaceName, err)
		}
		return nil, newNetworkInterfaceNotFoundError(machineID, networkInterfaceName)
	}
	return onmetalNetworkInterface, nil
}

// deleteOutdatedOnmetalNetworkInterfaceVirtualIPs deletes all ips of the network interface specified via
// machineID / networkInterfaceName that don't have the specified ipFamily.
func (s *Server) deleteOutdatedOnmetalNetworkInterfaceVirtualIPs(
	ctx context.Context,
	log logr.Logger,
	machineID string,
	networkInterfaceName string,
	ipFamily corev1.IPFamily,
) error {
	notIPFamilyRequirement, err := labels.NewRequirement(machinebrokerv1alpha1.IPFamilyLabel, selection.NotEquals, []string{string(ipFamily)})
	if err != nil {
		return fmt.Errorf("error creating ip family requirement: %w", err)
	}

	sel := labels.SelectorFromSet(labels.Set{
		machinebrokerv1alpha1.MachineIDLabel:            machineID,
		machinebrokerv1alpha1.NetworkInterfaceNameLabel: networkInterfaceName,
	}).Add(*notIPFamilyRequirement)

	log.V(1).Info("Deleting any outdated virtual ip")
	if err := s.client.DeleteAllOf(ctx, &networkingv1alpha1.VirtualIP{},
		client.InNamespace(s.namespace),
		client.MatchingLabelsSelector{Selector: sel},
	); err != nil {
		return fmt.Errorf("error deleting outdated virtual ips: %w", err)
	}
	return nil
}

func (s *Server) UpdateNetworkInterface(ctx context.Context, req *ori.UpdateNetworkInterfaceRequest) (res *ori.UpdateNetworkInterfaceResponse, retErr error) {
	machineID := req.MachineId
	networkInterfaceName := req.NetworkInterfaceName
	log := s.loggerFrom(ctx, "MachineID", machineID, "NetworkInterfaceName", networkInterfaceName)

	ips, err := s.parseIPs(req.Ips)
	if err != nil {
		return nil, err
	}

	var onmetalVirtualIPConfig *OnmetalVirtualIPConfig
	if req.VirtualIp != nil {
		onmetalVirtualIPConfig, err = s.getOnmetalVirtualIPConfig(machineID, networkInterfaceName, req.VirtualIp)
		if err != nil {
			return nil, err
		}
	}

	if err := s.updateOnmetalNetworkInterfaceVirtualIP(ctx, log, machineID, networkInterfaceName, onmetalVirtualIPConfig); err != nil {
		var networkInterfaceNotFound *networkInterfaceNotFoundError
		if !errors.As(err, &networkInterfaceNotFound) {
			return nil, err
		}
		return nil, status.Error(codes.NotFound, networkInterfaceNotFound.Error())
	}

	onmetalNetworkInterface, err := s.getOnmetalNetworkInterface(ctx, machineID, networkInterfaceName)
	if err != nil {
		return nil, err
	}

	log.V(1).Info("Patching network interface ips")
	baseOnmetalNetworkInterface := onmetalNetworkInterface.DeepCopy()
	onmetalNetworkInterface.Spec.IPs = s.onmetalIPsToOnmetalIPSources(ips)
	if err := s.client.Patch(ctx, onmetalNetworkInterface, client.MergeFrom(baseOnmetalNetworkInterface)); err != nil {
		return nil, fmt.Errorf("error patching network interface ips: %w", err)
	}

	return &ori.UpdateNetworkInterfaceResponse{}, nil
}

func (s *Server) updateOnmetalNetworkInterfaceVirtualIP(
	ctx context.Context,
	log logr.Logger,
	machineID string,
	networkInterfaceName string,
	virtualIPConfig *OnmetalVirtualIPConfig,
) (retErr error) {
	if virtualIPConfig == nil {
		onmetalNetworkInterface, err := s.getOnmetalNetworkInterface(ctx, machineID, networkInterfaceName)
		if err != nil {
			return err
		}

		if onmetalNetworkInterface.Spec.VirtualIP != nil {
			log.V(1).Info("Removing virtual ip from network interface")
			baseOnmetalNetworkInterface := onmetalNetworkInterface.DeepCopy()
			onmetalNetworkInterface.Spec.VirtualIP = nil
			if err := s.client.Patch(ctx, onmetalNetworkInterface, client.MergeFrom(baseOnmetalNetworkInterface)); err != nil {
				return fmt.Errorf("error removing virtual ip from network interface: %w", err)
			}

			log.V(1).Info("Deleting virtual ip")
			if err := s.client.DeleteAllOf(ctx, &networkingv1alpha1.VirtualIP{},
				client.InNamespace(s.namespace),
				client.MatchingLabels{
					machinebrokerv1alpha1.MachineIDLabel:            machineID,
					machinebrokerv1alpha1.NetworkInterfaceNameLabel: networkInterfaceName,
				},
			); err != nil {
				return fmt.Errorf("error deleting virtual ip: %w", err)
			}
		}
		return nil
	}

	cleaner, cleanup := s.setupCleaner(ctx, log, &retErr)
	defer cleanup()

	onmetalVirtualIP := &networkingv1alpha1.VirtualIP{}
	onmetalVirtualIPKey := client.ObjectKeyFromObject(virtualIPConfig.VirtualIP)
	if err := s.client.Get(ctx, onmetalVirtualIPKey, onmetalVirtualIP); err != nil {
		if !apierrors.IsNotFound(err) {
			return fmt.Errorf("error getting machine %s virtual ip: %w", machineID, err)
		}

		onmetalVirtualIP, err = s.createOnmetalVirtualIP(ctx, log, cleaner, virtualIPConfig)
		if err != nil {
			return err
		}
	}

	if err := s.setOnmetalVirtualIPIP(ctx, log, onmetalVirtualIP, virtualIPConfig.IP); err != nil {
		return err
	}

	onmetalNetworkInterface, err := s.getOnmetalNetworkInterface(ctx, machineID, networkInterfaceName)
	if err != nil {
		return err
	}

	log.V(1).Info("Patching network interface virtual ip")
	baseOnmetalNetworkInterface := onmetalNetworkInterface.DeepCopy()
	onmetalNetworkInterface.Spec.VirtualIP = &networkingv1alpha1.VirtualIPSource{
		VirtualIPRef: &corev1.LocalObjectReference{Name: onmetalVirtualIP.Name},
	}
	if err := s.client.Patch(ctx, onmetalNetworkInterface, client.MergeFrom(baseOnmetalNetworkInterface)); err != nil {
		return fmt.Errorf("error patching network interface virtual ip: %w", err)
	}

	if err := s.deleteOutdatedOnmetalNetworkInterfaceVirtualIPs(ctx, log, machineID, networkInterfaceName, virtualIPConfig.IP.Family()); err != nil {
		log.Error(err, "Error deleting outdated virtual ips")
	}
	return nil
}
