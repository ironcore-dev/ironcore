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
	"math/rand"
	"time"

	"github.com/go-logr/logr"
	commonv1alpha1 "github.com/onmetal/onmetal-api/api/common/v1alpha1"
	computev1alpha1 "github.com/onmetal/onmetal-api/api/compute/v1alpha1"
	storagev1alpha1 "github.com/onmetal/onmetal-api/api/storage/v1alpha1"
	computeclient "github.com/onmetal/onmetal-api/internal/client/compute"
	"github.com/onmetal/onmetal-api/internal/controllers/storage/events"
	apiequality "github.com/onmetal/onmetal-api/utils/equality"
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apimachinery/pkg/types"
	"k8s.io/client-go/tools/record"
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
	record.EventRecorder
	client.Client
	APIReader client.Reader
	Scheme    *runtime.Scheme
	// BindTimeout is the maximum duration until a Volume's Bound condition is considered to be timed out.
	BindTimeout time.Duration
}

//+kubebuilder:rbac:groups="",resources=events,verbs=create;patch
//+kubebuilder:rbac:groups=storage.api.onmetal.de,resources=volumes,verbs=get;list;watch;create;update;patch;delete
//+kubebuilder:rbac:groups=storage.api.onmetal.de,resources=volumes/status,verbs=get;update;patch
//+kubebuilder:rbac:groups=storage.api.onmetal.de,resources=volumes/finalizers,verbs=update
//+kubebuilder:rbac:groups=compute.api.onmetal.de,resources=machines,verbs=get;list;watch

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
		return r.reconcileUnbound(ctx, log, volume)
	}

	return r.reconcileBound(ctx, log, volume)
}

func (r *VolumeReconciler) assign(ctx context.Context, volume *storagev1alpha1.Volume, claimRef commonv1alpha1.LocalUIDReference) error {
	base := volume.DeepCopy()
	volume.Spec.ClaimRef = &claimRef
	if err := r.Patch(ctx, volume, client.MergeFrom(base)); err != nil {
		return fmt.Errorf("error assigning volume: %w", err)
	}
	return nil
}

func (r *VolumeReconciler) reconcileUnbound(ctx context.Context, log logr.Logger, volume *storagev1alpha1.Volume) (ctrl.Result, error) {
	log.V(1).Info("Reconcile unbound")

	if volume.Spec.Unclaimable {
		log.V(1).Info("Volume is marked as unclaimable, setting to unbound")
		if err := r.patchStatus(ctx, volume, storagev1alpha1.VolumePhaseUnbound); err != nil {
			return ctrl.Result{}, err
		}

		log.V(1).Info("Set volume to unbound")
		return ctrl.Result{}, nil
	}

	log.V(1).Info("Getting requesting machine")
	machine, err := r.getRequestingMachine(ctx, volume)
	if err != nil {
		return ctrl.Result{}, fmt.Errorf("error getting requesting machine: %w", err)
	}

	if machine == nil {
		log.V(1).Info("No requester found, setting volume to unbound")
		if err := r.patchStatus(ctx, volume, storagev1alpha1.VolumePhaseUnbound); err != nil {
			return ctrl.Result{}, err
		}

		log.V(1).Info("Set volume to unbound")
		return ctrl.Result{}, nil
	}

	machineKey := client.ObjectKeyFromObject(machine)
	log = log.WithValues("MachineKey", machineKey)

	claimRef := commonv1alpha1.LocalUIDReference{
		Name: machine.Name,
		UID:  machine.UID,
	}
	log.V(1).Info("Assigning volume")
	if err := r.assign(ctx, volume, claimRef); err != nil {
		return ctrl.Result{}, fmt.Errorf("error assigning volume: %w", err)
	}

	log.V(1).Info("Successfully assigned volume")
	return ctrl.Result{}, nil
}

func (r *VolumeReconciler) getRequestingMachine(ctx context.Context, volume *storagev1alpha1.Volume) (*computev1alpha1.Machine, error) {
	machineList := &computev1alpha1.MachineList{}
	if err := r.List(ctx, machineList,
		client.InNamespace(volume.Namespace),
		client.MatchingFields{computeclient.MachineSpecVolumeNamesField: volume.Name},
	); err != nil {
		return nil, fmt.Errorf("error listing machines specifying volume: %w", err)
	}

	var matches []computev1alpha1.Machine
	for _, machine := range machineList.Items {
		if !machine.DeletionTimestamp.IsZero() {
			continue
		}

		matches = append(matches, machine)
	}
	if len(matches) == 0 {
		return nil, nil
	}

	match := matches[rand.Intn(len(matches))]
	return &match, nil
}

