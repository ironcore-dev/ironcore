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
	corev1alpha1 "github.com/onmetal/onmetal-api/api/core/v1alpha1"
	ori "github.com/onmetal/onmetal-api/ori/apis/machine/v1alpha1"
	"k8s.io/apimachinery/pkg/api/resource"
)

func (s *Server) PoolInfo(ctx context.Context, req *ori.PoolInfoRequest) (*ori.PoolInfoResponse, error) {
	log := s.loggerFrom(ctx)

	log.V(1).Info("Listing onmetal machine pools")
	onmetalMachinePoolList := &computev1alpha1.MachinePoolList{}
	if err := s.cluster.Client().List(ctx, onmetalMachinePoolList); err != nil {
		return nil, fmt.Errorf("error listing onmetal machine pools: %w", err)
	}

	var (
		sharedCPU, staticCPU       int64
		sharedMemory, staticMemory uint64
	)
	for _, onmetalMachinePool := range onmetalMachinePoolList.Items {
		staticCPU += onmetalMachinePool.Status.Capacity.Name(corev1alpha1.ResourceCPU, resource.DecimalSI).AsDec().UnscaledBig().Int64()
		sharedCPU += onmetalMachinePool.Status.Capacity.Name(corev1alpha1.SharedResourceCPU, resource.DecimalSI).AsDec().UnscaledBig().Int64()

		staticMemory += onmetalMachinePool.Status.Capacity.Name(corev1alpha1.ResourceMemory, resource.BinarySI).AsDec().UnscaledBig().Uint64()
		sharedMemory += onmetalMachinePool.Status.Capacity.Name(corev1alpha1.SharedResourceMemory, resource.BinarySI).AsDec().UnscaledBig().Uint64()
	}

	return &ori.PoolInfoResponse{
		SharedCpu:    sharedCPU,
		StaticCpu:    staticCPU,
		SharedMemory: sharedMemory,
		StaticMemory: staticMemory,
	}, nil
}
