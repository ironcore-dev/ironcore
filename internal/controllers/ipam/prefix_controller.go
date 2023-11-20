// SPDX-FileCopyrightText: 2023 SAP SE or an SAP affiliate company and IronCore contributors
// SPDX-License-Identifier: Apache-2.0

package ipam

import (
	"context"
	"fmt"
	"net/netip"
	"sort"
	"sync"
	"time"

	"github.com/go-logr/logr"
	"github.com/ironcore-dev/controller-utils/clientutils"
	commonv1alpha1 "github.com/ironcore-dev/ironcore/api/common/v1alpha1"
	ipamv1alpha1 "github.com/ironcore-dev/ironcore/api/ipam/v1alpha1"
	ipamclient "github.com/ironcore-dev/ironcore/internal/client/ipam"
	"github.com/ironcore-dev/ironcore/utils/equality"
	"go4.org/netipx"
	apierrors "k8s.io/apimachinery/pkg/api/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/labels"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/client-go/util/workqueue"
	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/controller/controllerutil"
	"sigs.k8s.io/controller-runtime/pkg/handler"
)

const (
	prefixFinalizer                   = "ipam.ironcore.dev/prefix"
	prefixAllocationRequesterUIDLabel = "ipam.ironcore.dev/requester-uid"
)

func (r *PrefixReconciler) acquireAllocation(
	prefix *ipamv1alpha1.Prefix,
	set *netipx.IPSet,
	allocation *ipamv1alpha1.PrefixAllocation,
) (res netip.Prefix, newSet *netipx.IPSet, ok bool, terminal bool) {
	if !prefixCompatibleWithAllocation(prefix, allocation) {
		return netip.Prefix{}, set, false, true
	}

	switch {
	case allocation.Spec.Prefix.IsValid():
		requestedPrefix := allocation.Spec.Prefix.Prefix
		if set, ok := r.ipSetRemovePrefix(set, requestedPrefix); ok {
			return requestedPrefix, set, true, true
		}
		return netip.Prefix{}, set, false, false
	case allocation.Spec.PrefixLength > 0:
		requestedPrefixLength := allocation.Spec.PrefixLength
		if prefix, set, ok := set.RemoveFreePrefix(uint8(requestedPrefixLength)); ok {
			return prefix, set, true, true
		}
		return netip.Prefix{}, set, false, false
	default:
		panic(fmt.Sprintf("unhandled allocation %#v", allocation))
	}
}

func (r *PrefixReconciler) ipSetRemovePrefix(set *netipx.IPSet, prefix netip.Prefix) (*netipx.IPSet, bool) {
	if !prefix.IsValid() || !set.ContainsPrefix(prefix) {
		return set, false
	}
	var sb netipx.IPSetBuilder
	sb.AddSet(set)
	sb.RemovePrefix(prefix)
	set, _ = sb.IPSet()
	return set, true
}

// PrefixReconciler reconciles a Prefix object
type PrefixReconciler struct {
	client.Client
	APIReader               client.Reader
	Scheme                  *runtime.Scheme
	PrefixAllocationTimeout time.Duration

	allocationLimiter workqueue.RateLimiter
	waitTimeByKey     sync.Map
}

func (r *PrefixReconciler) allocationBackoffFor(key client.ObjectKey) time.Duration {
	now := time.Now()
	waitTimeIface, _ := r.waitTimeByKey.LoadOrStore(key, now.Add(r.allocationLimiter.When(key)))
	waitTime := waitTimeIface.(time.Time)
	if now.After(waitTime) {
		return 0
	}
	return waitTime.Sub(now)
}

func (r *PrefixReconciler) forgetAllocationBackoffFor(key client.ObjectKey) {
	r.waitTimeByKey.Delete(key)
	r.allocationLimiter.Forget(key)
}

//+kubebuilder:rbac:groups=ipam.ironcore.dev,resources=prefixes,verbs=get;list;watch;create;update;patch;delete
//+kubebuilder:rbac:groups=ipam.ironcore.dev,resources=prefixes/status,verbs=get;update;patch
//+kubebuilder:rbac:groups=ipam.ironcore.dev,resources=prefixes/finalizers,verbs=update
//+kubebuilder:rbac:groups=ipam.ironcore.dev,resources=prefixallocations,verbs=get;list;watch;create;update;patch;delete
//+kubebuilder:rbac:groups=ipam.ironcore.dev,resources=prefixallocations/status,verbs=get;update;patch

