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

//+kubebuilder:rbac:groups=storage.onmetal.de,resources=volumeclaims,verbs=get;list;watch;create;update;patch;delete
//+kubebuilder:rbac:groups=storage.onmetal.de,resources=volumeclaims/status,verbs=get;update;patch
//+kubebuilder:rbac:groups=storage.onmetal.de,resources=volumeclaims/finalizers,verbs=update
//+kubebuilder:rbac:groups=storage.onmetal.de,resources=volumes,verbs=get;list;watch;update;patch

func (s *VolumeClaimScheduler) Reconcile(ctx context.Context, req ctrl.Request) (ctrl.Result, error) {
	log := ctrl.LoggerFrom(ctx)
	log.Info("reconcile volume claim")
	volumeClaim := &storagev1alpha1.VolumeClaim{}
	if err := s.Get(ctx, req.NamespacedName, volumeClaim); err != nil {
		return ctrl.Result{}, client.IgnoreNotFound(err)
	}
	return s.reconileExists(ctx, log, volumeClaim)
}

func (s *VolumeClaimScheduler) reconileExists(ctx context.Context, log logr.Logger, claim *storagev1alpha1.VolumeClaim) (ctrl.Result, error) {
	if !claim.DeletionTimestamp.IsZero() {
		return s.delete(ctx, log, claim)
	}
	return s.reconcile(ctx, log, claim)
}

func (s *VolumeClaimScheduler) delete(ctx context.Context, log logr.Logger, claim *storagev1alpha1.VolumeClaim) (ctrl.Result, error) {
	return ctrl.Result{}, nil
}

func (s *VolumeClaimScheduler) reconcile(ctx context.Context, log logr.Logger, claim *storagev1alpha1.VolumeClaim) (ctrl.Result, error) {
	log.Info("reconcile volume claim")
	if claim.Spec.VolumeRef.Name != "" {
		log.Info("claim is already assigned to volume", "volume", claim.Spec.VolumeRef.Name)
		return ctrl.Result{}, nil
	}

	log.Info("listing suitable volumes")
	sel, err := metav1.LabelSelectorAsSelector(claim.Spec.Selector)
	if err != nil {
		return ctrl.Result{}, fmt.Errorf("invalid label selector: %w", err)
	}
	volumeList := &storagev1alpha1.VolumeList{}
	if err := s.List(ctx, volumeList, client.InNamespace(claim.Namespace), client.MatchingLabelsSelector{Selector: sel}); err != nil {
		return ctrl.Result{}, fmt.Errorf("failed to list matching volumes: %w", err)
	}

	matchingVolume := s.findVolumeForClaim(volumeList.Items, claim)
	if matchingVolume == nil {
		s.Event(claim, corev1.EventTypeNormal, "FailedScheduling", "no matching volume found for claim")
		log.Info("could not find a matching volume for claim")
		return ctrl.Result{}, nil
	}

	log.Info("found matching volume, assigning claim to volume", "Volume", client.ObjectKeyFromObject(matchingVolume))
	baseVolume := matchingVolume.DeepCopy()
	matchingVolume.Spec.ClaimRef = storagev1alpha1.ClaimReference{
		Name: claim.Name,
		UID:  claim.UID,
	}
	if err := s.Patch(ctx, matchingVolume, client.MergeFrom(baseVolume)); err != nil {
		return ctrl.Result{}, fmt.Errorf("could not assign claim to volume %s: %w", client.ObjectKeyFromObject(matchingVolume), err)
	}

	log.Info("assigning volume to claim", "Volume", client.ObjectKeyFromObject(matchingVolume))
	baseClaim := claim.DeepCopy()
	claim.Spec.VolumeRef = corev1.LocalObjectReference{Name: matchingVolume.Name}
	if err := s.Patch(ctx, claim, client.MergeFrom(baseClaim)); err != nil {
		return ctrl.Result{}, fmt.Errorf("could not assign volume %s to claim: %w", client.ObjectKeyFromObject(matchingVolume), err)
	}
	log.Info("successfully assigned volume to claim", "Volume", client.ObjectKeyFromObject(matchingVolume))
	return ctrl.Result{}, nil
}

func (s *VolumeClaimScheduler) findVolumeForClaim(volumes []storagev1alpha1.Volume, claim *storagev1alpha1.VolumeClaim) *storagev1alpha1.Volume {
	var matchingVolume *storagev1alpha1.Volume
	for _, vol := range volumes {
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
	if claim.Spec.StorageClassRef != volume.Spec.StorageClassRef {
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
	log := ctrl.Log.WithName("volume-claim-scheduler").WithName("setup")
	return ctrl.NewControllerManagedBy(mgr).
		Named("volume-claim-scheduler").
		For(&storagev1alpha1.VolumeClaim{}, builder.WithPredicates(predicate.NewPredicateFuncs(func(object client.Object) bool {
			// Only reconcile claims which haven't been scheduled
			claim := object.(*storagev1alpha1.VolumeClaim)
			return claim.Spec.VolumeRef.Name == ""
		}))).
		Watches(&source.Kind{Type: &storagev1alpha1.Volume{}}, handler.EnqueueRequestsFromMapFunc(
			func(object client.Object) []ctrl.Request {
				volume := object.(*storagev1alpha1.Volume)
				if claimName := volume.Spec.ClaimRef.Name; claimName != "" {
					claim := &storagev1alpha1.VolumeClaim{}
					claimKey := client.ObjectKey{Namespace: volume.Namespace, Name: claimName}
					if err := s.Get(ctx, claimKey, claim); err != nil {
						log.Error(err, "failed to get claim referenced by volume", "VolumeClaim", claimKey)
						return nil
					}
					if claim.Spec.VolumeRef.Name != "" {
						return nil
					}
					log.V(1).Info("enqueueing claim that has already been accepted by its volume", "VolumeClaim", claimKey)
					return []ctrl.Request{
						{NamespacedName: claimKey},
					}
				}
				volumeClaims := &storagev1alpha1.VolumeClaimList{}
				if err := s.List(ctx, volumeClaims, client.InNamespace(volume.Namespace), client.MatchingFields{
					VolumeClaimSpecVolumeRefNameField: "",
				}); err != nil {
					log.Error(err, "could not list empty VolumeClaims", "Namespace", volume.Namespace)
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
