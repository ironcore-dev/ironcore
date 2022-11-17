// Copyright 2022 OnMetal authors
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

package controllers

import (
	"context"
	"crypto/sha256"
	"encoding/hex"
	"fmt"

	"github.com/go-logr/logr"
	"github.com/onmetal/controller-utils/clientutils"
	storagev1alpha1 "github.com/onmetal/onmetal-api/api/storage/v1alpha1"
	onmetalapiclient "github.com/onmetal/onmetal-api/apiutils/client"
	"github.com/onmetal/onmetal-api/apiutils/predicates"
	ori "github.com/onmetal/onmetal-api/ori/apis/storage/v1alpha1"
	volumepoolletv1alpha1 "github.com/onmetal/onmetal-api/volumepoollet/api/v1alpha1"
	"github.com/onmetal/onmetal-api/volumepoollet/controllers/events"
	"github.com/onmetal/onmetal-api/volumepoollet/vcm"
	"github.com/onmetal/onmetal-api/volumepoollet/vleg"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
	corev1 "k8s.io/api/core/v1"
	apierrors "k8s.io/apimachinery/pkg/api/errors"
	"k8s.io/apimachinery/pkg/api/resource"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/client-go/tools/record"
	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/builder"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/controller/controllerutil"
	"sigs.k8s.io/controller-runtime/pkg/manager"
	"sigs.k8s.io/controller-runtime/pkg/predicate"
)

type VolumeReconciler struct {
	record.EventRecorder
	client.Client
	Scheme *runtime.Scheme

	VolumeRuntime ori.VolumeRuntimeClient

	VolumeClassMapper vcm.VolumeClassMapper

	VolumePoolName   string
	WatchFilterValue string
}

func (r *VolumeReconciler) Reconcile(ctx context.Context, req ctrl.Request) (ctrl.Result, error) {
	log := ctrl.LoggerFrom(ctx)
	volume := &storagev1alpha1.Volume{}
	if err := r.Get(ctx, req.NamespacedName, volume); err != nil {
		if !apierrors.IsNotFound(err) {
			return ctrl.Result{}, fmt.Errorf("error getting volume %s: %w", req.NamespacedName, err)
		}
		return r.deleteGone(ctx, log, req.NamespacedName)
	}
	return r.reconcileExists(ctx, log, volume)
}

func (r *VolumeReconciler) deleteGone(ctx context.Context, log logr.Logger, volumeKey client.ObjectKey) (ctrl.Result, error) {
	log.V(1).Info("Delete gone")

	log.V(1).Info("Listing volumes matching key")
	res, err := r.VolumeRuntime.ListVolumes(ctx, &ori.ListVolumesRequest{
		Filter: &ori.VolumeFilter{
			LabelSelector: map[string]string{
				volumepoolletv1alpha1.VolumeNamespaceLabel: volumeKey.Namespace,
				volumepoolletv1alpha1.VolumeNameLabel:      volumeKey.Name,
			},
		},
	})
	if err != nil {
		return ctrl.Result{}, fmt.Errorf("error listing volumes matching key: %w", err)
	}

	log.V(1).Info("Listed volumes matching key", "NoOfVolumes", len(res.Volumes))
	var errs []error
	for _, volume := range res.Volumes {
		log := log.WithValues("VolumeID", volume.Id)
		log.V(1).Info("Deleting volume")
		_, err := r.VolumeRuntime.DeleteVolume(ctx, &ori.DeleteVolumeRequest{
			VolumeId: volume.Id,
		})
		if err != nil {
			if status.Code(err) != codes.NotFound {
				errs = append(errs, fmt.Errorf("error deleting volume %s: %w", volume.Id, err))
			} else {
				log.V(1).Info("Volume is already gone")
			}
		}
	}

	if len(errs) > 0 {
		return ctrl.Result{}, fmt.Errorf("error(s) deleting volume(s): %v", errs)
	}

	log.V(1).Info("Deleted gone")
	return ctrl.Result{}, nil
}

