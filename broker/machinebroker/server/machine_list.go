// Copyright 2022 IronCore authors
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

	computev1alpha1 "github.com/ironcore-dev/ironcore/api/compute/v1alpha1"
	networkingv1alpha1 "github.com/ironcore-dev/ironcore/api/networking/v1alpha1"
	storagev1alpha1 "github.com/ironcore-dev/ironcore/api/storage/v1alpha1"
	machinebrokerv1alpha1 "github.com/ironcore-dev/ironcore/broker/machinebroker/api/v1alpha1"
	"github.com/ironcore-dev/ironcore/broker/machinebroker/apiutils"
	iri "github.com/ironcore-dev/ironcore/iri/apis/machine/v1alpha1"
	clientutils "github.com/ironcore-dev/ironcore/utils/client"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
	corev1 "k8s.io/api/core/v1"
	apierrors "k8s.io/apimachinery/pkg/api/errors"
	"k8s.io/apimachinery/pkg/labels"
	"sigs.k8s.io/controller-runtime/pkg/client"
)

func (s *Server) listAggregateIronCoreMachines(ctx context.Context) ([]AggregateIronCoreMachine, error) {
	ironcoreMachineList := &computev1alpha1.MachineList{}
	if err := s.cluster.Client().List(ctx, ironcoreMachineList,
		client.InNamespace(s.cluster.Namespace()),
		client.MatchingLabels{
			machinebrokerv1alpha1.ManagerLabel: machinebrokerv1alpha1.MachineBrokerManager,
			machinebrokerv1alpha1.CreatedLabel: "true",
		},
	); err != nil {
		return nil, fmt.Errorf("error listing ironcore machines: %w", err)
	}

	listOpts := []client.ListOption{
		client.InNamespace(s.cluster.Namespace()),
		client.MatchingLabels{
			machinebrokerv1alpha1.ManagerLabel: machinebrokerv1alpha1.MachineBrokerManager,
		},
	}
	rd, err := clientutils.NewCachingReaderBuilder(s.cluster.Client()).
		List(&corev1.SecretList{}, listOpts...).
		List(&networkingv1alpha1.NetworkList{}, listOpts...).
		Build(ctx)
	if err != nil {
		return nil, fmt.Errorf("error building caching reader: %w", err)
	}

	var res []AggregateIronCoreMachine
	for i := range ironcoreMachineList.Items {
		ironcoreMachine := &ironcoreMachineList.Items[i]
		aggregateIronCoreMachine, err := s.aggregateIronCoreMachine(ctx, rd, ironcoreMachine)
		if err != nil {
			return nil, fmt.Errorf("error aggregating ironcore machine %s: %w", ironcoreMachine.Name, err)
		}

		res = append(res, *aggregateIronCoreMachine)
	}
	return res, nil
}

func (s *Server) aggregateIronCoreMachine(
	ctx context.Context,
	rd client.Reader,
	ironcoreMachine *computev1alpha1.Machine,
) (*AggregateIronCoreMachine, error) {
	var ignitionSecret *corev1.Secret
	if ignitionRef := ironcoreMachine.Spec.IgnitionRef; ignitionRef != nil {
		secret := &corev1.Secret{}
		secretKey := client.ObjectKey{Namespace: s.cluster.Namespace(), Name: ignitionRef.Name}
		if err := rd.Get(ctx, secretKey, secret); err != nil {
			return nil, fmt.Errorf("error getting ironcore ignition secret: %w", err)
		}

		ignitionSecret = secret
	}

	aggIronCoreNics := make(map[string]*AggregateIronCoreNetworkInterface)
	for _, machineNic := range ironcoreMachine.Spec.NetworkInterfaces {
		switch {
		case machineNic.NetworkInterfaceRef != nil:
			ironcoreNic := &networkingv1alpha1.NetworkInterface{}
			ironcoreNicKey := client.ObjectKey{Namespace: s.cluster.Namespace(), Name: machineNic.NetworkInterfaceRef.Name}
			if err := rd.Get(ctx, ironcoreNicKey, ironcoreNic); err != nil {
				return nil, fmt.Errorf("error getting ironcore network interface: %w", err)
			}

			aggIronCoreNic, err := s.aggregateIronCoreNetworkInterface(ctx, rd, ironcoreNic)
			if err != nil {
				return nil, fmt.Errorf("error aggregating network interface: %w", err)
			}

			aggIronCoreNics[machineNic.Name] = aggIronCoreNic
		}
	}

	aggIronCoreVolumes := make(map[string]*AggregateIronCoreVolume)
	for _, machineVolume := range ironcoreMachine.Spec.Volumes {
		switch {
		case machineVolume.VolumeRef != nil:
			ironcoreVolume := &storagev1alpha1.Volume{}
			ironcoreVolumeKey := client.ObjectKey{Namespace: s.cluster.Namespace(), Name: machineVolume.VolumeRef.Name}
			if err := rd.Get(ctx, ironcoreVolumeKey, ironcoreVolume); err != nil {
				return nil, fmt.Errorf("error getting ironcore volume: %w", err)
			}

			aggIronCoreVolume, err := s.aggregateIronCoreVolume(ctx, rd, ironcoreVolume)
			if err != nil {
				return nil, fmt.Errorf("error aggregating volume: %w", err)
			}

			aggIronCoreVolumes[machineVolume.Name] = aggIronCoreVolume
		}
	}

	return &AggregateIronCoreMachine{
		IgnitionSecret:    ignitionSecret,
		Machine:           ironcoreMachine,
		NetworkInterfaces: aggIronCoreNics,
		Volumes:           aggIronCoreVolumes,
	}, nil
}

