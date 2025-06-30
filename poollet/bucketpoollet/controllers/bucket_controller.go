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

	storagev1alpha1 "github.com/ironcore-dev/ironcore/api/storage/v1alpha1"
	iriBucket "github.com/ironcore-dev/ironcore/iri/apis/bucket"
	iri "github.com/ironcore-dev/ironcore/iri/apis/bucket/v1alpha1"
	irimeta "github.com/ironcore-dev/ironcore/iri/apis/meta/v1alpha1"
	bucketpoolletv1alpha1 "github.com/ironcore-dev/ironcore/poollet/bucketpoollet/api/v1alpha1"
	"github.com/ironcore-dev/ironcore/poollet/bucketpoollet/bcm"
	"github.com/ironcore-dev/ironcore/poollet/bucketpoollet/controllers/events"
	poolletutils "github.com/ironcore-dev/ironcore/utils/poollet"

	ironcoreclient "github.com/ironcore-dev/ironcore/utils/client"
	"github.com/ironcore-dev/ironcore/utils/predicates"

	"github.com/ironcore-dev/controller-utils/clientutils"
	corev1 "k8s.io/api/core/v1"
	apierrors "k8s.io/apimachinery/pkg/api/errors"
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

type BucketReconciler struct {
	record.EventRecorder
	client.Client
	Scheme *runtime.Scheme

	BucketRuntime     iriBucket.RuntimeService
	BucketRuntimeName string

	BucketClassMapper bcm.BucketClassMapper

	BucketPoolName   string
	WatchFilterValue string

	MaxConcurrentReconciles int
}

func (r *BucketReconciler) iriBucketLabels(bucket *storagev1alpha1.Bucket) map[string]string {
	return map[string]string{
		bucketpoolletv1alpha1.BucketUIDLabel:       string(bucket.UID),
		bucketpoolletv1alpha1.BucketNamespaceLabel: bucket.Namespace,
		bucketpoolletv1alpha1.BucketNameLabel:      bucket.Name,
	}
}

func (r *BucketReconciler) iriBucketAnnotations(_ *storagev1alpha1.Bucket) map[string]string {
	return map[string]string{}
}

func (r *BucketReconciler) listIRIBucketsByKey(ctx context.Context, bucketKey client.ObjectKey) ([]*iri.Bucket, error) {
	res, err := r.BucketRuntime.ListBuckets(ctx, &iri.ListBucketsRequest{
		Filter: &iri.BucketFilter{
			LabelSelector: map[string]string{
				bucketpoolletv1alpha1.BucketNamespaceLabel: bucketKey.Namespace,
				bucketpoolletv1alpha1.BucketNameLabel:      bucketKey.Name,
			},
		},
	})
	if err != nil {
		return nil, fmt.Errorf("error listing buckets by key: %w", err)
	}
	buckets := res.Buckets
	return buckets, nil
}

func (r *BucketReconciler) listIRIBucketsByUID(ctx context.Context, bucketUID types.UID) ([]*iri.Bucket, error) {
	res, err := r.BucketRuntime.ListBuckets(ctx, &iri.ListBucketsRequest{
		Filter: &iri.BucketFilter{
			LabelSelector: map[string]string{
				bucketpoolletv1alpha1.BucketUIDLabel: string(bucketUID),
			},
		},
	})
	if err != nil {
		return nil, fmt.Errorf("error listing buckets by uid: %w", err)
	}
	return res.Buckets, nil
}

//+kubebuilder:rbac:groups="",resources=events,verbs=create;patch
//+kubebuilder:rbac:groups="",resources=secrets,verbs=get;list;watch
//+kubebuilder:rbac:groups=storage.ironcore.dev,resources=buckets,verbs=get;list;watch;update;patch
//+kubebuilder:rbac:groups=storage.ironcore.dev,resources=buckets/status,verbs=get;update;patch
//+kubebuilder:rbac:groups=storage.ironcore.dev,resources=buckets/finalizers,verbs=update

