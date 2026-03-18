// SPDX-FileCopyrightText: 2023 SAP SE or an SAP affiliate company and IronCore contributors
// SPDX-License-Identifier: Apache-2.0

package controllers

import (
	"context"
	"fmt"

	commonv1alpha1 "github.com/ironcore-dev/ironcore/api/common/v1alpha1"
	computev1alpha1apply "github.com/ironcore-dev/ironcore/client-go/applyconfigurations/compute/v1alpha1"
	poolletutils "github.com/ironcore-dev/ironcore/poollet/common/utils"
	machinepoolletv1alpha1 "github.com/ironcore-dev/ironcore/poollet/machinepoollet/api/v1alpha1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/client"
)

type MachinePoolInit struct {
	client.Client

	MachinePoolName string
	ProviderID      string

	TopologyLabels map[commonv1alpha1.TopologyLabel]string

	// TODO: Remove OnInitialized / OnFailed as soon as the controller-runtime provides support for pre-start hooks:
	// https://github.com/kubernetes-sigs/controller-runtime/pull/2044

	OnInitialized func(ctx context.Context) error
	OnFailed      func(ctx context.Context, reason error) error
}

//+kubebuilder:rbac:groups=compute.ironcore.dev,resources=machinepools,verbs=get;list;create;update;patch; apply;delete

func (i *MachinePoolInit) Start(ctx context.Context) error {
	log := ctrl.LoggerFrom(ctx).WithName("machinepool").WithName("init")

	log.V(1).Info("Applying machine pool")
	machinePoolApply := computev1alpha1apply.MachinePool(i.MachinePoolName).
		WithSpec(computev1alpha1apply.MachinePoolSpec().
			WithProviderID(i.ProviderID))

	log.V(1).Info("Initially setting topology labels")
	om := &metav1.ObjectMeta{}
	poolletutils.SetTopologyLabels(log, om, i.TopologyLabels)
	if len(om.Labels) > 0 {
		machinePoolApply.WithLabels(om.Labels)
	}
	if err := i.Apply(ctx, machinePoolApply, client.ForceOwnership, client.FieldOwner(machinepoolletv1alpha1.FieldOwner)); err != nil {
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

func (i *MachinePoolInit) SetupWithManager(mgr ctrl.Manager) error {
	return mgr.Add(i)
}