// Reconcile is part of the main kubernetes reconciliation loop which aims to
// move the current state of the cluster closer to the desired state.
func (r *PrefixReconciler) Reconcile(ctx context.Context, req ctrl.Request) (ctrl.Result, error) {
	log := ctrl.LoggerFrom(ctx)
	prefix := &ipamv1alpha1.Prefix{}
	if err := r.Get(ctx, req.NamespacedName, prefix); err != nil {
		r.forgetAllocationBackoffFor(req.NamespacedName)
		return ctrl.Result{}, client.IgnoreNotFound(err)
	}

	return r.reconcileExists(ctx, log, prefix)
}

func (r *PrefixReconciler) reconcileExists(ctx context.Context, log logr.Logger, prefix *ipamv1alpha1.Prefix) (ctrl.Result, error) {
	if !prefix.DeletionTimestamp.IsZero() {
		return r.delete(ctx, log, prefix)
	}
	return r.reconcile(ctx, log, prefix)
}

func (r *PrefixReconciler) delete(ctx context.Context, log logr.Logger, prefix *ipamv1alpha1.Prefix) (ctrl.Result, error) {
	if !controllerutil.ContainsFinalizer(prefix, prefixFinalizer) {
		r.forgetAllocationBackoffFor(client.ObjectKeyFromObject(prefix))
		return ctrl.Result{}, nil
	}

	log.V(1).Info("Listing prefix allocations")
	allocationList := &ipamv1alpha1.PrefixAllocationList{}
	if err := r.APIReader.List(ctx, allocationList, client.InNamespace(prefix.Namespace)); err != nil {
		return ctrl.Result{}, fmt.Errorf("error listing dependent allocations: %w", err)
	}

	var dependentAllocations []string
	for _, allocation := range allocationList.Items {
		if prefixRef := allocation.Spec.PrefixRef; prefixRef == nil || prefixRef.Name != prefix.Name {
			continue
		}

		if allocation.Status.Phase != ipamv1alpha1.PrefixAllocationPhaseAllocated {
			continue
		}

		dependentAllocations = append(dependentAllocations, allocation.Name)
	}
	if len(dependentAllocations) > 0 {
		log.V(1).Info("There are still dependent allocations", "DependentAllocations", dependentAllocations)
		return ctrl.Result{}, nil
	}

	log.V(1).Info("No dependent allocations, allowing deletion")
	if err := clientutils.PatchRemoveFinalizer(ctx, r.Client, prefix, prefixFinalizer); err != nil {
		return ctrl.Result{}, fmt.Errorf("error removing finalizer: %w", err)
	}

	log.V(1).Info("Successfully removed finalizer")
	r.forgetAllocationBackoffFor(client.ObjectKeyFromObject(prefix))
	return ctrl.Result{}, nil
}

func (r *PrefixReconciler) patchStatus(ctx context.Context, prefix *ipamv1alpha1.Prefix, phase ipamv1alpha1.PrefixPhase) error {
	now := metav1.Now()
	base := prefix.DeepCopy()

	if prefix.Status.Phase != phase {
		prefix.Status.LastPhaseTransitionTime = &now
	}
	prefix.Status.Phase = phase

	if err := r.Status().Patch(ctx, prefix, client.MergeFrom(base)); err != nil {
		return fmt.Errorf("error patching status: %w", err)
	}
	return nil
}