func (r *VolumeReconciler) reconcileExists(ctx context.Context, log logr.Logger, volume *storagev1alpha1.Volume) (ctrl.Result, error) {
	if !volume.DeletionTimestamp.IsZero() {
		return r.delete(ctx, log, volume)
	}
	return r.reconcile(ctx, log, volume)
}

func (r *VolumeReconciler) delete(ctx context.Context, log logr.Logger, volume *storagev1alpha1.Volume) (ctrl.Result, error) {
	log.V(1).Info("Delete")

	if !controllerutil.ContainsFinalizer(volume, volumepoolletv1alpha1.VolumeFinalizer) {
		log.V(1).Info("No finalizer present, nothing to do")
		return ctrl.Result{}, nil
	}

	log.V(1).Info("Finalizer present")

	log.V(1).Info("Listing volumes")
	res, err := r.VolumeRuntime.ListVolumes(ctx, &ori.ListVolumesRequest{
		Filter: &ori.VolumeFilter{
			LabelSelector: map[string]string{
				volumepoolletv1alpha1.VolumeUIDLabel: string(volume.UID),
			},
		},
	})
	if err != nil {
		return ctrl.Result{}, fmt.Errorf("error listing volumes: %w", err)
	}

	log.V(1).Info("Listed volumes", "NoOfVolumes", len(res.Volumes))
	var errs []error
	for _, volume := range res.Volumes {
		log := log.WithValues("VolumeID", volume.Id)
		log.V(1).Info("Deleting volume")
		_, err := r.VolumeRuntime.DeleteVolume(ctx, &ori.DeleteVolumeRequest{
			VolumeId: volume.Id,
		})
		if err != nil {
			if status.Code(err) != codes.NotFound {
				errs = append(errs, fmt.Errorf("error deleting volume %s: %w", volume.Id, err))
			} else {
				log.V(1).Info("Volume is already gone")
			}
		}
	}

	if len(errs) > 0 {
		return ctrl.Result{}, fmt.Errorf("error(s) deleting volume(s): %v", errs)
	}

	log.V(1).Info("Deleted all runtime volumes, removing finalizer")
	if err := clientutils.PatchRemoveFinalizer(ctx, r.Client, volume, volumepoolletv1alpha1.VolumeFinalizer); err != nil {
		return ctrl.Result{}, fmt.Errorf("error removing finalizer: %w", err)
	}

	log.V(1).Info("Deleted")
	return ctrl.Result{}, nil
}

func getORIVolumeClassCapabilities(volumeClass *storagev1alpha1.VolumeClass) (*ori.VolumeClassCapabilities, error) {
	tps := volumeClass.Capabilities.Name(storagev1alpha1.ResourceTPS, resource.DecimalSI)
	iops := volumeClass.Capabilities.Name(storagev1alpha1.ResourceIOPS, resource.DecimalSI)

	return &ori.VolumeClassCapabilities{
		Tps:  tps.Value(),
		Iops: iops.Value(),
	}, nil
}

func (r *VolumeReconciler) getORIVolumeClass(ctx context.Context, volumeClassName string) (string, error) {
	volumeClass := &storagev1alpha1.VolumeClass{}
	volumeClassKey := client.ObjectKey{Name: volumeClassName}
	if err := r.Get(ctx, volumeClassKey, volumeClass); err != nil {
		err = fmt.Errorf("error getting volume class %s: %w", volumeClassKey, err)
		if !apierrors.IsNotFound(err) {
			return "", err
		}

		return "", NewDependencyNotReadyError(
			storagev1alpha1.Resource("volumeclasses"),
			volumeClassKey.Name,
			err,
		)
	}

	caps, err := getORIVolumeClassCapabilities(volumeClass)
	if err != nil {
		return "", fmt.Errorf("error getting ori volume class capabilities: %w", err)
	}

	class, err := r.VolumeClassMapper.GetVolumeClassFor(ctx, volumeClassName, caps)
	if err != nil {
		return "", fmt.Errorf("error getting matching volume class: %w", err)
	}
	return class.Name, nil
}

