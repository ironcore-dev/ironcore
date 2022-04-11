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

package ipam

import (
	"context"
	"fmt"

	"github.com/go-logr/logr"
	"github.com/onmetal/controller-utils/conditionutils"
	commonv1alpha1 "github.com/onmetal/onmetal-api/apis/common/v1alpha1"
	ipamv1alpha1 "github.com/onmetal/onmetal-api/apis/ipam/v1alpha1"
	"inet.af/netaddr"
	corev1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/runtime"
	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/controller/controllerutil"
	"sigs.k8s.io/controller-runtime/pkg/handler"
	"sigs.k8s.io/controller-runtime/pkg/source"
)

const (
	ClusterPrefixFinalizer = "ipam.api.onmetal.de/clusterprefix"
)

func IsRootClusterPrefix(prefix *ipamv1alpha1.ClusterPrefix) bool {
	return prefix.Spec.ParentRef == nil &&
		prefix.Spec.ParentSelector == nil
}

func IsClusterPrefixReady(prefix *ipamv1alpha1.ClusterPrefix) bool {
	return conditionutils.MustFindSliceStatus(prefix.Status.Conditions, string(ipamv1alpha1.ClusterPrefixReady)) == corev1.ConditionTrue
}

func IsClusterPrefixAllocationSucceeded(allocation *ipamv1alpha1.ClusterPrefixAllocation) bool {
	return (*ClusterPrefixAllocation)(allocation).ReadyState() == PrefixAllocationSucceeded
}

func IsClusterPrefixAllocationFailed(allocation *ipamv1alpha1.ClusterPrefixAllocation) bool {
	return (*ClusterPrefixAllocation)(allocation).ReadyState() == PrefixAllocationFailed
}

// ClusterPrefixReconciler reconciles a ClusterPrefix object
type ClusterPrefixReconciler struct {
	client.Client
	Scheme *runtime.Scheme
}

//+kubebuilder:rbac:groups=ipam.api.onmetal.de,resources=clusterprefixes,verbs=get;list;watch;create;update;patch;delete
//+kubebuilder:rbac:groups=ipam.api.onmetal.de,resources=clusterprefixes/status,verbs=get;update;patch
//+kubebuilder:rbac:groups=ipam.api.onmetal.de,resources=clusterprefixes/finalizers,verbs=update
//+kubebuilder:rbac:groups=ipam.api.onmetal.de,resources=clusterprefixallocations,verbs=get;list;watch;create;update;patch;delete
//+kubebuilder:rbac:groups=ipam.api.onmetal.de,resources=clusterprefixallocations/status,verbs=get;update;patch
//+kubebuilder:rbac:groups=ipam.api.onmetal.de,resources=prefixallocations,verbs=get;list;watch;create;update;patch;delete
//+kubebuilder:rbac:groups=ipam.api.onmetal.de,resources=prefixallocations/status,verbs=get;update;patch

// Reconcile is part of the main kubernetes reconciliation loop which aims to
// move the current state of the cluster closer to the desired state.
func (r *ClusterPrefixReconciler) Reconcile(ctx context.Context, req ctrl.Request) (ctrl.Result, error) {
	log := ctrl.LoggerFrom(ctx)
	prefix := &ipamv1alpha1.ClusterPrefix{}
	if err := r.Get(ctx, req.NamespacedName, prefix); err != nil {
		return ctrl.Result{}, client.IgnoreNotFound(err)
	}

	return r.reconcileExists(ctx, log, prefix)
}

func (r *ClusterPrefixReconciler) reconcileExists(ctx context.Context, log logr.Logger, prefix *ipamv1alpha1.ClusterPrefix) (ctrl.Result, error) {
	if !prefix.DeletionTimestamp.IsZero() {
		return r.delete(ctx, log, prefix)
	}
	return r.reconcile(ctx, log, prefix)
}

