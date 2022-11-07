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

	computev1alpha1 "github.com/onmetal/onmetal-api/apis/compute/v1alpha1"
	machinebrokerv1alpha1 "github.com/onmetal/onmetal-api/machinebroker/api/v1alpha1"
	ori "github.com/onmetal/onmetal-api/ori/apis/runtime/v1alpha1"
	"k8s.io/apimachinery/pkg/labels"
	"sigs.k8s.io/controller-runtime/pkg/client"
)

func (s *Server) listOnmetalMachines(ctx context.Context) ([]computev1alpha1.Machine, error) {
	onmetalMachineList := &computev1alpha1.MachineList{}

	if err := s.client.List(ctx, onmetalMachineList,
		client.InNamespace(s.namespace),
		client.MatchingLabels{machinebrokerv1alpha1.MachineManagerLabel: machinebrokerv1alpha1.MachineBrokerManager},
	); err != nil {
		return nil, fmt.Errorf("error listing machines: %w", err)
	}

	return onmetalMachineList.Items, nil
}

func (s *Server) listMachines(ctx context.Context) ([]*ori.Machine, error) {
	onmetalMachines, err := s.listOnmetalMachines(ctx)
	if err != nil {
		return nil, fmt.Errorf("error listing machines: %w", err)
	}

	var res []*ori.Machine
	for _, onmetalMachine := range onmetalMachines {
		machine, err := MachineToORIMachine(&onmetalMachine)
		if err != nil {
			return nil, err
		}

		res = append(res, machine)
	}
	return res, nil
}

func (s *Server) filterMachines(machines []*ori.Machine, filter *ori.MachineFilter) []*ori.Machine {
	if filter == nil {
		return machines
	}

	var (
		res []*ori.Machine
		sel = labels.SelectorFromSet(filter.LabelSelector)
	)
	for _, oriMachine := range machines {
		if !sel.Matches(labels.Set(oriMachine.Labels)) {
			continue
		}

		res = append(res, oriMachine)
	}
	return res
}

func (s *Server) getOnmetalMachine(ctx context.Context, id string) (*computev1alpha1.Machine, error) {
	onmetalMachineList := &computev1alpha1.MachineList{}
	if err := s.client.List(ctx, onmetalMachineList,
		client.InNamespace(s.namespace),
		client.MatchingLabels{
			machinebrokerv1alpha1.MachineIDLabel:      id,
			machinebrokerv1alpha1.MachineManagerLabel: machinebrokerv1alpha1.MachineBrokerManager,
		},
		client.Limit(2),
	); err != nil {
		return nil, fmt.Errorf("error listing machines by id %s: %w", id, err)
	}

	switch len(onmetalMachineList.Items) {
	case 0:
		return nil, ErrMachineNotFound
	case 1:
		return &onmetalMachineList.Items[0], nil
	default:
		return nil, fmt.Errorf("more than 1 machine found for id %s", id)
	}
}

func (s *Server) getMachine(ctx context.Context, id string) (*ori.Machine, error) {
	onmetalMachine, err := s.getOnmetalMachine(ctx, id)
	if err != nil {
		return nil, err
	}

	return MachineToORIMachine(onmetalMachine)
}

func (s *Server) ListMachines(ctx context.Context, req *ori.ListMachinesRequest) (*ori.ListMachinesResponse, error) {
	if filter := req.Filter; filter != nil && filter.Id != "" {
		machine, err := s.getMachine(ctx, filter.Id)
		if err != nil {
			if !errors.Is(err, ErrMachineNotFound) {
				return nil, err
			}
			return &ori.ListMachinesResponse{
				Machines: []*ori.Machine{},
			}, nil
		}

		return &ori.ListMachinesResponse{
			Machines: []*ori.Machine{machine},
		}, nil
	}

	machines, err := s.listMachines(ctx)
	if err != nil {
		return nil, err
	}

	machines = s.filterMachines(machines, req.Filter)

	return &ori.ListMachinesResponse{
		Machines: machines,
	}, nil
}
