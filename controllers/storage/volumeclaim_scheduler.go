/*
 * Copyright (c) 2022 by the OnMetal authors.
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
	storagev1alpha1 "github.com/onmetal/onmetal-api/apis/storage/v1alpha1"
	corev1 "k8s.io/api/core/v1"
	apierrors "k8s.io/apimachinery/pkg/api/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	quotav1 "k8s.io/apiserver/pkg/quota/v1"
	"k8s.io/client-go/tools/record"
	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/builder"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/handler"
	"sigs.k8s.io/controller-runtime/pkg/predicate"
	"sigs.k8s.io/controller-runtime/pkg/source"
)

type VolumeClaimScheduler struct {
	client.Client
	record.EventRecorder
}

//+kubebuilder:rbac:groups=storage.api.onmetal.de,resources=volumeclaims,verbs=get;list;watch;create;update;patch;delete
//+kubebuilder:rbac:groups=storage.api.onmetal.de,resources=volumeclaims/status,verbs=get;update;patch
//+kubebuilder:rbac:groups=storage.api.onmetal.de,resources=volumeclaims/finalizers,verbs=update
//+kubebuilder:rbac:groups=storage.api.onmetal.de,resources=volumes,verbs=get;list;watch;update;patch

func (s *VolumeClaimScheduler) Reconcile(ctx context.Context, req ctrl.Request) (ctrl.Result, error) {
	log := ctrl.LoggerFrom(ctx)
	volumeClaim := &storagev1alpha1.VolumeClaim{}
	if err := s.Get(ctx, req.NamespacedName, volumeClaim); err != nil {
		return ctrl.Result{}, client.IgnoreNotFound(err)
	}
	return s.reconcileExists(ctx, log, volumeClaim)
}

func (s *VolumeClaimScheduler) reconcileExists(ctx context.Context, log logr.Logger, claim *storagev1alpha1.VolumeClaim) (ctrl.Result, error) {
	if !claim.DeletionTimestamp.IsZero() {
		return s.delete(ctx, log, claim)
	}
	return s.reconcile(ctx, log, claim)
}

func (s *VolumeClaimScheduler) delete(ctx context.Context, log logr.Logger, claim *storagev1alpha1.VolumeClaim) (ctrl.Result, error) {
	return ctrl.Result{}, nil
}

func (s *VolumeClaimScheduler) bind(ctx context.Context, volumeClaim *storagev1alpha1.VolumeClaim, volume *storagev1alpha1.Volume) error {
	baseVolume := volume.DeepCopy()
	volume.Spec.ClaimRef = storagev1alpha1.ClaimReference{
		Name: volumeClaim.Name,
		UID:  volumeClaim.UID,
	}
	if err := s.Patch(ctx, volume, client.MergeFrom(baseVolume)); err != nil {
		return fmt.Errorf("could not assign volume claim to volume %s: %w", client.ObjectKeyFromObject(volume), err)
	}

	baseClaim := volumeClaim.DeepCopy()
	volumeClaim.Spec.VolumeRef = corev1.LocalObjectReference{Name: volume.Name}
	if err := s.Patch(ctx, volumeClaim, client.MergeFrom(baseClaim)); err != nil {
		return fmt.Errorf("could not assign volume %s to volume claim: %w", client.ObjectKeyFromObject(volume), err)
	}

	return nil
}

func (s *VolumeClaimScheduler) reconcile(ctx context.Context, log logr.Logger, volumeClaim *storagev1alpha1.VolumeClaim) (ctrl.Result, error) {
	log.V(1).Info("Reconciling volume claim")
	if volumeRefName := volumeClaim.Spec.VolumeRef.Name; volumeRefName != "" {
		volume := &storagev1alpha1.Volume{}
		volumeKey := client.ObjectKey{Namespace: volumeClaim.Namespace, Name: volumeRefName}
		log = log.WithValues("VolumeKey", volumeKey)
		log.V(1).Info("Getting volume specified by volume claim")
		if err := s.Get(ctx, volumeKey, volume); err != nil {
			if !apierrors.IsNotFound(err) {
				return ctrl.Result{}, fmt.Errorf("error getting volume %s specified by volume claim: %w", volumeKey, err)
			}

			log.V(1).Info("Volume specified by claim does not exist")
			return ctrl.Result{}, nil
		}

		if volume.Spec.ClaimRef.Name != "" {
			log.V(1).Info("Volume already specifies a claim ref", "ClaimRef", volume.Spec.ClaimRef)
			return ctrl.Result{}, nil
		}

		log.V(1).Info("Volume is not yet bound, trying to bind it")
		if err := s.bind(ctx, volumeClaim, volume); err != nil {
			return ctrl.Result{}, fmt.Errorf("error binding: %w", err)
		}

		log.V(1).Info("Successfully bound volume and claim")
		return ctrl.Result{}, nil
	}

	log.V(1).Info("Listing suitable volumes")
	sel, err := metav1.LabelSelectorAsSelector(volumeClaim.Spec.Selector)
	if err != nil {
		return ctrl.Result{}, fmt.Errorf("invalid label selector: %w", err)
	}

	volumeList := &storagev1alpha1.VolumeList{}
	if err := s.List(ctx, volumeList, client.InNamespace(volumeClaim.Namespace), client.MatchingLabelsSelector{Selector: sel}); err != nil {
		return ctrl.Result{}, fmt.Errorf("failed to list matching volumes: %w", err)
	}

	log.V(1).Info("Searching for suitable volumes for volume claim")
	volume := s.findVolumeForClaim(volumeList.Items, volumeClaim)
	if volume == nil {
		s.Event(volumeClaim, corev1.EventTypeNormal, "FailedScheduling", "no matching volume found for volume claim")
		log.V(1).Info("Could not find a matching volume for volume claim")
		return ctrl.Result{}, nil
	}

	log = log.WithValues("VolumeKey", client.ObjectKeyFromObject(volume))
	log.V(1).Info("Found matching volume, binding volume and volume claim")
	if err := s.bind(ctx, volumeClaim, volume); err != nil {
		return ctrl.Result{}, fmt.Errorf("error binding volume %s to claim: %w", client.ObjectKeyFromObject(volume), err)
	}

	log.V(1).Info("Successfully bound volume to claim", "Volume", client.ObjectKeyFromObject(volume))
	return ctrl.Result{}, nil
}

func (s *VolumeClaimScheduler) findVolumeForClaim(volumes []storagev1alpha1.Volume, claim *storagev1alpha1.VolumeClaim) *storagev1alpha1.Volume {
	var matchingVolume *storagev1alpha1.Volume
	for _, vol := range volumes {
		if !vol.DeletionTimestamp.IsZero() {
			continue
		}

		if claimRefName := vol.Spec.ClaimRef.Name; claimRefName != "" {
			if claimRefName != claim.Name {
				continue
			}
			if vol.Spec.ClaimRef.UID != claim.UID {
				continue
			}
			// If we hit a Volume that matches exactly our claim we need to return immediately to avoid over-claiming
			// Volumes in the cluster.
			vol := vol
			return &vol
		}

		if !s.volumeSatisfiesClaim(&vol, claim) {
			continue
		}
		vol := vol
		matchingVolume = &vol
	}
	return matchingVolume
}

func (s *VolumeClaimScheduler) volumeSatisfiesClaim(volume *storagev1alpha1.Volume, claim *storagev1alpha1.VolumeClaim) bool {
	if volume.Status.State != storagev1alpha1.VolumeStateAvailable {
		return false
	}
	if claim.Spec.VolumeClassRef != volume.Spec.VolumeClassRef {
		return false
	}
	// Check if the volume can occupy the claim
	if ok, _ := quotav1.LessThanOrEqual(claim.Spec.Resources, volume.Spec.Resources); !ok {
		return false
	}
	return true
}

// SetupWithManager sets up the controller with the Manager.
func (s *VolumeClaimScheduler) SetupWithManager(mgr ctrl.Manager) error {
	ctx := context.Background()
	log := ctrl.Log.WithName("volumeclaim-scheduler").WithName("setup")
	return ctrl.NewControllerManagedBy(mgr).
		Named("volumeclaim-scheduler").
		For(&storagev1alpha1.VolumeClaim{}, builder.WithPredicates(predicate.NewPredicateFuncs(func(object client.Object) bool {
			// Only reconcile claims which haven't been bound
			claim := object.(*storagev1alpha1.VolumeClaim)
			return claim.Status.Phase == storagev1alpha1.VolumeClaimPending
		}))).
		Watches(&source.Kind{Type: &storagev1alpha1.Volume{}}, handler.EnqueueRequestsFromMapFunc(
			func(object client.Object) []ctrl.Request {
				volume := object.(*storagev1alpha1.Volume)
				if claimName := volume.Spec.ClaimRef.Name; claimName != "" {
					claim := &storagev1alpha1.VolumeClaim{}
					claimKey := client.ObjectKey{Namespace: volume.Namespace, Name: claimName}
					if err := s.Get(ctx, claimKey, claim); err != nil {
						if !apierrors.IsNotFound(err) {
							log.Error(err, "Failed to get claim referenced by volume", "VolumeClaimRef", claimKey)
							return nil
						}

						log.V(1).Info("Claim referenced by volume does not exist", "VolumeClaimRef", claimKey)
						return nil
					}

					if claim.Spec.VolumeRef.Name != "" {
						return nil
					}
					log.V(1).Info("Enqueueing claim that has already been accepted by its volume", "VolumeClaimRef", claimKey)
					return []ctrl.Request{{NamespacedName: claimKey}}
				}

				volumeClaims := &storagev1alpha1.VolumeClaimList{}
				if err := s.List(ctx, volumeClaims, client.InNamespace(volume.Namespace), client.MatchingFields{
					VolumeClaimSpecVolumeRefNameField: "",
				}); err != nil {
					log.Error(err, "Could not list empty VolumeClaims", "Namespace", volume.Namespace)
					return nil
				}
				var requests []ctrl.Request
				for _, claim := range volumeClaims.Items {
					if s.volumeSatisfiesClaim(volume, &claim) {
						requests = append(requests, ctrl.Request{
							NamespacedName: client.ObjectKeyFromObject(&claim),
						})
					}
				}
				return requests
			})).
		Complete(s)
}
