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
	"sort"

	"github.com/adracus/reflcompare"
	"github.com/go-logr/logr"
	"inet.af/netaddr"
	corev1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apimachinery/pkg/types"
	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/builder"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/controller/controllerutil"
	"sigs.k8s.io/controller-runtime/pkg/handler"
	"sigs.k8s.io/controller-runtime/pkg/manager"
	"sigs.k8s.io/controller-runtime/pkg/predicate"
	"sigs.k8s.io/controller-runtime/pkg/reconcile"
	"sigs.k8s.io/controller-runtime/pkg/source"

	commonv1alpha1 "github.com/onmetal/onmetal-api/apis/common/v1alpha1"
	"github.com/onmetal/onmetal-api/apis/network"
	networkv1alpha1 "github.com/onmetal/onmetal-api/apis/network/v1alpha1"
	"github.com/onmetal/onmetal-api/equality"
	"github.com/onmetal/onmetal-api/predicates"
)

const (
	ipamRangeFinalizerName = network.LabelDomain + "/ipamrange"
)

var (
	allocationComparisons = reflcompare.NewComparisonsOrDie()
)

// IPAMRangeReconciler reconciles a IPAMRange object
type IPAMRangeReconciler struct {
	client.Client
	Scheme *runtime.Scheme
}

//+kubebuilder:rbac:groups=network.onmetal.de,resources=ipamranges,verbs=get;list;watch;create;update;patch;delete
//+kubebuilder:rbac:groups=network.onmetal.de,resources=ipamranges/status,verbs=get;update;patch
//+kubebuilder:rbac:groups=network.onmetal.de,resources=ipamranges/finalizers,verbs=update

// Reconcile
// if parent -> handle request part and set range
func (r *IPAMRangeReconciler) Reconcile(ctx context.Context, req ctrl.Request) (ctrl.Result, error) {
	log := ctrl.LoggerFrom(ctx)
	ipamRange := &networkv1alpha1.IPAMRange{}
	if err := r.Get(ctx, req.NamespacedName, ipamRange); err != nil {
		return ctrl.Result{}, client.IgnoreNotFound(err)
	}

	return r.reconcileExists(ctx, log, ipamRange)
}

const parentField = ".spec.parent.name"

func (r *IPAMRangeReconciler) reconcileRequestsFromParent(ctx context.Context, log logr.Logger, object client.Object) []reconcile.Request {
	ipamRange := object.(*networkv1alpha1.IPAMRange)
	list := &networkv1alpha1.IPAMRangeList{}
	if err := r.List(ctx, list, client.InNamespace(ipamRange.Namespace), client.MatchingFields{parentField: ipamRange.Name}); err != nil {
		return nil
	}

	requests := make([]reconcile.Request, 0, len(list.Items))
	for _, child := range list.Items {
		requests = append(requests, reconcile.Request{NamespacedName: client.ObjectKeyFromObject(&child)})
	}
	return requests
}

func (r *IPAMRangeReconciler) reconcileRequestsFromChildren(ctx context.Context, log logr.Logger, object client.Object) []reconcile.Request {
	ipamRange := object.(*networkv1alpha1.IPAMRange)
	if ipamRange.Spec.Parent == nil {
		return nil
	}

	return []reconcile.Request{
		{
			NamespacedName: types.NamespacedName{
				Namespace: ipamRange.Namespace,
				Name:      ipamRange.Spec.Parent.Name,
			},
		},
	}
}

