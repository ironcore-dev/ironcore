// Copyright 2021 OnMetal authors
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

package storage

import (
	"context"
	"fmt"
	"math/rand"
	"sigs.k8s.io/controller-runtime/pkg/reconcile"

	"github.com/go-logr/logr"
	corev1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/labels"
	"k8s.io/apimachinery/pkg/util/sets"
	"k8s.io/client-go/tools/record"
	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/builder"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/handler"
	"sigs.k8s.io/controller-runtime/pkg/manager"
	"sigs.k8s.io/controller-runtime/pkg/predicate"
	"sigs.k8s.io/controller-runtime/pkg/source"

	"github.com/onmetal/onmetal-api/apis/common/v1alpha1"
	storagev1alpha1 "github.com/onmetal/onmetal-api/apis/storage/v1alpha1"
)

const (
	storagePoolStatusAvailableStorageClassesNameField = ".status.availableStorageClasses[*].name"
	volumeSpecStoragePoolNameField                    = ".spec.storagePool.name"
)

type VolumeScheduler struct {
	client.Client
	Events record.EventRecorder
}

//+kubebuilder:rbac:groups=storage.onmetal.de,resources=volumes,verbs=get;list;watch;update;patch
//+kubebuilder:rbac:groups=storage.onmetal.de,resources=volumes/status,verbs=get;update;patch
//+kubebuilder:rbac:groups=storage.onmetal.de,resources=storagepools,verbs=get;list;watch

// Reconcile reconciles the desired with the actual state.
func (s *VolumeScheduler) Reconcile(ctx context.Context, req ctrl.Request) (ctrl.Result, error) {
	log := ctrl.LoggerFrom(ctx)

	volume := &storagev1alpha1.Volume{}
	if err := s.Get(ctx, req.NamespacedName, volume); err != nil {
		return ctrl.Result{}, client.IgnoreNotFound(err)
	}

	if !volume.DeletionTimestamp.IsZero() {
		log.Info("Volume is already deleting")
		return ctrl.Result{}, nil
	}
	if volume.Spec.StoragePool.Name != "" {
		log.Info("Volume is already assigned")
		return ctrl.Result{}, nil
	}
	return s.schedule(ctx, log, volume)
}

func (s *VolumeScheduler) schedule(ctx context.Context, log logr.Logger, volume *storagev1alpha1.Volume) (ctrl.Result, error) {
	log.Info("Scheduling volume")
	if volume.Status.State != storagev1alpha1.VolumeStatePending {
		base := volume.DeepCopy()
		volume.Status.State = storagev1alpha1.VolumeStatePending
		if err := s.Status().Patch(ctx, volume, client.MergeFrom(base)); err != nil {
			return ctrl.Result{}, fmt.Errorf("error patching volume state to pending: %w", err)
		}
		return ctrl.Result{Requeue: true}, nil
	}

	list := &storagev1alpha1.StoragePoolList{}
	if err := s.List(ctx, list,
		client.MatchingFields{storagePoolStatusAvailableStorageClassesNameField: volume.Spec.StorageClassRef.Name},
		client.MatchingLabels(volume.Spec.StoragePoolSelector),
	); err != nil {
		return ctrl.Result{}, fmt.Errorf("error listing storage pools: %w", err)
	}

	var available []storagev1alpha1.StoragePool
	for _, storagePool := range list.Items {
		if storagePool.DeletionTimestamp.IsZero() {
			available = append(available, storagePool)
		}
	}
	if len(available) == 0 {
		log.Info("No storage pool available for storage class", "StorageClass", volume.Spec.StorageClassRef.Name)
		s.Events.Eventf(volume, corev1.EventTypeNormal, "CannotSchedule", "No StoragePool found for StorageClass %s", volume.Spec.StorageClassRef.Name)
		return ctrl.Result{}, nil
	}

	// Filter storage pools by checking if the volume tolerates all the taints of a storage pool
	var filtered []storagev1alpha1.StoragePool
	for _, pool := range available {
		if v1alpha1.TolerateTaints(volume.Spec.Tolerations, pool.Spec.Taints) {
			filtered = append(filtered, pool)
		}
	}
	if len(filtered) == 0 {
		log.Info("No storage pool tolerated by the volume", "Tolerations", volume.Spec.Tolerations)
		s.Events.Eventf(volume, corev1.EventTypeNormal, "CannotSchedule", "No StoragePool tolerated by %s", &volume.Spec.Tolerations)
		return ctrl.Result{}, nil
	}
	available = filtered

	// Get a random pool to distribute evenly.
	// TODO: Instead of random distribution, try to come up w/ metrics that include usage of each pool to
	// avoid unfortunate random distribution of items.
	pool := available[rand.Intn(len(available))]
	log = log.WithValues("StoragePool", pool.Name)
	base := volume.DeepCopy()
	volume.Spec.StoragePool.Name = pool.Name
	log.Info("Patching volume")
	if err := s.Patch(ctx, volume, client.MergeFrom(base)); err != nil {
		return ctrl.Result{}, fmt.Errorf("error scheduling volume on pool: %w", err)
	}

	log.Info("Successfully assigned volume")
	return ctrl.Result{}, nil
}

