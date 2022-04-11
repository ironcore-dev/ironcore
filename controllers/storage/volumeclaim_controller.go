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
	"sigs.k8s.io/controller-runtime/pkg/builder"
	"sigs.k8s.io/controller-runtime/pkg/handler"
	"sigs.k8s.io/controller-runtime/pkg/predicate"
	"sigs.k8s.io/controller-runtime/pkg/source"

	storagev1alpha1 "github.com/onmetal/onmetal-api/apis/storage/v1alpha1"
	"k8s.io/apimachinery/pkg/runtime"
	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/client"
)

// VolumeClaimReconciler reconciles a VolumeClaim object
type VolumeClaimReconciler struct {
	client.Client
	APIReader          client.Reader
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
	log.V(1).Info("Reconciling volume claim")
	if claim.Spec.VolumeRef.Name == "" {
		log.V(1).Info("Volume claim is not bound")
		if err := r.patchVolumeClaimStatus(ctx, claim, storagev1alpha1.VolumeClaimPending); err != nil {
			return ctrl.Result{}, fmt.Errorf("error setting volume claim to pending: %w", err)
		}
		return ctrl.Result{}, nil
	}

	volume := &storagev1alpha1.Volume{}
	volumeKey := client.ObjectKey{
		Namespace: claim.Namespace,
		Name:      claim.Spec.VolumeRef.Name,
	}
	log.V(1).Info("Getting volume for volume claim", "VolumeKey", volumeKey)
	// We have to use APIReader here as stale data might cause unbinding the already bound volume.
	if err := r.APIReader.Get(ctx, volumeKey, volume); err != nil {
		if !errors.IsNotFound(err) {
			return ctrl.Result{}, fmt.Errorf("error getting volume %s for volume claim: %w", volumeKey, err)
		}

		log.V(1).Info("Volume claim is lost as the corresponding volume cannot be found", "VolumeKey", volumeKey)
		if err := r.patchVolumeClaimStatus(ctx, claim, storagev1alpha1.VolumeClaimLost); err != nil {
			return ctrl.Result{}, fmt.Errorf("error setting volume claim to lost: %w", err)
		}
		return ctrl.Result{}, nil
	}

	if volume.Spec.ClaimRef.Name == claim.Name && volume.Spec.ClaimRef.UID == claim.UID {
		log.V(1).Info("Volume is bound to claim", "VolumeKey", volumeKey)
		if err := r.patchVolumeClaimStatus(ctx, claim, storagev1alpha1.VolumeClaimBound); err != nil {
			return ctrl.Result{}, fmt.Errorf("error setting volume claim to bound: %w", err)
		}
		return ctrl.Result{}, nil
	}

	log.V(1).Info("Volume is not (yet) bound to claim", "VolumeKey", volumeKey, "ClaimRef", volume.Spec.ClaimRef)
	if err := r.patchVolumeClaimStatus(ctx, claim, storagev1alpha1.VolumeClaimPending); err != nil {
		return ctrl.Result{}, fmt.Errorf("error setting volume claim to pending: %w", err)
	}
	return ctrl.Result{}, nil
}

func (r *VolumeClaimReconciler) patchVolumeClaimStatus(ctx context.Context, volumeClaim *storagev1alpha1.VolumeClaim, phase storagev1alpha1.VolumeClaimPhase) error {
	base := volumeClaim.DeepCopy()
	volumeClaim.Status.Phase = phase
	return r.Status().Patch(ctx, volumeClaim, client.MergeFrom(base))
}

const (
	VolumeClaimSpecVolumeRefNameField = ".spec.volumeRef.name"
)

// SetupWithManager sets up the controller with the Manager.
func (r *VolumeClaimReconciler) SetupWithManager(mgr ctrl.Manager) error {
	ctx := context.Background()
	log := ctrl.Log.WithName("volumeclaim").WithName("setup")

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
		For(&storagev1alpha1.VolumeClaim{}).
		Watches(&source.Kind{Type: &storagev1alpha1.Volume{}},
			handler.EnqueueRequestsFromMapFunc(func(object client.Object) []ctrl.Request {
				volume := object.(*storagev1alpha1.Volume)

				claims := &storagev1alpha1.VolumeClaimList{}
				if err := r.List(ctx, claims, &client.ListOptions{
					FieldSelector: fields.OneTermEqualSelector(VolumeClaimSpecVolumeRefNameField, volume.GetName()),
					Namespace:     volume.GetNamespace(),
				}); err != nil {
					log.Error(err, "error listing volume claims matching volume", "VolumeKey", client.ObjectKeyFromObject(volume))
					return []ctrl.Request{}
				}

				res := make([]ctrl.Request, 0, len(claims.Items))
				for _, item := range claims.Items {
					res = append(res, ctrl.Request{
						NamespacedName: client.ObjectKeyFromObject(&item),
					})
				}
				return res
			}),
			builder.WithPredicates(predicate.ResourceVersionChangedPredicate{}),
		).
		Complete(r)
}
