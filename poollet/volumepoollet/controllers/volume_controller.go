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
	storageclient "github.com/ironcore-dev/ironcore/internal/client/storage"
	irimeta "github.com/ironcore-dev/ironcore/iri/apis/meta/v1alpha1"
	iriVolume "github.com/ironcore-dev/ironcore/iri/apis/volume"
	iri "github.com/ironcore-dev/ironcore/iri/apis/volume/v1alpha1"
	poolletutils "github.com/ironcore-dev/ironcore/poollet/common/utils"
	volumepoolletv1alpha1 "github.com/ironcore-dev/ironcore/poollet/volumepoollet/api/v1alpha1"
	"github.com/ironcore-dev/ironcore/poollet/volumepoollet/controllers/events"
	"github.com/ironcore-dev/ironcore/poollet/volumepoollet/vcm"
	utilsmaps "github.com/ironcore-dev/ironcore/utils/maps"

	utilclient "github.com/ironcore-dev/ironcore/utils/client"
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
	"sigs.k8s.io/controller-runtime/pkg/handler"
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

func (r *VolumeReconciler) listIRIVolumesByKey(ctx context.Context, volumeKey client.ObjectKey) ([]*iri.Volume, error) {
	res, err := r.VolumeRuntime.ListVolumes(ctx, &iri.ListVolumesRequest{
		Filter: &iri.VolumeFilter{
			LabelSelector: map[string]string{
				volumepoolletv1alpha1.VolumeNamespaceLabel: volumeKey.Namespace,
				volumepoolletv1alpha1.VolumeNameLabel:      volumeKey.Name,
			},
		},
	})
	if err != nil {
		return nil, fmt.Errorf("error listing volumes by key: %w", err)
	}
	volumes := res.Volumes
	return volumes, nil
}

func (r *VolumeReconciler) listIRIVolumesByUID(ctx context.Context, volumeUID types.UID) ([]*iri.Volume, error) {
	res, err := r.VolumeRuntime.ListVolumes(ctx, &iri.ListVolumesRequest{
		Filter: &iri.VolumeFilter{
			LabelSelector: map[string]string{
				volumepoolletv1alpha1.VolumeUIDLabel: string(volumeUID),
			},
		},
	})
	if err != nil {
		return nil, fmt.Errorf("error listing volumes by uid: %w", err)
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
		return r.deleteGone(ctx, log, req.NamespacedName)
	}
	return r.reconcileExists(ctx, log, volume)
}

