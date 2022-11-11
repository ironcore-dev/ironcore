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

package controllers

import (
	"context"
	"fmt"

	computev1alpha1 "github.com/onmetal/onmetal-api/apis/compute/v1alpha1"
	machinepoolletv1alpha1 "github.com/onmetal/onmetal-api/machinepoollet/api/v1alpha1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/client"
)

type MachinePoolInit struct {
	client.Client

	MachinePoolName string
	ProviderID      string

	OnInitialized func(ctx context.Context) error
	OnFailed      func(ctx context.Context, reason error) error
}

func (i *MachinePoolInit) Start(ctx context.Context) error {
	log := ctrl.LoggerFrom(ctx).WithName("machinepool").WithName("init")

	log.V(1).Info("Applying machine pool")
	machinePool := &computev1alpha1.MachinePool{
		TypeMeta: metav1.TypeMeta{
			APIVersion: computev1alpha1.SchemeGroupVersion.String(),
			Kind:       "MachinePool",
		},
		ObjectMeta: metav1.ObjectMeta{
			Name: i.MachinePoolName,
		},
		Spec: computev1alpha1.MachinePoolSpec{
			ProviderID: i.ProviderID,
		},
	}
	if err := i.Patch(ctx, machinePool, client.Apply, client.ForceOwnership, client.FieldOwner(machinepoolletv1alpha1.FieldOwner)); err != nil {
		if i.OnFailed != nil {
			log.V(1).Info("Failed applying, calling OnFailed callback", "Error", err)
			return i.OnFailed(ctx, err)
		}
		return fmt.Errorf("error applying machine pool: %w", err)
	}

	log.V(1).Info("Successfully applied machine pool")
	if i.OnInitialized != nil {
		log.V(1).Info("Calling OnInitialized callback")
		return i.OnInitialized(ctx)
	}
	return nil
}
