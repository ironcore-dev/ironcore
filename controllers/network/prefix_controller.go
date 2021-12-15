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

package network

import (
	"context"
	"fmt"
	"github.com/go-logr/logr"
	"github.com/onmetal/controller-utils/conditionutils"
	commonv1alpha1 "github.com/onmetal/onmetal-api/apis/common/v1alpha1"
	networkv1alpha1 "github.com/onmetal/onmetal-api/apis/network/v1alpha1"
	"inet.af/netaddr"
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/labels"
	"k8s.io/apimachinery/pkg/runtime"
	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/controller/controllerutil"
	"sigs.k8s.io/controller-runtime/pkg/handler"
	"sigs.k8s.io/controller-runtime/pkg/source"
)

const (
	PrefixFinalizer = "network.onmetal.de/prefix"
)

func IsPrefixAllocationSucceeded(allocation *networkv1alpha1.PrefixAllocation) bool {
	return (*PrefixAllocation)(allocation).ReadyState() == PrefixAllocationSucceeded
}

func IsPrefixAllocationFailed(allocation *networkv1alpha1.PrefixAllocation) bool {
	return (*PrefixAllocation)(allocation).ReadyState() == PrefixAllocationFailed
}

func IsRootPrefix(prefix *networkv1alpha1.Prefix) bool {
	return prefix.Spec.ParentRef == nil &&
		prefix.Spec.ParentSelector == nil
}

func IsPrefixReady(prefix *networkv1alpha1.Prefix) bool {
	return conditionutils.MustFindSliceStatus(prefix.Status.Conditions, string(networkv1alpha1.PrefixReady)) == corev1.ConditionTrue
}

// PrefixReconciler reconciles a Prefix object
type PrefixReconciler struct {
	client.Client
	Scheme *runtime.Scheme
}

//+kubebuilder:rbac:groups=network.onmetal.de,resources=prefixes,verbs=get;list;watch;create;update;patch;delete
//+kubebuilder:rbac:groups=network.onmetal.de,resources=prefixes/status,verbs=get;update;patch
//+kubebuilder:rbac:groups=network.onmetal.de,resources=prefixes/finalizers,verbs=update
//+kubebuilder:rbac:groups=network.onmetal.de,resources=prefixallocations,verbs=get;list;watch;create;update;patch;delete
//+kubebuilder:rbac:groups=network.onmetal.de,resources=prefixallocations/status,verbs=get;update;patch

// Reconcile is part of the main kubernetes reconciliation loop which aims to
// move the current state of the cluster closer to the desired state.
func (r *PrefixReconciler) Reconcile(ctx context.Context, req ctrl.Request) (ctrl.Result, error) {
	log := ctrl.LoggerFrom(ctx)
	prefix := &networkv1alpha1.Prefix{}
	if err := r.Get(ctx, req.NamespacedName, prefix); err != nil {
		return ctrl.Result{}, client.IgnoreNotFound(err)
	}

	return r.reconcileExists(ctx, log, prefix)
}

func (r *PrefixReconciler) reconcileExists(ctx context.Context, log logr.Logger, prefix *networkv1alpha1.Prefix) (ctrl.Result, error) {
	if !prefix.DeletionTimestamp.IsZero() {
		return r.delete(ctx, log, prefix)
	}
	return r.reconcile(ctx, log, prefix)
}

func (r *PrefixReconciler) delete(ctx context.Context, log logr.Logger, prefix *networkv1alpha1.Prefix) (ctrl.Result, error) {
	if !controllerutil.ContainsFinalizer(prefix, PrefixFinalizer) {
		return ctrl.Result{}, nil
	}

	log.V(1).Info("Listing dependent prefix allocations")
	list := &networkv1alpha1.PrefixAllocationList{}
	if err := r.List(ctx, list, client.InNamespace(prefix.Namespace), client.MatchingFields{
		prefixAllocationSpecPrefixRefIfKindIsPrefixNameField: prefix.Name,
	}, client.Limit(1)); err != nil {
		return ctrl.Result{}, fmt.Errorf("error listing dependent allocations: %w", err)
	}

	if len(list.Items) > 0 {
		log.V(1).Info("There are still dependent allocations")
		return ctrl.Result{}, nil
	}

	log.V(1).Info("No dependent allocations, allowing deletion")
	base := prefix.DeepCopy()
	controllerutil.RemoveFinalizer(prefix, PrefixFinalizer)
	if err := r.Patch(ctx, prefix, client.MergeFrom(base)); err != nil {
		return ctrl.Result{}, fmt.Errorf("error removing finalizer: %w", err)
	}

	return ctrl.Result{}, nil
}