func (r *BucketReconciler) Reconcile(ctx context.Context, req ctrl.Request) (ctrl.Result, error) {
	log := ctrl.LoggerFrom(ctx)
	bucket := &storagev1alpha1.Bucket{}
	if err := r.Get(ctx, req.NamespacedName, bucket); err != nil {
		if !apierrors.IsNotFound(err) {
			return ctrl.Result{}, fmt.Errorf("error getting bucket %s: %w", req.NamespacedName, err)
		}
		return r.deleteGone(ctx, log, req.NamespacedName)
	}
	return r.reconcileExists(ctx, log, bucket)
}

func (r *BucketReconciler) deleteGone(ctx context.Context, log logr.Logger, bucketKey client.ObjectKey) (ctrl.Result, error) {
	log.V(1).Info("Delete gone")

	log.V(1).Info("Listing iri buckets by key")
	buckets, err := r.listIRIBucketsByKey(ctx, bucketKey)
	if err != nil {
		return ctrl.Result{}, fmt.Errorf("error listing iri buckets by key: %w", err)
	}

	ok, err := r.deleteIRIBuckets(ctx, log, buckets)
	if err != nil {
		return ctrl.Result{}, fmt.Errorf("error deleting iri buckets: %w", err)
	}
	if !ok {
		log.V(1).Info("Not all iri buckets are gone, requeueing")
		return ctrl.Result{Requeue: true}, nil
	}

	log.V(1).Info("Deleted gone")
	return ctrl.Result{}, nil
}

func (r *BucketReconciler) deleteIRIBuckets(ctx context.Context, log logr.Logger, buckets []*iri.Bucket) (bool, error) {
	var (
		errs                 []error
		deletingIRIBucketIDs []string
	)

	for _, bucket := range buckets {
		iriBucketID := bucket.Metadata.Id
		log := log.WithValues("IRIBucketID", iriBucketID)
		log.V(1).Info("Deleting iri bucket")
		_, err := r.BucketRuntime.DeleteBucket(ctx, &iri.DeleteBucketRequest{
			BucketId: iriBucketID,
		})
		if err != nil {
			if status.Code(err) != codes.NotFound {
				errs = append(errs, fmt.Errorf("error deleting iri bucket %s: %w", iriBucketID, err))
			} else {
				log.V(1).Info("IRI Bucket is already gone")
			}
		} else {
			log.V(1).Info("Issued iri bucket deletion")
			deletingIRIBucketIDs = append(deletingIRIBucketIDs, iriBucketID)
		}
	}

	switch {
	case len(errs) > 0:
		return false, fmt.Errorf("error(s) deleting iri bucket(s): %v", errs)
	case len(deletingIRIBucketIDs) > 0:
		log.V(1).Info("Buckets are in deletion", "DeletingIRIBucketIDs", deletingIRIBucketIDs)
		return false, nil
	default:
		log.V(1).Info("No iri buckets present")
		return true, nil
	}
}

func (r *BucketReconciler) reconcileExists(ctx context.Context, log logr.Logger, bucket *storagev1alpha1.Bucket) (ctrl.Result, error) {
	if !bucket.DeletionTimestamp.IsZero() {
		return r.delete(ctx, log, bucket)
	}
	return r.reconcile(ctx, log, bucket)
}

