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
	bucketpoolletv1alpha1 "github.com/onmetal/onmetal-api/poollet/bucketpoollet/api/v1alpha1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/client"
)

type BucketPoolInit struct {
	client.Client

	BucketPoolName string
	ProviderID     string

	OnInitialized func(ctx context.Context) error
	OnFailed      func(ctx context.Context, reason error) error
}

//+kubebuilder:rbac:groups=storage.api.onmetal.de,resources=bucketpools,verbs=get;list;create;update;patch

func (i *BucketPoolInit) Start(ctx context.Context) error {
	log := ctrl.LoggerFrom(ctx).WithName("bucketpool").WithName("init")

	log.V(1).Info("Applying bucket pool")
	bucketPool := &storagev1alpha1.BucketPool{
		TypeMeta: metav1.TypeMeta{
			APIVersion: storagev1alpha1.SchemeGroupVersion.String(),
			Kind:       "BucketPool",
		},
		ObjectMeta: metav1.ObjectMeta{
			Name: i.BucketPoolName,
		},
		Spec: storagev1alpha1.BucketPoolSpec{
			ProviderID: i.ProviderID,
		},
	}
	if err := i.Patch(ctx, bucketPool, client.Apply, client.ForceOwnership, client.FieldOwner(bucketpoolletv1alpha1.FieldOwner)); err != nil {
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