func (r *VolumeReconciler) getORIVolumeMetadata(volume *storagev1alpha1.Volume) *ori.VolumeMetadata {
	return &ori.VolumeMetadata{
		Namespace: volume.Namespace,
		Name:      volume.Name,
		Uid:       string(volume.UID),
	}
}

func (r *VolumeReconciler) getVolumeConfig(ctx context.Context, log logr.Logger, volume *storagev1alpha1.Volume) (*ori.VolumeConfig, error) {
	log.V(1).Info("Getting volume class")
	class, err := r.getORIVolumeClass(ctx, volume.Spec.VolumeClassRef.Name)
	if err != nil {
		if !IsDependencyNotReadyError(err) {
			r.Eventf(volume, corev1.EventTypeWarning, events.ErrorGettingVolumeClass, "Error getting volume class: %v", err)
			return nil, fmt.Errorf("error getting volume class: %w", err)
		}

		r.Eventf(volume, corev1.EventTypeNormal, events.VolumeClassNotReady, "Volume class not ready: %v", err)
		return nil, fmt.Errorf("volume class not ready: %w", err)
	}

	metadata := r.getORIVolumeMetadata(volume)

	return &ori.VolumeConfig{
		Metadata:    metadata,
		Image:       volume.Spec.Image,
		Class:       class,
		Annotations: map[string]string{},
		Labels: map[string]string{
			volumepoolletv1alpha1.VolumeUIDLabel:       string(volume.UID),
			volumepoolletv1alpha1.VolumeNamespaceLabel: volume.Namespace,
			volumepoolletv1alpha1.VolumeNameLabel:      volume.Name,
		},
	}, nil
}

func (r *VolumeReconciler) reconcile(ctx context.Context, log logr.Logger, volume *storagev1alpha1.Volume) (ctrl.Result, error) {
	log.V(1).Info("Reconcile")

	log.V(1).Info("Ensuring finalizer")
	modified, err := clientutils.PatchEnsureFinalizer(ctx, r.Client, volume, volumepoolletv1alpha1.VolumeFinalizer)
	if err != nil {
		return ctrl.Result{}, fmt.Errorf("error ensuring finalizer: %w", err)
	}
	if modified {
		log.V(1).Info("Added finalizer, requeueing")
		return ctrl.Result{Requeue: true}, nil
	}
	log.V(1).Info("Finalizer is present")

	log.V(1).Info("Ensuring no reconcile annotation")
	modified, err = onmetalapiclient.PatchEnsureNoReconcileAnnotation(ctx, r.Client, volume)
	if err != nil {
		return ctrl.Result{}, fmt.Errorf("error ensuring no reconcile annotation: %w", err)
	}
	if modified {
		log.V(1).Info("Removed reconcile annotation, requeueing")
		return ctrl.Result{Requeue: true}, nil
	}

	log.V(1).Info("Listing volumes")
	res, err := r.VolumeRuntime.ListVolumes(ctx, &ori.ListVolumesRequest{
		Filter: &ori.VolumeFilter{
			LabelSelector: map[string]string{
				volumepoolletv1alpha1.VolumeUIDLabel: string(volume.UID),
			},
		},
	})
	if err != nil {
		return ctrl.Result{}, fmt.Errorf("error listing volumes: %w", err)
	}

	switch len(res.Volumes) {
	case 0:
		return r.create(ctx, log, volume)
	case 1:
		oriVolume := res.Volumes[0]
		if err := r.updateStatus(ctx, log, volume, oriVolume.Id); err != nil {
			return ctrl.Result{}, fmt.Errorf("error updating volume status: %w", err)
		}
		return ctrl.Result{}, nil
	default:
		panic("unhandled multiple volumes")
	}
}