// SetupWithManager sets up the controller with the Manager.
func (r *IPAMRangeReconciler) SetupWithManager(mgr manager.Manager) error {
	ctx := context.Background()
	log := ctrl.Log.WithName("ipamrange").WithName("setup-with-manager")

	if err := mgr.GetFieldIndexer().IndexField(ctx, &networkv1alpha1.IPAMRange{}, parentField, func(object client.Object) []string {
		ipamRange := object.(*networkv1alpha1.IPAMRange)
		if ipamRange.Spec.Parent == nil {
			return nil
		}
		return []string{ipamRange.Spec.Parent.Name}
	}); err != nil {
		return fmt.Errorf("could not index field %s: %w", parentField, err)
	}

	return ctrl.NewControllerManagedBy(mgr).
		For(&networkv1alpha1.IPAMRange{}).
		Watches(
			// Parents trigger their children.
			&source.Kind{Type: &networkv1alpha1.IPAMRange{}},
			handler.EnqueueRequestsFromMapFunc(func(object client.Object) []reconcile.Request {
				return r.reconcileRequestsFromParent(ctx, log, object)
			}),
			builder.WithPredicates(predicates.IPAMRangeAllocationsChangedPredicate{}),
		).
		Watches(
			// Children trigger their parents.
			&source.Kind{Type: &networkv1alpha1.IPAMRange{}},
			handler.EnqueueRequestsFromMapFunc(func(object client.Object) []reconcile.Request {
				return r.reconcileRequestsFromChildren(ctx, log, object)
			}),
			builder.WithPredicates(predicate.GenerationChangedPredicate{}),
		).
		Complete(r)
}

func (r *IPAMRangeReconciler) reconcileExists(ctx context.Context, log logr.Logger, ipamRange *networkv1alpha1.IPAMRange) (ctrl.Result, error) {
	if !ipamRange.DeletionTimestamp.IsZero() {
		if !controllerutil.ContainsFinalizer(ipamRange, ipamRangeFinalizerName) {
			return ctrl.Result{}, nil
		}
		return r.delete(ctx, log, ipamRange)
	}

	if !controllerutil.ContainsFinalizer(ipamRange, ipamRangeFinalizerName) {
		controllerutil.AddFinalizer(ipamRange, ipamRangeFinalizerName)
		if err := r.Update(ctx, ipamRange); err != nil {
			return ctrl.Result{}, fmt.Errorf("error setting finalizer: %w", err)
		}

		return ctrl.Result{}, nil
	}
	return r.reconcile(ctx, log, ipamRange)
}

func (r *IPAMRangeReconciler) delete(ctx context.Context, log logr.Logger, ipamRange *networkv1alpha1.IPAMRange) (ctrl.Result, error) {
	childList := &networkv1alpha1.IPAMRangeList{}
	if err := r.List(ctx, childList, client.InNamespace(ipamRange.Namespace), client.MatchingFields{parentField: ipamRange.Name}); err != nil {
		return ctrl.Result{}, fmt.Errorf("could not list children: %w", err)
	}

	if len(childList.Items) > 0 {
		log.Info("Children are still present, delaying deletion.")
		return ctrl.Result{}, nil
	}

	log.Info("No children present anymore, removing finalizer.")
	updated := ipamRange.DeepCopy()
	controllerutil.RemoveFinalizer(updated, ipamRangeFinalizerName)
	if err := r.Patch(ctx, updated, client.MergeFrom(ipamRange)); err != nil {
		return ctrl.Result{}, fmt.Errorf("could not remove finalizer: %w", err)
	}

	log.V(1).Info("Successfully removed finalizer")
	return ctrl.Result{}, nil
}

func ipSetFromCIDRs(cidrs []commonv1alpha1.IPPrefix) (*netaddr.IPSet, error) {
	var bldr netaddr.IPSetBuilder
	for _, cidr := range cidrs {
		bldr.AddPrefix(cidr.IPPrefix)
	}
	return bldr.IPSet()
}

func (r *IPAMRangeReconciler) mapChildNameToChild(items []networkv1alpha1.IPAMRange) map[string]networkv1alpha1.IPAMRange {
	nameToChild := make(map[string]networkv1alpha1.IPAMRange)
	for _, child := range items {
		nameToChild[child.Name] = child
	}

	return nameToChild
}

func (r *IPAMRangeReconciler) fulfilledRequests(nameToChild map[string]networkv1alpha1.IPAMRange, ipamRange *networkv1alpha1.IPAMRange) map[string]map[networkv1alpha1.IPAMRangeRequest]allocation {
	res := make(map[string]map[networkv1alpha1.IPAMRangeRequest]allocation)
	for _, allocStatus := range ipamRange.Status.Allocations {
		if allocStatus.State != networkv1alpha1.IPAMRangeAllocationUsed {
			continue
		}

		user, request := allocStatus.User, allocStatus.Request
		if user == nil || request == nil {
			continue
		}

		child, ok := nameToChild[user.Name]
		if !ok {
			continue
		}

		for _, childRequest := range child.Spec.Requests {
			if equality.Semantic.DeepEqual(childRequest, *request) {
				requests := res[child.Name]
				if requests == nil {
					requests = make(map[networkv1alpha1.IPAMRangeRequest]allocation)
				}

				requests[*request] = allocation{
					cidr: allocStatus.CIDR,
					ips:  allocStatus.IPs,
				}

				res[child.Name] = requests
				break
			}
		}
	}
	return res
}

