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
	"github.com/onmetal/onmetal-api/apis/common/v1alpha1"
	api "github.com/onmetal/onmetal-api/apis/network/v1alpha1"
	"github.com/onmetal/onmetal-api/pkg/cache/usagecache"
	"github.com/onmetal/onmetal-api/pkg/ipam"
	"github.com/onmetal/onmetal-api/pkg/utils"
	"net"
	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/client"
)

func (r *Reconciler) reconcileRequest(ctx context.Context, log *utils.Logger, current *IPAM) (ctrl.Result, error) {
	log.Infof("reconcile request for %s/%s", current.object.Namespace, current.object.Name)
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

	requestId := utils.NewObjectId(current.object)
	r.GetUsageCache().ReplaceObjectUsageInfo(requestId, usagecache.NewObjectUsageInfo("uses", rangeId))

	// check for cycles and self reference
	if cycle := r.GetUsageCache().IsCyclicForRelationForGK(requestId, "uses", api.IPAMRangeGK); len(cycle) > 0 {
		return r.invalid(ctx, log, current.object, "reference cycle not allowed: %v", cycle)
	}

	ipr, err := r.cache.getRange(ctx, rangeId.ObjectKey, nil)
	if ipr == nil {
		return r.setStatus(ctx, log, current.object, v1alpha1.StateError, "IPAMRange %s not found", rangeId.ObjectKey)
	}
	defer r.cache.release(rangeId.ObjectKey)
	if ipr.error != "" {
		return r.setStatus(ctx, log, current.object, v1alpha1.StateError, "IPAMRange %s not valid: %s", rangeId.ObjectKey, ipr.error)
	}

	if !ipr.object.DeletionTimestamp.IsZero() {
		return r.setStatus(ctx, log, current.object, v1alpha1.StateError, "IPAMRange %s is already deleting", rangeId.ObjectKey)
	}

	if ipr.ipam == nil {
		if len(ipr.object.Status.CIDRs) == 0 {
			log.Infof("parent %s is not yet ready", ipr.objectId)
			return utils.Succeeded()
		}
	}

	if len(current.requestSpecs) > 0 {
		for i, c := range current.requestSpecs {
			if c.Bits() > ipr.ipam.Bits() {
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

		if len(current.object.Status.CIDRs) > 0 {
			log.Infof("already allocated")
			return utils.Succeeded()
		}

		allocated := []*net.IPNet{}
		if ipr.pendingRequest != nil {
			if ipr.pendingRequest.key != requestId.ObjectKey {
				return utils.Succeeded()
			}
			log.Infof("found pending request: %s", ipr.pendingRequest.key)
			allocated = ipr.pendingRequest.CIDRs
		} else {
			allocated, err = ipr.Alloc(ctx, log, r.Client, current.requestSpecs, current.objectId.ObjectKey)
			if err != nil {
				return utils.Requeue(err)
			}
		}
		if len(allocated) != 0 {
			log.Infof("allocated %v", allocated)
			newObj := current.object.DeepCopy()
			newObj.Status.CIDRs = nil
			for _, a := range allocated {
				newObj.Status.CIDRs = append(newObj.Status.CIDRs, a.String())
			}
			if err := r.Client.Status().Patch(ctx, newObj, client.MergeFrom(current.object)); err != nil {
				return utils.Requeue(err)
			}
			log.Infof("allocation finished. trigger range %s", ipr.objectId.ObjectKey)
			// make sure to update cache with new object
			current.object = newObj
			r.Trigger(ipr.objectId)
		} else {
			if err != nil {
				return r.setStatus(ctx, log, current.object, v1alpha1.StateError, err.Error())
			}
			return r.setStatus(ctx, log, current.object, v1alpha1.StateBusy, "requested range is busy")
		}
	}
	return r.ready(ctx, log, current.object, "")
}

func (r *Reconciler) deleteRequest(ctx context.Context, log *utils.Logger, current *IPAM) (ctrl.Result, error) {
	requestId := utils.NewObjectId(current.object)
	for _, f := range current.object.GetFinalizers() {
		if f != finalizerName {
			log.Infof("object %s still in use by others -> delay deletion", requestId.ObjectKey)
			return utils.Succeeded()
		}
	}

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
			ipr, err := r.cache.getRange(ctx, rangeId.ObjectKey, nil)
			if err != nil {
				return utils.Requeue(err)
			}
			if ipr != nil {
				defer r.cache.release(rangeId.ObjectKey)
				allocated := []*net.IPNet{}

				if ipr.pendingRequest != nil {
					if ipr.pendingRequest.key != requestId.ObjectKey {
						log.Infof("operation on ipamrange still pending -> delay delete")
						return utils.Succeeded()
					}
					allocated = ipr.pendingRequest.CIDRs
					log.Infof("continuing release %s", allocated)
				} else {
					var allocated ipam.CIDRList
					for _, c := range current.object.Status.CIDRs {
						_, cidr, err := net.ParseCIDR(c)
						if err == nil {
							allocated = append(allocated, cidr)
						}
					}
					err = ipr.Free(ctx, log, r.Client, allocated, current.objectId.ObjectKey)
					if err != nil {
						return utils.Requeue(err)
					}
				}
				newObj := current.object.DeepCopy()
				newObj.Status.CIDRs = nil
				if err := r.Client.Status().Patch(ctx, newObj, client.MergeFrom(current.object)); err != nil {
					return utils.Requeue(err)
				}
				current.object = newObj
				r.Trigger(ipr.objectId)
			}
		}
	}
	if err := r.AssureFinalizerRemoved(ctx, log, current.object); err != nil {
		return utils.Requeue(err)
	}
	return utils.Succeeded()
}