func (r *BucketReconciler) delete(ctx context.Context, log logr.Logger, bucket *storagev1alpha1.Bucket) (ctrl.Result, error) {
	log.V(1).Info("Delete")

	if !controllerutil.ContainsFinalizer(bucket, bucketpoolletv1alpha1.BucketFinalizer) {
		log.V(1).Info("No finalizer present, nothing to do")
		return ctrl.Result{}, nil
	}

	log.V(1).Info("Finalizer present")

	log.V(1).Info("Listing buckets")
	buckets, err := r.listIRIBucketsByUID(ctx, bucket.UID)
	if err != nil {
		return ctrl.Result{}, fmt.Errorf("error listing buckets by uid: %w", err)
	}

	ok, err := r.deleteIRIBuckets(ctx, log, buckets)
	if err != nil {
		return ctrl.Result{}, fmt.Errorf("error deleting iri buckets: %w", err)
	}
	if !ok {
		log.V(1).Info("Not all iri buckets are gone, requeueing")
		return ctrl.Result{Requeue: true}, nil
	}

	log.V(1).Info("Deleted all iri buckets, removing finalizer")
	if err := clientutils.PatchRemoveFinalizer(ctx, r.Client, bucket, bucketpoolletv1alpha1.BucketFinalizer); err != nil {
		return ctrl.Result{}, fmt.Errorf("error removing finalizer: %w", err)
	}

	log.V(1).Info("Deleted")
	return ctrl.Result{}, nil
}

func getIRIBucketClassCapabilities(bucketClass *storagev1alpha1.BucketClass) *iri.BucketClassCapabilities {
	tps := bucketClass.Capabilities.TPS()
	iops := bucketClass.Capabilities.IOPS()

	return &iri.BucketClassCapabilities{
		Tps:  tps.Value(),
		Iops: iops.Value(),
	}
}

func (r *BucketReconciler) prepareIRIBucketMetadata(bucket *storagev1alpha1.Bucket) *irimeta.ObjectMetadata {
	return &irimeta.ObjectMetadata{
		Labels:      r.iriBucketLabels(bucket),
		Annotations: r.iriBucketAnnotations(bucket),
	}
}

func (r *BucketReconciler) prepareIRIBucketClass(ctx context.Context, bucket *storagev1alpha1.Bucket, bucketClassName string) (string, bool, error) {
	bucketClass := &storagev1alpha1.BucketClass{}
	bucketClassKey := client.ObjectKey{Name: bucketClassName}
	if err := r.Get(ctx, bucketClassKey, bucketClass); err != nil {
		err = fmt.Errorf("error getting bucket class %s: %w", bucketClassKey, err)
		if !apierrors.IsNotFound(err) {
			return "", false, fmt.Errorf("error getting bucket class %s: %w", bucketClassName, err)
		}

		r.Eventf(bucket, corev1.EventTypeNormal, events.BucketClassNotReady, "Bucket class %s not found", bucketClassName)
		return "", false, nil
	}

	caps := getIRIBucketClassCapabilities(bucketClass)

	class, err := r.BucketClassMapper.GetBucketClassFor(ctx, bucketClassName, caps)
	if err != nil {
		return "", false, fmt.Errorf("error getting matching bucket class: %w", err)
	}
	return class.Name, true, nil
}

func (r *BucketReconciler) prepareIRIBucket(ctx context.Context, log logr.Logger, bucket *storagev1alpha1.Bucket) (*iri.Bucket, bool, error) {
	var (
		ok   = true
		errs []error
	)

	log.V(1).Info("Getting bucket class")
	class, classOK, err := r.prepareIRIBucketClass(ctx, bucket, bucket.Spec.BucketClassRef.Name)
	switch {
	case err != nil:
		errs = append(errs, fmt.Errorf("error preparing iri bucket class: %w", err))
	case !classOK:
		ok = false
	}

	metadata := r.prepareIRIBucketMetadata(bucket)

	if len(errs) > 0 {
		return nil, false, fmt.Errorf("error(s) preparing iri bucket: %v", errs)
	}
	if !ok {
		return nil, false, nil
	}

	return &iri.Bucket{
		Metadata: metadata,
		Spec: &iri.BucketSpec{
			Class: class,
		},
	}, true, nil
}

