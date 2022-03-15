/*
 * Copyright (c) 2021 by the OnMetal authors.
 *
 * Licensed under the Apache License, Version 2.0 (the "License");
 * you may not use this file except in compliance with the License.
 * You may obtain a copy of the License at
 *
 *     http://www.apache.org/licenses/LICENSE-2.0
 *
 * Unless required by applicable law or agreed to in writing, software
 * distributed under the License is distributed on an "AS IS" BASIS,
 * WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
 * See the License for the specific language governing permissions and
 * limitations under the License.
 */

package storage

import (
	"context"
	"fmt"

	"github.com/go-logr/logr"
	"github.com/onmetal/controller-utils/clientutils"
	storagev1alpha1 "github.com/onmetal/onmetal-api/apis/storage/v1alpha1"
	"k8s.io/apimachinery/pkg/api/errors"
	"k8s.io/apimachinery/pkg/fields"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apimachinery/pkg/types"
	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/builder"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/handler"
	"sigs.k8s.io/controller-runtime/pkg/predicate"
	"sigs.k8s.io/controller-runtime/pkg/reconcile"
	"sigs.k8s.io/controller-runtime/pkg/source"
)

// VolumeReconciler reconciles a Volume object
type VolumeReconciler struct {
	client.Client
	Scheme             *runtime.Scheme
	SharedFieldIndexer *clientutils.SharedFieldIndexer
}

//+kubebuilder:rbac:groups=storage.onmetal.de,resources=volumes,verbs=get;list;watch;create;update;patch;delete
//+kubebuilder:rbac:groups=storage.onmetal.de,resources=volumes/status,verbs=get;update;patch
//+kubebuilder:rbac:groups=storage.onmetal.de,resources=volumes/finalizers,verbs=update
//+kubebuilder:rbac:groups=storage.onmetal.de,resources=volumeclaims,verbs=get;list

// Reconcile is part of the main reconciliation loop for Volume types
func (r *VolumeReconciler) Reconcile(ctx context.Context, req ctrl.Request) (ctrl.Result, error) {
	log := ctrl.LoggerFrom(ctx)
	volume := &storagev1alpha1.Volume{}
	if err := r.Get(ctx, req.NamespacedName, volume); err != nil {
		return ctrl.Result{}, client.IgnoreNotFound(err)
	}
	return r.reconcileExists(ctx, log, volume)
}

func (r *VolumeReconciler) reconcileExists(ctx context.Context, log logr.Logger, volume *storagev1alpha1.Volume) (ctrl.Result, error) {
	if !volume.DeletionTimestamp.IsZero() {
		return r.delete(ctx, log, volume)
	}
	return r.reconcile(ctx, log, volume)
}

func (r *VolumeReconciler) delete(ctx context.Context, log logr.Logger, volume *storagev1alpha1.Volume) (ctrl.Result, error) {
	return ctrl.Result{}, nil
}

func (r *VolumeReconciler) reconcile(ctx context.Context, log logr.Logger, volume *storagev1alpha1.Volume) (ctrl.Result, error) {
	log.V(1).Info("Reconciling volume")
	if volume.Spec.ClaimRef.Name == "" {
		if err := r.updateVolumePhase(ctx, log, volume, storagev1alpha1.VolumeAvailable); err != nil {
			return ctrl.Result{}, err
		}
		return ctrl.Result{}, nil
	}

	volumeClaim := &storagev1alpha1.VolumeClaim{}
	volumeClaimKey := client.ObjectKey{
		Namespace: volume.Namespace,
		Name:      volume.Spec.ClaimRef.Name,
	}
	log.V(1).Info("Volume is bound to claim", "VolumeClaimKey", volumeClaimKey)
	if err := r.Get(ctx, volumeClaimKey, volumeClaim); err != nil {
		if !errors.IsNotFound(err) {
			return ctrl.Result{}, fmt.Errorf("failed to get volumeclaim %s: %w", volumeClaimKey, err)
		}

		log.V(1).Info("volume is released as the corresponding claim can not be found", "Volume", client.ObjectKeyFromObject(volume), "VolumeClaim", volumeClaimKey)
		if err := r.updateVolumePhase(ctx, log, volume, storagev1alpha1.VolumeAvailable); err != nil {
			return ctrl.Result{}, err
		}
		baseVolume := volume.DeepCopy()
		volume.Spec.ClaimRef = storagev1alpha1.ClaimReference{}
		if err := r.Patch(ctx, volume, client.MergeFrom(baseVolume)); err != nil {
			return ctrl.Result{}, fmt.Errorf("could not remove claim to volume %s: %w", client.ObjectKeyFromObject(volume), err)
		}
		return ctrl.Result{}, nil
	}
	if volumeClaim.Spec.VolumeRef.Name == volume.Name && volume.Spec.ClaimRef.UID == volumeClaim.UID {
		log.Info("synchronizing Volume: all is bound", "Volume", client.ObjectKeyFromObject(volume))
		if err := r.updateVolumePhase(ctx, log, volume, storagev1alpha1.VolumeBound); err != nil {
			return ctrl.Result{}, err
		}
	}
	return ctrl.Result{}, nil
}

func (r *VolumeReconciler) updateVolumePhase(ctx context.Context, log logr.Logger, volume *storagev1alpha1.Volume, phase storagev1alpha1.VolumePhase) error {
	log.V(1).Info("patching volume phase", "Volume", client.ObjectKeyFromObject(volume), "Phase", phase)
	if volume.Status.Phase == phase {
		// Nothing to do.
		log.V(1).Info("updating Volume: phase already set", "Volume", client.ObjectKeyFromObject(volume), "Phase", phase)
		return nil
	}
	volumeBase := volume.DeepCopy()
	volume.Status.Phase = phase
	if err := r.Status().Patch(ctx, volume, client.MergeFrom(volumeBase)); err != nil {
		return fmt.Errorf("updating Volume %s: set phase %s failed: %w", volume.Name, phase, err)
	}
	log.V(1).Info("patched volume phase", "Volume", client.ObjectKeyFromObject(volume), "Phase", phase)
	return nil
}

// SetupWithManager sets up the controller with the Manager.
func (r *VolumeReconciler) SetupWithManager(mgr ctrl.Manager) error {
	ctx := context.Background()
	if err := r.SharedFieldIndexer.IndexField(ctx, &storagev1alpha1.Volume{}, VolumeSpecVolumeClaimNameRefField); err != nil {
		return err
	}

	return ctrl.NewControllerManagedBy(mgr).
		Named("volume-controller").
		For(&storagev1alpha1.Volume{}).
		Watches(&source.Kind{Type: &storagev1alpha1.VolumeClaim{}},
			handler.EnqueueRequestsFromMapFunc(func(object client.Object) []reconcile.Request {
				volumeClaim := object.(*storagev1alpha1.VolumeClaim)
				volumes := &storagev1alpha1.VolumeList{}
				if err := r.List(ctx, volumes, &client.ListOptions{
					FieldSelector: fields.OneTermEqualSelector(VolumeSpecVolumeClaimNameRefField, volumeClaim.GetName()),
					Namespace:     volumeClaim.GetNamespace(),
				}); err != nil {
					return []reconcile.Request{}
				}
				requests := make([]reconcile.Request, len(volumes.Items))
				for i, item := range volumes.Items {
					requests[i] = reconcile.Request{
						NamespacedName: types.NamespacedName{
							Name:      item.GetName(),
							Namespace: item.GetNamespace(),
						},
					}
				}
				return requests
			}),
			builder.WithPredicates(predicate.ResourceVersionChangedPredicate{}),
		).
		Complete(r)
}
