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
	common "github.com/onmetal/onmetal-api/apis/common/v1alpha1"
	api "github.com/onmetal/onmetal-api/apis/network/v1alpha1"
	"github.com/onmetal/onmetal-api/pkg/logging"
	"github.com/onmetal/onmetal-api/pkg/utils"
	"reflect"
	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/client"
)

func (r *Reconciler) reconcileRange(ctx context.Context, log *logging.Logger, current *IPAM) (ctrl.Result, error) {
	if len(r.GetUsageCache().GetUsersForRelationToGK(utils.NewObjectId(current.object), "uses", api.IPAMRangeGK)) > 0 {
		if err := r.AssureFinalizer(ctx, log, current.object); err != nil {
			return utils.Requeue(err)
		}
	}
	if current.object.Spec.Mode == "" && current.ipam != nil {
		mode := api.ModeFirstMatch
		if current.ipam.IsRoundRobin() {
			mode = api.ModeRoundRobin
		}
		newObj := current.object.DeepCopy()
		newObj.Spec.Mode = mode
		if err := r.Patch(ctx, newObj, client.MergeFrom(current.object)); err != nil {
			return utils.Requeue(err)
		}
	}
	if current.pendingRequest != nil {
		log.Infof("found pending request for %s", current.pendingRequest.key)
		req, err := r.cache.getRange(ctx, log, current.pendingRequest.key, nil)
		if err != nil {
			return utils.Requeue(err)
		}
		newCurrent := current.object.DeepCopy()
		if req != nil {
			log.Infof("found status for pending request:    %v", req.object.Status.CIDRs)
			defer r.cache.release(log, current.pendingRequest.key)
			var list []string
			for _, c := range current.pendingRequest.CIDRs {
				list = append(list, c.String())
			}
			log.Infof("expected status for pending request: %v", list)
			if !reflect.DeepEqual(req.object.Status.CIDRs, list) {
				log.Infof("expected status not yet set in pending request")
				return utils.Succeeded()
			}
		} else {
			log.Infof("pending request already deleted")
			if len(current.pendingRequest.CIDRs) > 0 {
				log.Infof("releasing pending allocations: %s", current.pendingRequest.CIDRs)
			}
			for _, c := range current.pendingRequest.CIDRs {
				current.ipam.Free(c)
			}
		}
		newCurrent.Status.PendingRequest = nil
		log.Infof("finalizing pending request for %s", current.pendingRequest.key)
		if err := r.Status().Patch(ctx, newCurrent, client.MergeFrom(current.object)); err != nil {
			return utils.Requeue(err)
		}
		current.pendingRequest = nil
		// trigger all users of this ipamrange
		log.Infof("trigger all users of %s", current.objectId.ObjectKey)
		users := r.GetUsageCache().GetUsersForRelationToGK(current.objectId, "uses", api.IPAMRangeGK)
		r.TriggerAll(users)
	}
	return r.setStatus(ctx, log, current.object, common.StateReady, "")
}
