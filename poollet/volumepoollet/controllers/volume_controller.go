// SPDX-FileCopyrightText: 2023 SAP SE or an SAP affiliate company and IronCore contributors
// SPDX-License-Identifier: Apache-2.0

package controllers

import (
	"context"
	"crypto/sha256"
	"encoding/hex"
	"fmt"

	"github.com/go-logr/logr"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"

	corev1alpha1 "github.com/ironcore-dev/ironcore/api/core/v1alpha1"
	storagev1alpha1 "github.com/ironcore-dev/ironcore/api/storage/v1alpha1"
	irimeta "github.com/ironcore-dev/ironcore/iri/apis/meta/v1alpha1"
	iriVolume "github.com/ironcore-dev/ironcore/iri/apis/volume"
	iri "github.com/ironcore-dev/ironcore/iri/apis/volume/v1alpha1"
	poolletutils "github.com/ironcore-dev/ironcore/poollet/common/utils"
	volumepoolletv1alpha1 "github.com/ironcore-dev/ironcore/poollet/volumepoollet/api/v1alpha1"
	"github.com/ironcore-dev/ironcore/poollet/volumepoollet/controllers/events"
	"github.com/ironcore-dev/ironcore/poollet/volumepoollet/vcm"
	utilsmaps "github.com/ironcore-dev/ironcore/utils/maps"

	ironcoreclient "github.com/ironcore-dev/ironcore/utils/client"
	"github.com/ironcore-dev/ironcore/utils/predicates"

	"github.com/ironcore-dev/controller-utils/clientutils"
	corev1 "k8s.io/api/core/v1"
	apierrors "k8s.io/apimachinery/pkg/api/errors"
	"k8s.io/apimachinery/pkg/api/resource"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apimachinery/pkg/types"
	"k8s.io/client-go/tools/record"
	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/builder"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/controller"
	"sigs.k8s.io/controller-runtime/pkg/controller/controllerutil"
	"sigs.k8s.io/controller-runtime/pkg/predicate"
)

type VolumeReconciler struct {
	record.EventRecorder
	client.Client
	Scheme *runtime.Scheme

	VolumeRuntime     iriVolume.RuntimeService
	VolumeRuntimeName string

	VolumeClassMapper vcm.VolumeClassMapper

	VolumePoolName string

	DownwardAPILabels      map[string]string
	DownwardAPIAnnotations map[string]string

	WatchFilterValue string

	MaxConcurrentReconciles int
}

func (r *VolumeReconciler) iriVolumeLabels(volume *storagev1alpha1.Volume) (map[string]string, error) {
	labels := map[string]string{
		volumepoolletv1alpha1.VolumeUIDLabel:       string(volume.UID),
		volumepoolletv1alpha1.VolumeNamespaceLabel: volume.Namespace,
		volumepoolletv1alpha1.VolumeNameLabel:      volume.Name,
	}
	apiLabels, err := poolletutils.PrepareDownwardAPILabels(volume, r.DownwardAPILabels, volumepoolletv1alpha1.VolumeDownwardAPIPrefix)
	if err != nil {
		return nil, err
	}
	labels = utilsmaps.AppendMap(labels, apiLabels)
	return labels, nil
}

func (r *VolumeReconciler) iriVolumeAnnotations(_ *storagev1alpha1.Volume) map[string]string {
	return map[string]string{}
}

func (r *VolumeReconciler) getIRIVolumeByID(ctx context.Context, volumeID types.UID) ([]*iri.Volume, error) {
	res, err := r.VolumeRuntime.ListVolumes(ctx, &iri.ListVolumesRequest{
		Filter: &iri.VolumeFilter{
			Id: string(volumeID),
		},
	})
	if err != nil {
		return nil, fmt.Errorf("error listing volumes by id: %w", err)
	}
	return res.Volumes, nil
}

//+kubebuilder:rbac:groups="",resources=events,verbs=create;patch
//+kubebuilder:rbac:groups="",resources=secrets,verbs=get;list;watch
//+kubebuilder:rbac:groups=storage.ironcore.dev,resources=volumes,verbs=get;list;watch;update;patch
//+kubebuilder:rbac:groups=storage.ironcore.dev,resources=volumes/status,verbs=get;update;patch
//+kubebuilder:rbac:groups=storage.ironcore.dev,resources=volumes/finalizers,verbs=update