func (s *Server) getIronCoreMachine(ctx context.Context, id string) (*computev1alpha1.Machine, error) {
	ironcoreMachine := &computev1alpha1.Machine{}
	ironcoreMachineKey := client.ObjectKey{Namespace: s.cluster.Namespace(), Name: id}
	if err := s.cluster.Client().Get(ctx, ironcoreMachineKey, ironcoreMachine); err != nil {
		if !apierrors.IsNotFound(err) {
			return nil, fmt.Errorf("error getting ironcore machine %s: %w", id, err)
		}
		return nil, status.Errorf(codes.NotFound, "machine %s not found", id)
	}
	if !apiutils.IsManagedBy(ironcoreMachine, machinebrokerv1alpha1.MachineBrokerManager) || !apiutils.IsCreated(ironcoreMachine) {
		return nil, status.Errorf(codes.NotFound, "machine %s not found", id)
	}
	return ironcoreMachine, nil
}

func (s *Server) getAggregateIronCoreMachine(ctx context.Context, id string) (*AggregateIronCoreMachine, error) {
	ironcoreMachine, err := s.getIronCoreMachine(ctx, id)
	if err != nil {
		return nil, err
	}

	return s.aggregateIronCoreMachine(ctx, s.cluster.Client(), ironcoreMachine)
}

func (s *Server) listMachines(ctx context.Context) ([]*iri.Machine, error) {
	ironcoreMachines, err := s.listAggregateIronCoreMachines(ctx)
	if err != nil {
		return nil, fmt.Errorf("error listing machines: %w", err)
	}

	var res []*iri.Machine
	for _, ironcoreMachine := range ironcoreMachines {
		machine, err := s.convertAggregateIronCoreMachine(&ironcoreMachine)
		if err != nil {
			return nil, err
		}

		res = append(res, machine)
	}
	return res, nil
}

func (s *Server) filterMachines(machines []*iri.Machine, filter *iri.MachineFilter) []*iri.Machine {
	if filter == nil {
		return machines
	}

	var (
		res []*iri.Machine
		sel = labels.SelectorFromSet(filter.LabelSelector)
	)
	for _, iriMachine := range machines {
		if !sel.Matches(labels.Set(iriMachine.Metadata.Labels)) {
			continue
		}

		res = append(res, iriMachine)
	}
	return res
}

func (s *Server) getMachine(ctx context.Context, id string) (*iri.Machine, error) {
	aggregateIronCoreMachine, err := s.getAggregateIronCoreMachine(ctx, id)
	if err != nil {
		return nil, err
	}

	return s.convertAggregateIronCoreMachine(aggregateIronCoreMachine)
}

func (s *Server) ListMachines(ctx context.Context, req *iri.ListMachinesRequest) (*iri.ListMachinesResponse, error) {
	if filter := req.Filter; filter != nil && filter.Id != "" {
		machine, err := s.getMachine(ctx, filter.Id)
		if err != nil {
			if status.Code(err) != codes.NotFound {
				return nil, err
			}
			return &iri.ListMachinesResponse{
				Machines: []*iri.Machine{},
			}, nil
		}

		return &iri.ListMachinesResponse{
			Machines: []*iri.Machine{machine},
		}, nil
	}

	machines, err := s.listMachines(ctx)
	if err != nil {
		return nil, err
	}

	machines = s.filterMachines(machines, req.Filter)

	return &iri.ListMachinesResponse{
		Machines: machines,
	}, nil
}