func (r *PrefixReconciler) allocationMatchesPrefix(allocation *ipamv1alpha1.PrefixAllocation, prefix *ipamv1alpha1.Prefix) bool {
	if prefix.Spec.IPFamily != allocation.Spec.IPFamily {
		return false
	}

	if prefix.Spec.ParentRef != nil && (allocation.Spec.PrefixRef == nil || *allocation.Spec.PrefixRef != *prefix.Spec.ParentRef) {
		return false
	}

	if !equality.Semantic.DeepEqual(prefix.Spec.ParentSelector, allocation.Spec.PrefixSelector) {
		return false
	}

	equalPrefixPointers := func(p1, p2 *commonv1alpha1.IPPrefix) bool {
		return (p1 == nil) == (p2 == nil) && (p1 == nil || *p1 == *p2)
	}

	if allocation.Status.Phase == ipamv1alpha1.PrefixAllocationPhaseAllocated {
		switch {
		case prefix.Spec.Prefix.IsValid():
			return *prefix.Spec.Prefix == *allocation.Status.Prefix
		case prefix.Spec.PrefixLength > 0:
			return prefix.Spec.PrefixLength == int32(allocation.Status.Prefix.Bits())
		default:
			return false
		}
	}

	switch {
	case prefix.Spec.Prefix.IsValid():
		return equalPrefixPointers(prefix.Spec.Prefix, allocation.Spec.Prefix)
	case prefix.Spec.PrefixLength > 0:
		return prefix.Spec.PrefixLength == allocation.Spec.PrefixLength
	default:
		return false
	}
}

var prefixAllocationPhaseValue = map[ipamv1alpha1.PrefixAllocationPhase]int{
	ipamv1alpha1.PrefixAllocationPhaseAllocated: 2,
	ipamv1alpha1.PrefixAllocationPhaseFailed:    -1,
	ipamv1alpha1.PrefixAllocationPhasePending:   1,
}

// prefixAllocationLess determines if allocation is less than other by first comparing the phase of both
// and then, if both phases are the same, prefers the older object.
func (r *PrefixReconciler) prefixAllocationLess(allocation, other *ipamv1alpha1.PrefixAllocation) bool {
	return prefixAllocationPhaseValue[allocation.Status.Phase] < prefixAllocationPhaseValue[other.Status.Phase] ||
		allocation.GetCreationTimestamp().Time.After(other.GetCreationTimestamp().Time)
}

func (r *PrefixReconciler) newAllocationForPrefix(prefix *ipamv1alpha1.Prefix) (*ipamv1alpha1.PrefixAllocation, error) {
	var (
		allocationPrefix       *commonv1alpha1.IPPrefix
		allocationPrefixLength int32
	)
	if prefixPrefix := prefix.Spec.Prefix; prefixPrefix.IsValid() {
		allocationPrefix = prefixPrefix
	} else {
		allocationPrefixLength = prefix.Spec.PrefixLength
	}

	allocation := &ipamv1alpha1.PrefixAllocation{
		ObjectMeta: metav1.ObjectMeta{
			Namespace:    prefix.Namespace,
			GenerateName: prefix.Name + "-",
			Labels: map[string]string{
				prefixAllocationRequesterUIDLabel: string(prefix.UID),
			},
		},
		Spec: ipamv1alpha1.PrefixAllocationSpec{
			IPFamily:       prefix.Spec.IPFamily,
			Prefix:         allocationPrefix,
			PrefixLength:   allocationPrefixLength,
			PrefixRef:      prefix.Spec.ParentRef,
			PrefixSelector: prefix.Spec.ParentSelector,
		},
	}
	if err := controllerutil.SetControllerReference(prefix, allocation, r.Scheme); err != nil {
		return nil, err
	}
	return allocation, nil
}

func (r *PrefixReconciler) findActiveAllocation(prefix *ipamv1alpha1.Prefix, allocations []ipamv1alpha1.PrefixAllocation) (active *ipamv1alpha1.PrefixAllocation, outdated []ipamv1alpha1.PrefixAllocation) {
	for _, allocation := range allocations {
		if !r.allocationMatchesPrefix(&allocation, prefix) {
			outdated = append(outdated, allocation)
			continue
		}

		if active == nil || r.prefixAllocationLess(active, &allocation) {
			if active != nil {
				outdated = append(outdated, *active)
			}

			newActive := allocation
			active = &newActive
		}
	}
	return active, outdated
}