func (r *VolumeReconciler) Reconcile(ctx context.Context, req ctrl.Request) (ctrl.Result, error) {
	log := ctrl.LoggerFrom(ctx)
	volume := &storagev1alpha1.Volume{}
	if err := r.Get(ctx, req.NamespacedName, volume); err != nil {
		if !apierrors.IsNotFound(err) {
			return ctrl.Result{}, fmt.Errorf("error getting volume %s: %w", req.NamespacedName, err)
		}
		return r.deleteGone(ctx, log, volume.UID)
	}
	return r.reconcileExists(ctx, log, volume)
}

func (r *VolumeReconciler) deleteGone(ctx context.Context, log logr.Logger, volumeUID types.UID) (ctrl.Result, error) {
	log.V(1).Info("Delete gone")

	log.V(1).Info("Listing iri volumes by Id")
	volumes, err := r.getIRIVolumeByID(ctx, volumeUID)
	if err != nil {
		return ctrl.Result{}, fmt.Errorf("error listing iri volumes by Id: %w", err)
	}

	ok, err := r.deleteIRIVolumes(ctx, log, volumes)
	if err != nil {
		return ctrl.Result{}, fmt.Errorf("error deleting iri volumes: %w", err)
	}
	if !ok {
		log.V(1).Info("Not all iri volumes are gone, requeueing")
		return ctrl.Result{Requeue: true}, nil
	}

	log.V(1).Info("Deleted gone")
	return ctrl.Result{}, nil
}

func (r *VolumeReconciler) deleteIRIVolumes(ctx context.Context, log logr.Logger, volumes []*iri.Volume) (bool, error) {
	var (
		errs                 []error
		deletingIRIVolumeIDs []string
	)

	for _, volume := range volumes {
		iriVolumeID := volume.Metadata.Id
		log := log.WithValues("IRIVolumeID", iriVolumeID)
		log.V(1).Info("Deleting iri volume")
		_, err := r.VolumeRuntime.DeleteVolume(ctx, &iri.DeleteVolumeRequest{
			VolumeId: iriVolumeID,
		})
		if err != nil {
			if status.Code(err) != codes.NotFound {
				errs = append(errs, fmt.Errorf("error deleting iri volume %s: %w", iriVolumeID, err))
			} else {
				log.V(1).Info("IRI Volume is already gone")
			}
		} else {
			log.V(1).Info("Issued iri volume deletion")
			deletingIRIVolumeIDs = append(deletingIRIVolumeIDs, iriVolumeID)
		}
	}

	switch {
	case len(errs) > 0:
		return false, fmt.Errorf("error(s) deleting iri volume(s): %v", errs)
	case len(deletingIRIVolumeIDs) > 0:
		log.V(1).Info("Volumes are in deletion", "DeletingIRIVolumeIDs", deletingIRIVolumeIDs)
		return false, nil
	default:
		log.V(1).Info("No iri volumes present")
		return true, nil
	}
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
	volumes, err := r.getIRIVolumeByID(ctx, volume.UID)
	if err != nil {
		return ctrl.Result{}, fmt.Errorf("error listing volumes by uid: %w", err)
	}

	ok, err := r.deleteIRIVolumes(ctx, log, volumes)
	if err != nil {
		return ctrl.Result{}, fmt.Errorf("error deleting iri volumes: %w", err)
	}
	if !ok {
		log.V(1).Info("Not all iri volumes are gone, requeueing")
		return ctrl.Result{Requeue: true}, nil
	}

	log.V(1).Info("Deleted all iri volumes, removing finalizer")
	if err := clientutils.PatchRemoveFinalizer(ctx, r.Client, volume, volumepoolletv1alpha1.VolumeFinalizer); err != nil {
		return ctrl.Result{}, fmt.Errorf("error removing finalizer: %w", err)
	}

	log.V(1).Info("Deleted")
	return ctrl.Result{}, nil
}

