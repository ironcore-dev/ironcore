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

	"github.com/go-logr/logr"
	"github.com/onmetal/onmetal-api/apis/common/v1alpha1"
	api "github.com/onmetal/onmetal-api/apis/network/v1alpha1"
	"github.com/onmetal/onmetal-api/pkg/cache/usagecache"
	"github.com/onmetal/onmetal-api/pkg/utils"
	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/client"
)

func (r *Reconciler) reconcileRequest(ctx context.Context, log logr.Logger, current *IPAM) (ctrl.Result, error) {
	log.Info("reconcile request", "namespace", current.object.Namespace, "name", current.object.Name)
	rangeId, failed, err := r.ScopeEvaluator.EvaluateScopedReferenceToObjectId(ctx, current.object.Namespace, api.IPAMRangeGK, current.object.Spec.Parent)
	if err != nil {
		if !failed {
			return utils.Requeue(err)
		}
		return r.setStatus(ctx, log, current.object, v1alpha1.StateError, err.Error())
	}

	if rangeId.Name == "" {
		return utils.Succeeded()
	}

	log.Info("found parent", "key", rangeId.ObjectKey)

	requestId := utils.NewObjectId(current.object)
	// Update usage cache
	r.GetUsageCache().ReplaceObjectUsageInfo(requestId, usagecache.NewObjectUsageInfo("uses", rangeId))

	// check for cycles and self reference
	if cycle := r.GetUsageCache().IsCyclicForRelationForGK(requestId, "uses", api.IPAMRangeGK); len(cycle) > 0 {
		return r.invalid(ctx, log, current.object, "reference cycle not allowed: %v", cycle)
	}

	ipr, err := r.cache.getRange(ctx, log, rangeId.ObjectKey, nil)
	if ipr == nil {
		if err != nil {
			return utils.Requeue(err)
		}
		return r.setStatus(ctx, log, current.object, v1alpha1.StateError, "IPAMRange %s not found", rangeId.ObjectKey)
	}
	defer r.cache.release(log, rangeId.ObjectKey)
	if ipr.error != "" {
		return r.setStatus(ctx, log, current.object, v1alpha1.StateError, "IPAMRange %s not valid: %s", rangeId.ObjectKey, ipr.error)
	}

	if !ipr.object.DeletionTimestamp.IsZero() {
		return r.setStatus(ctx, log, current.object, v1alpha1.StateError, "IPAMRange %s is already deleting", rangeId.ObjectKey)
	}

	if ipr.ipam == nil {
		if len(ipr.allocations) == 0 {
			log.Info("parent is not yet ready", "parent", ipr.objectId)
			return utils.Succeeded()
		}
	}

	for i, c := range current.requestSpecs {
		if c.Spec != nil && c.Spec.Bits() > ipr.ipam.Bits() {
			return r.setStatus(ctx, log, current.object, v1alpha1.StateInvalid, "size (entry %d) %d bits too large", i, ipr.ipam.Bits())
		}
	}
	// set finalizer current object
	if err := r.AssureFinalizer(ctx, log, current.object); err != nil {
		return utils.Requeue(err)
	}
	// set finalizer on parent
	if found, err := r.CheckAndAssureFinalizer(ctx, log, ipr.object); err != nil || !found {
		return utils.Requeue(err)
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
			requeue = true
		}
	}
	if len(allocated) == 0 && ipr.pendingRequest == nil && len(deleted) == 0 {
		log.Info("nothing new to allocate or free")
		return utils.Succeeded()
	}
	log.Info("allocation changes", "allocated", allocated, "deleted", deleted)
	newObj := current.determineState(allocations, deletions)
	if err := r.Client.Status().Patch(ctx, newObj, client.MergeFrom(current.object)); err != nil {
		return utils.Requeue(err)
	}
	log.Info("allocation finished. trigger range", "objectkey", ipr.objectId.ObjectKey)
	// make sure to update cache with new object -> trigger range object -> triggers its users
	current.object = newObj
	current.updateAllocations(allocations, deletions)
	r.Trigger(ipr.objectId)
	if requeue {
		return utils.Requeue(nil)
	}
	return utils.Succeeded()
}

func (r *Reconciler) deleteRequest(ctx context.Context, log logr.Logger, current *IPAM) (ctrl.Result, error) {
	requestId := utils.NewObjectId(current.object)
	for _, f := range current.object.GetFinalizers() {
		if f != finalizerName {
			log.Info("object still in use by others, delaying deletion", "key", requestId.ObjectKey)
			return utils.Succeeded()
		}
	}
	var ipr *IPAM
	if len(r.GetUsageCache().GetUsersForRelationToGK(requestId, "uses", api.IPAMRangeGK)) > 0 {
		// TODO: reject deletion in validation webhook if there still users of this object
		return utils.Succeeded()
	}
	if len(current.object.Status.CIDRs) > 0 {
		rangeId, failed, err := r.ScopeEvaluator.EvaluateScopedReferenceToObjectId(ctx, current.object.Namespace, api.IPAMRangeGK, current.object.Spec.Parent)
		if err != nil {
			if !failed {
				return utils.Requeue(err)
			}
		}
		if rangeId.Name != "" {
			ipr, err = r.cache.getRange(ctx, log, rangeId.ObjectKey, nil)
			if err != nil {
				return utils.Requeue(err)
			}
			if ipr != nil {
				defer r.cache.release(log, rangeId.ObjectKey)
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
				log.Info("updating status for", "object", current.objectId)
				if err := r.Client.Status().Patch(ctx, newObj, client.MergeFrom(current.object)); err != nil {
					return utils.Requeue(err)
				}
				current.object = newObj
				current.allocations = nil
				current.deletions = nil
			}
		}
	}
	if err := r.AssureFinalizerRemoved(ctx, log, current.object); err != nil {
		return utils.Requeue(err)
	}
	if ipr != nil {
		log.Info("trigger range update", "range", ipr.objectId)
		r.Trigger(ipr.objectId)
	}
	return utils.Succeeded()
}
