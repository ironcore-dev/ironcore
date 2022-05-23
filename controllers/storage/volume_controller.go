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
	"time"

	"github.com/go-logr/logr"
	"github.com/onmetal/controller-utils/clientutils"
	storagev1alpha1 "github.com/onmetal/onmetal-api/apis/storage/v1alpha1"
	apiequality "github.com/onmetal/onmetal-api/equality"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apimachinery/pkg/types"
	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/builder"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/event"
	"sigs.k8s.io/controller-runtime/pkg/handler"
	"sigs.k8s.io/controller-runtime/pkg/predicate"
	"sigs.k8s.io/controller-runtime/pkg/source"
)

// VolumeReconciler reconciles a Volume object
type VolumeReconciler struct {
	client.Client
	APIReader client.Reader
	Scheme    *runtime.Scheme
	// BindTimeout is the maximum duration until a Volume's Bound condition is considered to be timed out.
	BindTimeout        time.Duration
	SharedFieldIndexer *clientutils.SharedFieldIndexer
}

//+kubebuilder:rbac:groups=storage.api.onmetal.de,resources=volumes,verbs=get;list;watch;create;update;patch;delete
//+kubebuilder:rbac:groups=storage.api.onmetal.de,resources=volumes/status,verbs=get;update;patch
//+kubebuilder:rbac:groups=storage.api.onmetal.de,resources=volumes/finalizers,verbs=update
//+kubebuilder:rbac:groups=storage.api.onmetal.de,resources=volumeclaims,verbs=get;list

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

func (r *VolumeReconciler) phaseTransitionTimedOut(timestamp *metav1.Time) bool {
	if timestamp.IsZero() {
		return false
	}
	return timestamp.Add(r.BindTimeout).Before(time.Now())
}

func (r *VolumeReconciler) reconcile(ctx context.Context, log logr.Logger, volume *storagev1alpha1.Volume) (ctrl.Result, error) {
	log.V(1).Info("Reconciling volume")
	if volume.Spec.ClaimRef == nil {
		log.V(1).Info("Volume is not bound and not referencing any claim")
		if err := r.patchVolumeStatus(ctx, volume, storagev1alpha1.VolumePhaseUnbound); err != nil {
			return ctrl.Result{}, err
		}

		log.V(1).Info("Successfully marked volume as unbound")
		return ctrl.Result{}, nil
	}

	volumeClaim := &storagev1alpha1.VolumeClaim{}
	volumeClaimKey := client.ObjectKey{
		Namespace: volume.Namespace,
		Name:      volume.Spec.ClaimRef.Name,
	}
	log = log.WithValues("VolumeClaimKey", volumeClaimKey)
	log.V(1).Info("Volume references claim")
	// We have to use APIReader here as stale data might cause unbinding a volume for a short duration.
	err := r.APIReader.Get(ctx, volumeClaimKey, volumeClaim)
	if client.IgnoreNotFound(err) != nil {
		return ctrl.Result{}, fmt.Errorf("error getting volume claim %s: %w", volumeClaimKey, err)
	}

	if err == nil && volume.Spec.ClaimRef.UID == "" {
		log = log.WithValues("VolumeClaimUID", volumeClaim.UID)
		log.V(1).Info("Setting claim ref uid")
		if err := r.setClaimRefUID(ctx, volume, volumeClaim.UID); err != nil {
			return ctrl.Result{}, err
		}

		log.V(1).Info("Set claim ref uid")
		return ctrl.Result{}, nil
	}

	volumeClaimExists := err == nil
	validReferences := volumeClaimExists && r.validReferences(volume, volumeClaim)
	volumePhase := volume.Status.Phase
	volumePhaseLastTransitionTime := volume.Status.LastPhaseTransitionTime
	volumeState := volume.Status.State

	bindOK := volumeState == storagev1alpha1.VolumeStateAvailable && validReferences

	log = log.WithValues(
		"VolumeClaimExists", volumeClaimExists,
		"ValidReferences", validReferences,
		"VolumeState", volumeState,
		"VolumePhase", volumePhase,
		"VolumePhaseLastTransitionTime", volumePhaseLastTransitionTime,
	)
	switch {
	case bindOK:
		log.V(1).Info("Setting volume to bound")
		if err := r.patchVolumeStatus(ctx, volume, storagev1alpha1.VolumePhaseBound); err != nil {
			return ctrl.Result{}, fmt.Errorf("error binding volume: %w", err)
		}

		log.V(1).Info("Successfully set volume to bound.")
		return ctrl.Result{}, nil
	case !bindOK && volumePhase == storagev1alpha1.VolumePhasePending && r.phaseTransitionTimedOut(volumePhaseLastTransitionTime):
		log.V(1).Info("Bind is not ok and timed out, releasing volume")
		if err := r.releaseVolume(ctx, volume); err != nil {
			return ctrl.Result{}, fmt.Errorf("error releasing volume: %w", err)
		}

		log.V(1).Info("Successfully released volume")
		return ctrl.Result{}, nil
	default:
		log.V(1).Info("Bind is not ok and not yet timed out, setting to pending")
		if err := r.patchVolumeStatus(ctx, volume, storagev1alpha1.VolumePhasePending); err != nil {
			return ctrl.Result{}, fmt.Errorf("error setting volume to pending: %w", err)
		}

		log.V(1).Info("Successfully set volume to pending")
		return r.requeueAfterBoundTimeout(volume), nil
	}
}