func (r *VolumeReconciler) reconcileBound(ctx context.Context, log logr.Logger, volume *storagev1alpha1.Volume) (ctrl.Result, error) {
	machine := &computev1alpha1.Machine{}
	machineKey := client.ObjectKey{
		Namespace: volume.Namespace,
		Name:      volume.Spec.ClaimRef.Name,
	}
	log = log.WithValues("MachineKey", machineKey)
	log.V(1).Info("Reconcile bound")
	// We have to use APIReader here as stale data might cause unbinding a volume for a short duration.
	err := r.APIReader.Get(ctx, machineKey, machine)
	if client.IgnoreNotFound(err) != nil {
		return ctrl.Result{}, fmt.Errorf("error getting machine %s: %w", machineKey, err)
	}

	machineExists := err == nil
	validReferences := machineExists && r.validReferences(volume, machine)
	volumePhase := volume.Status.Phase
	volumePhaseLastTransitionTime := volume.Status.LastPhaseTransitionTime
	volumeState := volume.Status.State

	log = log.WithValues(
		"MachineExists", machineExists,
		"ValidReferences", validReferences,
		"VolumeState", volumeState,
		"VolumePhase", volumePhase,
		"VolumePhaseLastTransitionTime", volumePhaseLastTransitionTime,
	)

	if !machineExists {
		r.Eventf(volume, corev1.EventTypeWarning, events.FailedBindingMachine, "Machine %s not found", machineKey.Name)
	}

	switch {
	case validReferences:
		log.V(1).Info("Setting volume to bound")
		if err := r.patchStatus(ctx, volume, storagev1alpha1.VolumePhaseBound); err != nil {
			return ctrl.Result{}, fmt.Errorf("error binding volume: %w", err)
		}

		log.V(1).Info("Successfully set volume to bound.")
		return ctrl.Result{}, nil
	case !validReferences && volumePhase == storagev1alpha1.VolumePhasePending && r.phaseTransitionTimedOut(volumePhaseLastTransitionTime):
		log.V(1).Info("Bind is not ok and timed out, releasing volume")
		if err := r.releaseVolume(ctx, volume); err != nil {
			return ctrl.Result{}, fmt.Errorf("error releasing volume: %w", err)
		}

		log.V(1).Info("Successfully released volume")
		return ctrl.Result{}, nil
	default:
		log.V(1).Info("Bind is not ok and not yet timed out, setting to pending")
		if err := r.patchStatus(ctx, volume, storagev1alpha1.VolumePhasePending); err != nil {
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

func machineContainsVolumeName(machine *computev1alpha1.Machine, volumeName string) bool {
	for _, name := range computev1alpha1.MachineVolumeNames(machine) {
		if name == volumeName {
			return true
		}
	}
	return false
}

func (r *VolumeReconciler) validReferences(volume *storagev1alpha1.Volume, machine *computev1alpha1.Machine) bool {
	if !machineContainsVolumeName(machine, volume.Name) {
		return false
	}

	claimRef := volume.Spec.ClaimRef
	if claimRef == nil {
		return false
	}
	return claimRef.Name == machine.Name && claimRef.UID == machine.UID
}

func (r *VolumeReconciler) releaseVolume(ctx context.Context, volume *storagev1alpha1.Volume) error {
	baseVolume := volume.DeepCopy()
	volume.Spec.ClaimRef = nil
	return r.Patch(ctx, volume, client.MergeFrom(baseVolume))
}

func (r *VolumeReconciler) patchStatus(ctx context.Context, volume *storagev1alpha1.Volume, phase storagev1alpha1.VolumePhase) error {
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
		Watches(&source.Kind{Type: &computev1alpha1.Machine{}},
			handler.EnqueueRequestsFromMapFunc(func(obj client.Object) []ctrl.Request {
				machine := obj.(*computev1alpha1.Machine)

				volumeNames := computev1alpha1.MachineVolumeNames(machine)
				res := make([]ctrl.Request, 0, len(volumeNames))
				for _, volumeName := range volumeNames {
					res = append(res, ctrl.Request{
						NamespacedName: types.NamespacedName{
							Namespace: machine.Namespace,
							Name:      volumeName,
						},
					})
				}
				return res
			}),
		).
		Complete(r)
}