func getIRIVolumeClassCapabilities(volumeClass *storagev1alpha1.VolumeClass) *iri.VolumeClassCapabilities {
	tps := volumeClass.Capabilities.TPS()
	iops := volumeClass.Capabilities.IOPS()

	return &iri.VolumeClassCapabilities{
		Tps:  tps.Value(),
		Iops: iops.Value(),
	}
}

func (r *VolumeReconciler) prepareIRIVolumeMetadata(volume *storagev1alpha1.Volume, errs []error) (*irimeta.ObjectMetadata, []error) {
	labels, err := r.iriVolumeLabels(volume)
	if err != nil {
		errs = append(errs, fmt.Errorf("error preparing iri volume labels: %w", err))
	}
	return &irimeta.ObjectMetadata{
		Id:          string(volume.UID),
		Labels:      labels,
		Annotations: r.iriVolumeAnnotations(volume),
	}, errs
}

func (r *VolumeReconciler) prepareIRIVolumeClass(ctx context.Context, volume *storagev1alpha1.Volume, volumeClassName string) (string, bool, error) {
	volumeClass := &storagev1alpha1.VolumeClass{}
	volumeClassKey := client.ObjectKey{Name: volumeClassName}
	if err := r.Get(ctx, volumeClassKey, volumeClass); err != nil {
		err = fmt.Errorf("error getting volume class %s: %w", volumeClassKey, err)
		if !apierrors.IsNotFound(err) {
			return "", false, fmt.Errorf("error getting volume class %s: %w", volumeClassName, err)
		}

		r.Eventf(volume, corev1.EventTypeNormal, events.VolumeClassNotReady, "Volume class %s not found", volumeClassName)
		return "", false, nil
	}

	caps := getIRIVolumeClassCapabilities(volumeClass)

	class, _, err := r.VolumeClassMapper.GetVolumeClassFor(ctx, volumeClassName, caps)
	if err != nil {
		return "", false, fmt.Errorf("error getting matching volume class: %w", err)
	}
	return class.Name, true, nil
}

func (r *VolumeReconciler) prepareIRIVolumeEncryption(ctx context.Context, volume *storagev1alpha1.Volume) (*iri.EncryptionSpec, bool, error) {
	encryption := volume.Spec.Encryption
	if encryption == nil {
		return nil, true, nil
	}

	encryptionSecret := &corev1.Secret{}
	encryptionSecretKey := client.ObjectKey{Name: encryption.SecretRef.Name, Namespace: volume.Namespace}
	if err := r.Get(ctx, encryptionSecretKey, encryptionSecret); err != nil {
		err = fmt.Errorf("error getting volume encryption secret %s: %w", encryptionSecretKey, err)
		if !apierrors.IsNotFound(err) {
			return nil, false, fmt.Errorf("error getting volume encryption secret %s: %w", encryption.SecretRef.Name, err)
		}

		r.Eventf(volume, corev1.EventTypeNormal, events.VolumeEncryptionSecretNotReady, "Volume encryption secret %s not found", encryption.SecretRef.Name)
		return nil, false, nil
	}

	return &iri.EncryptionSpec{
		SecretData: encryptionSecret.Data,
	}, true, nil
}

func (r *VolumeReconciler) prepareIRIVolumeResources(resources corev1alpha1.ResourceList) *iri.VolumeResources {
	storageBytes := resources.Storage().Value()

	return &iri.VolumeResources{
		StorageBytes: storageBytes,
	}
}

