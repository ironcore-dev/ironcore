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

package ipamrange

import (
	"context"
	"fmt"
	"net"
	"reflect"
	"sigs.k8s.io/controller-runtime/pkg/controller/controllerutil"

	"github.com/go-logr/logr"
	"github.com/mandelsoft/kubipam/pkg/ipam"
	"github.com/onmetal/onmetal-api/apis/common/v1alpha1"
	networkv1alpha1 "github.com/onmetal/onmetal-api/apis/network/v1alpha1"
	"github.com/onmetal/onmetal-api/pkg/utils"
	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/client"
)

func (r *Reconciler) reconcileRootRequest(ctx context.Context, log logr.Logger, current *IPAM) (ctrl.Result, error) {
	log.Info("reconcile root request")
	if len(current.requestSpecs) == 0 {
		return r.invalid(ctx, log, current, "at least one cidr must be specified for root (no parent) ipam range")
	}
	var cidrs AllocationStatusList
	for i, c := range current.requestSpecs {
		if !c.IsValid() {
			cidrs = append(cidrs, &AllocationStatus{
				Allocation: Allocation{
					Request: c.Request,
					CIDR:    nil,
				},
				Status:  networkv1alpha1.AllocationStateFailed,
				Message: c.Error,
			})
			continue
		}
		if !c.Spec.IsCIDR() {
			return r.invalid(ctx, log, current, "request spec %d is not a valid cidr %s for a root ipam range", i, c)
		}
		_, cidr, _ := net.ParseCIDR(c.Request)
		if ipam.CIDRHostMaskSize(cidr) == 0 {
			// TODO: rethink IP delegation
			return r.invalid(ctx, log, current, "root cidr must have more than one ip address")
		}
		cidrs = append(cidrs, &AllocationStatus{
			Allocation: Allocation{
				Request: c.Request,
				CIDR:    cidr,
			},
			Status:  networkv1alpha1.AllocationStateAllocated,
			Message: SuccessfulUsageMessage,
		})
	}

	if err := updateStatus(ctx, log, r, current, func(obj *networkv1alpha1.IPAMRange) {
		obj.Status.CIDRs = cidrs.AsCIDRAllocationStatusList()
	}, func() {
		current.updateAllocations(cidrs, current.deletions)
		// trigger all users of this ipamrange
		log.Info("trigger all users of range", "key", current.objectId.ObjectKey)
	}); err != nil {
		return utils.Requeue(err)
	}
	return utils.Succeeded()
}

func (r *Reconciler) reconcileChildRequest(ctx context.Context, log logr.Logger, current *IPAM) (ctrl.Result, error) {
	log.Info("reconcile request")
	parent := &networkv1alpha1.IPAMRange{}
	parentKey := client.ObjectKey{Namespace: current.object.Namespace, Name: current.object.Spec.Parent.Name}
	log.V(1).Info("Getting parent", "parent", parentKey)
	if err := r.Get(ctx, parentKey, parent); err != nil {
		return ctrl.Result{}, fmt.Errorf("error getting parent %s: %w", parentKey, err)
	}

	log.V(1).Info("Found parent", "parent", parentKey)
	requestId := utils.NewObjectId(current.object)

	ipr, err := r.cache.getRange(ctx, log, client.ObjectKeyFromObject(parent), nil)
	if ipr == nil {
		if err != nil {
			return utils.Requeue(err)
		}
		return r.setStatus(ctx, log, current, v1alpha1.StateError, "IPAMRange %s not found", parentKey)
	}
	defer r.cache.release(log, parentKey)
	if ipr.error != "" {
		return r.setStatus(ctx, log, current, v1alpha1.StateError, "IPAMRange %s not valid: %s", parentKey, ipr.error)
	}

	if !ipr.object.DeletionTimestamp.IsZero() {
		return r.setStatus(ctx, log, current, v1alpha1.StateError, "IPAMRange %s is already deleting", parentKey)
	}

	if ipr.ipam == nil {
		if len(ipr.allocations) == 0 {
			log.Info("parent is not yet ready", "parent", ipr.objectId)
			return utils.Succeeded()
		}
	}

	for i, c := range current.requestSpecs {
		if c.Spec != nil && c.Spec.Bits() > ipr.ipam.Bits() {
			return r.setStatus(ctx, log, current, v1alpha1.StateInvalid, "size (entry %d) %d bits too large", i, ipr.ipam.Bits())
		}
	}

	var allocated AllocationStatusList
	var allocations AllocationStatusList
	var deleted AllocationStatusList
	var deletions AllocationStatusList
	requeue := false
	if ipr.pendingRequest != nil {
		if ipr.pendingRequest.key != requestId.ObjectKey {
			return utils.Succeeded()
		}
		log.Info("found pending request", "request", ipr.pendingRequest.key)
		allocations = NewAllocationStatusListFromAllocations(ipr.pendingRequest.CIDRs)
		deletions = NewAllocationStatusListFromAllocations(ipr.pendingRequest.Deletions)
	} else {
		reqSpecs, delAllocs, oldAllocs := current.requestSpecs.PendingActions(current.allocations)
		if err := current.HandleRelease(ctx, log, r.Client, current, delAllocs); err != nil {
			return utils.Requeue(err)
		}
		deletions, deleted, err = ipr.Free(ctx, log, r.Client, current)
		if err != nil {
			return utils.Requeue(err)
		}
		if len(deleted) == 0 {
			allocated, err = ipr.Alloc(ctx, log, r.Client, current, reqSpecs)
			if err != nil {
				return utils.Requeue(err)
			}
			allocations = append(oldAllocs, allocated...)
		} else {
			// check for additional allocation after successful deletion
			log.Info("requeue for check for new allocations requests after a deletion")
			requeue = true
		}
	}
	if len(allocated) == 0 && ipr.pendingRequest == nil && len(deleted) == 0 {
		log.Info("nothing new to allocate or free")
		return utils.Succeeded()
	}
	log.Info("allocation changes", "allocated", allocated, "deleted", deleted)
	newObj := current.determineState(allocations, deletions)
	if !reflect.DeepEqual(newObj, current.object) {
		if err := r.Client.Status().Patch(ctx, newObj, client.MergeFrom(current.object)); err != nil {
			return utils.Requeue(err)
		}
		log.Info("allocation changes successfully updated. trigger range", "objectkey", ipr.objectId.ObjectKey)
	}
	// make sure to update cache with new object -> trigger range object -> triggers its users
	current.object = newObj
	current.updateAllocations(allocations, deletions)
	if requeue {
		return utils.Requeue(nil)
	}
	return utils.Succeeded()
}