func (r *PrefixReconciler) allocateSubPrefix(ctx context.Context, log logr.Logger, prefix *ipamv1alpha1.Prefix) (*ipamv1alpha1.PrefixAllocation, time.Duration, error) {
	log.V(1).Info("Listing prefix allocations controlled by prefix")
	allocationList := &ipamv1alpha1.PrefixAllocationList{}
	if err := clientutils.ListAndFilterControlledBy(ctx, clientutils.ReaderClient(r.APIReader, r.Client), prefix, allocationList,
		client.InNamespace(prefix.Namespace),
		client.MatchingLabels{
			prefixAllocationRequesterUIDLabel: string(prefix.UID),
		},
	); err != nil {
		return nil, 0, err
	}
	log.V(1).Info("Successfully listed prefix allocations", "NoOfItems", len(allocationList.Items))

	active, outdated := r.findActiveAllocation(prefix, allocationList.Items)
	defer r.pruneOutdatedAllocations(ctx, log, outdated)

	if active == nil {
		log.V(1).Info("Creating new allocation")
		created, err := r.createNewAllocation(ctx, prefix)
		return created, 0, err
	}

	allocationPhase := r.adjustedAllocationPhase(active)
	switch {
	case allocationPhase == ipamv1alpha1.PrefixAllocationPhaseAllocated:
		log.V(1).Info("Allocation is allocated")
		r.forgetAllocationBackoffFor(client.ObjectKeyFromObject(prefix))
		return active, 0, nil
	case allocationPhase == ipamv1alpha1.PrefixAllocationPhaseFailed:
		log.V(1).Info("Allocation is failed")
		retry, err := r.canRetryAllocation(ctx, prefix, active)
		if err != nil || !retry {
			return active, 0, err
		}

		// We protect against over-allocating with a per-prefix backoff.
		backoff := r.allocationBackoffFor(client.ObjectKeyFromObject(prefix))
		if backoff > 0 {
			return active, backoff, nil
		}

		// Delete / free up the old allocation
		log.V(1).Info("Deleting old allocation")
		if err := r.Delete(ctx, active); client.IgnoreNotFound(err) != nil {
			return nil, 0, err
		}

		// Actually initiate the retry
		log.V(1).Info("Retrying allocation")
		created, err := r.createNewAllocation(ctx, prefix)
		return created, 0, err
	default:
		log.V(1).Info("Allocation is not allocated / failed", "Phase", allocationPhase)
		return active, 0, nil
	}
}

// adjustedAllocationPhase calculates an adjusted phase of a PrefixAllocation.
// For the adjusted phase, it is considered
//   - If the allocation is in a terminal phase, that state is returned.
//   - If the allocation is not scheduled, that state is returned.
//   - If the allocation is scheduled but no lastTransitionTime has been recorded, that state is returned
//   - If the allocation is in a non-terminal state, and it has been scheduled, once a configurable timeout has passed,
//     it is considered to be failed.
func (r *PrefixReconciler) adjustedAllocationPhase(allocation *ipamv1alpha1.PrefixAllocation) ipamv1alpha1.PrefixAllocationPhase {
	allocationPhase := allocation.Status.Phase
	if allocationPhase.IsTerminal() || allocation.Spec.PrefixRef == nil {
		return allocationPhase
	}

	lastTransitionTime := allocation.Status.LastPhaseTransitionTime
	if lastTransitionTime.IsZero() {
		return allocationPhase
	}

	if lastTransitionTime.Add(r.PrefixAllocationTimeout).After(time.Now()) {
		return ipamv1alpha1.PrefixAllocationPhaseFailed
	}
	return allocationPhase
}

func (r *PrefixReconciler) createNewAllocation(ctx context.Context, prefix *ipamv1alpha1.Prefix) (*ipamv1alpha1.PrefixAllocation, error) {
	active, err := r.newAllocationForPrefix(prefix)
	if err != nil {
		return nil, err
	}

	if err := r.Create(ctx, active); err != nil {
		return nil, err
	}
	return active, nil
}

func (r *PrefixReconciler) canRetryAllocation(ctx context.Context, prefix *ipamv1alpha1.Prefix, allocation *ipamv1alpha1.PrefixAllocation) (bool, error) {
	// We can always retry if we ended up on a bad prefix with scheduling.
	if prefix.Spec.ParentRef == nil {
		return true, nil
	}

	// If the user request a specific parent prefix to host a prefix, we have to check whether the
	// parent prefix now can host the prefix.
	parentPrefix := &ipamv1alpha1.Prefix{}
	parentKey := client.ObjectKey{Namespace: prefix.Namespace, Name: prefix.Spec.ParentRef.Name}
	if err := r.Get(ctx, parentKey, parentPrefix); err != nil {
		if !apierrors.IsNotFound(err) {
			return false, err
		}
		return false, nil
	}

	return prefixFitsAllocation(parentPrefix, allocation), nil
}

