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

	"k8s.io/apimachinery/pkg/api/resource"

	computev1alpha1 "github.com/ironcore-dev/ironcore/api/compute/v1alpha1"
	corev1alpha1 "github.com/ironcore-dev/ironcore/api/core/v1alpha1"
	ori "github.com/ironcore-dev/ironcore/ori/apis/machine/v1alpha1"
	apierrors "k8s.io/apimachinery/pkg/api/errors"
	"k8s.io/apimachinery/pkg/util/sets"
	"sigs.k8s.io/controller-runtime/pkg/client"
)

func (s *Server) getTargetIronCoreMachinePools(ctx context.Context) ([]computev1alpha1.MachinePool, error) {
	if s.cluster.MachinePoolName() != "" {
		ironcoreMachinePool := &computev1alpha1.MachinePool{}
		ironcoreMachinePoolKey := client.ObjectKey{Name: s.cluster.MachinePoolName()}
		if err := s.cluster.Client().Get(ctx, ironcoreMachinePoolKey, ironcoreMachinePool); err != nil {
			if !apierrors.IsNotFound(err) {
				return nil, fmt.Errorf("error getting machine pool %s: %w", s.cluster.MachinePoolName(), err)
			}
			return nil, nil
		}
	}

	machinePoolList := &computev1alpha1.MachinePoolList{}
	if err := s.cluster.Client().List(ctx, machinePoolList,
		client.MatchingLabels(s.cluster.MachinePoolSelector()),
	); err != nil {
		return nil, fmt.Errorf("error listing machine pools: %w", err)
	}
	return machinePoolList.Items, nil
}

func (s *Server) gatherAvailableMachineClassNames(ironcoreMachinePools []computev1alpha1.MachinePool) sets.Set[string] {
	res := sets.New[string]()
	for _, ironcoreMachinePool := range ironcoreMachinePools {
		for _, availableMachineClass := range ironcoreMachinePool.Status.AvailableMachineClasses {
			res.Insert(availableMachineClass.Name)
		}
	}
	return res
}

func (s *Server) gatherMachineClassQuantity(ironcoreMachinePools []computev1alpha1.MachinePool) map[string]*resource.Quantity {
	res := map[string]*resource.Quantity{}
	for _, ironcoreMachinePool := range ironcoreMachinePools {
		for resourceName, resourceQuantity := range ironcoreMachinePool.Status.Capacity {
			if corev1alpha1.IsClassCountResource(resourceName) {
				if _, ok := res[string(resourceName)]; !ok {
					res[string(resourceName)] = resource.NewQuantity(0, resource.DecimalSI)
				}
				res[string(resourceName)].Add(resourceQuantity)
			}
		}
	}
	return res
}

func (s *Server) filterIronCoreMachineClasses(
	availableMachineClassNames sets.Set[string],
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

func (s *Server) convertIronCoreMachineClassStatus(machineClass *computev1alpha1.MachineClass, quantity *resource.Quantity) (*ori.MachineClassStatus, error) {
	cpu := machineClass.Capabilities.CPU()
	memory := machineClass.Capabilities.Memory()

	return &ori.MachineClassStatus{
		MachineClass: &ori.MachineClass{
			Name: machineClass.Name,
			Capabilities: &ori.MachineClassCapabilities{
				CpuMillis:   cpu.MilliValue(),
				MemoryBytes: memory.Value(),
			},
		},
		Quantity: quantity.Value(),
	}, nil
}

func (s *Server) Status(ctx context.Context, req *ori.StatusRequest) (*ori.StatusResponse, error) {
	log := s.loggerFrom(ctx)

	log.V(1).Info("Getting target ironcore machine pools")
	ironcoreMachinePools, err := s.getTargetIronCoreMachinePools(ctx)
	if err != nil {
		return nil, fmt.Errorf("error getting target ironcore machine pools: %w", err)
	}

	log.V(1).Info("Gathering available machine class names")
	availableIronCoreMachineClassNames := s.gatherAvailableMachineClassNames(ironcoreMachinePools)

	if len(availableIronCoreMachineClassNames) == 0 {
		log.V(1).Info("No available machine classes")
		return &ori.StatusResponse{MachineClassStatus: []*ori.MachineClassStatus{}}, nil
	}

	log.V(1).Info("Gathering machine class quantity")
	machineClassQuantity := s.gatherMachineClassQuantity(ironcoreMachinePools)

	log.V(1).Info("Listing ironcore machine classes")
	ironcoreMachineClassList := &computev1alpha1.MachineClassList{}
	if err := s.cluster.Client().List(ctx, ironcoreMachineClassList); err != nil {
		return nil, fmt.Errorf("error listing ironcore machine classes: %w", err)
	}

	availableIronCoreMachineClasses := s.filterIronCoreMachineClasses(availableIronCoreMachineClassNames, ironcoreMachineClassList.Items)
	machineClassStatus := make([]*ori.MachineClassStatus, 0, len(availableIronCoreMachineClasses))
	for _, ironcoreMachineClass := range availableIronCoreMachineClasses {
		quantity, ok := machineClassQuantity[string(corev1alpha1.ClassCountFor(corev1alpha1.ClassTypeMachineClass, ironcoreMachineClass.Name))]
		if !ok {
			log.V(1).Info("Ignored class - missing quantity", "MachineClass", ironcoreMachineClass.Name)
			continue
		}

		machineClass, err := s.convertIronCoreMachineClassStatus(&ironcoreMachineClass, quantity)
		if err != nil {
			return nil, fmt.Errorf("error converting ironcore machine class %s: %w", ironcoreMachineClass.Name, err)
		}

		machineClassStatus = append(machineClassStatus, machineClass)
	}

	log.V(1).Info("Returning machine classes")
	return &ori.StatusResponse{
		MachineClassStatus: machineClassStatus,
	}, nil
}