type childNameAndRequest struct {
	childName string
	request   networkv1alpha1.IPAMRangeRequest
}

func (r *IPAMRangeReconciler) sortedRequests(items []networkv1alpha1.IPAMRange, fulfilledRequests map[string]map[networkv1alpha1.IPAMRangeRequest]allocation) []childNameAndRequest {
	var requests []childNameAndRequest
	for _, item := range items {
		for _, request := range item.Spec.Requests {
			requests = append(requests, childNameAndRequest{
				childName: item.Name,
				request:   request,
			})
		}
	}
	// sort requests so requests that already have been allocated.
	// this means, requests *with* already allocated CIDR / IP range appear first in the
	// sorted slice.
	sort.Slice(requests, func(i, j int) bool {
		req1, req2 := requests[i], requests[j]
		_, ok1 := fulfilledRequests[req1.childName][req1.request]
		_, ok2 := fulfilledRequests[req2.childName][req2.request]
		return ok1 || !ok2
	})
	return requests
}

type allocation struct {
	ips  *commonv1alpha1.IPRange
	cidr *commonv1alpha1.IPPrefix
}

func (r *IPAMRangeReconciler) gatherAvailable(ctx context.Context, ipamRange *networkv1alpha1.IPAMRange) (available *netaddr.IPSet, parentAllocations []allocation, failed []networkv1alpha1.IPAMRangeAllocationStatus, err error) {
	if ipamRange.Spec.Parent == nil {
		available, err := ipSetFromCIDRs(ipamRange.Spec.CIDRs)
		if err != nil {
			return nil, nil, nil, err
		}

		return available, nil, nil, nil
	}

	parent := &networkv1alpha1.IPAMRange{}
	if err := r.Get(ctx, client.ObjectKey{Namespace: ipamRange.Namespace, Name: ipamRange.Spec.Parent.Name}, parent); err != nil {
		return nil, nil, nil, fmt.Errorf("could not retrieve parent %s: %w", ipamRange.Spec.Parent.Name, err)
	}

	var (
		availableBldr netaddr.IPSetBuilder
		other         []networkv1alpha1.IPAMRangeAllocationStatus
	)
	for _, allocStatus := range parent.Status.Allocations {
		allocStatus := allocStatus
		if activeRequest, user := allocStatus.Request, allocStatus.User; allocStatus.Request != nil && user != nil && user.Name == ipamRange.Name {
			for _, request := range ipamRange.Spec.Requests {
				if equality.Semantic.DeepEqual(*activeRequest, request) {
					if allocStatus.State == networkv1alpha1.IPAMRangeAllocationUsed {
						switch {
						case allocStatus.CIDR != nil:
							availableBldr.AddPrefix(allocStatus.CIDR.IPPrefix)
						case allocStatus.IPs != nil:
							availableBldr.AddRange(allocStatus.IPs.Range())
						}
						parentAllocations = append(parentAllocations, allocation{allocStatus.IPs, allocStatus.CIDR})
					} else {
						other = append(other, networkv1alpha1.IPAMRangeAllocationStatus{
							CIDR:    allocStatus.CIDR,
							IPs:     allocStatus.IPs,
							State:   allocStatus.State,
							Request: allocStatus.Request,
						})
					}
				}
			}
		}
	}
	available, err = availableBldr.IPSet()
	if err != nil {
		return nil, nil, nil, err
	}

	return available, parentAllocations, other, nil
}

func isNetIP(ip netaddr.IP) bool {
	switch {
	case ip.Is4():
		return ip.As4()[3] == 0
	case ip.Is6():
		return ip.As16()[15] == 0
	default:
		return false
	}
}