func (r *PrefixReconciler) patchReadiness(ctx context.Context, prefix *networkv1alpha1.Prefix, readyCond networkv1alpha1.PrefixCondition) error {
	base := prefix.DeepCopy()
	conditionutils.MustUpdateSlice(&prefix.Status.Conditions, string(networkv1alpha1.PrefixReady),
		conditionutils.UpdateStatus(readyCond.Status),
		conditionutils.UpdateReason(readyCond.Reason),
		conditionutils.UpdateMessage(readyCond.Message),
	)
	if err := r.Status().Patch(ctx, prefix, client.MergeFrom(base)); err != nil {
		return fmt.Errorf("error patching ready condition: %w", err)
	}
	return nil
}

func (r *PrefixReconciler) allocateSelf(ctx context.Context, log logr.Logger, prefix *networkv1alpha1.Prefix) (ok bool, err error) {
	switch {
	case IsPrefixReady(prefix):
		log.V(1).Info("Prefix is allocated")
		return true, nil
	case IsRootPrefix(prefix):
		log.V(1).Info("Root prefix status needs to be patched to report ready for allocation")
		if err := r.patchReadiness(ctx, prefix, networkv1alpha1.PrefixCondition{
			Status:  corev1.ConditionTrue,
			Reason:  "Allocated",
			Message: "The prefix is a root prefix and thus allocated by default.",
		}); err != nil {
			return false, fmt.Errorf("error marking root prefix as allocated: %w", err)
		}
		return false, nil
	default:
		log.V(1).Info("Prefix is a sub-prefix")
		m := NewPrefixAllocator(r.Client, r.Scheme)
		res, err := m.Apply(ctx, (*Prefix)(prefix))
		switch {
		case err != nil:
			if err := r.patchReadiness(ctx, prefix, networkv1alpha1.PrefixCondition{
				Status:  corev1.ConditionUnknown,
				Reason:  "ErrorAllocating",
				Message: fmt.Sprintf("Allocating resulted in an error: %v.", err),
			}); err != nil {
				log.Error(err, "Error patching readiness")
			}
			return false, fmt.Errorf("error applying allocation: %w", err)
		case res.ReadyState() == PrefixAllocationSucceeded && !prefix.Spec.Prefix.IsValid():
			if err := r.assignPrefixValues(ctx, prefix, res); err != nil {
				return false, fmt.Errorf("error patching prefix assignment: %w", err)
			}
			return false, nil
		case res.ReadyState() == PrefixAllocationSucceeded:
			log.V(1).Info("Marking sub prefix as allocated")
			if err := r.patchReadiness(ctx, prefix, networkv1alpha1.PrefixCondition{
				Status:  corev1.ConditionTrue,
				Reason:  networkv1alpha1.PrefixReadyReasonAllocated,
				Message: "Prefix has been successfully allocated.",
			}); err != nil {
				return false, fmt.Errorf("error marking prefix as allocated: %w", err)
			}
			return false, nil
		default:
			if err := r.patchReadiness(ctx, prefix, networkv1alpha1.PrefixCondition{
				Status:  corev1.ConditionFalse,
				Reason:  networkv1alpha1.PrefixReadyReasonPending,
				Message: "Prefix is not yet allocated.",
			}); err != nil {
				return false, fmt.Errorf("error marking prefix as pending: %w", err)
			}
			return false, nil
		}
	}
}

func (r *PrefixReconciler) assignPrefixValues(ctx context.Context, prefix *networkv1alpha1.Prefix, res PrefixAllocationer) error {
	base := prefix.DeepCopy()
	prefix.Spec.Prefix = res.Result().Prefix
	set, err := IPSetFromNetaddrPrefix(res.Result().Prefix.IPPrefix)
	if err != nil {
		return fmt.Errorf("error computing ip set: %w", err)
	}

	for i, reservationLength := range prefix.Spec.ReservationLengths {
		var (
			ok          bool
			reservation netaddr.IPPrefix
		)
		reservation, set, ok = set.RemoveFreePrefix(uint8(reservationLength))
		if !ok {
			return fmt.Errorf("could not fit reservation %d of length %d", i, reservationLength)
		}

		prefix.Spec.Reservations = append(prefix.Spec.Reservations, commonv1alpha1.IPPrefix{IPPrefix: reservation})
	}

	parentRef := res.PrefixRef()
	prefix.Spec.ParentRef = parentRef
	if err := r.Patch(ctx, prefix, client.MergeFrom(base)); err != nil {
		return fmt.Errorf("error assigning prefix: %w", err)
	}
	return nil
}

