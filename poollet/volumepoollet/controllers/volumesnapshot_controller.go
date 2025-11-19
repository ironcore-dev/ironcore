// SPDX-FileCopyrightText: 2025 SAP SE or an SAP affiliate company and IronCore contributors
// SPDX-License-Identifier: Apache-2.0

package controllers

import (
	"context"
	"fmt"

	"github.com/go-logr/logr"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
	"k8s.io/apimachinery/pkg/api/resource"

	storagev1alpha1 "github.com/ironcore-dev/ironcore/api/storage/v1alpha1"
	irimeta "github.com/ironcore-dev/ironcore/iri/apis/meta/v1alpha1"
	iriVolume "github.com/ironcore-dev/ironcore/iri/apis/volume"
	iri "github.com/ironcore-dev/ironcore/iri/apis/volume/v1alpha1"
	poolletutils "github.com/ironcore-dev/ironcore/poollet/common/utils"
	volumepoolletv1alpha1 "github.com/ironcore-dev/ironcore/poollet/volumepoollet/api/v1alpha1"
	"github.com/ironcore-dev/ironcore/poollet/volumepoollet/controllers/events"
	ironcoreclient "github.com/ironcore-dev/ironcore/utils/client"
	utilsmaps "github.com/ironcore-dev/ironcore/utils/maps"
	"github.com/ironcore-dev/ironcore/utils/predicates"

	"github.com/ironcore-dev/controller-utils/clientutils"
	corev1 "k8s.io/api/core/v1"
	apierrors "k8s.io/apimachinery/pkg/api/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apimachinery/pkg/types"
	"k8s.io/client-go/tools/record"
	"k8s.io/kubectl/pkg/util/fieldpath"
	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/builder"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/controller"
	"sigs.k8s.io/controller-runtime/pkg/controller/controllerutil"
)

type VolumeSnapshotReconciler struct {
	record.EventRecorder
	client.Client
	Scheme *runtime.Scheme

	VolumeRuntime     iriVolume.RuntimeService
	VolumeRuntimeName string

	DownwardAPILabels      map[string]string
	DownwardAPIAnnotations map[string]string

	WatchFilterValue string

	MaxConcurrentReconciles int
}

//+kubebuilder:rbac:groups="",resources=events,verbs=create;patch
//+kubebuilder:rbac:groups=storage.ironcore.dev,resources=volumesnapshots,verbs=get;list;watch;update;patch
//+kubebuilder:rbac:groups=storage.ironcore.dev,resources=volumesnapshots/status,verbs=get;update;patch
//+kubebuilder:rbac:groups=storage.ironcore.dev,resources=volumesnapshots/finalizers,verbs=update

func (r *VolumeSnapshotReconciler) Reconcile(ctx context.Context, req ctrl.Request) (ctrl.Result, error) {
	log := ctrl.LoggerFrom(ctx)
	volumeSnapshot := &storagev1alpha1.VolumeSnapshot{}
	if err := r.Get(ctx, req.NamespacedName, volumeSnapshot); err != nil {
		if !apierrors.IsNotFound(err) {
			return ctrl.Result{}, fmt.Errorf("error getting volume snapshot %s: %w", req.NamespacedName, err)
		}
		return r.deleteGone(ctx, log, req.NamespacedName)
	}
	return r.reconcileExists(ctx, log, volumeSnapshot)
}

func (r *VolumeSnapshotReconciler) reconcileExists(ctx context.Context, log logr.Logger, volumeSnapshot *storagev1alpha1.VolumeSnapshot) (ctrl.Result, error) {
	if !volumeSnapshot.DeletionTimestamp.IsZero() {
		return r.delete(ctx, log, volumeSnapshot)
	}
	return r.reconcile(ctx, log, volumeSnapshot)
}