func (r *VolumeReconciler) create(ctx context.Context, log logr.Logger, volume *storagev1alpha1.Volume) (ctrl.Result, error) {
	log.V(1).Info("Create")

	log.V(1).Info("Getting volume config")
	volumeConfig, err := r.getVolumeConfig(ctx, log, volume)
	if err != nil {
		err = fmt.Errorf("error getting volume config: %w", err)
		return ctrl.Result{}, IgnoreDependencyNotReadyError(err)
	}

	log.V(1).Info("Creating volume")
	res, err := r.VolumeRuntime.CreateVolume(ctx, &ori.CreateVolumeRequest{
		Config: volumeConfig,
	})
	if err != nil {
		return ctrl.Result{}, fmt.Errorf("error creating volume: %w", err)
	}
	volumeID := res.Volume.Id
	log = log.WithValues("VolumeID", volumeID)
	log.V(1).Info("Created")

	log.V(1).Info("Updating status")
	if err := r.updateStatus(ctx, log, volume, volumeID); err != nil {
		return ctrl.Result{}, fmt.Errorf("error updating volume status: %w", err)
	}

	log.V(1).Info("Created")
	return ctrl.Result{}, nil
}

func (r *VolumeReconciler) volumeSecretName(volumeName string, volumeHandle string) string {
	sum := sha256.Sum256([]byte(fmt.Sprintf("%s/%s", volumeName, volumeHandle)))
	return hex.EncodeToString(sum[:])[:63]
}

var oriVolumeStateToComputeV1Alpha1VolumeState = map[ori.VolumeState]storagev1alpha1.VolumeState{
	ori.VolumeState_VOLUME_PENDING:   storagev1alpha1.VolumeStatePending,
	ori.VolumeState_VOLUME_AVAILABLE: storagev1alpha1.VolumeStateAvailable,
	ori.VolumeState_VOLUME_ERROR:     storagev1alpha1.VolumeStateError,
}

func (r *VolumeReconciler) oriVolumeStateToComputeV1Alpha1VolumeState(oriState ori.VolumeState) storagev1alpha1.VolumeState {
	return oriVolumeStateToComputeV1Alpha1VolumeState[oriState]
}

func (r *VolumeReconciler) updateStatus(
	ctx context.Context,
	log logr.Logger,
	volume *storagev1alpha1.Volume,
	volumeID string,
) error {
	log.V(1).Info("Getting runtime status")
	res, err := r.VolumeRuntime.VolumeStatus(ctx, &ori.VolumeStatusRequest{
		VolumeId: volumeID,
	})
	if err != nil {
		return fmt.Errorf("error getting volume status: %w", err)
	}

	runtimeStatus := res.Status

	var access *storagev1alpha1.VolumeAccess

	if runtimeStatus.State == ori.VolumeState_VOLUME_AVAILABLE {
		if oriAccess := runtimeStatus.Access; oriAccess != nil {
			var secretRef *corev1.LocalObjectReference

			if oriAccess.SecretData != nil {
				log.V(1).Info("Applying volume secret")
				volumeSecret := &corev1.Secret{
					TypeMeta: metav1.TypeMeta{
						APIVersion: corev1.SchemeGroupVersion.String(),
						Kind:       "Secret",
					},
					ObjectMeta: metav1.ObjectMeta{
						Namespace: volume.Namespace,
						Name:      r.volumeSecretName(volume.Name, oriAccess.Handle),
						Labels: map[string]string{
							volumepoolletv1alpha1.VolumeUIDLabel: string(volume.UID[:63]),
						},
					},
					Data: oriAccess.SecretData,
				}
				_ = ctrl.SetControllerReference(volume, volumeSecret, r.Scheme)
				if err := r.Patch(ctx, volumeSecret, client.Apply, client.FieldOwner(volumepoolletv1alpha1.FieldOwner)); err != nil {
					return fmt.Errorf("error applying volume secret: %w", err)
				}
				secretRef = &corev1.LocalObjectReference{Name: volumeSecret.Name}
			} else {
				log.V(1).Info("Deleting any corresponding volume secret")
				if err := r.DeleteAllOf(ctx, &corev1.Secret{},
					client.InNamespace(volume.Namespace),
					client.MatchingLabels{
						volumepoolletv1alpha1.VolumeUIDLabel: string(volume.UID[:63]),
					},
				); err != nil {
					return fmt.Errorf("error deleting any corresponding volume secret: %w", err)
				}
			}

			access = &storagev1alpha1.VolumeAccess{
				SecretRef:        secretRef,
				Driver:           oriAccess.Driver,
				Handle:           oriAccess.Handle,
				VolumeAttributes: oriAccess.Attributes,
			}
		}

	}

	base := volume.DeepCopy()
	now := metav1.Now()

	volume.Status.Access = access
	newState := r.oriVolumeStateToComputeV1Alpha1VolumeState(runtimeStatus.State)
	if newState != volume.Status.State {
		volume.Status.LastStateTransitionTime = &now
	}
	volume.Status.State = newState

	if err := r.Status().Patch(ctx, volume, client.MergeFrom(base)); err != nil {
		return fmt.Errorf("error patching volume status: %w", err)
	}
	return nil
}

