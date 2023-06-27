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
	machinebrokerv1alpha1 "github.com/onmetal/onmetal-api/broker/machinebroker/api/v1alpha1"
	"github.com/onmetal/onmetal-api/broker/machinebroker/apiutils"
	ori "github.com/onmetal/onmetal-api/ori/apis/machine/v1alpha1"
	clientutils "github.com/onmetal/onmetal-api/utils/client"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
	corev1 "k8s.io/api/core/v1"
	apierrors "k8s.io/apimachinery/pkg/api/errors"
	"k8s.io/apimachinery/pkg/labels"
	"sigs.k8s.io/controller-runtime/pkg/client"
)

func (s *Server) listAggregateOnmetalMachines(ctx context.Context) ([]AggregateOnmetalMachine, error) {
	onmetalMachineList := &computev1alpha1.MachineList{}
	if err := s.cluster.Client().List(ctx, onmetalMachineList,
		client.InNamespace(s.cluster.Namespace()),
		client.MatchingLabels{
			machinebrokerv1alpha1.ManagerLabel: machinebrokerv1alpha1.MachineBrokerManager,
			machinebrokerv1alpha1.CreatedLabel: "true",
		},
	); err != nil {
		return nil, fmt.Errorf("error listing onmetal machines: %w", err)
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

	var res []AggregateOnmetalMachine
	for i := range onmetalMachineList.Items {
		onmetalMachine := &onmetalMachineList.Items[i]
		aggregateOnmetalMachine, err := s.aggregateOnmetalMachine(ctx, rd, onmetalMachine)
		if err != nil {
			return nil, fmt.Errorf("error aggregating onmetal machine %s: %w", onmetalMachine.Name, err)
		}

		res = append(res, *aggregateOnmetalMachine)
	}
	return res, nil
}

func (s *Server) aggregateOnmetalMachine(
	ctx context.Context,
	rd client.Reader,
	onmetalMachine *computev1alpha1.Machine,
) (*AggregateOnmetalMachine, error) {
	var ignitionSecret *corev1.Secret
	if ignitionRef := onmetalMachine.Spec.IgnitionRef; ignitionRef != nil {
		secret := &corev1.Secret{}
		secretKey := client.ObjectKey{Namespace: s.cluster.Namespace(), Name: ignitionRef.Name}
		if err := rd.Get(ctx, secretKey, secret); err != nil {
			return nil, fmt.Errorf("error getting onmetal ignition secret: %w", err)
		}

		ignitionSecret = secret
	}

	aggOnmetalNics := make(map[string]*AggregateOnmetalNetworkInterface)
	for _, machineNic := range onmetalMachine.Spec.NetworkInterfaces {
		switch {
		case machineNic.NetworkInterfaceRef != nil:
			onmetalNic := &networkingv1alpha1.NetworkInterface{}
			onmetalNicKey := client.ObjectKey{Namespace: s.cluster.Namespace(), Name: machineNic.NetworkInterfaceRef.Name}
			if err := rd.Get(ctx, onmetalNicKey, onmetalNic); err != nil {
				return nil, fmt.Errorf("error getting onmetal network interface: %w", err)
			}

			aggOnmetalNic, err := s.aggregateOnmetalNetworkInterface(ctx, rd, onmetalNic)
			if err != nil {
				return nil, fmt.Errorf("error aggregating network interface: %w", err)
			}

			aggOnmetalNics[machineNic.Name] = aggOnmetalNic
		}
	}

	aggOnmetalVolumes := make(map[string]*AggregateOnmetalVolume)
	for _, machineVolume := range onmetalMachine.Spec.Volumes {
		switch {
		case machineVolume.VolumeRef != nil:
			onmetalVolume := &storagev1alpha1.Volume{}
			onmetalVolumeKey := client.ObjectKey{Namespace: s.cluster.Namespace(), Name: machineVolume.VolumeRef.Name}
			if err := rd.Get(ctx, onmetalVolumeKey, onmetalVolume); err != nil {
				return nil, fmt.Errorf("error getting onmetal volume: %w", err)
			}

			aggOnmetalVolume, err := s.aggregateOnmetalVolume(ctx, rd, onmetalVolume)
			if err != nil {
				return nil, fmt.Errorf("error aggregating volume: %w", err)
			}

			aggOnmetalVolumes[machineVolume.Name] = aggOnmetalVolume
		}
	}

	return &AggregateOnmetalMachine{
		IgnitionSecret:    ignitionSecret,
		Machine:           onmetalMachine,
		NetworkInterfaces: aggOnmetalNics,
		Volumes:           aggOnmetalVolumes,
	}, nil
}

func (s *Server) getOnmetalMachine(ctx context.Context, id string) (*computev1alpha1.Machine, error) {
	onmetalMachine := &computev1alpha1.Machine{}
	onmetalMachineKey := client.ObjectKey{Namespace: s.cluster.Namespace(), Name: id}
	if err := s.cluster.Client().Get(ctx, onmetalMachineKey, onmetalMachine); err != nil {
		if !apierrors.IsNotFound(err) {
			return nil, fmt.Errorf("error getting onmetal machine %s: %w", id, err)
		}
		return nil, status.Errorf(codes.NotFound, "machine %s not found", id)
	}
	if !apiutils.IsManagedBy(onmetalMachine, machinebrokerv1alpha1.MachineBrokerManager) || !apiutils.IsCreated(onmetalMachine) {
		return nil, status.Errorf(codes.NotFound, "machine %s not found", id)
	}
	return onmetalMachine, nil
}

func (s *Server) getAggregateOnmetalMachine(ctx context.Context, id string) (*AggregateOnmetalMachine, error) {
	onmetalMachine, err := s.getOnmetalMachine(ctx, id)
	if err != nil {
		return nil, err
	}

	return s.aggregateOnmetalMachine(ctx, s.cluster.Client(), onmetalMachine)
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
