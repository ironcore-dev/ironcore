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
	machinebrokerv1alpha1 "github.com/onmetal/onmetal-api/broker/machinebroker/api/v1alpha1"
	ori "github.com/onmetal/onmetal-api/ori/apis/machine/v1alpha1"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
	corev1 "k8s.io/api/core/v1"
	apierrors "k8s.io/apimachinery/pkg/api/errors"
	"k8s.io/apimachinery/pkg/labels"
	"sigs.k8s.io/controller-runtime/pkg/client"
)

func (s *Server) listAggregateOnmetalMachines(ctx context.Context) ([]AggregateOnmetalMachine, error) {
	onmetalMachineList := &computev1alpha1.MachineList{}
	if err := s.listManagedAndCreated(ctx, onmetalMachineList); err != nil {
		return nil, fmt.Errorf("error listing onmetal machines: %w", err)
	}

	ignitionSecretList := &corev1.SecretList{}
	if err := s.listWithPurpose(ctx, ignitionSecretList, machinebrokerv1alpha1.IgnitionPurpose); err != nil {
		return nil, fmt.Errorf("error listing ignition secrets: %w", err)
	}

	ignitionSecretByName := objectStructsToObjectPtrByNameMap(ignitionSecretList.Items)
	getIgnitionSecret := objectByNameMapGetter(corev1.Resource("secrets"), ignitionSecretByName)

	var res []AggregateOnmetalMachine
	for i := range onmetalMachineList.Items {
		onmetalMachine := &onmetalMachineList.Items[i]
		aggregateOnmetalMachine, err := s.aggregateOnmetalMachine(onmetalMachine, getIgnitionSecret)
		if err != nil {
			return nil, fmt.Errorf("error assembling onmetal machine %s: %w", onmetalMachine.Name, err)
		}

		res = append(res, *aggregateOnmetalMachine)
	}
	return res, nil
}

func (s *Server) aggregateOnmetalMachine(
	machine *computev1alpha1.Machine,
	getIgnitionSecret func(name string) (*corev1.Secret, error),
) (*AggregateOnmetalMachine, error) {
	var ignitionSecret *corev1.Secret
	if ignitionRef := machine.Spec.IgnitionRef; ignitionRef != nil {
		secret, err := getIgnitionSecret(ignitionRef.Name)
		if err != nil {
			return nil, fmt.Errorf("error getting machine ignition secret %s: %w", ignitionRef.Name, err)
		}

		ignitionSecret = secret
	}

	return &AggregateOnmetalMachine{
		IgnitionSecret: ignitionSecret,
		Machine:        machine,
	}, nil
}

func (s *Server) getAggregateOnmetalMachine(ctx context.Context, id string) (*AggregateOnmetalMachine, error) {
	onmetalMachine := &computev1alpha1.Machine{}
	if err := s.getManagedAndCreated(ctx, id, onmetalMachine); err != nil {
		if !apierrors.IsNotFound(err) {
			return nil, fmt.Errorf("error getting onmetal machine %s: %w", id, err)
		}
		return nil, status.Errorf(codes.NotFound, "machine %s not found", id)
	}

	return s.aggregateOnmetalMachine(onmetalMachine, func(name string) (*corev1.Secret, error) {
		secret := &corev1.Secret{}
		if err := s.client.Get(ctx, client.ObjectKey{Namespace: s.namespace, Name: name}, secret); err != nil {
			return nil, err
		}
		return secret, nil
	})
}

func (s *Server) listMachines(ctx context.Context) ([]*ori.Machine, error) {
	onmetalMachines, err := s.listAggregateOnmetalMachines(ctx)
	if err != nil {
		return nil, fmt.Errorf("error listing machines: %w", err)
	}

	var res []*ori.Machine
	for _, onmetalMachine := range onmetalMachines {
		machine, err := s.convertAggregateOnmetalMachine(&onmetalMachine)
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
		if !sel.Matches(labels.Set(oriMachine.Metadata.Labels)) {
			continue
		}

		res = append(res, oriMachine)
	}
	return res
}

func (s *Server) getMachine(ctx context.Context, id string) (*ori.Machine, error) {
	aggregateOnmetalMachine, err := s.getAggregateOnmetalMachine(ctx, id)
	if err != nil {
		return nil, err
	}

	return s.convertAggregateOnmetalMachine(aggregateOnmetalMachine)
}

func (s *Server) ListMachines(ctx context.Context, req *ori.ListMachinesRequest) (*ori.ListMachinesResponse, error) {
	if filter := req.Filter; filter != nil && filter.Id != "" {
		machine, err := s.getMachine(ctx, filter.Id)
		if err != nil {
			if status.Code(err) != codes.NotFound {
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
