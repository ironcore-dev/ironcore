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

	storagev1alpha1 "github.com/onmetal/onmetal-api/api/storage/v1alpha1"
	volumepoolletv1alpha1 "github.com/onmetal/onmetal-api/volumepoollet/api/v1alpha1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/client"
)

type VolumePoolInit struct {
	client.Client

	VolumePoolName string
	ProviderID     string

	OnInitialized func(ctx context.Context) error
	OnFailed      func(ctx context.Context, reason error) error
}

func (i *VolumePoolInit) Start(ctx context.Context) error {
	log := ctrl.LoggerFrom(ctx).WithName("volumepool").WithName("init")

	log.V(1).Info("Applying volume pool")
	volumePool := &storagev1alpha1.VolumePool{
		TypeMeta: metav1.TypeMeta{
			APIVersion: storagev1alpha1.SchemeGroupVersion.String(),
			Kind:       "VolumePool",
		},
		ObjectMeta: metav1.ObjectMeta{
			Name: i.VolumePoolName,
		},
		Spec: storagev1alpha1.VolumePoolSpec{
			ProviderID: i.ProviderID,
		},
	}
	if err := i.Patch(ctx, volumePool, client.Apply, client.ForceOwnership, client.FieldOwner(volumepoolletv1alpha1.FieldOwner)); err != nil {
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
