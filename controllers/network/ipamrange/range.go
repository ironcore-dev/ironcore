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
	"sigs.k8s.io/controller-runtime/pkg/controller/controllerutil"

	"github.com/go-logr/logr"
	networkv1alpha1 "github.com/onmetal/onmetal-api/apis/network/v1alpha1"
	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/client"
)

func (r *Reconciler) reconcileRange(ctx context.Context, log logr.Logger, current *IPAM) (ctrl.Result, error) {
	log.Info("reconcile range")
	if !controllerutil.ContainsFinalizer(current.object, finalizerName) {
		controllerutil.AddFinalizer(current.object, finalizerName)
		if err := r.Update(ctx, current.object, fieldOwner); err != nil {
			return ctrl.Result{}, fmt.Errorf("could not add finalizer: %w", err)
		}
		return ctrl.Result{}, nil
	}
	if current.object.Spec.Mode == "" && current.ipam != nil {
		mode := networkv1alpha1.ModeFirstMatch
		if current.ipam.IsRoundRobin() {
			mode = networkv1alpha1.ModeRoundRobin
		}
		newObj := current.object.DeepCopy()
		newObj.Spec.Mode = mode
		if err := r.Patch(ctx, newObj, client.MergeFrom(current.object)); err != nil {
			return ctrl.Result{}, fmt.Errorf("could not patch range: %w", err)
		}
	}
	if current.pendingRequest != nil {
		log.Info("found pending request", "key", current.pendingRequest.key)
		req, err := r.cache.getRange(ctx, log, current.pendingRequest.key, nil)
		if err != nil {
			return ctrl.Result{}, fmt.Errorf("could not get range: %w", err)
		}
		newCurrent := current.object.DeepCopy()
		if req != nil {
			defer r.cache.release(log, current.pendingRequest.key)
			if !current.pendingRequest.MatchState(log, req) {
				return ctrl.Result{}, nil
			}
		} else {
			log.Info("pending request already deleted")
			if len(current.pendingRequest.CIDRs) > 0 {
				log.Info("releasing pending allocations", "cidrs", current.pendingRequest.CIDRs)
			}
			for _, c := range current.pendingRequest.CIDRs {
				current.ipam.Free(c.CIDR)
			}
		}
		newCurrent.Status.PendingRequest = nil
		log.Info("finalizing pending request", "key", current.pendingRequest.key)
		if err := r.Status().Patch(ctx, newCurrent, client.MergeFrom(current.object)); err != nil {
			return ctrl.Result{}, fmt.Errorf("could not finalize pending request: %w", err)
		}
		current.pendingRequest = nil
		// trigger all users of this ipamrange
		log.Info("trigger all users of range", "key", client.ObjectKeyFromObject(current.object))
	}
	return r.setStatus(ctx, log, current, networkv1alpha1.IPAMRangeReady, "")
}
