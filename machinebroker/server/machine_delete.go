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
	storagev1alpha1 "github.com/onmetal/onmetal-api/api/storage/v1alpha1"
	machinebrokerv1alpha1 "github.com/onmetal/onmetal-api/machinebroker/api/v1alpha1"
	ori "github.com/onmetal/onmetal-api/ori/apis/compute/v1alpha1"
	corev1 "k8s.io/api/core/v1"
	"sigs.k8s.io/controller-runtime/pkg/client"
)

func (s *Server) DeleteMachine(ctx context.Context, req *ori.DeleteMachineRequest) (*ori.DeleteMachineResponse, error) {
	machineID := req.MachineId
	log := s.loggerFrom(ctx, "MachineID", machineID)

	var errs []error

	log.V(1).Info("Deleting machine")
	if err := s.client.DeleteAllOf(ctx, &computev1alpha1.Machine{},
		client.InNamespace(s.namespace),
		client.MatchingLabels{machinebrokerv1alpha1.MachineIDLabel: machineID},
	); err != nil {
		errs = append(errs, fmt.Errorf("error deleting machine: %w", err))
	}

	log.V(1).Info("Deleting network interfaces")
	if err := s.client.DeleteAllOf(ctx, &networkingv1alpha1.NetworkInterface{},
		client.InNamespace(s.namespace),
		client.MatchingLabels{machinebrokerv1alpha1.MachineIDLabel: machineID},
	); err != nil {
		errs = append(errs, fmt.Errorf("error deleting network interfaces: %w", err))
	}

	log.V(1).Info("Deleting virtual ips")
	if err := s.client.DeleteAllOf(ctx, &networkingv1alpha1.VirtualIP{},
		client.InNamespace(s.namespace),
		client.MatchingLabels{machinebrokerv1alpha1.MachineIDLabel: machineID},
	); err != nil {
		errs = append(errs, fmt.Errorf("error deleting virtual ips"))
	}

	log.V(1).Info("Deleting networks")
	if err := s.client.DeleteAllOf(ctx, &networkingv1alpha1.Network{},
		client.InNamespace(s.namespace),
		client.MatchingLabels{machinebrokerv1alpha1.MachineIDLabel: machineID},
	); err != nil {
		errs = append(errs, fmt.Errorf("error deleting networks: %w", err))
	}

	log.V(1).Info("Deleting volumes")
	if err := s.client.DeleteAllOf(ctx, &storagev1alpha1.Volume{},
		client.InNamespace(s.namespace),
		client.MatchingLabels{machinebrokerv1alpha1.MachineIDLabel: machineID},
	); err != nil {
		errs = append(errs, fmt.Errorf("error deleting volumes: %w", err))
	}

	log.V(1).Info("Deleting secrets")
	if err := s.client.DeleteAllOf(ctx, &corev1.Secret{},
		client.InNamespace(s.namespace),
		client.MatchingLabels{machinebrokerv1alpha1.MachineIDLabel: machineID},
	); err != nil {
		errs = append(errs, fmt.Errorf("error deleting secrets: %w", err))
	}

	if len(errs) > 0 {
		return &ori.DeleteMachineResponse{}, fmt.Errorf("error(s) deleting machine: %v", errs)
	}
	return &ori.DeleteMachineResponse{}, nil
}
