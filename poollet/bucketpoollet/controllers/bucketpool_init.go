// SPDX-FileCopyrightText: 2023 SAP SE or an SAP affiliate company and IronCore contributors
// SPDX-License-Identifier: Apache-2.0

package controllers

import (
	"context"
	"fmt"

	commonv1alpha1 "github.com/ironcore-dev/ironcore/api/common/v1alpha1"

	storagev1alpha1apply "github.com/ironcore-dev/ironcore/client-go/applyconfigurations/storage/v1alpha1"
	bucketpoolletv1alpha1 "github.com/ironcore-dev/ironcore/poollet/bucketpoollet/api/v1alpha1"
	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/client"
)

type BucketPoolInit struct {
	client.Client

	BucketPoolName string
	ProviderID     string

	TopologyLabels map[commonv1alpha1.TopologyLabel]string

	OnInitialized func(ctx context.Context) error
	OnFailed      func(ctx context.Context, reason error) error
}

//+kubebuilder:rbac:groups=storage.ironcore.dev,resources=bucketpools,verbs=get;list;create;update;patch;apply;delete

func (i *BucketPoolInit) Start(ctx context.Context) error {
	log := ctrl.LoggerFrom(ctx).WithName("bucketpool").WithName("init")

	log.V(1).Info("Applying bucket pool")
	bucketPoolApply := storagev1alpha1apply.BucketPool(i.BucketPoolName).
		WithKind("BucketPool").
		WithAPIVersion("storage.ironcore.dev/v1alpha1").
		WithSpec(storagev1alpha1apply.BucketPoolSpec().
			WithProviderID(i.ProviderID))

	log.V(1).Info("Initially setting topology labels")
	// Convert topology labels to string map and set them
	labels := make(map[string]string)
	for key, val := range i.TopologyLabels {
		log.V(1).Info("Setting topology label", "Label", key, "Value", val)
		labels[string(key)] = val
	}
	if len(labels) > 0 {
		bucketPoolApply.WithLabels(labels)
	}

	if err := i.Apply(ctx, bucketPoolApply, client.ForceOwnership, client.FieldOwner(bucketpoolletv1alpha1.FieldOwner)); err != nil {
		if i.OnFailed != nil {
			log.V(1).Info("Failed applying, calling OnFailed callback", "Error", err)
			return i.OnFailed(ctx, err)
		}
		return fmt.Errorf("error applying bucket pool: %w", err)
	}

	log.V(1).Info("Successfully applied bucket pool")
	if i.OnInitialized != nil {
		log.V(1).Info("Calling OnInitialized callback")
		return i.OnInitialized(ctx)
	}
	return nil
}

func (i *BucketPoolInit) SetupWithManager(mgr ctrl.Manager) error {
	return mgr.Add(i)
}