func (r *PrefixReconciler) processAllocations(ctx context.Context, log logr.Logger, prefix *networkv1alpha1.Prefix) (available, reserved []commonv1alpha1.IPPrefix, err error) {
	list := &networkv1alpha1.PrefixAllocationList{}
	log.V(1).Info("Listing referencing allocations")
	if err := r.List(ctx, list, client.InNamespace(prefix.Namespace), client.MatchingFields{
		prefixAllocationSpecPrefixRefIfKindIsPrefixNameField: prefix.Name,
	}); err != nil {
		return nil, nil, fmt.Errorf("error listing allocations: %w", err)
	}

	var (
		availableBuilder netaddr.IPSetBuilder
		newAllocations   []networkv1alpha1.PrefixAllocation
	)
	availableBuilder.AddPrefix(prefix.Spec.Prefix.IPPrefix)
	for _, allocation := range list.Items {
		switch {
		case IsPrefixAllocationSucceeded(&allocation):
			if allocation.Status.Prefix.IsValid() {
				availableBuilder.RemovePrefix(allocation.Status.Prefix.IPPrefix)
			} else {
				availableBuilder.RemoveRange(allocation.Status.Range.Range())
			}
		case !IsPrefixAllocationFailed(&allocation):
			newAllocations = append(newAllocations, allocation)
		}
	}

	availableSet, err := availableBuilder.IPSet()
	if err != nil {
		return nil, nil, fmt.Errorf("error building available set: %w", err)
	}

	for _, reservation := range prefix.Spec.Reservations {
		var reservedPrefixes []netaddr.IPPrefix
		reservedPrefixes, availableSet = RemoveIntersection(availableSet, reservation.IPPrefix)
		for _, reservedPrefix := range reservedPrefixes {
			reserved = append(reserved, commonv1alpha1.IPPrefix{IPPrefix: reservedPrefix})
		}
	}

	for _, newAllocation := range newAllocations {
		newAvailableSet, err := r.processAllocation(ctx, log, availableSet, &newAllocation)
		if err != nil {
			log.V(1).Error(err, "Error processing allocation", "Allocation", newAllocation)
			continue
		}

		availableSet = newAvailableSet
	}

	availablePrefixes := availableSet.Prefixes()
	for _, availablePrefix := range availablePrefixes {
		available = append(available, commonv1alpha1.IPPrefix{IPPrefix: availablePrefix})
	}
	return available, reserved, nil
}

func (r *PrefixReconciler) patchAllocationReadiness(
	ctx context.Context,
	allocation *networkv1alpha1.PrefixAllocation,
	prefix netaddr.IPPrefix,
	rng netaddr.IPRange,
	readyCond networkv1alpha1.PrefixAllocationCondition,
) error {
	base := allocation.DeepCopy()
	if prefix.IsValid() {
		allocation.Status.Prefix = commonv1alpha1.IPPrefix{IPPrefix: prefix}
	}
	if rng.IsValid() {
		allocation.Status.Range = commonv1alpha1.PtrToIPRange(commonv1alpha1.NewIPRange(rng))
	}
	conditionutils.MustUpdateSlice(&allocation.Status.Conditions, string(networkv1alpha1.PrefixAllocationReady),
		conditionutils.UpdateStatus(readyCond.Status),
		conditionutils.UpdateReason(readyCond.Reason),
		conditionutils.UpdateMessage(readyCond.Message),
	)
	if err := r.Status().Patch(ctx, allocation, client.MergeFrom(base)); err != nil {
		return fmt.Errorf("error updating allocation status: %w", err)
	}
	return nil
}