func (r *IPAMRangeReconciler) acquireRequest(set *netaddr.IPSet, request networkv1alpha1.IPAMRangeRequest) (prefix *netaddr.IPPrefix, ipRange *netaddr.IPRange, newSet *netaddr.IPSet, ok bool) {
	switch {
	case request.CIDR != nil:
		if !set.ContainsPrefix(request.CIDR.IPPrefix) {
			return nil, nil, nil, false
		}
		var bldr netaddr.IPSetBuilder
		bldr.AddSet(set)
		bldr.RemovePrefix(request.CIDR.IPPrefix)
		newSet, err := bldr.IPSet()
		newSet.Ranges()
		if err != nil {
			return nil, nil, nil, false
		}
		return &request.CIDR.IPPrefix, nil, newSet, true
	case request.IPs != nil:
		ipRange := request.IPs.Range()
		if !set.ContainsRange(ipRange) {
			return nil, nil, nil, false
		}

		var bldr netaddr.IPSetBuilder
		bldr.AddSet(set)
		bldr.RemoveRange(ipRange)
		newSet, err := bldr.IPSet()
		if err != nil {
			return nil, nil, nil, false
		}
		return nil, &ipRange, newSet, true
	case request.Size > 0:
		prefix, newSet, ok := set.RemoveFreePrefix(uint8(request.Size))
		if !ok {
			return nil, nil, nil, false
		}
		return &prefix, nil, newSet, true
	// TODO: It should be possible to request arbitrarily large ip ranges.
	// Additionally, the allocation has to be enhanced to account for ip ranges of 'size' 1.
	case request.IPCount == 1:
		ranges := set.Ranges()
		for _, rng := range ranges {
			ip := rng.From()
			if isNetIP(ip) {
				ip = ip.Next()
			}
			if ip.IsZero() || !rng.Contains(ip) {
				continue
			}

			var bldr netaddr.IPSetBuilder
			bldr.AddSet(set)
			bldr.Remove(ip)
			newSet, err := bldr.IPSet()
			if err != nil {
				return nil, nil, nil, false
			}
			ipRange := netaddr.IPRangeFrom(ip, ip)
			return nil, &ipRange, newSet, true
		}
		return nil, nil, nil, false
	default:
		return nil, nil, nil, false
	}
}

func (r *IPAMRangeReconciler) computeChildAllocations(
	available *netaddr.IPSet,
	fulfilledRequests map[string]map[networkv1alpha1.IPAMRangeRequest]allocation,
	requests []childNameAndRequest,
) (newAvailable *netaddr.IPSet, childAllocations []networkv1alpha1.IPAMRangeAllocationStatus) {
	for _, requestAndName := range requests {
		originalRequest, name := requestAndName.request, requestAndName.childName

		// it's possible that no request from IPAMRange is fullfilled
		oldRequests := fulfilledRequests[name]
		// we copy the original request since we're modifying it
		// below to force re-acquiring already allocated IPs.
		request := originalRequest
		if allocated, ok := oldRequests[request]; ok {
			request.CIDR = allocated.cidr
			request.IPs = allocated.ips
		}

		prefix, ipRange, newSet, ok := r.acquireRequest(available, request)
		if !ok {
			childAllocations = append(childAllocations, networkv1alpha1.IPAMRangeAllocationStatus{
				State:   networkv1alpha1.IPAMRangeAllocationFailed,
				Request: &originalRequest,
				User:    &corev1.LocalObjectReference{Name: name},
			})
		} else {
			available = newSet
			var cidr *commonv1alpha1.IPPrefix
			if prefix != nil {
				cidr = commonv1alpha1.NewIPPrefixPtr(*prefix)
			}
			var ips *commonv1alpha1.IPRange
			if ipRange != nil {
				ips = commonv1alpha1.NewIPRangePtr(*ipRange)
			}

			childAllocations = append(childAllocations, networkv1alpha1.IPAMRangeAllocationStatus{
				State:   networkv1alpha1.IPAMRangeAllocationUsed,
				CIDR:    cidr,
				IPs:     ips,
				Request: &originalRequest,
				User:    &corev1.LocalObjectReference{Name: name},
			})
		}
	}
	return available, childAllocations
}