func (r *PrefixReconciler) pruneOutdatedAllocations(ctx context.Context, log logr.Logger, outdated []ipamv1alpha1.PrefixAllocation) {
	for _, outdated := range outdated {
		if _, err := clientutils.DeleteIfExists(ctx, r.Client, &outdated); err != nil {
			log.Error(err, "Error deleting outdated allocation %s: %w", client.ObjectKeyFromObject(&outdated), err)
		}
	}
}

func (r *PrefixReconciler) allocateSelf(ctx context.Context, log logr.Logger, prefix *ipamv1alpha1.Prefix) (ok bool, backoff time.Duration, err error) {
	switch {
	case prefix.Status.Phase == ipamv1alpha1.PrefixPhaseAllocated:
		log.V(1).Info("Prefix is allocated")
		return true, 0, nil
	case prefix.Spec.ParentRef == nil && prefix.Spec.ParentSelector == nil:
		log.V(1).Info("Marking root prefix as allocated")
		if err := r.patchStatus(ctx, prefix, ipamv1alpha1.PrefixPhaseAllocated); err != nil {
			return false, 0, fmt.Errorf("error marking root prefix as allocated: %w", err)
		}
		return false, 0, nil
	default:
		log.V(1).Info("Allocating sub-prefix")
		allocation, backoff, err := r.allocateSubPrefix(ctx, log, prefix)
		if err != nil {
			log.Error(err, "Error allocating sub-prefix")
			if err := r.patchStatus(ctx, prefix, ipamv1alpha1.PrefixPhasePending); err != nil {
				log.Error(err, "Error patching status")
			}
			return false, 0, fmt.Errorf("error applying allocation: %w", err)
		}

		allocationPhase := allocation.Status.Phase
		switch {
		case allocationPhase == ipamv1alpha1.PrefixAllocationPhaseAllocated && !prefix.Spec.Prefix.IsValid():
			if err := r.assignPrefix(ctx, prefix, allocation); err != nil {
				return false, 0, fmt.Errorf("error patching prefix assignment: %w", err)
			}
			return false, 0, nil
		case allocationPhase == ipamv1alpha1.PrefixAllocationPhaseAllocated:
			log.V(1).Info("Marking sub prefix as allocated")
			if err := r.patchStatus(ctx, prefix, ipamv1alpha1.PrefixPhaseAllocated); err != nil {
				return false, 0, fmt.Errorf("error marking prefix as allocated: %w", err)
			}
			return false, 0, nil
		default:
			if err := r.patchStatus(ctx, prefix, ipamv1alpha1.PrefixPhasePending); err != nil {
				return false, backoff, fmt.Errorf("error marking prefix as pending: %w", err)
			}
			return false, backoff, nil
		}
	}
}

func (r *PrefixReconciler) assignPrefix(ctx context.Context, prefix *ipamv1alpha1.Prefix, allocation *ipamv1alpha1.PrefixAllocation) error {
	base := prefix.DeepCopy()
	prefix.Spec.Prefix = allocation.Status.Prefix
	prefix.Spec.ParentRef = allocation.Spec.PrefixRef
	if err := r.Patch(ctx, prefix, client.MergeFrom(base)); err != nil {
		return fmt.Errorf("error assigning prefix: %w", err)
	}
	return nil
}