func (r *VolumeReconciler) deleteGone(ctx context.Context, log logr.Logger, volumeKey client.ObjectKey) (ctrl.Result, error) {
	log.V(1).Info("Delete gone")

	log.V(1).Info("Listing iri volumes by key")
	volumes, err := r.listIRIVolumesByKey(ctx, volumeKey)
	if err != nil {
		return ctrl.Result{}, fmt.Errorf("error listing iri volumes by key: %w", err)
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
	volumes, err := r.listIRIVolumesByUID(ctx, volume.UID)
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

func (r *VolumeReconciler) prepareIRIVolumeResources(resources corev1alpha1.ResourceList) *iri.VolumeResources {
	storageBytes := resources.Storage().Value()

	return &iri.VolumeResources{
		StorageBytes: storageBytes,
	}
}

func (r *VolumeReconciler) prepareIRIVolumeSnapshotDataSource(log logr.Logger, volume *storagev1alpha1.Volume, volumeSnapshot *storagev1alpha1.VolumeSnapshot) (*iri.VolumeDataSource, bool, error) {
	log.V(1).Info("Processing volume snapshot for data source")

	switch volumeSnapshot.Status.State {
	case storagev1alpha1.VolumeSnapshotStateFailed:
		r.Eventf(volume, corev1.EventTypeWarning, events.VolumeSnapshotFailed,
			"VolumeSnapshot %s is in failed state", volumeSnapshot.Name)
		return nil, false, fmt.Errorf("volume snapshot %s is in failed state", volumeSnapshot.Name)

	case storagev1alpha1.VolumeSnapshotStatePending:
		log.V(1).Info("Volume snapshot is pending, waiting for state change", "snapshot", volumeSnapshot.Name)
		return nil, false, nil

	case storagev1alpha1.VolumeSnapshotStateReady:
		if volumeSnapshot.Status.SnapshotID == "" {
			log.V(1).Info("Volume snapshot is ready but has no snapshot ID, waiting for ID", "snapshot", volumeSnapshot.Name)
			return nil, false, nil
		}
		return &iri.VolumeDataSource{
			SnapshotDataSource: &iri.SnapshotDataSource{
				SnapshotId: volumeSnapshot.Status.SnapshotID,
			},
		}, true, nil

	default:
		return nil, false, fmt.Errorf("unknown volume snapshot state %v", volumeSnapshot.Status.State)
	}
}

func (r *VolumeReconciler) prepareIRIVolumeDataSource(log logr.Logger, volume *storagev1alpha1.Volume, volumeSnapshot *storagev1alpha1.VolumeSnapshot) (*iri.VolumeDataSource, bool, error) {
	if volume.Spec.VolumeSnapshotRef != nil {
		return r.prepareIRIVolumeSnapshotDataSource(log, volume, volumeSnapshot)
	}

	if volume.Spec.OSImage != nil && *volume.Spec.OSImage != "" {
		return &iri.VolumeDataSource{
			ImageDataSource: &iri.ImageDataSource{
				Image: *volume.Spec.OSImage,
			},
		}, true, nil
	}

	return nil, true, nil
}

func (r *VolumeReconciler) prepareIRIVolumeSpecEncryption(ctx context.Context, volume *storagev1alpha1.Volume) (*iri.EncryptionSpec, bool, error) {
	secretName := volume.Spec.Encryption.SecretRef.Name
	if secretName == "" {
		return nil, false, fmt.Errorf("volume encryption secret name is empty")
	}

	encryptionSecret := &corev1.Secret{}
	encryptionSecretKey := client.ObjectKey{Name: secretName, Namespace: volume.Namespace}
	if err := r.Get(ctx, encryptionSecretKey, encryptionSecret); err != nil {
		if apierrors.IsNotFound(err) {
			r.Eventf(volume, corev1.EventTypeNormal, events.VolumeEncryptionSecretNotReady, "Volume encryption secret %s not found", secretName)
			return nil, false, nil
		}
		return nil, false, fmt.Errorf("error getting volume encryption secret %s: %w", secretName, err)
	}

	return &iri.EncryptionSpec{
		SecretData: encryptionSecret.Data,
	}, true, nil
}

func (r *VolumeReconciler) prepareIRIVolumeInheritedEncryption(ctx context.Context, volumeSnapshot *storagev1alpha1.VolumeSnapshot) (*iri.EncryptionSpec, error) {
	if volumeSnapshot.Spec.VolumeRef == nil {
		return nil, nil
	}

	sourceVolume := &storagev1alpha1.Volume{}
	sourceVolumeKey := client.ObjectKey{Namespace: volumeSnapshot.Namespace, Name: volumeSnapshot.Spec.VolumeRef.Name}
	if err := r.Get(ctx, sourceVolumeKey, sourceVolume); err != nil {
		if apierrors.IsNotFound(err) {
			return nil, fmt.Errorf("source volume %s not found, cannot determine encryption inheritance", volumeSnapshot.Spec.VolumeRef.Name)
		}
		return nil, fmt.Errorf("error getting source volume %s: %w", sourceVolumeKey, err)
	}

	if sourceVolume.Spec.Encryption == nil {
		return nil, nil
	}

	encryptionSecret := &corev1.Secret{}
	encryptionSecretKey := client.ObjectKey{Name: sourceVolume.Spec.Encryption.SecretRef.Name, Namespace: sourceVolume.Namespace}
	if err := r.Get(ctx, encryptionSecretKey, encryptionSecret); err != nil {
		if apierrors.IsNotFound(err) {
			return nil, fmt.Errorf("source volume encryption secret %s not found, cannot inherit encryption", sourceVolume.Spec.Encryption.SecretRef.Name)
		}
		return nil, fmt.Errorf("error getting encryption secret %s: %w", encryptionSecretKey, err)
	}

	return &iri.EncryptionSpec{
		SecretData: encryptionSecret.Data,
	}, nil
}

func (r *VolumeReconciler) prepareIRIVolumeEncryption(ctx context.Context, volume *storagev1alpha1.Volume, volumeSnapshot *storagev1alpha1.VolumeSnapshot) (*iri.EncryptionSpec, bool, error) {
	if volume.Spec.VolumeSnapshotRef != nil {
		inheritedEncryption, err := r.prepareIRIVolumeInheritedEncryption(ctx, volumeSnapshot)
		if err != nil {
			return nil, false, fmt.Errorf("error getting encryption from source volume: %w", err)
		}

		if inheritedEncryption != nil {
			if volume.Spec.Encryption != nil {
				return nil, false, fmt.Errorf("cannot specify encryption when creating volume from encrypted snapshot: source volume is encrypted, encryption will be inherited from source volume")
			}

			r.Eventf(volume, corev1.EventTypeNormal, "VolumeEncryptionInherited",
				"Inheriting encryption from encrypted source volume %s", volumeSnapshot.Spec.VolumeRef.Name)
			return inheritedEncryption, true, nil
		}
	}

	if volume.Spec.Encryption != nil {
		return r.prepareIRIVolumeSpecEncryption(ctx, volume)
	}

	return nil, true, nil
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

	var volumeSnapshot *storagev1alpha1.VolumeSnapshot
	if volume.Spec.VolumeSnapshotRef != nil {
		log.V(1).Info("Getting volume snapshot")
		volumeSnapshot = &storagev1alpha1.VolumeSnapshot{}
		volumeSnapshotKey := client.ObjectKey{Namespace: volume.Namespace, Name: volume.Spec.VolumeSnapshotRef.Name}
		if err := r.Get(ctx, volumeSnapshotKey, volumeSnapshot); err != nil {
			if apierrors.IsNotFound(err) {
				r.Eventf(volume, corev1.EventTypeWarning, events.VolumeSnapshotNotFound,
					"VolumeSnapshot %s not found", volume.Spec.VolumeSnapshotRef.Name)
				return nil, false, fmt.Errorf("volume snapshot %s not found", volume.Spec.VolumeSnapshotRef.Name)
			}
			return nil, false, fmt.Errorf("error getting volume snapshot %s: %w", volume.Spec.VolumeSnapshotRef.Name, err)
		}
	}

	log.V(1).Info("Getting encryption secret")
	encryption, encryptionOK, err := r.prepareIRIVolumeEncryption(ctx, volume, volumeSnapshot)
	switch {
	case err != nil:
		errs = append(errs, fmt.Errorf("error preparing iri volume encryption: %w", err))
	case !encryptionOK:
		ok = false
	}

	log.V(1).Info("Getting volume data source")
	dataSource, dataSourceOK, err := r.prepareIRIVolumeDataSource(log, volume, volumeSnapshot)
	switch {
	case err != nil:
		errs = append(errs, fmt.Errorf("error preparing iri volume data source: %w", err))
	case !dataSourceOK:
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
			Image:            volume.Spec.Image,
			Class:            class,
			Resources:        resources,
			Encryption:       encryption,
			VolumeDataSource: dataSource,
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
	modified, err = utilclient.PatchEnsureNoReconcileAnnotation(ctx, r.Client, volume)
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

	log.V(1).Info("Creating volume")
	res, err := r.VolumeRuntime.CreateVolume(ctx, &iri.CreateVolumeRequest{
		Volume: iriVolume,
	})
	if err != nil {
		return ctrl.Result{}, fmt.Errorf("error creating volume: %w", err)
	}

	iriVolume = res.Volume

	volumeID := iriVolume.Metadata.Id
	log = log.WithValues("VolumeID", volumeID)
	log.V(1).Info("Created")

	log.V(1).Info("Updating status")
	if err := r.updateStatus(ctx, log, volume, iriVolume); err != nil {
		return ctrl.Result{}, fmt.Errorf("error updating volume status: %w", err)
	}

	log.V(1).Info("Created")
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

func (r *VolumeReconciler) enqueueVolumesReferencingSnapshot() handler.EventHandler {
	return handler.EnqueueRequestsFromMapFunc(func(ctx context.Context, obj client.Object) []ctrl.Request {
		volumeSnapshot := obj.(*storagev1alpha1.VolumeSnapshot)
		log := ctrl.LoggerFrom(ctx)

		volumeList := &storagev1alpha1.VolumeList{}
		if err := r.List(ctx, volumeList,
			client.InNamespace(volumeSnapshot.Namespace),
			client.MatchingFields{
				storageclient.VolumeSpecVolumeSnapshotRefNameField: volumeSnapshot.Name,
			},
		); err != nil {
			log.Error(err, "Error listing volumes referencing snapshot", "VolumeSnapshotKey", client.ObjectKeyFromObject(volumeSnapshot))
			return nil
		}

		return utilclient.ReconcileRequestsFromObjectStructSlice(volumeList.Items)
	})
}

func (r *VolumeReconciler) SetupWithManager(mgr ctrl.Manager) error {
	log := ctrl.Log.WithName("volumepoollet").WithName("volume")

	return ctrl.NewControllerManagedBy(mgr).
		For(
			&storagev1alpha1.Volume{},
			builder.WithPredicates(
				VolumeRunsInVolumePoolPredicate(r.VolumePoolName),
				predicates.ResourceHasFilterLabel(log, r.WatchFilterValue),
				predicates.ResourceIsNotExternallyManaged(log),
			),
		).
		Watches(
			&storagev1alpha1.VolumeSnapshot{},
			r.enqueueVolumesReferencingSnapshot(),
			builder.WithPredicates(
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