func (r *VolumeReconciler) prepareIRIVolume(ctx context.Context, log logr.Logger, volume *storagev1alpha1.Volume) (*iri.Volume, bool, error) {
	var (
		ok   = true
		errs []error
	)

	log.V(1).Info("Getting volume class")
	class, classOK, err := r.prepareIRIVolumeClass(ctx, volume, volume.Spec.VolumeClassRef.Name)
	switch {
	case err != nil:
		errs = append(errs, fmt.Errorf("error preparing iri volume class: %w", err))
	case !classOK:
		ok = false
	}

	log.V(1).Info("Getting encryption secret")
	encryption, encryptionOK, err := r.prepareIRIVolumeEncryption(ctx, volume)
	switch {
	case err != nil:
		errs = append(errs, fmt.Errorf("error preparing iri volume class: %w", err))
	case !encryptionOK:
		ok = false
	}

	resources := r.prepareIRIVolumeResources(volume.Spec.Resources)
	metadata, errs := r.prepareIRIVolumeMetadata(volume, errs)

	if len(errs) > 0 {
		return nil, false, fmt.Errorf("error(s) preparing iri volume: %v", errs)
	}
	if !ok {
		return nil, false, nil
	}

	return &iri.Volume{
		Metadata: metadata,
		Spec: &iri.VolumeSpec{
			Image:      volume.Spec.Image,
			Class:      class,
			Resources:  resources,
			Encryption: encryption,
		},
	}, true, nil
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
	modified, err = ironcoreclient.PatchEnsureNoReconcileAnnotation(ctx, r.Client, volume)
	if err != nil {
		return ctrl.Result{}, fmt.Errorf("error ensuring no reconcile annotation: %w", err)
	}
	if modified {
		log.V(1).Info("Removed reconcile annotation, requeueing")
		return ctrl.Result{Requeue: true}, nil
	}

	log.V(1).Info("Listing volumes")
	res, err := r.VolumeRuntime.ListVolumes(ctx, &iri.ListVolumesRequest{
		Filter: &iri.VolumeFilter{
			Id: string(volume.UID),
		},
	})
	if err != nil {
		return ctrl.Result{}, fmt.Errorf("error listing volumes: %w", err)
	}

	switch len(res.Volumes) {
	case 0:
		return r.create(ctx, log, volume)
	case 1:
		iriVolume := res.Volumes[0]
		if err := r.update(ctx, log, volume, iriVolume); err != nil {
			return ctrl.Result{}, fmt.Errorf("error updating volume: %w", err)
		}

		if err := r.updateStatus(ctx, log, volume, iriVolume); err != nil {
			return ctrl.Result{}, fmt.Errorf("error updating volume status: %w", err)
		}
		return ctrl.Result{}, nil
	default:
		panic("unhandled multiple volumes")
	}
}

func (r *VolumeReconciler) create(ctx context.Context, log logr.Logger, volume *storagev1alpha1.Volume) (ctrl.Result, error) {
	log.V(1).Info("Create")

	log.V(1).Info("Preparing iri volume")
	iriVolume, ok, err := r.prepareIRIVolume(ctx, log, volume)
	if err != nil {
		return ctrl.Result{}, fmt.Errorf("error preparing iri volume: %w", err)
	}
	if !ok {
		log.V(1).Info("IRI volume is not yet ready to be prepared")
		return ctrl.Result{}, nil
	}

	log.V(1).Info("Creating volume", "volume before ", iriVolume)
	res, err := r.VolumeRuntime.CreateVolume(ctx, &iri.CreateVolumeRequest{
		Volume: iriVolume,
	})
	if err != nil {
		return ctrl.Result{}, fmt.Errorf("error creating volume: %w", err)
	}

	iriVolume = res.Volume

	volumeID := iriVolume.Metadata.Id
	log = log.WithValues("VolumeID", volumeID)
	log.V(1).Info("Created", "volume after ", iriVolume)

	log.V(1).Info("Updating status")
	if err := r.updateStatus(ctx, log, volume, iriVolume); err != nil {
		return ctrl.Result{}, fmt.Errorf("error updating volume status: %w", err)
	}

	log.V(1).Info("Created", "volume after ", iriVolume)
	return ctrl.Result{}, nil
}

func (r *VolumeReconciler) update(ctx context.Context, log logr.Logger, volume *storagev1alpha1.Volume, iriVolume *iri.Volume) error {
	storageBytes := volume.Spec.Resources.Storage().Value()
	oldStorageBytes := iriVolume.Spec.Resources.StorageBytes
	if storageBytes != oldStorageBytes {
		log.V(1).Info("Expanding volume", "StorageBytes", storageBytes, "OldStorageBytes", oldStorageBytes)
		if _, err := r.VolumeRuntime.ExpandVolume(ctx, &iri.ExpandVolumeRequest{
			VolumeId: iriVolume.Metadata.Id,
			Resources: &iri.VolumeResources{
				StorageBytes: storageBytes,
			},
		}); err != nil {
			return fmt.Errorf("failed to expand volume: %w", err)
		}
	}

	return nil
}