func (s *VolumeScheduler) SetupWithManager(mgr manager.Manager) error {
	ctx := context.Background()
	log := ctrl.Log.WithName("volume-scheduler").WithName("setup")
	ctx = ctrl.LoggerInto(ctx, log)

	if err := mgr.GetFieldIndexer().IndexField(ctx, &storagev1alpha1.StoragePool{}, storagePoolStatusAvailableStorageClassesNameField, func(object client.Object) []string {
		pool := object.(*storagev1alpha1.StoragePool)
		names := make([]string, 0, len(pool.Status.AvailableStorageClasses))
		for _, availableStorageClass := range pool.Status.AvailableStorageClasses {
			names = append(names, availableStorageClass.Name)
		}
		return names
	}); err != nil {
		return fmt.Errorf("could not setup field indexer for %s: %w", storagePoolStatusAvailableStorageClassesNameField, err)
	}

	if err := mgr.GetFieldIndexer().IndexField(ctx, &storagev1alpha1.Volume{}, volumeSpecStoragePoolNameField, func(object client.Object) []string {
		volume := object.(*storagev1alpha1.Volume)
		return []string{volume.Spec.StoragePool.Name}
	}); err != nil {
		return fmt.Errorf("could not setup field indexer for %s: %w", volumeSpecStoragePoolNameField, err)
	}

	return ctrl.NewControllerManagedBy(mgr).
		Named("volume-scheduler").
		// Enqueue unscheduled volumes.
		For(&storagev1alpha1.Volume{},
			builder.WithPredicates(
				predicate.NewPredicateFuncs(func(object client.Object) bool {
					volume := object.(*storagev1alpha1.Volume)
					return volume.DeletionTimestamp.IsZero() && volume.Spec.StoragePool.Name == ""
				}),
			),
		).
		// Enqueue unscheduled volumes if a storage pool w/ required storage classes becomes available.
		Watches(&source.Kind{Type: &storagev1alpha1.StoragePool{}},
			handler.EnqueueRequestsFromMapFunc(func(object client.Object) []reconcile.Request {
				pool := object.(*storagev1alpha1.StoragePool)
				if !pool.DeletionTimestamp.IsZero() {
					return nil
				}

				list := &storagev1alpha1.VolumeList{}
				if err := s.List(ctx, list, client.MatchingFields{volumeSpecStoragePoolNameField: ""}); err != nil {
					log.Error(err, "error listing unscheduled volumes")
					return nil
				}

				availableClassNames := sets.NewString()
				for _, availableStorageClass := range pool.Status.AvailableStorageClasses {
					availableClassNames.Insert(availableStorageClass.Name)
				}

				var requests []reconcile.Request
				for _, volume := range list.Items {
					storagePoolSelector := labels.SelectorFromSet(volume.Spec.StoragePoolSelector)
					if availableClassNames.Has(volume.Spec.StorageClassRef.Name) && storagePoolSelector.Matches(labels.Set(pool.Labels)) {
						requests = append(requests, ctrl.Request{NamespacedName: client.ObjectKeyFromObject(&volume)})
					}
				}
				return requests
			}),
		).
		Complete(s)
}