func (r *VolumeSnapshotReconciler) deleteGone(ctx context.Context, log logr.Logger, volumeSnapshotKey client.ObjectKey) (ctrl.Result, error) {
	log.V(1).Info("Delete gone")

	log.V(1).Info("Listing iri volume snapshots by key")
	volumeSnapshots, err := r.listIRIVolumeSnapshotsByKey(ctx, volumeSnapshotKey)
	if err != nil {
		return ctrl.Result{}, fmt.Errorf("error listing iri volume snapshots by key: %w", err)
	}

	ok, err := r.deleteIRIVolumeSnapshots(ctx, log, volumeSnapshots)
	if err != nil {
		return ctrl.Result{}, fmt.Errorf("error deleting iri volume snapshots: %w", err)
	}
	if !ok {
		log.V(1).Info("Not all iri volume snapshots are gone, requeueing")
		return ctrl.Result{RequeueAfter: 1}, nil
	}

	log.V(1).Info("Deleted gone")
	return ctrl.Result{}, nil
}

func (r *VolumeSnapshotReconciler) listIRIVolumeSnapshotsByKey(ctx context.Context, volumeSnapshotKey client.ObjectKey) ([]*iri.VolumeSnapshot, error) {
	res, err := r.VolumeRuntime.ListVolumeSnapshots(ctx, &iri.ListVolumeSnapshotsRequest{
		Filter: &iri.VolumeSnapshotFilter{
			LabelSelector: map[string]string{
				volumepoolletv1alpha1.VolumeSnapshotNamespaceLabel: volumeSnapshotKey.Namespace,
				volumepoolletv1alpha1.VolumeSnapshotNameLabel:      volumeSnapshotKey.Name,
			},
		},
	})
	if err != nil {
		return nil, fmt.Errorf("error listing volume snapshots by key: %w", err)
	}
	volumeSnapshots := res.VolumeSnapshots
	return volumeSnapshots, nil
}

func (r *VolumeSnapshotReconciler) delete(ctx context.Context, log logr.Logger, volumeSnapshot *storagev1alpha1.VolumeSnapshot) (ctrl.Result, error) {
	log.V(1).Info("Delete")

	if !controllerutil.ContainsFinalizer(volumeSnapshot, volumepoolletv1alpha1.VolumeSnapshotFinalizer) {
		log.V(1).Info("No finalizer present, nothing to do")
		return ctrl.Result{}, nil
	}

	log.V(1).Info("Finalizer present")

	log.V(1).Info("Listing volume snapshots")
	volumeSnapshots, err := r.listIRIVolumeSnapshotsByUID(ctx, volumeSnapshot.UID)
	if err != nil {
		return ctrl.Result{}, fmt.Errorf("error listing volume snapshots by uid: %w", err)
	}

	ok, err := r.deleteIRIVolumeSnapshots(ctx, log, volumeSnapshots)
	if err != nil {
		return ctrl.Result{}, fmt.Errorf("error deleting iri volume snapshots: %w", err)
	}
	if !ok {
		log.V(1).Info("Not all iri volume snapshots are gone, requeueing")
		return ctrl.Result{RequeueAfter: 1}, nil
	}

	log.V(1).Info("Deleted all iri volume snapshots, removing finalizer")
	if err := clientutils.PatchRemoveFinalizer(ctx, r.Client, volumeSnapshot, volumepoolletv1alpha1.VolumeSnapshotFinalizer); err != nil {
		return ctrl.Result{}, fmt.Errorf("error removing finalizer: %w", err)
	}

	log.V(1).Info("Deleted")
	return ctrl.Result{}, nil
}

func (r *VolumeSnapshotReconciler) listIRIVolumeSnapshotsByUID(ctx context.Context, uid types.UID) ([]*iri.VolumeSnapshot, error) {
	res, err := r.VolumeRuntime.ListVolumeSnapshots(ctx, &iri.ListVolumeSnapshotsRequest{
		Filter: &iri.VolumeSnapshotFilter{
			LabelSelector: map[string]string{
				volumepoolletv1alpha1.VolumeSnapshotUIDLabel: string(uid),
			},
		},
	})
	if err != nil {
		return nil, fmt.Errorf("error listing volume snapshots by uid: %w", err)
	}

	return res.VolumeSnapshots, nil
}