func (r *VolumeReconciler) volumeSecretName(volumeName, volumeHandle string) string {
	sum := sha256.Sum256([]byte(fmt.Sprintf("%s/%s", volumeName, volumeHandle)))
	return hex.EncodeToString(sum[:])[:63]
}

var iriVolumeStateToVolumeState = map[iri.VolumeState]storagev1alpha1.VolumeState{
	iri.VolumeState_VOLUME_PENDING:   storagev1alpha1.VolumeStatePending,
	iri.VolumeState_VOLUME_AVAILABLE: storagev1alpha1.VolumeStateAvailable,
	iri.VolumeState_VOLUME_ERROR:     storagev1alpha1.VolumeStateError,
}

func (r *VolumeReconciler) convertIRIVolumeState(iriState iri.VolumeState) (storagev1alpha1.VolumeState, error) {
	if res, ok := iriVolumeStateToVolumeState[iriState]; ok {
		return res, nil
	}
	return "", fmt.Errorf("unknown volume state %v", iriState)
}

func (r *VolumeReconciler) updateStatus(ctx context.Context, log logr.Logger, volume *storagev1alpha1.Volume, iriVolume *iri.Volume) error {
	var access *storagev1alpha1.VolumeAccess

	if iriVolume.Status.State == iri.VolumeState_VOLUME_AVAILABLE {
		if iriAccess := iriVolume.Status.Access; iriAccess != nil {
			var secretRef *corev1.LocalObjectReference

			if iriAccess.SecretData != nil {
				log.V(1).Info("Applying volume secret")
				volumeSecret := &corev1.Secret{
					TypeMeta: metav1.TypeMeta{
						APIVersion: corev1.SchemeGroupVersion.String(),
						Kind:       "Secret",
					},
					ObjectMeta: metav1.ObjectMeta{
						Namespace: volume.Namespace,
						Name:      r.volumeSecretName(volume.Name, iriAccess.Handle),
						Labels: map[string]string{
							volumepoolletv1alpha1.VolumeUIDLabel: string(volume.UID),
						},
					},
					Data: iriAccess.SecretData,
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
						volumepoolletv1alpha1.VolumeUIDLabel: string(volume.UID),
					},
				); err != nil {
					return fmt.Errorf("error deleting any corresponding volume secret: %w", err)
				}
			}

			access = &storagev1alpha1.VolumeAccess{
				SecretRef:        secretRef,
				Driver:           iriAccess.Driver,
				Handle:           iriAccess.Handle,
				VolumeAttributes: iriAccess.Attributes,
			}
		}
	}

	base := volume.DeepCopy()
	now := metav1.Now()

	volumeID := poolletutils.MakeID(r.VolumeRuntimeName, iriVolume.Metadata.Id)

	volume.Status.Access = access
	newState, err := r.convertIRIVolumeState(iriVolume.Status.State)
	if err != nil {
		return err
	}
	if newState != volume.Status.State {
		volume.Status.LastStateTransitionTime = &now
	}
	volume.Status.State = newState
	volume.Status.VolumeID = volumeID.String()
	if iriVolume.Status.Resources != nil {
		volume.Status.Resources = corev1alpha1.ResourceList{
			corev1alpha1.ResourceStorage: *resource.NewQuantity(iriVolume.Status.Resources.StorageBytes, resource.DecimalSI),
		}
	}

	if err := r.Status().Patch(ctx, volume, client.MergeFrom(base)); err != nil {
		return fmt.Errorf("error patching volume status: %w", err)
	}
	return nil
}

func (r *VolumeReconciler) SetupWithManager(mgr ctrl.Manager) error {
	log := ctrl.Log.WithName("volumepoollet")

	return ctrl.NewControllerManagedBy(mgr).
		For(
			&storagev1alpha1.Volume{},
			builder.WithPredicates(
				VolumeRunsInVolumePoolPredicate(r.VolumePoolName),
				predicates.ResourceHasFilterLabel(log, r.WatchFilterValue),
				predicates.ResourceIsNotExternallyManaged(log),
			),
		).
		WithOptions(
			controller.Options{
				MaxConcurrentReconciles: r.MaxConcurrentReconciles,
			}).
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
