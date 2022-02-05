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
	"k8s.io/apimachinery/pkg/api/errors"
	"k8s.io/apimachinery/pkg/fields"
	"k8s.io/apimachinery/pkg/types"
	"sigs.k8s.io/controller-runtime/pkg/builder"
	"sigs.k8s.io/controller-runtime/pkg/handler"
	"sigs.k8s.io/controller-runtime/pkg/predicate"
	"sigs.k8s.io/controller-runtime/pkg/reconcile"
	"sigs.k8s.io/controller-runtime/pkg/source"

	storagev1alpha1 "github.com/onmetal/onmetal-api/apis/storage/v1alpha1"
	"k8s.io/apimachinery/pkg/runtime"
	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/client"
)

// VolumeClaimReconciler reconciles a VolumeClaim object
type VolumeClaimReconciler struct {
	client.Client
	Scheme             *runtime.Scheme
	SharedFieldIndexer *clientutils.SharedFieldIndexer
}

//+kubebuilder:rbac:groups=storage.onmetal.de,resources=volumeclaims,verbs=get;list;watch;create;update;patch;delete
//+kubebuilder:rbac:groups=storage.onmetal.de,resources=volumeclaims/status,verbs=get;update;patch
//+kubebuilder:rbac:groups=storage.onmetal.de,resources=volumeclaims/finalizers,verbs=update
//+kubebuilder:rbac:groups=storage.onmetal.de,resources=volumes,verbs=get;list

// Reconcile is part of the main reconciliation loop for VolumeClaim types
func (r *VolumeClaimReconciler) Reconcile(ctx context.Context, req ctrl.Request) (ctrl.Result, error) {
	log := ctrl.LoggerFrom(ctx)
	claim := &storagev1alpha1.VolumeClaim{}
	if err := r.Get(ctx, req.NamespacedName, claim); err != nil {
		return ctrl.Result{}, client.IgnoreNotFound(err)
	}
	return r.reconcileExists(ctx, log, claim)
}

func (r *VolumeClaimReconciler) reconcileExists(ctx context.Context, log logr.Logger, claim *storagev1alpha1.VolumeClaim) (ctrl.Result, error) {
	if !claim.DeletionTimestamp.IsZero() {
		return r.delete(ctx, log, claim)
	}
	return r.reconcile(ctx, log, claim)
}

func (r *VolumeClaimReconciler) delete(ctx context.Context, log logr.Logger, claim *storagev1alpha1.VolumeClaim) (ctrl.Result, error) {
	return ctrl.Result{}, nil
}

func (r *VolumeClaimReconciler) reconcile(ctx context.Context, log logr.Logger, claim *storagev1alpha1.VolumeClaim) (ctrl.Result, error) {
	log.Info("synchronizing VolumeClaim", "VolumeClaim", client.ObjectKeyFromObject(claim))
	if claim.Spec.VolumeRef.Name == "" {
		if err := r.updateVolumeClaimPhase(ctx, log, claim, storagev1alpha1.VolumeClaimPending); err != nil {
			return ctrl.Result{}, err
		}
		return ctrl.Result{}, nil
	}

	log.Info("synchronizing VolumeClaim: claim is bound to volume",
		"VolumeClaim", client.ObjectKeyFromObject(claim), "Volume", claim.Spec.VolumeRef.Name)
	volume := &storagev1alpha1.Volume{}
	volumeKey := types.NamespacedName{
		Namespace: claim.Namespace,
		Name:      claim.Spec.VolumeRef.Name,
	}
	if err := r.Get(ctx, volumeKey, volume); err != nil {
		if !errors.IsNotFound(err) {
			return ctrl.Result{}, fmt.Errorf("failed to get volume %s for volumeclaim %s: %w", volumeKey, client.ObjectKeyFromObject(claim), err)
		}
		log.Info("volumeclaim is released as the corresponding volume can not be found", "VolumeClaim", client.ObjectKeyFromObject(claim), "Volume", volumeKey)
		if err := r.updateVolumeClaimPhase(ctx, log, claim, storagev1alpha1.VolumeClaimLost); err != nil {
			return ctrl.Result{}, err
		}
		return ctrl.Result{}, nil
	}
	if volume.Spec.ClaimRef.Name == claim.Name && volume.Spec.ClaimRef.UID == claim.UID {
		log.Info("synchronizing VolumeClaim: all is bound", "VolumeClaim", client.ObjectKeyFromObject(claim))
		if err := r.updateVolumeClaimPhase(ctx, log, claim, storagev1alpha1.VolumeClaimBound); err != nil {
			return ctrl.Result{}, err
		}
	}
	return ctrl.Result{}, nil
}

func (r *VolumeClaimReconciler) updateVolumeClaimPhase(ctx context.Context, log logr.Logger, claim *storagev1alpha1.VolumeClaim, phase storagev1alpha1.VolumeClaimPhase) error {
	log.V(1).Info("patching volumeclaim phase", "VolumeClaim", client.ObjectKeyFromObject(claim), "Phase", phase)
	if claim.Status.Phase == phase {
		// Nothing to do.
		log.V(1).Info("updating VolumeClaim: phase already set", "VolumeClaim", client.ObjectKeyFromObject(claim), "Phase", phase)
		return nil
	}
	volumeClaimBase := claim.DeepCopy()
	claim.Status.Phase = phase
	if err := r.Status().Patch(ctx, claim, client.MergeFrom(volumeClaimBase)); err != nil {
		return fmt.Errorf("updating VolumeClaim %s: set phase %s failed: %w", claim.Name, phase, err)
	}
	log.V(1).Info("patched volumeclaim phase", "VolumeClaim", client.ObjectKeyFromObject(claim), "Phase", phase)
	return nil
}

const (
	VolumeClaimSpecVolumeRefNameField = ".spec.volumeRef.name"
)

// SetupWithManager sets up the controller with the Manager.
func (r *VolumeClaimReconciler) SetupWithManager(mgr ctrl.Manager) error {
	ctx := context.Background()
	if err := r.SharedFieldIndexer.IndexField(ctx, &storagev1alpha1.Volume{}, VolumeSpecVolumeClaimNameRefField); err != nil {
		return err
	}
	if err := mgr.GetFieldIndexer().IndexField(ctx, &storagev1alpha1.VolumeClaim{}, VolumeClaimSpecVolumeRefNameField, func(object client.Object) []string {
		claim := object.(*storagev1alpha1.VolumeClaim)
		if claim.Spec.VolumeRef.Name == "" {
			return nil
		}
		return []string{claim.Spec.VolumeRef.Name}
	}); err != nil {
		return err
	}
	return ctrl.NewControllerManagedBy(mgr).
		Named("volumeclaim-controller").
		For(&storagev1alpha1.VolumeClaim{}).
		Watches(&source.Kind{Type: &storagev1alpha1.Volume{}},
			handler.EnqueueRequestsFromMapFunc(func(object client.Object) []reconcile.Request {
				volume := object.(*storagev1alpha1.Volume)
				claims := &storagev1alpha1.VolumeClaimList{}
				if err := r.List(ctx, claims, &client.ListOptions{
					FieldSelector: fields.OneTermEqualSelector(VolumeClaimSpecVolumeRefNameField, volume.GetName()),
					Namespace:     volume.GetNamespace(),
				}); err != nil {
					return []reconcile.Request{}
				}
				requests := make([]reconcile.Request, len(claims.Items))
				for i, item := range claims.Items {
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