func (r *VolumeSnapshotReconciler) deleteIRIVolumeSnapshots(ctx context.Context, log logr.Logger, volumeSnapshots []*iri.VolumeSnapshot) (bool, error) {
	var (
		errs                         []error
		deletingIRIVolumeSnapshotIDs []string
	)

	for _, volumeSnapshot := range volumeSnapshots {
		iriVolumeSnapshotID := volumeSnapshot.Metadata.Id
		log := log.WithValues("IRIVolumeSnapshotID", iriVolumeSnapshotID)
		log.V(1).Info("Deleting iri volume snapshot")
		_, err := r.VolumeRuntime.DeleteVolumeSnapshot(ctx, &iri.DeleteVolumeSnapshotRequest{
			VolumeSnapshotId: iriVolumeSnapshotID,
		})
		if err != nil {
			if status.Code(err) != codes.NotFound {
				errs = append(errs, fmt.Errorf("error deleting iri volume snapshot %s: %w", iriVolumeSnapshotID, err))
			} else {
				log.V(1).Info("IRI Volume snapshot is already gone")
			}
		} else {
			log.V(1).Info("Issued iri volume snapshot deletion")
			deletingIRIVolumeSnapshotIDs = append(deletingIRIVolumeSnapshotIDs, iriVolumeSnapshotID)
		}
	}

	switch {
	case len(errs) > 0:
		return false, fmt.Errorf("error(s) deleting iri volume snapshot(s): %v", errs)
	case len(deletingIRIVolumeSnapshotIDs) > 0:
		log.V(1).Info("Volume snapshots are in deletion", "DeletingIRIVolumeSnapshotIDs", deletingIRIVolumeSnapshotIDs)
		return false, nil
	default:
		log.V(1).Info("No iri volume snapshots present")
		return true, nil
	}
}

func (r *VolumeSnapshotReconciler) iriVolumeSnapshotLabels(volumeSnapshot *storagev1alpha1.VolumeSnapshot) (map[string]string, error) {
	labels := map[string]string{
		volumepoolletv1alpha1.VolumeSnapshotUIDLabel:       string(volumeSnapshot.UID),
		volumepoolletv1alpha1.VolumeSnapshotNamespaceLabel: volumeSnapshot.Namespace,
		volumepoolletv1alpha1.VolumeSnapshotNameLabel:      volumeSnapshot.Name,
	}
	apiLabels, err := poolletutils.PrepareDownwardAPILabels(volumeSnapshot, r.DownwardAPILabels, volumepoolletv1alpha1.VolumeSnapshotDownwardAPIPrefix)
	if err != nil {
		return nil, err
	}
	labels = utilsmaps.AppendMap(labels, apiLabels)
	return labels, nil
}

func (r *VolumeSnapshotReconciler) iriVolumeSnapshotAnnotations(volumeSnapshot *storagev1alpha1.VolumeSnapshot) (map[string]string, error) {
	annotations := map[string]string{}

	for name, fieldPath := range r.DownwardAPIAnnotations {
		value, err := fieldpath.ExtractFieldPathAsString(volumeSnapshot, fieldPath)
		if err != nil {
			return nil, fmt.Errorf("error extracting downward api annotation %q: %w", name, err)
		}
		annotations[poolletutils.DownwardAPIAnnotation(volumepoolletv1alpha1.VolumeSnapshotDownwardAPIPrefix, name)] = value
	}

	return annotations, nil
}

func (r *VolumeSnapshotReconciler) prepareIRIVolumeSnapshotMetadata(volumeSnapshot *storagev1alpha1.VolumeSnapshot, errs []error) (*irimeta.ObjectMetadata, []error) {
	labels, err := r.iriVolumeSnapshotLabels(volumeSnapshot)
	if err != nil {
		errs = append(errs, fmt.Errorf("error preparing iri volume snapshot labels: %w", err))
	}

	annotations, err := r.iriVolumeSnapshotAnnotations(volumeSnapshot)
	if err != nil {
		errs = append(errs, fmt.Errorf("error preparing iri volume snapshot annotations: %w", err))
	}

	return &irimeta.ObjectMetadata{
		Labels:      labels,
		Annotations: annotations,
	}, errs
}