func (r *PrefixReconciler) processAllocations(ctx context.Context, log logr.Logger, prefix *ipamv1alpha1.Prefix) (used []commonv1alpha1.IPPrefix, err error) {
	list := &ipamv1alpha1.PrefixAllocationList{}
	log.V(1).Info("Listing referencing allocations")
	if err := r.List(ctx, list,
		client.InNamespace(prefix.Namespace),
		client.MatchingFields{ipamclient.PrefixAllocationSpecPrefixRefNameField: prefix.Name},
	); err != nil {
		return nil, fmt.Errorf("error listing allocations: %w", err)
	}

	var (
		availableBuilder netipx.IPSetBuilder
		newAllocations   []ipamv1alpha1.PrefixAllocation
	)
	availableBuilder.AddPrefix(prefix.Spec.Prefix.Prefix)
	for _, allocation := range list.Items {
		allocationPhase := allocation.Status.Phase
		switch {
		case allocationPhase == ipamv1alpha1.PrefixAllocationPhaseAllocated:
			used = append(used, *allocation.Status.Prefix)
			availableBuilder.RemovePrefix(allocation.Status.Prefix.Prefix)
		case allocationPhase != ipamv1alpha1.PrefixAllocationPhaseFailed:
			newAllocations = append(newAllocations, allocation)
		}
	}

	availableSet, err := availableBuilder.IPSet()
	if err != nil {
		return nil, fmt.Errorf("error building available set: %w", err)
	}

	for _, newAllocation := range newAllocations {
		newAvailableSet, res, err := r.processAllocation(ctx, log, prefix, availableSet, &newAllocation)
		if err != nil {
			log.V(1).Error(err, "Error processing allocation", "Allocation", newAllocation)
			continue
		}

		availableSet = newAvailableSet
		if res.IsValid() {
			used = append(used, commonv1alpha1.IPPrefix{Prefix: res})
		}
	}

	// Sort for deterministic status
	sort.Slice(used, func(i, j int) bool { return used[i].String() < used[j].String() })
	return used, nil
}

func (r *PrefixReconciler) patchAllocationStatus(
	ctx context.Context,
	allocation *ipamv1alpha1.PrefixAllocation,
	prefix *commonv1alpha1.IPPrefix,
	phase ipamv1alpha1.PrefixAllocationPhase,
) error {
	now := metav1.Now()
	base := allocation.DeepCopy()

	allocation.Status.Prefix = prefix
	if allocation.Status.Phase != phase {
		allocation.Status.LastPhaseTransitionTime = &now
	}
	allocation.Status.Phase = phase
	if err := r.Status().Patch(ctx, allocation, client.MergeFrom(base)); err != nil {
		return fmt.Errorf("error updating allocation status: %w", err)
	}
	return nil
}

func (r *PrefixReconciler) processAllocation(
	ctx context.Context,
	log logr.Logger,
	prefix *ipamv1alpha1.Prefix,
	available *netipx.IPSet,
	allocation *ipamv1alpha1.PrefixAllocation,
) (*netipx.IPSet, netip.Prefix, error) {
	log = log.WithValues("AllocationKey", client.ObjectKeyFromObject(allocation))
	if !allocation.DeletionTimestamp.IsZero() {
		return available, netip.Prefix{}, nil
	}

	res, newAvailableSet, ok, terminal := r.acquireAllocation(prefix, available, allocation)
	switch {
	case !ok && terminal:
		log.V(1).Info("Marking terminally non-allocatable allocation as failed")
		if err := r.patchAllocationStatus(ctx, allocation, allocation.Status.Prefix, ipamv1alpha1.PrefixAllocationPhaseFailed); client.IgnoreNotFound(err) != nil {
			return available, netip.Prefix{}, fmt.Errorf("could not mark allocation as failed: %w", err)
		}
		return available, netip.Prefix{}, nil
	case !ok:
		log.V(1).Info("Marking non-allocatable allocation as pending")
		if err := r.patchAllocationStatus(ctx, allocation, allocation.Status.Prefix, ipamv1alpha1.PrefixAllocationPhasePending); client.IgnoreNotFound(err) != nil {
			return available, netip.Prefix{}, fmt.Errorf("could not mark allocation as pending: %w", err)
		}
		return available, netip.Prefix{}, nil
	default:
		log.V(1).Info("Marking allocation as allocated")
		if err := r.patchAllocationStatus(ctx, allocation, commonv1alpha1.NewIPPrefix(res), ipamv1alpha1.PrefixAllocationPhaseAllocated); err != nil {
			return available, netip.Prefix{}, fmt.Errorf("error marking allocation as succeeded: %w", err)
		}
		return newAvailableSet, res, nil
	}
}

