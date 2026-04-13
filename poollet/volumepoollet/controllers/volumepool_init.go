// SPDX-FileCopyrightText: 2023 SAP SE or an SAP affiliate company and IronCore contributors
// SPDX-License-Identifier: Apache-2.0

package controllers

import (
	"context"
	"fmt"

	commonv1alpha1 "github.com/ironcore-dev/ironcore/api/common/v1alpha1"
	storagev1alpha1apply "github.com/ironcore-dev/ironcore/client-go/applyconfigurations/storage/v1alpha1"
	poolletutils "github.com/ironcore-dev/ironcore/poollet/common/utils"
	volumepoolletv1alpha1 "github.com/ironcore-dev/ironcore/poollet/volumepoollet/api/v1alpha1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/client"
)

type VolumePoolInit struct {
	client.Client

	VolumePoolName string
	ProviderID     string

	TopologyLabels map[commonv1alpha1.TopologyLabel]string

	OnInitialized func(ctx context.Context) error
	OnFailed      func(ctx context.Context, reason error) error
}

//+kubebuilder:rbac:groups=storage.ironcore.dev,resources=volumepools,verbs=get;list;create;update;patch;apply;delete

func (i *VolumePoolInit) Start(ctx context.Context) error {
	log := ctrl.LoggerFrom(ctx).WithName("volumepool").WithName("init")

	log.V(1).Info("Applying volume pool")
	volumePoolApply := storagev1alpha1apply.VolumePool(i.VolumePoolName).
		WithSpec(storagev1alpha1apply.VolumePoolSpec().
			WithProviderID(i.ProviderID))

	log.V(1).Info("Initially setting topology labels")
	om := &metav1.ObjectMeta{}
	poolletutils.SetTopologyLabels(log, om, i.TopologyLabels)
	if len(om.Labels) > 0 {
		volumePoolApply.WithLabels(om.Labels)
	}

	if err := i.Apply(ctx, volumePoolApply, client.ForceOwnership, client.FieldOwner(volumepoolletv1alpha1.FieldOwner)); err != nil {
		if i.OnFailed != nil {
			log.V(1).Info("Failed applying, calling OnFailed callback", "Error", err)
			return i.OnFailed(ctx, err)
		}
		return fmt.Errorf("error applying volume pool: %w", err)
	}

	log.V(1).Info("Successfully applied volume pool")
	if i.OnInitialized != nil {
		log.V(1).Info("Calling OnInitialized callback")
		return i.OnInitialized(ctx)
	}
	return nil
}

func (i *VolumePoolInit) SetupWithManager(mgr ctrl.Manager) error {
	return mgr.Add(i)
}