func (r *IPAMRangeReconciler) intersectUsed(available *netaddr.IPSet, used allocation) (newAvailable, intersection *netaddr.IPSet) {
	var (
		intersectionBuilder netaddr.IPSetBuilder
		newAvailableBuilder netaddr.IPSetBuilder
	)
	newAvailableBuilder.AddSet(available)
	intersectionBuilder.AddSet(available)
	var ipsSetBuilder netaddr.IPSetBuilder
	switch {
	case used.ips != nil:
		ipsSetBuilder.AddRange(used.ips.Range())
	case used.cidr != nil:
		ipsSetBuilder.AddPrefix(used.cidr.IPPrefix)
	}
	ipsSet, _ := ipsSetBuilder.IPSet()
	intersectionBuilder.Intersect(ipsSet)
	intersection, _ = intersectionBuilder.IPSet()
	newAvailableBuilder.RemoveSet(intersection)
	newAvailable, _ = newAvailableBuilder.IPSet()
	return newAvailable, intersection
}

func (r *IPAMRangeReconciler) computeFreeAllocations(available *netaddr.IPSet, parentAllocations []allocation) []networkv1alpha1.IPAMRangeAllocationStatus {
	var res []networkv1alpha1.IPAMRangeAllocationStatus
	for _, info := range parentAllocations {
		var intersection *netaddr.IPSet
		available, intersection = r.intersectUsed(available, info)
		switch {
		case info.ips != nil:
			for _, ipRange := range intersection.Ranges() {
				res = append(res, networkv1alpha1.IPAMRangeAllocationStatus{
					IPs:   commonv1alpha1.NewIPRangePtr(ipRange),
					State: networkv1alpha1.IPAMRangeAllocationFree,
				})
			}
		case info.cidr != nil:
			for _, cidr := range intersection.Prefixes() {
				res = append(res, networkv1alpha1.IPAMRangeAllocationStatus{
					CIDR:  commonv1alpha1.NewIPPrefixPtr(cidr),
					State: networkv1alpha1.IPAMRangeAllocationFree,
				})
			}
		}
	}

	for _, cidr := range available.Prefixes() {
		res = append(res, networkv1alpha1.IPAMRangeAllocationStatus{
			CIDR:  commonv1alpha1.NewIPPrefixPtr(cidr),
			State: networkv1alpha1.IPAMRangeAllocationFree,
		})
	}
	return res
}

func (r *IPAMRangeReconciler) reconcile(ctx context.Context, log logr.Logger, ipamRange *networkv1alpha1.IPAMRange) (ctrl.Result, error) {
	available, parentAllocations, otherAllocations, err := r.gatherAvailable(ctx, ipamRange)
	if err != nil {
		return ctrl.Result{}, fmt.Errorf("could not gather available: %w", err)
	}

	list := &networkv1alpha1.IPAMRangeList{}
	if err := r.List(ctx, list, client.InNamespace(ipamRange.Namespace), client.MatchingFields{parentField: ipamRange.Name}); err != nil {
		return ctrl.Result{}, fmt.Errorf("could not list children: %w", err)
	}

	nameToChild := r.mapChildNameToChild(list.Items)
	fulfilledRequests := r.fulfilledRequests(nameToChild, ipamRange)
	requests := r.sortedRequests(list.Items, fulfilledRequests)
	newAvailable, childAllocations := r.computeChildAllocations(available, fulfilledRequests, requests)
	freeAllocations := r.computeFreeAllocations(newAvailable, parentAllocations)

	var newAllocations []networkv1alpha1.IPAMRangeAllocationStatus
	newAllocations = append(newAllocations, childAllocations...)
	newAllocations = append(newAllocations, freeAllocations...)
	newAllocations = append(newAllocations, otherAllocations...)

	// Sort the allocations to ensure idempotency when watching for allocation changes.
	sort.Slice(newAllocations, func(i, j int) bool {
		ai, aj := newAllocations[i], newAllocations[j]
		return allocationComparisons.DeepCompare(ai, aj) < 0
	})

	updated := ipamRange.DeepCopy()
	updated.Status.ObservedGeneration = updated.Generation
	updated.Status.Allocations = newAllocations
	if err := r.Status().Patch(ctx, updated, client.MergeFrom(ipamRange)); err != nil {
		return reconcile.Result{}, fmt.Errorf("could not update ipam range status: %w", err)
	}

	return ctrl.Result{}, nil
}
