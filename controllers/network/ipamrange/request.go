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
	"github.com/onmetal/onmetal-api/apis/common/v1alpha1"
	api "github.com/onmetal/onmetal-api/apis/network/v1alpha1"
	"github.com/onmetal/onmetal-api/pkg/cache/usagecache"
	"github.com/onmetal/onmetal-api/pkg/utils"
	"net"
	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/client"
)

//var assignedCIDRField = fieldpath.RequiredField(&api.IPAMRequest{}, ".Status.CIDR")
//var rangeFilter = resources.NewGroupKindFilter(api.IPAMRANGE)
//
func (r *Reconciler) setupRequest(ctx context.Context, log *utils.Logger, current *IPAM) (utils.ObjectId, ctrl.Result, error) {
	objId, failed, err := r.ScopeEvaluator.EvaluateScopedReferenceToObjectId(ctx, current.object.Namespace, api.IPAMRangeGK, current.object.Spec.Parent)
	if err != nil {
		if !failed {
			result, err := utils.Requeue(err)
			return objId, result, err
		}
		result, err := r.setStatus(ctx, log, current.object, v1alpha1.StateError, err.Error())
		return objId, result, err
	}

	if objId.Name != "" {
		ipam := r.cache.ipams[objId.ObjectKey]
		if ipam != nil {
			if len(current.object.Status.CIDRs) != 0 {
				for _, c := range current.object.Status.CIDRs {
					_, cidr, err := net.ParseCIDR(c)
					if err != nil {
						log.Errorf(err, "invalid state of ipam request %v: invalid cidr: %s", current.object, c)
					} else {
						ipam.ipam.Busy(cidr)
					}
				}
			}
		}
		return objId, ctrl.Result{}, nil
	}
	return utils.ObjectId{}, ctrl.Result{}, nil
}

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
	r.manager.GetUsageCache().ReplaceObjectUsageInfo(requestId, usagecache.NewObjectUsageInfo("uses", rangeId))

	// check for cycles and self reference
	if cycle := r.manager.GetUsageCache().IsCyclicForRelationForGK(requestId, "uses", api.IPAMRangeGK); len(cycle) > 0 {
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

	if len(current.requestSpecs) > 0 {
		for i, c := range current.requestSpecs {
			if c.Bits() > ipr.ipam.Bits() {
				return r.setStatus(ctx, log, current.object, v1alpha1.StateInvalid, "size (entry %d) %d too large: network %d", i, c.NetBits(), ipr.ipam.Bits())
			}
		}
		// set finalizer current object
		if err := utils.AssureFinalizer(ctx, log, r.Client, finalizerName, current.object); err != nil {
			return utils.Requeue(err)
		}
		// set finalizer on parent
		if err := utils.AssureFinalizer(ctx, log, r.Client, finalizerName, ipr.object); err != nil {
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
			allocated = ipr.pendingRequest.CIDRs
		} else {
			if len(current.requestSpecs) > 0 {
				for _, c := range current.requestSpecs {
					cidr, err := c.Alloc(ipr.ipam)
					if cidr == nil || err != nil {
						for _, a := range allocated {
							ipr.ipam.Free(a)
						}
						allocated = nil
						break
					} else {
						allocated = append(allocated, cidr)
					}
				}
			}
			if len(allocated) != 0 {
				if err := r.updateRange(ctx, ipr, requestId, allocated); err != nil {
					for _, a := range allocated {
						ipr.ipam.Free(a)
					}
					return utils.Requeue(err)
				}
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
		} else {
			if err != nil {
				return r.setStatus(ctx, log, current.object, v1alpha1.StateError, err.Error())
			}
			return r.setStatus(ctx, log, current.object, v1alpha1.StateBusy, "requested range is busy")
		}
	}
	return r.ready(ctx, log, current.object, "")
}

func (r *Reconciler) setIPAMState(ipr *IPAM, newIpr *api.IPAMRange) {
	blocks, round := ipr.ipam.State()
	var state []string
	for i := 0; i < len(round); i++ {
		state = append(state, fmt.Sprintf("%s/%d", round[i], i))
	}
	newIpr.Status.RoundRobinState = state
	newIpr.Status.AllocationState = blocks
}

func (r *Reconciler) updateRange(ctx context.Context, ipr *IPAM, id utils.ObjectId, allocated []*net.IPNet) error {
	newIpr := ipr.object.DeepCopy()
	r.setIPAMState(ipr, newIpr)

	newIpr.Status.PendingRequest = &api.IPAMPendingRequest{
		Name:      id.Name,
		Namespace: id.Namespace,
		CIDRs:     nil,
	}

	for _, a := range allocated {
		newIpr.Status.PendingRequest.CIDRs = append(newIpr.Status.PendingRequest.CIDRs, a.String())
	}
	err := r.Client.Status().Patch(ctx, newIpr, client.MergeFrom(ipr.object))
	if err == nil {
		ipr.object = newIpr
		ipr.pendingRequest = &PendingRequest{
			key:   id.ObjectKey,
			CIDRs: allocated,
		}
	}
	return err
}

func (r *Reconciler) deleteRequest(ctx context.Context, log *utils.Logger, current *IPAM) (ctrl.Result, error) {
	requestId := utils.NewObjectId(current.object)
	for _, f := range current.object.GetFinalizers() {
		if f != finalizerName {
			log.Infof("object %s still in use by others -> delay deletion", requestId.ObjectKey)
			return utils.Succeeded()
		}
	}

	if len(r.manager.GetUsageCache().GetUsersForRelationToGK(requestId, "uses", api.IPAMRangeGK)) > 0 {
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
					log.Infof("releasing %s", current.object.Status.CIDRs)
					for _, c := range current.object.Status.CIDRs {
						_, cidr, err := net.ParseCIDR(c)
						if err == nil {
							allocated = append(allocated, cidr)
						}
					}
				}

				if len(allocated) != 0 {
					for _, a := range allocated {
						ipr.ipam.Free(a)
					}
					if err := r.updateRange(ctx, ipr, requestId, allocated); err != nil {
						for _, a := range allocated {
							ipr.ipam.Busy(a)
						}
						return utils.Requeue(err)
					}
					log.Infof("releasing %v", allocated)
				}
				newObj := current.object.DeepCopy()
				newObj.Status.CIDRs = nil
				if err := r.Client.Status().Patch(ctx, newObj, client.MergeFrom(current.object)); err != nil {
					return utils.Requeue(err)
				}
			}
		}
	}
	if err := utils.AssureFinalizerRemoved(ctx, log, r.Client, finalizerName, current.object); err != nil {
		return utils.Requeue(err)
	}
	return utils.Succeeded()
}