func (r *VolumeReconciler) requeueAfterBoundTimeout(volume *storagev1alpha1.Volume) ctrl.Result {
	boundTimeoutExpirationDuration := time.Until(volume.Status.LastPhaseTransitionTime.Add(r.BindTimeout)).Round(time.Second)
	if boundTimeoutExpirationDuration <= 0 {
		return ctrl.Result{Requeue: true}
	}
	return ctrl.Result{RequeueAfter: boundTimeoutExpirationDuration}
}

func (r *VolumeReconciler) validReferences(volume *storagev1alpha1.Volume, volumeClaim *storagev1alpha1.VolumeClaim) bool {
	volumeRef := volumeClaim.Spec.VolumeRef
	if volumeRef == nil || volumeRef.Name != volume.Name {
		return false
	}

	claimRef := volume.Spec.ClaimRef
	if claimRef == nil {
		return false
	}
	return claimRef.Name == volumeClaim.Name && claimRef.UID == volumeClaim.UID
}

func (r *VolumeReconciler) releaseVolume(ctx context.Context, volume *storagev1alpha1.Volume) error {
	baseVolume := volume.DeepCopy()
	volume.Spec.ClaimRef = nil
	return r.Patch(ctx, volume, client.MergeFrom(baseVolume))
}

func (r *VolumeReconciler) setClaimRefUID(ctx context.Context, volume *storagev1alpha1.Volume, uid types.UID) error {
	base := volume.DeepCopy()
	volume.Spec.ClaimRef.UID = uid
	if err := r.Patch(ctx, volume, client.MergeFrom(base)); err != nil {
		return fmt.Errorf("error setting claim ref uid: %w", err)
	}
	return nil
}

func (r *VolumeReconciler) patchVolumeStatus(ctx context.Context, volume *storagev1alpha1.Volume, phase storagev1alpha1.VolumePhase) error {
	now := metav1.Now()
	volumeBase := volume.DeepCopy()

	if volume.Status.Phase != phase {
		volume.Status.LastPhaseTransitionTime = &now
	}
	volume.Status.Phase = phase

	return r.Status().Patch(ctx, volume, client.MergeFrom(volumeBase))
}

// SetupWithManager sets up the controller with the Manager.
func (r *VolumeReconciler) SetupWithManager(mgr ctrl.Manager) error {
	ctx := context.Background()
	log := ctrl.Log.WithName("volume").WithName("setup")
	if err := r.SharedFieldIndexer.IndexField(ctx, &storagev1alpha1.Volume{}, VolumeSpecVolumeClaimNameRefField); err != nil {
		return err
	}

	return ctrl.NewControllerManagedBy(mgr).
		For(
			&storagev1alpha1.Volume{},
			builder.WithPredicates(predicate.Funcs{
				UpdateFunc: func(event event.UpdateEvent) bool {
					oldVolume, newVolume := event.ObjectOld.(*storagev1alpha1.Volume), event.ObjectNew.(*storagev1alpha1.Volume)
					return !apiequality.Semantic.DeepEqual(oldVolume.Spec, newVolume.Spec) ||
						oldVolume.Status.State != newVolume.Status.State ||
						oldVolume.Status.Phase != newVolume.Status.Phase
				},
			}),
		).
		Watches(&source.Kind{Type: &storagev1alpha1.VolumeClaim{}},
			handler.EnqueueRequestsFromMapFunc(func(obj client.Object) []ctrl.Request {
				volumeClaim := obj.(*storagev1alpha1.VolumeClaim)

				volumes := &storagev1alpha1.VolumeList{}
				if err := r.List(ctx, volumes, client.InNamespace(volumeClaim.Namespace), client.MatchingFields{
					VolumeSpecVolumeClaimNameRefField: volumeClaim.Name,
				}); err != nil {
					log.Error(err, "Error listing claims using volume")
					return []ctrl.Request{}
				}

				res := make([]ctrl.Request, 0, len(volumes.Items))
				for _, item := range volumes.Items {
					res = append(res, ctrl.Request{
						NamespacedName: types.NamespacedName{
							Name:      item.GetName(),
							Namespace: item.GetNamespace(),
						},
					})
				}
				return res
			}),
			builder.WithPredicates(predicate.Funcs{
				UpdateFunc: func(event event.UpdateEvent) bool {
					oldVolumeClaim, newVolumeClaim := event.ObjectOld.(*storagev1alpha1.VolumeClaim), event.ObjectNew.(*storagev1alpha1.VolumeClaim)
					return !apiequality.Semantic.DeepEqual(oldVolumeClaim.Spec, newVolumeClaim.Spec) ||
						oldVolumeClaim.Status.Phase != newVolumeClaim.Status.Phase
				},
			}),
		).
		Complete(r)
}