func (r *ClusterPrefixReconciler) delete(ctx context.Context, log logr.Logger, prefix *ipamv1alpha1.ClusterPrefix) (ctrl.Result, error) {
	if !controllerutil.ContainsFinalizer(prefix, ClusterPrefixFinalizer) {
		return ctrl.Result{}, nil
	}

	log.V(1).Info("Listing dependent cluster prefix allocations")
	clusterList := &ipamv1alpha1.ClusterPrefixAllocationList{}
	if err := r.List(ctx, clusterList, client.MatchingFields{
		clusterPrefixAllocationSpecPrefixRefNameField: prefix.Name,
	}, client.Limit(1)); err != nil {
		return ctrl.Result{}, fmt.Errorf("error listing dependent cluster allocations: %w", err)
	}

	if len(clusterList.Items) > 0 {
		log.V(1).Info("There are still dependent cluster allocations")
		return ctrl.Result{}, nil
	}

	log.V(1).Info("Listing dependent prefix allocations")
	list := &ipamv1alpha1.PrefixAllocationList{}
	if err := r.List(ctx, list, client.MatchingFields{
		prefixAllocationSpecPrefixRefIfKindIsClusterPrefixNameField: prefix.Name,
	}, client.Limit(1)); err != nil {
		return ctrl.Result{}, fmt.Errorf("error listing dependent allocations: %w", err)
	}

	if len(list.Items) > 0 {
		log.V(1).Info("There are still dependent allocations")
		return ctrl.Result{}, nil
	}

	log.V(1).Info("No dependent allocations, allowing deletion")
	base := prefix.DeepCopy()
	controllerutil.RemoveFinalizer(prefix, ClusterPrefixFinalizer)
	if err := r.Patch(ctx, prefix, client.MergeFrom(base)); err != nil {
		return ctrl.Result{}, fmt.Errorf("error removing finalizer: %w", err)
	}

	return ctrl.Result{}, nil
}

func (r *ClusterPrefixReconciler) patchReadiness(ctx context.Context, prefix *ipamv1alpha1.ClusterPrefix, readyCond ipamv1alpha1.ClusterPrefixCondition) error {
	base := prefix.DeepCopy()
	conditionutils.MustUpdateSlice(&prefix.Status.Conditions, string(ipamv1alpha1.ClusterPrefixReady),
		conditionutils.UpdateStatus(readyCond.Status),
		conditionutils.UpdateReason(readyCond.Reason),
		conditionutils.UpdateMessage(readyCond.Message),
	)
	if err := r.Status().Patch(ctx, prefix, client.MergeFrom(base)); err != nil {
		return fmt.Errorf("error patching ready condition: %w", err)
	}
	return nil
}

func (r *ClusterPrefixReconciler) assignValues(ctx context.Context, prefix *ipamv1alpha1.ClusterPrefix, res PrefixAllocationer) error {
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
	if parentRef.Kind != ipamv1alpha1.ClusterPrefixKind {
		return fmt.Errorf("invalid parent ref kind %q, expected %q", parentRef.Kind, ipamv1alpha1.ClusterPrefixKind)
	}
	prefix.Spec.ParentRef = &corev1.LocalObjectReference{Name: parentRef.Name}
	if err := r.Patch(ctx, prefix, client.MergeFrom(base)); err != nil {
		return fmt.Errorf("error assigning prefix: %w", err)
	}
	return nil
}