func (r *VolumeSnapshotReconciler) prepareIRIVolumeSnapshot(ctx context.Context, volumeSnapshot *storagev1alpha1.VolumeSnapshot) (*iri.VolumeSnapshot, bool, error) {

	if volumeSnapshot.Spec.VolumeRef == nil {
		return nil, false, fmt.Errorf("volumeRef is required")
	}

	volume := &storagev1alpha1.Volume{}
	if err := r.Get(ctx, client.ObjectKey{Namespace: volumeSnapshot.Namespace, Name: volumeSnapshot.Spec.VolumeRef.Name}, volume); err != nil {
		if apierrors.IsNotFound(err) {
			r.Eventf(volumeSnapshot, corev1.EventTypeNormal, events.SourceVolumeNotFound,
				"Source volume %s not found", volumeSnapshot.Spec.VolumeRef.Name)
			return nil, false, nil
		}
		return nil, false, fmt.Errorf("error getting referenced volume: %w", err)
	}

	if volume.Status.State != storagev1alpha1.VolumeStateAvailable {
		r.Eventf(volumeSnapshot, corev1.EventTypeNormal, events.SourceVolumeNotAvailable,
			"Source volume %s is not available (state: %s)", volumeSnapshot.Spec.VolumeRef.Name, volume.Status.State)
		return nil, false, nil
	}

	metadata, errs := r.prepareIRIVolumeSnapshotMetadata(volumeSnapshot, []error{})

	if len(errs) > 0 {
		return nil, false, fmt.Errorf("error(s) preparing iri volume snapshot metadata: %v", errs)
	}

	return &iri.VolumeSnapshot{
		Metadata: metadata,
		Spec: &iri.VolumeSnapshotSpec{
			VolumeId: volume.Name,
		},
	}, true, nil
}

func (r *VolumeSnapshotReconciler) reconcile(ctx context.Context, log logr.Logger, volumeSnapshot *storagev1alpha1.VolumeSnapshot) (ctrl.Result, error) {
	log.V(1).Info("Reconcile")

	log.V(1).Info("Ensuring finalizer")
	modified, err := clientutils.PatchEnsureFinalizer(ctx, r.Client, volumeSnapshot, volumepoolletv1alpha1.VolumeSnapshotFinalizer)
	if err != nil {
		return ctrl.Result{}, fmt.Errorf("error ensuring finalizer: %w", err)
	}
	if modified {
		log.V(1).Info("Added finalizer, requeueing")
		return ctrl.Result{RequeueAfter: 1}, nil
	}
	log.V(1).Info("Finalizer is present")

	log.V(1).Info("Ensuring no reconcile annotation")
	modified, err = ironcoreclient.PatchEnsureNoReconcileAnnotation(ctx, r.Client, volumeSnapshot)
	if err != nil {
		return ctrl.Result{}, fmt.Errorf("error ensuring no reconcile annotation: %w", err)
	}
	if modified {
		log.V(1).Info("Removed reconcile annotation, requeueing")
		return ctrl.Result{RequeueAfter: 1}, nil
	}

	log.V(1).Info("Listing volume snapshots")
	res, err := r.VolumeRuntime.ListVolumeSnapshots(ctx, &iri.ListVolumeSnapshotsRequest{
		Filter: &iri.VolumeSnapshotFilter{
			LabelSelector: map[string]string{
				volumepoolletv1alpha1.VolumeSnapshotUIDLabel: string(volumeSnapshot.UID),
			},
		},
	})
	if err != nil {
		return ctrl.Result{}, fmt.Errorf("error listing volume snapshots: %w", err)
	}

	switch len(res.VolumeSnapshots) {
	case 0:
		return r.create(ctx, log, volumeSnapshot)
	case 1:
		iriVolumeSnapshot := res.VolumeSnapshots[0]
		if err := r.updateStatus(ctx, log, volumeSnapshot, iriVolumeSnapshot); err != nil {
			return ctrl.Result{}, fmt.Errorf("error updating volume snapshot status: %w", err)
		}
		return ctrl.Result{}, nil
	default:
		panic("unhandled multiple volume snapshots")
	}
}