func (r *BucketReconciler) reconcile(ctx context.Context, log logr.Logger, bucket *storagev1alpha1.Bucket) (ctrl.Result, error) {
	log.V(1).Info("Reconcile")

	log.V(1).Info("Ensuring finalizer")
	modified, err := clientutils.PatchEnsureFinalizer(ctx, r.Client, bucket, bucketpoolletv1alpha1.BucketFinalizer)
	if err != nil {
		return ctrl.Result{}, fmt.Errorf("error ensuring finalizer: %w", err)
	}
	if modified {
		log.V(1).Info("Added finalizer, requeueing")
		return ctrl.Result{Requeue: true}, nil
	}
	log.V(1).Info("Finalizer is present")

	log.V(1).Info("Ensuring no reconcile annotation")
	modified, err = ironcoreclient.PatchEnsureNoReconcileAnnotation(ctx, r.Client, bucket)
	if err != nil {
		return ctrl.Result{}, fmt.Errorf("error ensuring no reconcile annotation: %w", err)
	}
	if modified {
		log.V(1).Info("Removed reconcile annotation, requeueing")
		return ctrl.Result{Requeue: true}, nil
	}

	log.V(1).Info("Listing buckets")
	res, err := r.BucketRuntime.ListBuckets(ctx, &iri.ListBucketsRequest{
		Filter: &iri.BucketFilter{
			LabelSelector: map[string]string{
				bucketpoolletv1alpha1.BucketUIDLabel: string(bucket.UID),
			},
		},
	})
	if err != nil {
		return ctrl.Result{}, fmt.Errorf("error listing buckets: %w", err)
	}

	switch len(res.Buckets) {
	case 0:
		return r.create(ctx, log, bucket)
	case 1:
		iriBucket := res.Buckets[0]
		if err := r.updateStatus(ctx, log, bucket, iriBucket); err != nil {
			return ctrl.Result{}, fmt.Errorf("error updating bucket status: %w", err)
		}
		return ctrl.Result{}, nil
	default:
		panic("unhandled multiple buckets")
	}
}

func (r *BucketReconciler) create(ctx context.Context, log logr.Logger, bucket *storagev1alpha1.Bucket) (ctrl.Result, error) {
	log.V(1).Info("Create")

	log.V(1).Info("Preparing iri bucket")
	iriBucket, ok, err := r.prepareIRIBucket(ctx, log, bucket)
	if err != nil {
		return ctrl.Result{}, fmt.Errorf("error preparing iri bucket: %w", err)
	}
	if !ok {
		log.V(1).Info("IRI bucket is not yet ready to be prepared")
		return ctrl.Result{}, nil
	}

	log.V(1).Info("Creating bucket")
	res, err := r.BucketRuntime.CreateBucket(ctx, &iri.CreateBucketRequest{
		Bucket: iriBucket,
	})
	if err != nil {
		return ctrl.Result{}, fmt.Errorf("error creating bucket: %w", err)
	}

	iriBucket = res.Bucket

	bucketID := iriBucket.Metadata.Id
	log = log.WithValues("BucketID", bucketID)
	log.V(1).Info("Created")

	log.V(1).Info("Updating status")
	if err := r.updateStatus(ctx, log, bucket, iriBucket); err != nil {
		return ctrl.Result{}, fmt.Errorf("error updating bucket status: %w", err)
	}

	log.V(1).Info("Created")
	return ctrl.Result{}, nil
}

func (r *BucketReconciler) bucketSecretName(bucketName string) string {
	sum := sha256.Sum256([]byte(bucketName))
	return hex.EncodeToString(sum[:])[:63]
}

var iriBucketStateToBucketState = map[iri.BucketState]storagev1alpha1.BucketState{
	iri.BucketState_BUCKET_PENDING:   storagev1alpha1.BucketStatePending,
	iri.BucketState_BUCKET_AVAILABLE: storagev1alpha1.BucketStateAvailable,
	iri.BucketState_BUCKET_ERROR:     storagev1alpha1.BucketStateError,
}

func (r *BucketReconciler) convertIRIBucketState(iriState iri.BucketState) (storagev1alpha1.BucketState, error) {
	if res, ok := iriBucketStateToBucketState[iriState]; ok {
		return res, nil
	}
	return "", fmt.Errorf("unknown bucket state %v", iriState)
}