func (r *ClusterPrefixReconciler) allocateSelf(ctx context.Context, log logr.Logger, prefix *ipamv1alpha1.ClusterPrefix) (bool, error) {
	switch {
	case IsClusterPrefixReady(prefix):
		log.V(1).Info("Prefix is ready")
		return true, nil
	case IsRootClusterPrefix(prefix):
		log.V(1).Info("Marking root prefix as ready")
		if err := r.patchReadiness(ctx, prefix, ipamv1alpha1.ClusterPrefixCondition{
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
		res, err := m.Apply(ctx, (*ClusterPrefix)(prefix))
		switch {
		case err != nil:
			if err := r.patchReadiness(ctx, prefix, ipamv1alpha1.ClusterPrefixCondition{
				Status:  corev1.ConditionUnknown,
				Reason:  "ErrorAllocating",
				Message: fmt.Sprintf("Allocating resulted in an error: %v.", err),
			}); err != nil {
				log.Error(err, "Error patching readiness")
			}
			return false, fmt.Errorf("error applying allocation: %w", err)
		case res.ReadyState() == PrefixAllocationSucceeded && !prefix.Spec.Prefix.IsValid():
			if err := r.assignValues(ctx, prefix, res); err != nil {
				return false, fmt.Errorf("error patching prefix assignment: %w", err)
			}
			return false, nil
		case res.ReadyState() == PrefixAllocationSucceeded:
			log.V(1).Info("Marking sub prefix as allocated")
			if err := r.patchReadiness(ctx, prefix, ipamv1alpha1.ClusterPrefixCondition{
				Status:  corev1.ConditionTrue,
				Reason:  ipamv1alpha1.PrefixReadyReasonAllocated,
				Message: "Prefix has been successfully allocated.",
			}); err != nil {
				return false, fmt.Errorf("error marking prefix as allocated: %w", err)
			}
			return false, nil
		default:
			if err := r.patchReadiness(ctx, prefix, ipamv1alpha1.ClusterPrefixCondition{
				Status:  corev1.ConditionFalse,
				Reason:  ipamv1alpha1.PrefixReadyReasonPending,
				Message: "Prefix is not yet allocated.",
			}); err != nil {
				return false, fmt.Errorf("error marking prefix as pending: %w", err)
			}
			return false, nil
		}
	}
}

func (r *ClusterPrefixReconciler) processAllocations(ctx context.Context, log logr.Logger, prefix *ipamv1alpha1.ClusterPrefix) (available, reserved []commonv1alpha1.IPPrefix, err error) {
	var availableBuilder netaddr.IPSetBuilder
	availableBuilder.AddPrefix(prefix.Spec.Prefix.IPPrefix)

	clusterList := &ipamv1alpha1.ClusterPrefixAllocationList{}
	log.V(1).Info("Listing referencing cluster prefix allocations")
	if err := r.List(ctx, clusterList, client.MatchingFields{
		clusterPrefixAllocationSpecPrefixRefNameField: prefix.Name,
	}); err != nil {
		return nil, nil, fmt.Errorf("error listing cluster allocations: %w", err)
	}

	var newClusterAllocations []ipamv1alpha1.ClusterPrefixAllocation
	for _, allocation := range clusterList.Items {
		switch {
		case IsClusterPrefixAllocationSucceeded(&allocation):
			availableBuilder.RemovePrefix(allocation.Status.Prefix.IPPrefix)
		case !IsClusterPrefixAllocationFailed(&allocation):
			newClusterAllocations = append(newClusterAllocations, allocation)
		}
	}

	log.V(1).Info("Listing referencing namespaced allocations")
	list := &ipamv1alpha1.PrefixAllocationList{}
	if err := r.List(ctx, list, client.MatchingFields{
		prefixAllocationSpecPrefixRefIfKindIsClusterPrefixNameField: prefix.Name,
	}); err != nil {
		return nil, nil, fmt.Errorf("error listing namespaced allocations: %w", err)
	}

	var newAllocations []ipamv1alpha1.PrefixAllocation
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

	for _, newAllocation := range newClusterAllocations {
		newAvailableSet, err := r.processClusterAllocation(ctx, log, availableSet, &newAllocation)
		if err != nil {
			log.V(1).Error(err, "Error processing cluster allocation", "ClusterAllocation", newAllocation)
			continue
		}

		availableSet = newAvailableSet
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

func (r *ClusterPrefixReconciler) patchAllocationReadiness(
	ctx context.Context,
	allocation *ipamv1alpha1.PrefixAllocation,
	prefix netaddr.IPPrefix,
	rng netaddr.IPRange,
	readyCond ipamv1alpha1.PrefixAllocationCondition,
) error {
	base := allocation.DeepCopy()
	if prefix.IsValid() {
		allocation.Status.Prefix = commonv1alpha1.IPPrefix{IPPrefix: prefix}
	}
	if rng.IsValid() {
		allocation.Status.Range = commonv1alpha1.NewIPRangePtr(rng)
	}
	conditionutils.MustUpdateSlice(&allocation.Status.Conditions, string(ipamv1alpha1.PrefixAllocationReady),
		conditionutils.UpdateStatus(readyCond.Status),
		conditionutils.UpdateReason(readyCond.Reason),
		conditionutils.UpdateMessage(readyCond.Message),
	)
	if err := r.Status().Patch(ctx, allocation, client.MergeFrom(base)); err != nil {
		return fmt.Errorf("error updating allocation status: %w", err)
	}
	return nil
}

func (r *ClusterPrefixReconciler) processAllocation(ctx context.Context, log logr.Logger, available *netaddr.IPSet, allocation *ipamv1alpha1.PrefixAllocation) (*netaddr.IPSet, error) {
	log = log.WithValues("AllocationKey", client.ObjectKeyFromObject(allocation))
	prefix, rng, newAvailableSet, ok := RemoveFreePrefixForRequest(available, allocation.Spec.PrefixAllocationRequest)
	switch {
	case !ok && allocation.Spec.PrefixSelector != nil:
		log.V(1).Info("Evicting scheduled non-allocatable allocation")
		if err := r.Delete(ctx, allocation); client.IgnoreNotFound(err) != nil {
			return available, fmt.Errorf("error evicting allocation: %w", err)
		}
		return available, nil
	case !ok:
		log.V(1).Info("Marking non-allocatable allocation as pending")
		if err := r.patchAllocationReadiness(ctx, allocation, netaddr.IPPrefix{}, netaddr.IPRange{}, ipamv1alpha1.PrefixAllocationCondition{
			Status:  corev1.ConditionFalse,
			Reason:  ipamv1alpha1.PrefixAllocationReadyReasonPending,
			Message: "Could not allocate request.",
		}); client.IgnoreNotFound(err) != nil {
			return available, fmt.Errorf("could not mark allocation as pending")
		}
		return available, nil
	default:
		log.V(1).Info("Marking allocation as allocated")
		if err := r.patchAllocationReadiness(ctx, allocation, prefix, rng, ipamv1alpha1.PrefixAllocationCondition{
			Status:  corev1.ConditionTrue,
			Reason:  ipamv1alpha1.PrefixAllocationReadyReasonSucceeded,
			Message: "The request has been successfully allocated.",
		}); err != nil {
			return available, fmt.Errorf("error marking allocation as ready: %w", err)
		}
		return newAvailableSet, nil
	}
}

func (r *ClusterPrefixReconciler) patchClusterAllocationReadiness(
	ctx context.Context,
	allocation *ipamv1alpha1.ClusterPrefixAllocation,
	prefix netaddr.IPPrefix,
	readyCond ipamv1alpha1.ClusterPrefixAllocationCondition,
) error {
	base := allocation.DeepCopy()
	allocation.Status.Prefix = commonv1alpha1.IPPrefix{IPPrefix: prefix}
	conditionutils.MustUpdateSlice(&allocation.Status.Conditions, string(ipamv1alpha1.ClusterPrefixAllocationReady),
		conditionutils.UpdateStatus(readyCond.Status),
		conditionutils.UpdateReason(readyCond.Reason),
		conditionutils.UpdateMessage(readyCond.Message),
	)
	if err := r.Status().Patch(ctx, allocation, client.MergeFrom(base)); err != nil {
		return fmt.Errorf("error updating allocation status: %w", err)
	}
	return nil
}

func (r *ClusterPrefixReconciler) processClusterAllocation(ctx context.Context, log logr.Logger, available *netaddr.IPSet, allocation *ipamv1alpha1.ClusterPrefixAllocation) (*netaddr.IPSet, error) {
	log = log.WithValues("AllocationKey", client.ObjectKeyFromObject(allocation))
	prefix, _, newAvailableSet, ok := RemoveFreePrefixForRequest(available, ipamv1alpha1.PrefixAllocationRequest{
		Prefix:       allocation.Spec.Prefix,
		PrefixLength: allocation.Spec.PrefixLength,
	})
	switch {
	case !ok && allocation.Spec.PrefixSelector != nil:
		log.V(1).Info("Evicting non-allocatable allocation")
		if err := r.Delete(ctx, allocation); client.IgnoreNotFound(err) != nil {
			return available, fmt.Errorf("error evicting allocation: %w", err)
		}
		return available, nil
	case !ok:
		log.V(1).Info("Marking non-allocatable allocation as pending")
		if err := r.patchClusterAllocationReadiness(ctx, allocation, netaddr.IPPrefix{}, ipamv1alpha1.ClusterPrefixAllocationCondition{
			Status:  corev1.ConditionFalse,
			Reason:  ipamv1alpha1.ClusterPrefixAllocationReadyReasonPending,
			Message: "Could not allocate request.",
		}); client.IgnoreNotFound(err) != nil {
			return available, fmt.Errorf("could not mark allocation as pending")
		}
		return available, nil
	default:
		log.V(1).Info("Marking allocation as allocated")
		if err := r.patchClusterAllocationReadiness(ctx, allocation, prefix, ipamv1alpha1.ClusterPrefixAllocationCondition{
			Status:  corev1.ConditionTrue,
			Reason:  ipamv1alpha1.ClusterPrefixAllocationReadyReasonSucceeded,
			Message: "The request has been successfully allocated.",
		}); err != nil {
			return available, fmt.Errorf("error marking allocation as ready: %w", err)
		}
		return newAvailableSet, nil
	}
}

func (r *ClusterPrefixReconciler) reconcile(ctx context.Context, log logr.Logger, prefix *ipamv1alpha1.ClusterPrefix) (ctrl.Result, error) {
	if !controllerutil.ContainsFinalizer(prefix, ClusterPrefixFinalizer) {
		base := prefix.DeepCopy()
		controllerutil.AddFinalizer(prefix, ClusterPrefixFinalizer)
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
	clusterPrefixAllocationSpecPrefixRefNameField               = ".spec.prefixRef.name"
	prefixAllocationSpecPrefixRefIfKindIsClusterPrefixNameField = ".spec.prefixRef[?(@.kind == 'ClusterPrefix')].name"
)

// SetupWithManager sets up the controller with the Manager.
func (r *ClusterPrefixReconciler) SetupWithManager(mgr ctrl.Manager) error {
	ctx := context.Background()

	if err := mgr.GetFieldIndexer().IndexField(ctx, &ipamv1alpha1.ClusterPrefixAllocation{}, clusterPrefixAllocationSpecPrefixRefNameField, func(obj client.Object) []string {
		prefix := obj.(*ipamv1alpha1.ClusterPrefixAllocation)
		prefixRef := prefix.Spec.PrefixRef
		if prefixRef == nil {
			return nil
		}
		return []string{prefixRef.Name}
	}); err != nil {
		return err
	}

	if err := mgr.GetFieldIndexer().IndexField(ctx, &ipamv1alpha1.PrefixAllocation{}, prefixAllocationSpecPrefixRefIfKindIsClusterPrefixNameField, func(obj client.Object) []string {
		prefix := obj.(*ipamv1alpha1.PrefixAllocation)
		prefixRef := prefix.Spec.PrefixRef
		if prefixRef == nil || prefixRef.Kind != ipamv1alpha1.ClusterPrefixKind {
			return nil
		}
		return []string{prefixRef.Name}
	}); err != nil {
		return err
	}

	return ctrl.NewControllerManagedBy(mgr).
		For(&ipamv1alpha1.ClusterPrefix{}).
		Owns(&ipamv1alpha1.ClusterPrefixAllocation{}).
		Watches(
			&source.Kind{Type: &ipamv1alpha1.PrefixAllocation{}},
			r.enqueueByAllocationPrefixRef(),
		).
		Watches(
			&source.Kind{Type: &ipamv1alpha1.ClusterPrefixAllocation{}},
			r.enqueueByClusterAllocationPrefixRef(),
		).
		Complete(r)
}

func (r *ClusterPrefixReconciler) enqueueByAllocationPrefixRef() handler.EventHandler {
	return handler.EnqueueRequestsFromMapFunc(func(obj client.Object) []ctrl.Request {
		allocation := obj.(*ipamv1alpha1.PrefixAllocation)
		prefixRef := allocation.Spec.PrefixRef
		if prefixRef == nil || prefixRef.Kind != ipamv1alpha1.ClusterPrefixKind {
			return nil
		}
		return []ctrl.Request{
			{
				NamespacedName: client.ObjectKey{
					Name: prefixRef.Name,
				},
			},
		}
	})
}

func (r *ClusterPrefixReconciler) enqueueByClusterAllocationPrefixRef() handler.EventHandler {
	return handler.EnqueueRequestsFromMapFunc(func(obj client.Object) []ctrl.Request {
		allocation := obj.(*ipamv1alpha1.ClusterPrefixAllocation)
		prefixRef := allocation.Spec.PrefixRef
		if prefixRef == nil {
			return nil
		}
		return []ctrl.Request{
			{
				NamespacedName: client.ObjectKey{
					Name: prefixRef.Name,
				},
			},
		}
	})
}

type ClusterPrefix ipamv1alpha1.ClusterPrefix

func (p *ClusterPrefix) Object() client.Object {
	return (*ipamv1alpha1.ClusterPrefix)(p)
}

const (
	ClusterPrefixAllocationPrefixLabel = "ipam.api.onmetal.de/clusterprefix"
)

func (p *ClusterPrefix) Label() string {
	return ClusterPrefixAllocationPrefixLabel
}

func (p *ClusterPrefix) Available() []commonv1alpha1.IPPrefix {
	return p.Status.Available
}

func (p *ClusterPrefix) Request() ipamv1alpha1.PrefixAllocationRequest {
	return ipamv1alpha1.PrefixAllocationRequest{
		Prefix:       p.Spec.Prefix,
		PrefixLength: p.Spec.PrefixLength,
	}
}

func (p *ClusterPrefix) PrefixRef() *ipamv1alpha1.PrefixReference {
	if p.Spec.ParentRef == nil {
		return nil
	}
	return &ipamv1alpha1.PrefixReference{
		Kind: ipamv1alpha1.ClusterPrefixKind,
		Name: p.Spec.ParentRef.Name,
	}
}

func (p *ClusterPrefix) PrefixSelector() *ipamv1alpha1.PrefixSelector {
	if p.Spec.ParentSelector == nil {
		return nil
	}
	return &ipamv1alpha1.PrefixSelector{
		Kind:          ipamv1alpha1.ClusterPrefixKind,
		LabelSelector: *p.Spec.ParentSelector,
	}
}
