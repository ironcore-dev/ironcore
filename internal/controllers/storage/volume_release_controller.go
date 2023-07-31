// Copyright 2023 OnMetal authors
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

	"github.com/go-logr/logr"
	computev1alpha1 "github.com/onmetal/onmetal-api/api/compute/v1alpha1"
	storagev1alpha1 "github.com/onmetal/onmetal-api/api/storage/v1alpha1"
	apierrors "k8s.io/apimachinery/pkg/api/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/utils/lru"
	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/builder"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/handler"
	"sigs.k8s.io/controller-runtime/pkg/predicate"
)

type VolumeReleaseReconciler struct {
	client.Client
	APIReader client.Reader

	AbsenceCache *lru.Cache
}

//+kubebuilder:rbac:groups=storage.api.onmetal.de,resources=volumes,verbs=get;list;watch;update;patch
//+kubebuilder:rbac:groups=compute.api.onmetal.de,resources=machines,verbs=get;list;watch

func (r *VolumeReleaseReconciler) Reconcile(ctx context.Context, req ctrl.Request) (ctrl.Result, error) {
	log := ctrl.LoggerFrom(ctx)
	volume := &storagev1alpha1.Volume{}
	if err := r.Get(ctx, req.NamespacedName, volume); err != nil {
		return ctrl.Result{}, client.IgnoreNotFound(err)
	}

	return r.reconcileExists(ctx, log, volume)
}

func (r *VolumeReleaseReconciler) reconcileExists(ctx context.Context, log logr.Logger, volume *storagev1alpha1.Volume) (ctrl.Result, error) {
	if !volume.DeletionTimestamp.IsZero() {
		log.V(1).Info("Volume is already deleting, nothing to do")
		return ctrl.Result{}, nil
	}

	return r.reconcile(ctx, log, volume)
}

func (r *VolumeReleaseReconciler) volumeClaimExists(ctx context.Context, volume *storagev1alpha1.Volume) (bool, error) {
	claimRef := volume.Spec.ClaimRef
	if _, ok := r.AbsenceCache.Get(claimRef.UID); ok {
		return false, nil
	}

	claimer := &metav1.PartialObjectMetadata{
		TypeMeta: metav1.TypeMeta{
			APIVersion: computev1alpha1.SchemeGroupVersion.String(),
			Kind:       "Machine",
		},
	}
	claimerKey := client.ObjectKey{Namespace: volume.Namespace, Name: claimRef.Name}
	if err := r.APIReader.Get(ctx, claimerKey, claimer); err != nil {
		if !apierrors.IsNotFound(err) {
			return false, fmt.Errorf("error getting claiming machine %s: %w", claimRef.Name, err)
		}

		r.AbsenceCache.Add(claimRef.UID, nil)
		return false, nil
	}
	return true, nil
}

func (r *VolumeReleaseReconciler) releaseVolume(ctx context.Context, volume *storagev1alpha1.Volume) error {
	baseNic := volume.DeepCopy()
	volume.Spec.ClaimRef = nil
	if err := r.Patch(ctx, volume, client.StrategicMergeFrom(baseNic, client.MergeFromWithOptimisticLock{})); err != nil {
		return fmt.Errorf("error patching volume: %w", err)
	}
	return nil
}

func (r *VolumeReleaseReconciler) reconcile(ctx context.Context, log logr.Logger, volume *storagev1alpha1.Volume) (ctrl.Result, error) {
	log.V(1).Info("Reconcile")

	if volume.Spec.ClaimRef == nil {
		log.V(1).Info("Volume is not claimed, nothing to do")
		return ctrl.Result{}, nil
	}

	log.V(1).Info("Checking whether volume claimer exists")
	ok, err := r.volumeClaimExists(ctx, volume)
	if err != nil {
		return ctrl.Result{}, fmt.Errorf("error checking whether volume claimer exists: %w", err)
	}
	if ok {
		log.V(1).Info("Volume claimer is still present")
		return ctrl.Result{}, nil
	}

	log.V(1).Info("Volume claimer does not exist, releasing volume")
	if err := r.releaseVolume(ctx, volume); err != nil {
		if !apierrors.IsConflict(err) {
			return ctrl.Result{}, fmt.Errorf("error releasing volume: %w", err)
		}
		log.V(1).Info("Volume was updated, requeueing")
		return ctrl.Result{Requeue: true}, nil
	}

	log.V(1).Info("Reconciled")
	return ctrl.Result{}, nil
}

func (r *VolumeReleaseReconciler) volumeClaimedPredicate() predicate.Predicate {
	return predicate.NewPredicateFuncs(func(obj client.Object) bool {
		volume := obj.(*storagev1alpha1.Volume)
		return volume.Spec.ClaimRef != nil
	})
}

func (r *VolumeReleaseReconciler) enqueueByMachine() handler.EventHandler {
	return handler.EnqueueRequestsFromMapFunc(func(ctx context.Context, obj client.Object) []ctrl.Request {
		machine := obj.(*computev1alpha1.Machine)
		log := ctrl.LoggerFrom(ctx)

		volumeList := &storagev1alpha1.VolumeList{}
		if err := r.List(ctx, volumeList,
			client.InNamespace(machine.Namespace),
		); err != nil {
			log.Error(err, "Error listing volumes")
			return nil
		}

		var reqs []ctrl.Request
		for _, volume := range volumeList.Items {
			claimRef := volume.Spec.ClaimRef
			if claimRef == nil {
				continue
			}

			if claimRef.UID != machine.UID {
				continue
			}

			reqs = append(reqs, ctrl.Request{NamespacedName: client.ObjectKeyFromObject(&volume)})
		}
		return reqs
	})
}

func (r *VolumeReleaseReconciler) machineDeletingPredicate() predicate.Predicate {
	return predicate.NewPredicateFuncs(func(obj client.Object) bool {
		machine := obj.(*computev1alpha1.Machine)
		return !machine.DeletionTimestamp.IsZero()
	})
}

func (r *VolumeReleaseReconciler) SetupWithManager(mgr ctrl.Manager) error {
	return ctrl.NewControllerManagedBy(mgr).
		Named("volumerelease").
		For(
			&storagev1alpha1.Volume{},
			builder.WithPredicates(r.volumeClaimedPredicate()),
		).
		Watches(
			&computev1alpha1.Machine{},
			r.enqueueByMachine(),
			builder.WithPredicates(r.machineDeletingPredicate()),
		).
		Complete(r)
}