func (r *PrefixReconciler) processAllocation(ctx context.Context, log logr.Logger, available *netaddr.IPSet, allocation *networkv1alpha1.PrefixAllocation) (*netaddr.IPSet, error) {
	log = log.WithValues("AllocationKey", client.ObjectKeyFromObject(allocation))
	prefix, rng, newAvailableSet, ok := RemoveFreePrefixForRequest(available, allocation.Spec.PrefixAllocationRequest)
	switch {
	case !ok && allocation.Spec.PrefixSelector != nil:
		log.V(1).Info("Evicting non-allocatable allocation")
		if err := r.Delete(ctx, allocation); client.IgnoreNotFound(err) != nil {
			return available, fmt.Errorf("error evicting allocation: %w", err)
		}
		return available, nil
	case !ok:
		log.V(1).Info("Marking non-allocatable allocation as pending")
		if err := r.patchAllocationReadiness(ctx, allocation, netaddr.IPPrefix{}, netaddr.IPRange{}, networkv1alpha1.PrefixAllocationCondition{
			Status:  corev1.ConditionFalse,
			Reason:  networkv1alpha1.PrefixAllocationReadyReasonPending,
			Message: "Could not allocate request.",
		}); err != nil {
			return available, fmt.Errorf("could not mark allocation as pending")
		}
		return available, nil
	default:
		log.V(1).Info("Marking allocation as allocated")
		if err := r.patchAllocationReadiness(ctx, allocation, prefix, rng, networkv1alpha1.PrefixAllocationCondition{
			Status:  corev1.ConditionTrue,
			Reason:  networkv1alpha1.PrefixAllocationReadyReasonSucceeded,
			Message: "The request has been successfully allocated.",
		}); err != nil {
			return available, fmt.Errorf("error marking allocation as ready: %w", err)
		}
		return newAvailableSet, nil
	}
}

func (r *PrefixReconciler) reconcile(ctx context.Context, log logr.Logger, prefix *networkv1alpha1.Prefix) (ctrl.Result, error) {
	if !controllerutil.ContainsFinalizer(prefix, PrefixFinalizer) {
		base := prefix.DeepCopy()
		controllerutil.AddFinalizer(prefix, PrefixFinalizer)
		if err := r.Patch(ctx, prefix, client.MergeFrom(base)); err != nil {
			return ctrl.Result{}, fmt.Errorf("error adding finalizer: %w", err)
		}
		return ctrl.Result{}, nil
	}

	ok, err := r.allocateSelf(ctx, log, prefix)
	if err != nil {
		return ctrl.Result{}, fmt.Errorf("error allocating prefix: %w", err)
	}
	if !ok {
		return ctrl.Result{}, nil
	}

	available, reserved, err := r.processAllocations(ctx, log, prefix)
	if err != nil {
		return ctrl.Result{}, fmt.Errorf("error computing usages: %w", err)
	}

	log.V(1).Info("Updating status", "Available", available, "Reserved", reserved)
	base := prefix.DeepCopy()
	prefix.Status.Available = available
	prefix.Status.Reserved = reserved
	if err := r.Status().Patch(ctx, prefix, client.MergeFrom(base)); err != nil {
		return ctrl.Result{}, fmt.Errorf("error patching status: %w", err)
	}
	return ctrl.Result{}, nil
}

const (
	prefixAllocationSpecPrefixRefIfKindIsPrefixNameField = ".spec.prefixRef[?(@.kind == 'Prefix')].name"
	prefixSpecParentRefIfKindIsPrefixNameField           = ".spec.parentRef[?(@.kind == 'Prefix')].name"
	prefixSpecParentRefIfKindIsClusterPrefixName         = ".spec.parentRef[?(@.kind == 'ClusterPrefix')].name"
	prefixSpecParentSelectorKind                         = ".spec.parentSelector.kind"
)

