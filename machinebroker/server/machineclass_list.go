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

	"github.com/onmetal/controller-utils/set"
	computev1alpha1 "github.com/onmetal/onmetal-api/api/compute/v1alpha1"
	ori "github.com/onmetal/onmetal-api/ori/apis/compute/v1alpha1"
	apierrors "k8s.io/apimachinery/pkg/api/errors"
	"sigs.k8s.io/controller-runtime/pkg/client"
)

func (s *Server) getTargetOnmetalMachinePools(ctx context.Context) ([]computev1alpha1.MachinePool, error) {
	if s.machinePoolName != "" {
		onmetalMachinePool := &computev1alpha1.MachinePool{}
		onmetalMachinePoolKey := client.ObjectKey{Name: s.machinePoolName}
		if err := s.client.Get(ctx, onmetalMachinePoolKey, onmetalMachinePool); err != nil {
			if !apierrors.IsNotFound(err) {
				return nil, fmt.Errorf("error getting machine pool %s: %w", s.machinePoolName, err)
			}
			return nil, nil
		}
	}

	machinePoolList := &computev1alpha1.MachinePoolList{}
	if err := s.client.List(ctx, machinePoolList,
		client.MatchingLabels(s.machinePoolSelector),
	); err != nil {
		return nil, fmt.Errorf("error listing machine pools: %w", err)
	}
	return machinePoolList.Items, nil
}

func (s *Server) gatherAvailableMachineClassNames(onmetalMachinePools []computev1alpha1.MachinePool) set.Set[string] {
	res := set.New[string]()
	for _, onmetalMachinePool := range onmetalMachinePools {
		for _, availableMachineClass := range onmetalMachinePool.Status.AvailableMachineClasses {
			res.Insert(availableMachineClass.Name)
		}
	}
	return res
}

func (s *Server) filterOnmetalMachineClasses(
	availableMachineClassNames set.Set[string],
	machineClasses []computev1alpha1.MachineClass,
) []computev1alpha1.MachineClass {
	var filtered []computev1alpha1.MachineClass
	for _, machineClass := range machineClasses {
		if !availableMachineClassNames.Has(machineClass.Name) {
			continue
		}

		filtered = append(filtered, machineClass)
	}
	return filtered
}

func (s *Server) convertOnmetalMachineClass(machineClass *computev1alpha1.MachineClass) (*ori.MachineClass, error) {
	cpu := machineClass.Capabilities.Cpu()
	memory := machineClass.Capabilities.Memory()

	return &ori.MachineClass{
		Name: machineClass.Name,
		Capabilities: &ori.MachineClassCapabilities{
			CpuMillis:   cpu.MilliValue(),
			MemoryBytes: uint64(memory.Value()),
		},
	}, nil
}

func (s *Server) ListMachineClasses(ctx context.Context, req *ori.ListMachineClassesRequest) (*ori.ListMachineClassesResponse, error) {
	log := s.loggerFrom(ctx)

	log.V(1).Info("Getting target onmetal machine pools")
	onmetalMachinePools, err := s.getTargetOnmetalMachinePools(ctx)
	if err != nil {
		return nil, fmt.Errorf("error getting target onmetal machine pools: %w", err)
	}

	log.V(1).Info("Gathering available machine class names")
	availableOnmetalMachineClassNames := s.gatherAvailableMachineClassNames(onmetalMachinePools)

	if len(availableOnmetalMachineClassNames) == 0 {
		log.V(1).Info("No available machine classes")
		return &ori.ListMachineClassesResponse{MachineClasses: []*ori.MachineClass{}}, nil
	}

	log.V(1).Info("Listing onmetal machine classes")
	onmetalMachineClassList := &computev1alpha1.MachineClassList{}
	if err := s.client.List(ctx, onmetalMachineClassList); err != nil {
		return nil, fmt.Errorf("error listing onmetal machine classes: %w", err)
	}

	availableOnmetalMachineClasses := s.filterOnmetalMachineClasses(availableOnmetalMachineClassNames, onmetalMachineClassList.Items)
	machineClasses := make([]*ori.MachineClass, 0, len(availableOnmetalMachineClasses))
	for _, onmetalMachineClass := range availableOnmetalMachineClasses {
		machineClass, err := s.convertOnmetalMachineClass(&onmetalMachineClass)
		if err != nil {
			return nil, fmt.Errorf("error converting onmetal machine class %s: %w", onmetalMachineClass.Name, err)
		}

		machineClasses = append(machineClasses, machineClass)
	}

	log.V(1).Info("Returning machine classes")
	return &ori.ListMachineClassesResponse{
		MachineClasses: machineClasses,
	}, nil
}