func (r *VolumeReconciler) SetupWithManager(mgr ctrl.Manager) error {
	log := ctrl.Log.WithName("volumepoollet")

	mapper := vcm.NewGeneric(r.VolumeRuntime, vcm.GenericOptions{})

	if err := mgr.Add(mapper); err != nil {
		return fmt.Errorf("error adding volume class mapper: %w", err)
	}

	gen := vleg.NewGeneric(r.VolumeRuntime, vleg.GenericOptions{})

	if err := mgr.Add(gen); err != nil {
		return fmt.Errorf("error adding volume lifecycle event generator: %w", err)
	}

	if err := mgr.Add(manager.RunnableFunc(func(ctx context.Context) error {
		log := log.WithName("volume-lifecycle-event")
		for {
			select {
			case <-ctx.Done():
				return nil
			case evt := <-gen.Watch():
				log = log.WithValues(
					"EventType", evt.Type,
					"VolumeID", evt.ID,
					"Namespace", evt.Metadata.Namespace,
					"Name", evt.Metadata.Name,
					"UID", evt.Metadata.UID,
				)
				log.V(5).Info("Received lifecycle event")

				volume := &storagev1alpha1.Volume{
					ObjectMeta: metav1.ObjectMeta{
						Namespace: evt.Metadata.Namespace,
						Name:      evt.Metadata.Name,
					},
				}
				if err := onmetalapiclient.PatchAddReconcileAnnotation(ctx, r.Client, volume); client.IgnoreNotFound(err) != nil {
					log.Error(err, "Error adding reconcile annotation")
				}
			}
		}
	})); err != nil {
		return fmt.Errorf("error adding volume lifecycle event handler: %w", err)
	}

	return ctrl.NewControllerManagedBy(mgr).
		For(
			&storagev1alpha1.Volume{},
			builder.WithPredicates(
				VolumeRunsInVolumePoolPredicate(r.VolumePoolName),
				predicates.ResourceHasFilterLabel(log, r.WatchFilterValue),
				predicates.ResourceIsNotExternallyManaged(log),
			),
		).
		Complete(r)
}

func VolumeRunsInVolumePool(volume *storagev1alpha1.Volume, volumePoolName string) bool {
	volumePoolRef := volume.Spec.VolumePoolRef
	if volumePoolRef == nil {
		return false
	}

	return volumePoolRef.Name == volumePoolName
}

func VolumeRunsInVolumePoolPredicate(volumePoolName string) predicate.Predicate {
	return predicate.NewPredicateFuncs(func(object client.Object) bool {
		volume := object.(*storagev1alpha1.Volume)
		return VolumeRunsInVolumePool(volume, volumePoolName)
	})
}
