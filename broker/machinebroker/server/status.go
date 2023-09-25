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

	"k8s.io/apimachinery/pkg/api/resource"

	computev1alpha1 "github.com/onmetal/onmetal-api/api/compute/v1alpha1"
	corev1alpha1 "github.com/onmetal/onmetal-api/api/core/v1alpha1"
	ori "github.com/onmetal/onmetal-api/ori/apis/machine/v1alpha1"
	apierrors "k8s.io/apimachinery/pkg/api/errors"
	"k8s.io/apimachinery/pkg/util/sets"
	"sigs.k8s.io/controller-runtime/pkg/client"
)

func (s *Server) getTargetOnmetalMachinePools(ctx context.Context) ([]computev1alpha1.MachinePool, error) {
	if s.cluster.MachinePoolName() != "" {
		onmetalMachinePool := &computev1alpha1.MachinePool{}
		onmetalMachinePoolKey := client.ObjectKey{Name: s.cluster.MachinePoolName()}
		if err := s.cluster.Client().Get(ctx, onmetalMachinePoolKey, onmetalMachinePool); err != nil {
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

func (s *Server) gatherAvailableMachineClassNames(onmetalMachinePools []computev1alpha1.MachinePool) sets.Set[string] {
	res := sets.New[string]()
	for _, onmetalMachinePool := range onmetalMachinePools {
		for _, availableMachineClass := range onmetalMachinePool.Status.AvailableMachineClasses {
			res.Insert(availableMachineClass.Name)
		}
	}
	return res
}

func (s *Server) gatherMachineClassQuantity(onmetalMachinePools []computev1alpha1.MachinePool) map[string]*resource.Quantity {
	res := map[string]*resource.Quantity{}
	for _, onmetalMachinePool := range onmetalMachinePools {
		for resourceName, resourceQuantity := range onmetalMachinePool.Status.Capacity {
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

func (s *Server) filterOnmetalMachineClasses(
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

func (s *Server) convertOnmetalMachineClassStatus(machineClass *computev1alpha1.MachineClass, quantity *resource.Quantity) (*ori.MachineClassStatus, error) {
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

	log.V(1).Info("Getting target onmetal machine pools")
	onmetalMachinePools, err := s.getTargetOnmetalMachinePools(ctx)
	if err != nil {
		return nil, fmt.Errorf("error getting target onmetal machine pools: %w", err)
	}

	log.V(1).Info("Gathering available machine class names")
	availableOnmetalMachineClassNames := s.gatherAvailableMachineClassNames(onmetalMachinePools)

	if len(availableOnmetalMachineClassNames) == 0 {
		log.V(1).Info("No available machine classes")
		return &ori.StatusResponse{MachineClassStatus: []*ori.MachineClassStatus{}}, nil
	}

	log.V(1).Info("Gathering machine class quantity")
	machineClassQuantity := s.gatherMachineClassQuantity(onmetalMachinePools)

	log.V(1).Info("Listing onmetal machine classes")
	onmetalMachineClassList := &computev1alpha1.MachineClassList{}
	if err := s.cluster.Client().List(ctx, onmetalMachineClassList); err != nil {
		return nil, fmt.Errorf("error listing onmetal machine classes: %w", err)
	}

	availableOnmetalMachineClasses := s.filterOnmetalMachineClasses(availableOnmetalMachineClassNames, onmetalMachineClassList.Items)
	machineClassStatus := make([]*ori.MachineClassStatus, 0, len(availableOnmetalMachineClasses))
	for _, onmetalMachineClass := range availableOnmetalMachineClasses {
		quantity, ok := machineClassQuantity[string(corev1alpha1.ClassCountFor(corev1alpha1.ClassTypeMachineClass, onmetalMachineClass.Name))]
		if !ok {
			log.V(1).Info("Ignored class - missing quantity", "MachineClass", onmetalMachineClass.Name)
			continue
		}

		machineClass, err := s.convertOnmetalMachineClassStatus(&onmetalMachineClass, quantity)
		if err != nil {
			return nil, fmt.Errorf("error converting onmetal machine class %s: %w", onmetalMachineClass.Name, err)
		}

		machineClassStatus = append(machineClassStatus, machineClass)
	}

	log.V(1).Info("Returning machine classes")
	return &ori.StatusResponse{
		MachineClassStatus: machineClassStatus,
	}, nil
}