func (r *VolumeSnapshotReconciler) create(ctx context.Context, log logr.Logger, volumeSnapshot *storagev1alpha1.VolumeSnapshot) (ctrl.Result, error) {
	log.V(1).Info("Create")

	log.V(1).Info("Preparing iri volume snapshot")
	iriVolumeSnapshot, ok, err := r.prepareIRIVolumeSnapshot(ctx, volumeSnapshot)
	if err != nil {
		return ctrl.Result{}, fmt.Errorf("error preparing iri volume snapshot: %w", err)
	}
	if !ok {
		log.V(1).Info("IRI volume snapshot is not yet ready to be prepared")
		return ctrl.Result{}, nil
	}

	log.V(1).Info("Creating volume snapshot")
	res, err := r.VolumeRuntime.CreateVolumeSnapshot(ctx, &iri.CreateVolumeSnapshotRequest{
		VolumeSnapshot: iriVolumeSnapshot,
	})
	if err != nil {
		return ctrl.Result{}, fmt.Errorf("error creating volume snapshot: %w", err)
	}

	iriVolumeSnapshot = res.VolumeSnapshot

	volumeSnapshotID := iriVolumeSnapshot.Metadata.Id
	log = log.WithValues("VolumeSnapshotID", volumeSnapshotID)
	log.V(1).Info("Created")

	log.V(1).Info("Updating status")
	if err := r.updateStatus(ctx, log, volumeSnapshot, iriVolumeSnapshot); err != nil {
		return ctrl.Result{}, fmt.Errorf("error updating volume snapshot status: %w", err)
	}

	log.V(1).Info("Created")
	return ctrl.Result{}, nil
}

var iriVolumeSnapshotStateToVolumeSnapshotState = map[iri.VolumeSnapshotState]storagev1alpha1.VolumeSnapshotState{
	iri.VolumeSnapshotState_VOLUME_SNAPSHOT_PENDING: storagev1alpha1.VolumeSnapshotStatePending,
	iri.VolumeSnapshotState_VOLUME_SNAPSHOT_READY:   storagev1alpha1.VolumeSnapshotStateReady,
	iri.VolumeSnapshotState_VOLUME_SNAPSHOT_FAILED:  storagev1alpha1.VolumeSnapshotStateFailed,
}

func (r *VolumeSnapshotReconciler) convertIRIVolumeSnapshotState(iriState iri.VolumeSnapshotState) (storagev1alpha1.VolumeSnapshotState, error) {
	if res, ok := iriVolumeSnapshotStateToVolumeSnapshotState[iriState]; ok {
		return res, nil
	}
	return "", fmt.Errorf("unknown volume snapshot state %v", iriState)
}

func (r *VolumeSnapshotReconciler) updateStatus(ctx context.Context, log logr.Logger, volumeSnapshot *storagev1alpha1.VolumeSnapshot, iriVolumeSnapshot *iri.VolumeSnapshot) error {
	log.V(1).Info("Updating status")

	base := volumeSnapshot.DeepCopy()
	now := metav1.Now()

	newState, err := r.convertIRIVolumeSnapshotState(iriVolumeSnapshot.Status.State)
	if err != nil {
		return fmt.Errorf("error converting iri volume snapshot state: %w", err)
	}

	if newState != volumeSnapshot.Status.State {
		volumeSnapshot.Status.LastStateTransitionTime = &now
	}
	volumeSnapshot.Status.State = newState

	snapshotID := poolletutils.MakeID(r.VolumeRuntimeName, iriVolumeSnapshot.Metadata.Id)
	volumeSnapshot.Status.SnapshotID = snapshotID.String()

	if iriVolumeSnapshot.Status.Size > 0 {
		volumeSnapshot.Status.Size = resource.NewQuantity(iriVolumeSnapshot.Status.Size, resource.DecimalSI)
	}

	if err := r.Client.Status().Patch(ctx, volumeSnapshot, client.MergeFrom(base)); err != nil {
		return fmt.Errorf("error patching volume snapshot status: %w", err)
	}

	log.V(1).Info("Updated status")
	return nil
}

func (r *VolumeSnapshotReconciler) SetupWithManager(mgr ctrl.Manager) error {
	log := ctrl.Log.WithName("volumepoollet").WithName("volumesnapshot")

	return ctrl.NewControllerManagedBy(mgr).
		For(
			&storagev1alpha1.VolumeSnapshot{},
			builder.WithPredicates(
				predicates.ResourceHasFilterLabel(log, r.WatchFilterValue),
				predicates.ResourceIsNotExternallyManaged(log),
			),
		).
		WithOptions(controller.Options{
			MaxConcurrentReconciles: r.MaxConcurrentReconciles,
		}).
		Complete(r)
}