// SetupWithManager sets up the controller with the Manager.
func (r *PrefixReconciler) SetupWithManager(mgr ctrl.Manager) error {
	ctx := context.Background()
	log := ctrl.Log.WithName("prefix").WithName("setup")
	if err := mgr.GetFieldIndexer().IndexField(ctx, &networkv1alpha1.PrefixAllocation{}, prefixAllocationSpecPrefixRefIfKindIsPrefixNameField, func(obj client.Object) []string {
		allocation := obj.(*networkv1alpha1.PrefixAllocation)
		prefixRef := allocation.Spec.PrefixRef
		if prefixRef == nil || prefixRef.Kind != networkv1alpha1.PrefixKind {
			return nil
		}
		return []string{prefixRef.Name}
	}); err != nil {
		return fmt.Errorf("error indexing field %q: %w", prefixAllocationSpecPrefixRefIfKindIsPrefixNameField, err)
	}

	if err := mgr.GetFieldIndexer().IndexField(ctx, &networkv1alpha1.Prefix{}, prefixSpecParentRefIfKindIsPrefixNameField, func(obj client.Object) []string {
		prefix := obj.(*networkv1alpha1.Prefix)
		parentRef := prefix.Spec.ParentRef
		if parentRef == nil || parentRef.Kind != networkv1alpha1.PrefixKind {
			return nil
		}
		return []string{parentRef.Name}
	}); err != nil {
		return fmt.Errorf("error indexing field %q: %w", prefixSpecParentRefIfKindIsPrefixNameField, err)
	}

	if err := mgr.GetFieldIndexer().IndexField(ctx, &networkv1alpha1.Prefix{}, prefixSpecParentSelectorKind, func(obj client.Object) []string {
		prefix := obj.(*networkv1alpha1.Prefix)
		if prefix.Spec.ParentSelector == nil {
			return nil
		}
		return []string{prefix.Spec.ParentSelector.Kind}
	}); err != nil {
		return fmt.Errorf("error indexing field %q: %w", prefixSpecParentRefIfKindIsPrefixNameField, err)
	}

	if err := mgr.GetFieldIndexer().IndexField(ctx, &networkv1alpha1.Prefix{}, prefixSpecParentRefIfKindIsClusterPrefixName, func(obj client.Object) []string {
		prefix := obj.(*networkv1alpha1.Prefix)
		parentRef := prefix.Spec.ParentRef
		if parentRef == nil || parentRef.Kind != networkv1alpha1.ClusterPrefixKind {
			return nil
		}
		return []string{parentRef.Name}
	}); err != nil {
		return fmt.Errorf("error indexing field")
	}

	return ctrl.NewControllerManagedBy(mgr).
		For(&networkv1alpha1.Prefix{}).
		Owns(&networkv1alpha1.PrefixAllocation{}).
		Watches(
			&source.Kind{Type: &networkv1alpha1.PrefixAllocation{}},
			r.enqueueByAllocationPrefixRef(),
		).
		Watches(
			&source.Kind{Type: &networkv1alpha1.Prefix{}},
			r.enqueueByPrefixParentRef(ctx, log),
		).
		Watches(
			&source.Kind{Type: &networkv1alpha1.Prefix{}},
			r.enqueueByPrefixParentSelector(ctx, log),
		).
		Watches(
			&source.Kind{Type: &networkv1alpha1.ClusterPrefix{}},
			r.enqueueByPrefixClusterParentRef(ctx, log),
		).
		Watches(
			&source.Kind{Type: &networkv1alpha1.ClusterPrefix{}},
			r.enqueueByPrefixClusterParentSelector(ctx, log),
		).
		Complete(r)
}

func (r *PrefixReconciler) enqueueByAllocationPrefixRef() handler.EventHandler {
	return handler.EnqueueRequestsFromMapFunc(func(obj client.Object) []ctrl.Request {
		allocation := obj.(*networkv1alpha1.PrefixAllocation)
		prefixRef := allocation.Spec.PrefixRef
		if prefixRef == nil || prefixRef.Kind != networkv1alpha1.PrefixKind {
			return nil
		}
		return []ctrl.Request{
			{
				NamespacedName: client.ObjectKey{
					Namespace: allocation.Namespace,
					Name:      prefixRef.Name,
				},
			},
		}
	})
}

func (r *PrefixReconciler) enqueueByPrefixClusterParentSelector(ctx context.Context, log logr.Logger) handler.EventHandler {
	return handler.EnqueueRequestsFromMapFunc(func(obj client.Object) []ctrl.Request {
		clusterPrefix := obj.(*networkv1alpha1.ClusterPrefix)
		if !IsClusterPrefixReady(clusterPrefix) {
			return nil
		}

		list := &networkv1alpha1.PrefixList{}
		if err := r.List(ctx, list, client.MatchingFields{prefixSpecParentSelectorKind: networkv1alpha1.ClusterPrefixKind}); err != nil {
			log.V(1).Error(err, "Error listing prefixes")
			return nil
		}

		var requests []ctrl.Request
		for _, other := range list.Items {
			if other.Spec.ParentRef != nil {
				continue
			}

			sel, err := metav1.LabelSelectorAsSelector(&other.Spec.ParentSelector.LabelSelector)
			if err != nil {
				log.Error(err, "Invalid label selector", "Key", client.ObjectKeyFromObject(&other))
				continue
			}

			if sel.Matches(labels.Set(clusterPrefix.Labels)) {
				log.V(6).Info("Enqueueing prefix", "Key", client.ObjectKeyFromObject(&other))
				requests = append(requests, ctrl.Request{
					NamespacedName: client.ObjectKeyFromObject(&other),
				})
			}
		}

		return requests
	})
}