func (r *PrefixReconciler) reconcile(ctx context.Context, log logr.Logger, prefix *ipamv1alpha1.Prefix) (ctrl.Result, error) {
	if !controllerutil.ContainsFinalizer(prefix, prefixFinalizer) {
		base := prefix.DeepCopy()
		controllerutil.AddFinalizer(prefix, prefixFinalizer)
		if err := r.Patch(ctx, prefix, client.MergeFrom(base)); err != nil {
			return ctrl.Result{}, fmt.Errorf("error adding finalizer: %w", err)
		}
		return ctrl.Result{}, nil
	}

	ok, backoff, err := r.allocateSelf(ctx, log, prefix)
	if err != nil {
		return ctrl.Result{}, fmt.Errorf("error allocating prefix: %w", err)
	}
	if !ok {
		return ctrl.Result{RequeueAfter: backoff}, nil
	}

	used, err := r.processAllocations(ctx, log, prefix)
	if err != nil {
		return ctrl.Result{}, fmt.Errorf("error computing usages: %w", err)
	}

	log.V(1).Info("Updating status", "Used", used)
	base := prefix.DeepCopy()
	prefix.Status.Used = used
	if err := r.Status().Patch(ctx, prefix, client.MergeFrom(base)); err != nil {
		return ctrl.Result{}, fmt.Errorf("error patching status: %w", err)
	}
	return ctrl.Result{}, nil
}

// SetupWithManager sets up the controller with the Manager.
func (r *PrefixReconciler) SetupWithManager(mgr ctrl.Manager) error {
	r.allocationLimiter = workqueue.NewItemExponentialFailureRateLimiter(5*time.Millisecond, 1000*time.Second)

	return ctrl.NewControllerManagedBy(mgr).
		For(&ipamv1alpha1.Prefix{}).
		Owns(&ipamv1alpha1.PrefixAllocation{}).
		Watches(
			&ipamv1alpha1.PrefixAllocation{},
			r.enqueueByAllocationPrefixRef(),
		).
		Watches(
			&ipamv1alpha1.Prefix{},
			r.enqueueByPrefixParentRef(),
		).
		Watches(
			&ipamv1alpha1.Prefix{},
			r.enqueueByPrefixParentSelector(),
		).
		Complete(r)
}

func (r *PrefixReconciler) enqueueByAllocationPrefixRef() handler.EventHandler {
	return handler.EnqueueRequestsFromMapFunc(func(ctx context.Context, obj client.Object) []ctrl.Request {
		allocation := obj.(*ipamv1alpha1.PrefixAllocation)
		prefixRef := allocation.Spec.PrefixRef
		if prefixRef == nil {
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

func (r *PrefixReconciler) enqueueByPrefixParentSelector() handler.EventHandler {
	return handler.EnqueueRequestsFromMapFunc(func(ctx context.Context, obj client.Object) []ctrl.Request {
		prefix := obj.(*ipamv1alpha1.Prefix)
		log := ctrl.LoggerFrom(ctx)
		if !isPrefixAllocatedAndNotDeleting(prefix) {
			return nil
		}

		list := &ipamv1alpha1.PrefixList{}
		if err := r.List(ctx, list,
			client.InNamespace(prefix.Namespace),
			client.MatchingFields{ipamclient.PrefixSpecIPFamilyField: string(prefix.Spec.IPFamily)},
		); err != nil {
			log.V(1).Error(err, "Error listing prefixes", "Namespace", prefix.Namespace, "IPFamily", prefix.Spec.IPFamily)
			return nil
		}

		var requests []ctrl.Request
		for _, other := range list.Items {
			if other.Spec.ParentRef != nil {
				continue
			}

			sel, err := metav1.LabelSelectorAsSelector(other.Spec.ParentSelector)
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

func (r *PrefixReconciler) enqueueByPrefixParentRef() handler.EventHandler {
	return handler.EnqueueRequestsFromMapFunc(func(ctx context.Context, obj client.Object) []ctrl.Request {
		prefix := obj.(*ipamv1alpha1.Prefix)
		log := ctrl.LoggerFrom(ctx)
		if !isPrefixAllocatedAndNotDeleting(prefix) {
			return nil
		}

		list := &ipamv1alpha1.PrefixList{}
		if err := r.List(ctx, list,
			client.InNamespace(prefix.Namespace),
			client.MatchingFields{ipamclient.PrefixSpecParentRefNameField: prefix.Name},
		); err != nil {
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