func (r *Reconciler) deleteRequest(ctx context.Context, log logr.Logger, current *IPAM) (ctrl.Result, error) {
	requestId := utils.NewObjectId(current.object)
	if controllerutil.ContainsFinalizer(current.object, finalizerName) {
		log.Info("object still in use by others, delaying deletion", "key", requestId.ObjectKey)
		return ctrl.Result{}, nil
	}

	usersList := &networkv1alpha1.IPAMRangeList{}
	if err := r.List(ctx, usersList,
		client.MatchingFields{parentField: current.object.Name},
		client.InNamespace(current.object.Namespace),
	); err != nil {
		return ctrl.Result{}, fmt.Errorf("could not list using ipam ranges: %w", err)
	}

	if len(usersList.Items) > 0 {
		log.V(1).Info("Range is still in use")
		return ctrl.Result{}, nil
	}

	if len(current.object.Status.CIDRs) > 0 {
		parent := &networkv1alpha1.IPAMRange{}
		parentKey := client.ObjectKey{Namespace: current.object.Namespace, Name: current.object.Spec.Parent.Name}
		log.V(1).Info("Getting parent", "parent", parentKey)
		if err := r.Get(ctx, parentKey, parent); err != nil {
			return ctrl.Result{}, fmt.Errorf("error getting parent %s: %w", parentKey, err)
		}

		ipr, err := r.cache.getRange(ctx, log, parentKey, nil)
		if err != nil {
			return ctrl.Result{}, fmt.Errorf("error getting parent %s ipam: %w", parentKey, err)
		}

		if ipr != nil {
			defer r.cache.release(log, parentKey)
			var allocated AllocationStatusList
			if ipr.pendingRequest != nil {
				if ipr.pendingRequest.key != requestId.ObjectKey {
					log.Info("operation on ipamrange still pending, delaying deletion")
					return utils.Succeeded()
				}
				allocated = NewAllocationStatusListFromAllocations(ipr.pendingRequest.CIDRs)
				log.Info("continuing release", "allocated", allocated)
			} else {
				err = ipr.FreeAll(ctx, log, r.Client, current)
				if err != nil {
					return utils.Requeue(err)
				}
			}
			newObj := current.object.DeepCopy()
			newObj.Status.CIDRs = nil
			newObj.Status.PendingDeletions = nil
			if !reflect.DeepEqual(newObj, current.object) {
				log.Info("updating status for", "object", current.objectId)
				if err := r.Client.Status().Patch(ctx, newObj, client.MergeFrom(current.object)); err != nil {
					return utils.Requeue(err)
				}
			}
			current.object = newObj
			current.allocations = nil
			current.deletions = nil
		}
	}

	log.Info("Releasing range")
	controllerutil.RemoveFinalizer(current.object, finalizerName)
	if err := r.Update(ctx, current.object); err != nil {
		return ctrl.Result{}, fmt.Errorf("could not remove finalizer: %w", err)
	}
	return ctrl.Result{}, nil
}