func (r *PrefixReconciler) enqueueByPrefixParentSelector(ctx context.Context, log logr.Logger) handler.EventHandler {
	return handler.EnqueueRequestsFromMapFunc(func(obj client.Object) []ctrl.Request {
		prefix := obj.(*networkv1alpha1.Prefix)
		if !IsPrefixReady(prefix) {
			return nil
		}

		list := &networkv1alpha1.PrefixList{}
		if err := r.List(ctx, list, client.InNamespace(prefix.Namespace), client.MatchingFields{prefixSpecParentSelectorKind: networkv1alpha1.PrefixKind}); err != nil {
			log.V(1).Error(err, "Error listing prefixes in namespace", "Namespace", prefix.Namespace)
			return nil
		}

		var requests []ctrl.Request
		for _, other := range list.Items {
			if other.Spec.ParentRef != nil {
				continue
			}

			sel, err := metav1.LabelSelectorAsSelector(&other.Spec.ParentSelector.LabelSelector)
			if err != nil {
				log.Error(err, "Invalid label selector", "Key", client.ObjectKeyFromObject(&other))
				continue
			}

			if sel.Matches(labels.Set(prefix.Labels)) {
				log.V(6).Info("Enqueueing prefix", "Key", client.ObjectKeyFromObject(&other))
				requests = append(requests, ctrl.Request{
					NamespacedName: client.ObjectKeyFromObject(&other),
				})
			}
		}

		return requests
	})
}

func (r *PrefixReconciler) enqueueByPrefixParentRef(ctx context.Context, log logr.Logger) handler.EventHandler {
	return handler.EnqueueRequestsFromMapFunc(func(obj client.Object) []ctrl.Request {
		prefix := obj.(*networkv1alpha1.Prefix)
		if !IsPrefixReady(prefix) {
			return nil
		}

		list := &networkv1alpha1.PrefixList{}
		if err := r.List(ctx, list, client.InNamespace(prefix.Namespace), client.MatchingFields{
			prefixSpecParentRefIfKindIsPrefixNameField: prefix.Name,
		}); err != nil {
			log.Error(err, "Error listing prefixes with parent", "Key", client.ObjectKeyFromObject(prefix))
			return nil
		}

		requests := make([]ctrl.Request, 0, len(list.Items))
		for _, child := range list.Items {
			requests = append(requests, ctrl.Request{NamespacedName: client.ObjectKeyFromObject(&child)})
		}
		return requests
	})
}

func (r *PrefixReconciler) enqueueByPrefixClusterParentRef(ctx context.Context, log logr.Logger) handler.EventHandler {
	return handler.EnqueueRequestsFromMapFunc(func(obj client.Object) []ctrl.Request {
		clusterPrefix := obj.(*networkv1alpha1.ClusterPrefix)
		if !IsClusterPrefixReady(clusterPrefix) {
			return nil
		}

		list := &networkv1alpha1.PrefixList{}
		if err := r.List(ctx, list, client.MatchingFields{
			prefixSpecParentRefIfKindIsClusterPrefixName: clusterPrefix.Name,
		}); err != nil {
			log.Error(err, "Error listing prefixes with parent", "Key", client.ObjectKeyFromObject(clusterPrefix))
			return nil
		}

		requests := make([]ctrl.Request, 0, len(list.Items))
		for _, child := range list.Items {
			requests = append(requests, ctrl.Request{NamespacedName: client.ObjectKeyFromObject(&child)})
		}
		return requests
	})
}

type Prefix networkv1alpha1.Prefix

func (p *Prefix) Object() client.Object {
	return (*networkv1alpha1.Prefix)(p)
}

const (
	PrefixAllocationPrefixLabel = "network.onmetal.de/prefix"
)

func (p *Prefix) Label() string {
	return PrefixAllocationPrefixLabel
}

func (p *Prefix) Request() networkv1alpha1.PrefixAllocationRequest {
	return networkv1alpha1.PrefixAllocationRequest{
		Prefix:       p.Spec.Prefix,
		PrefixLength: p.Spec.PrefixLength,
	}
}

func (p *Prefix) PrefixRef() *networkv1alpha1.PrefixReference {
	return p.Spec.ParentRef
}

func (p *Prefix) PrefixSelector() *networkv1alpha1.PrefixSelector {
	return p.Spec.ParentSelector
}

func (p *Prefix) Available() []commonv1alpha1.IPPrefix {
	return p.Status.Available
}
