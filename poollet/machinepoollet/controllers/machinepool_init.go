// SPDX-FileCopyrightText: 2023 SAP SE or an SAP affiliate company and IronCore contributors
// SPDX-License-Identifier: Apache-2.0

package controllers

import (
	"context"
	"fmt"

	"github.com/go-logr/logr"
	commonv1alpha1 "github.com/ironcore-dev/ironcore/api/common/v1alpha1"
	computev1alpha1 "github.com/ironcore-dev/ironcore/api/compute/v1alpha1"
	machinepoolletv1alpha1 "github.com/ironcore-dev/ironcore/poollet/machinepoollet/api/v1alpha1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/client"
)

type MachinePoolInit struct {
	client.Client

	MachinePoolName string
	ProviderID      string

	TopologyRegionLabel string
	TopologyZoneLabel   string
	TopologyLabels      map[commonv1alpha1.TopologyLabel]string

	// TODO: Remove OnInitialized / OnFailed as soon as the controller-runtime provides support for pre-start hooks:
	// https://github.com/kubernetes-sigs/controller-runtime/pull/2044

	OnInitialized func(ctx context.Context) error
	OnFailed      func(ctx context.Context, reason error) error
}

//+kubebuilder:rbac:groups=compute.ironcore.dev,resources=machinepools,verbs=get;list;create;update;patch

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

	i.setTopologyLabels(log, machinePool)

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

func (i *MachinePoolInit) setTopologyLabels(log logr.Logger, machinePool *computev1alpha1.MachinePool) {
	log.V(1).Info("Initially setting topology labels")
	for key, val := range i.TopologyLabels {
		if machinePool.ObjectMeta.Labels == nil {
			machinePool.ObjectMeta.Labels = make(map[string]string)
		}
		log.V(1).Info("Setting topology label", "Label", key, "Value", val)
		machinePool.ObjectMeta.Labels[string(key)] = val
	}
}

func (i *MachinePoolInit) SetupWithManager(mgr ctrl.Manager) error {
	return mgr.Add(i)
}