func (r *BucketReconciler) updateStatus(ctx context.Context, log logr.Logger, bucket *storagev1alpha1.Bucket, iriBucket *iri.Bucket) error {
	var access *storagev1alpha1.BucketAccess

	if iriBucket.Status.State == iri.BucketState_BUCKET_AVAILABLE {
		if iriAccess := iriBucket.Status.Access; iriAccess != nil {
			var secretRef *corev1.LocalObjectReference

			if iriAccess.SecretData != nil {
				log.V(1).Info("Applying bucket secret")
				bucketSecret := &corev1.Secret{
					TypeMeta: metav1.TypeMeta{
						APIVersion: corev1.SchemeGroupVersion.String(),
						Kind:       "Secret",
					},
					ObjectMeta: metav1.ObjectMeta{
						Namespace: bucket.Namespace,
						Name:      r.bucketSecretName(bucket.Name),
						Labels: map[string]string{
							bucketpoolletv1alpha1.BucketUIDLabel: string(bucket.UID),
						},
					},
					Data: iriAccess.SecretData,
				}
				_ = ctrl.SetControllerReference(bucket, bucketSecret, r.Scheme)
				if err := r.Patch(ctx, bucketSecret, client.Apply, client.FieldOwner(bucketpoolletv1alpha1.FieldOwner)); err != nil {
					return fmt.Errorf("error applying bucket secret: %w", err)
				}
				secretRef = &corev1.LocalObjectReference{Name: bucketSecret.Name}
			} else {
				log.V(1).Info("Deleting any corresponding bucket secret")
				if err := r.DeleteAllOf(ctx, &corev1.Secret{},
					client.InNamespace(bucket.Namespace),
					client.MatchingLabels{
						bucketpoolletv1alpha1.BucketUIDLabel: string(bucket.UID),
					},
				); err != nil {
					return fmt.Errorf("error deleting any corresponding bucket secret: %w", err)
				}
			}

			access = &storagev1alpha1.BucketAccess{
				SecretRef: secretRef,
				Endpoint:  iriAccess.Endpoint,
			}
		}
	}

	base := bucket.DeepCopy()
	now := metav1.Now()

	bucketID := poolletutils.MakeID(r.BucketRuntimeName, iriBucket.Metadata.Id)

	bucket.Status.Access = access
	newState, err := r.convertIRIBucketState(iriBucket.Status.State)
	if err != nil {
		return err
	}
	if newState != bucket.Status.State {
		bucket.Status.LastStateTransitionTime = &now
	}
	bucket.Status.State = newState
	bucket.Status.BucketID = bucketID.String()

	if err := r.Status().Patch(ctx, bucket, client.MergeFrom(base)); err != nil {
		return fmt.Errorf("error patching bucket status: %w", err)
	}
	return nil
}

func (r *BucketReconciler) SetupWithManager(mgr ctrl.Manager) error {
	log := ctrl.Log.WithName("bucketpoollet")

	return ctrl.NewControllerManagedBy(mgr).
		For(
			&storagev1alpha1.Bucket{},
			builder.WithPredicates(
				BucketRunsInBucketPoolPredicate(r.BucketPoolName),
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

func BucketRunsInBucketPool(bucket *storagev1alpha1.Bucket, bucketPoolName string) bool {
	bucketPoolRef := bucket.Spec.BucketPoolRef
	if bucketPoolRef == nil {
		return false
	}

	return bucketPoolRef.Name == bucketPoolName
}

func BucketRunsInBucketPoolPredicate(bucketPoolName string) predicate.Predicate {
	return predicate.NewPredicateFuncs(func(object client.Object) bool {
		bucket := object.(*storagev1alpha1.Bucket)
		return BucketRunsInBucketPool(bucket, bucketPoolName)
	})
}
