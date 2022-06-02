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
	volumePoolStatusAvailableVolumeClassesNameField = ".status.availableVolumeClasses[*].name"
	volumeSpecVolumePoolNameField                   = ".spec.volumePool.name"
)

type VolumeScheduler struct {
	client.Client
	Events record.EventRecorder
}

//+kubebuilder:rbac:groups=storage.api.onmetal.de,resources=volumes,verbs=get;list;watch;update;patch
//+kubebuilder:rbac:groups=storage.api.onmetal.de,resources=volumes/status,verbs=get;update;patch
//+kubebuilder:rbac:groups=storage.api.onmetal.de,resources=volumepools,verbs=get;list;watch

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
	if volume.Spec.VolumePoolRef != nil {
		log.Info("Volume is already assigned", "VolumePoolRef", volume.Spec.VolumePoolRef)
		return ctrl.Result{}, nil
	}
	return s.schedule(ctx, log, volume)
}

func (s *VolumeScheduler) schedule(ctx context.Context, log logr.Logger, volume *storagev1alpha1.Volume) (ctrl.Result, error) {
	log.Info("Scheduling volume")
	list := &storagev1alpha1.VolumePoolList{}
	if err := s.List(ctx, list,
		client.MatchingFields{volumePoolStatusAvailableVolumeClassesNameField: volume.Spec.VolumeClassRef.Name},
		client.MatchingLabels(volume.Spec.VolumePoolSelector),
	); err != nil {
		return ctrl.Result{}, fmt.Errorf("error listing volume pools: %w", err)
	}

	var available []storagev1alpha1.VolumePool
	for _, volumePool := range list.Items {
		if volumePool.DeletionTimestamp.IsZero() {
			available = append(available, volumePool)
		}
	}
	if len(available) == 0 {
		log.Info("No volume pool available for volume class", "VolumeClass", volume.Spec.VolumeClassRef.Name)
		s.Events.Eventf(volume, corev1.EventTypeNormal, "CannotSchedule", "No VolumePoolRef found for VolumeClass %s", volume.Spec.VolumeClassRef.Name)
		return ctrl.Result{}, nil
	}

	// Filter volume pools by checking if the volume tolerates all the taints of a volume pool
	var filtered []storagev1alpha1.VolumePool
	for _, pool := range available {
		if v1alpha1.TolerateTaints(volume.Spec.Tolerations, pool.Spec.Taints) {
			filtered = append(filtered, pool)
		}
	}
	if len(filtered) == 0 {
		log.Info("No volume pool tolerated by the volume", "Tolerations", volume.Spec.Tolerations)
		s.Events.Eventf(volume, corev1.EventTypeNormal, "CannotSchedule", "No VolumePoolRef tolerated by %s", &volume.Spec.Tolerations)
		return ctrl.Result{}, nil
	}
	available = filtered

	// Get a random pool to distribute evenly.
	// TODO: Instead of random distribution, try to come up w/ metrics that include usage of each pool to
	// avoid unfortunate random distribution of items.
	pool := available[rand.Intn(len(available))]
	log = log.WithValues("VolumePoolRef", pool.Name)
	base := volume.DeepCopy()
	volume.Spec.VolumePoolRef = &corev1.LocalObjectReference{Name: pool.Name}
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

	if err := mgr.GetFieldIndexer().IndexField(ctx, &storagev1alpha1.VolumePool{}, volumePoolStatusAvailableVolumeClassesNameField, func(obj client.Object) []string {
		pool := obj.(*storagev1alpha1.VolumePool)
		names := make([]string, 0, len(pool.Status.AvailableVolumeClasses))
		for _, availableVolumeClass := range pool.Status.AvailableVolumeClasses {
			names = append(names, availableVolumeClass.Name)
		}
		return names
	}); err != nil {
		return err
	}

	if err := mgr.GetFieldIndexer().IndexField(ctx, &storagev1alpha1.Volume{}, volumeSpecVolumePoolNameField, func(obj client.Object) []string {
		volume := obj.(*storagev1alpha1.Volume)
		volumePoolRef := volume.Spec.VolumePoolRef
		if volumePoolRef == nil {
			return []string{""}
		}
		return []string{volumePoolRef.Name}
	}); err != nil {
		return err
	}

	return ctrl.NewControllerManagedBy(mgr).
		Named("volume-scheduler").
		// Enqueue unscheduled volumes.
		For(&storagev1alpha1.Volume{},
			builder.WithPredicates(
				predicate.NewPredicateFuncs(func(object client.Object) bool {
					volume := object.(*storagev1alpha1.Volume)
					return volume.DeletionTimestamp.IsZero() && volume.Spec.VolumePoolRef == nil
				}),
			),
		).
		// Enqueue unscheduled volumes if a volume pool w/ required volume classes becomes available.
		Watches(&source.Kind{Type: &storagev1alpha1.VolumePool{}},
			handler.EnqueueRequestsFromMapFunc(func(object client.Object) []ctrl.Request {
				pool := object.(*storagev1alpha1.VolumePool)
				if !pool.DeletionTimestamp.IsZero() {
					return nil
				}

				list := &storagev1alpha1.VolumeList{}
				if err := s.List(ctx, list, client.MatchingFields{volumeSpecVolumePoolNameField: ""}); err != nil {
					log.Error(err, "error listing unscheduled volumes")
					return nil
				}

				availableClassNames := sets.NewString()
				for _, availableVolumeClass := range pool.Status.AvailableVolumeClasses {
					availableClassNames.Insert(availableVolumeClass.Name)
				}

				var requests []ctrl.Request
				for _, volume := range list.Items {
					volumePoolSelector := labels.SelectorFromSet(volume.Spec.VolumePoolSelector)
					if availableClassNames.Has(volume.Spec.VolumeClassRef.Name) && volumePoolSelector.Matches(labels.Set(pool.Labels)) {
						requests = append(requests, ctrl.Request{NamespacedName: client.ObjectKeyFromObject(&volume)})
					}
				}
				return requests
			}),
		).
		Complete(s)
}
