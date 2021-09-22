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
	"github.com/go-logr/logr"
	common "github.com/onmetal/onmetal-api/apis/common/v1alpha1"
	"github.com/onmetal/onmetal-api/apis/core"
	api "github.com/onmetal/onmetal-api/apis/network/v1alpha1"
	"github.com/onmetal/onmetal-api/pkg/cache/usagecache"
	"github.com/onmetal/onmetal-api/pkg/controllerutils"
	"github.com/onmetal/onmetal-api/pkg/ipam"
	"github.com/onmetal/onmetal-api/pkg/manager"
	"github.com/onmetal/onmetal-api/pkg/utils"
	"k8s.io/apimachinery/pkg/api/errors"
	"net"
	"reflect"
	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/log"
)

const (
	finalizerName = core.LabelDomain + "/ipamrange"
)

// Reconciler reconciles a IPAMRange object
type Reconciler struct {
	controllerutils.Reconciler
	cache *IPAMCache
}

func NewReconciler() *Reconciler {
	return &Reconciler{
		Reconciler: controllerutils.NewReconciler("ipamrange", finalizerName),
		cache:      nil,
	}
}

//+kubebuilder:rbac:groups=network.onmetal.de,resources=ipamranges,verbs=get;list;watch;create;update;patch;delete
//+kubebuilder:rbac:groups=network.onmetal.de,resources=ipamranges/status,verbs=get;update;patch
//+kubebuilder:rbac:groups=network.onmetal.de,resources=ipamranges/finalizers,verbs=update

// Reconcile
// if parent -> handle request part and set range
func (r *Reconciler) Reconcile(ctx context.Context, req ctrl.Request) (ctrl.Result, error) {
	r.Wait() // wait until all caches are initialized
	log := log.FromContext(ctx)
	var obj api.IPAMRange
	if err := r.Get(ctx, req.NamespacedName, &obj); err != nil {
		if errors.IsNotFound(err) {
			return r.HandleDeleted(ctx, log, client.ObjectKey{
				Namespace: req.Namespace,
				Name:      req.Name,
			})
		}
		return utils.Requeue(err)
	}
	if obj.ObjectMeta.DeletionTimestamp.IsZero() {
		return r.HandleReconcile(ctx, log, &obj)
	} else {
		return r.HandleDelete(ctx, log, &obj)
	}
}

// SetupWithManager sets up the controller with the Manager.
func (r *Reconciler) SetupWithManager(mgr *manager.Manager) error {
	r.Reconciler.SetupWithManager(mgr)
	r.cache = NewIPAMCache(r.Client)
	c, err := ctrl.NewControllerManagedBy(mgr).
		For(&api.IPAMRange{}).
		Build(r)
	if err == nil {
		mgr.RegisterControllerFor(api.IPAMRangeGK, c)
		if err := mgr.GetUsageCache().RegisterExtractor(context.Background(), api.IPAMRangeGK, usagecache.RelationUses, api.IPAMRangeGK, &IPAMUsageExtractor{mgr.GetScopeEvaluator()}); err != nil {
			return err
		}
	}
	return err
}

func (r *Reconciler) HandleReconcile(ctx context.Context, log logr.Logger, obj *api.IPAMRange) (ctrl.Result, error) {
	log.Info("handle reconcile")
	newObjectKey := utils.NewObjectKey(obj)
	current, err := r.cache.getRange(ctx, log, newObjectKey, obj)
	if err != nil {
		return utils.Requeue(err)
	}
	defer r.cache.release(log, newObjectKey)
	if current.error != "" {
		return r.invalid(ctx, log, obj, current.error)
	}
	if obj.Spec.Parent != nil {
		// handle IPAMRange as request first
		result, err := r.reconcileRequest(ctx, log, current)
		if err != nil || !result.IsZero() {
			return result, err
		}
	} else {
		if len(current.requestSpecs) == 0 {
			return r.invalid(ctx, log, obj, "at least one cidr must be specified for root (no parent) ipam range")
		}
		var cidrs AllocationStatusList
		for i, c := range current.requestSpecs {
			if !c.IsValid() {
				cidrs = append(cidrs, &AllocationStatus{
					Allocation: Allocation{
						Request: c.Request,
						CIDR:    nil,
					},
					Status:  api.AllocationStateFailed,
					Message: c.Error,
				})
				continue
			}
			if !c.Spec.IsCIDR() {
				return r.invalid(ctx, log, obj, "request spec %d does is not a valid cidr %s for a root ipam range", i, c)
			}
			_, cidr, _ := net.ParseCIDR(c.Request)
			if ipam.CIDRHostMaskSize(cidr) == 0 {
				// TODO: rethink IP delegation
				return r.invalid(ctx, log, obj, "root cidr must have more than one ip address")
			}
			cidrs = append(cidrs, &AllocationStatus{
				Allocation: Allocation{
					Request: c.Request,
					CIDR:    cidr,
				},
				Status:  api.AllocationStateAllocated,
				Message: SuccessfulUsageMessage,
			})
		}
		newObj := current.object.DeepCopy()
		newObj.Status.CIDRs = cidrs.GetAllocationStatusList()
		if !reflect.DeepEqual(newObj, obj) {
			log.Info("setting range status", "requests", current.requestSpecs)
			if err := r.Status().Patch(ctx, newObj, client.MergeFrom(obj)); err != nil {
				return utils.Requeue(err)
			}
			if current.ipam == nil {
				// update self in cache
				current.updateFrom(log, newObj)
			} else {
				current.updateAllocations(cidrs)
			}
			// trigger all users of this ipamrange
			log.Info("trigger all users of range", "key", current.objectId.ObjectKey)
			users := r.GetUsageCache().GetUsersForRelationToGK(current.objectId, "uses", api.IPAMRangeGK)
			r.TriggerAll(users)
		}
	}
	if current.ipam != nil {
		return r.reconcileRange(ctx, log, current)
	}
	return ctrl.Result{}, nil
}

func (r *Reconciler) HandleDelete(ctx context.Context, log logr.Logger, obj *api.IPAMRange) (ctrl.Result, error) {
	log.Info("handle deletion")
	newObjectKey := utils.NewObjectKey(obj)
	current, err := r.cache.getRange(ctx, log, newObjectKey, obj)
	if err != nil {
		return utils.Requeue(err)
	}
	defer r.cache.release(log, newObjectKey)
	return r.deleteRequest(ctx, log, current)
}

// TODO: generalize state handling
func (r *Reconciler) invalid(ctx context.Context, log logr.Logger, obj *api.IPAMRange, message string, args ...interface{}) (ctrl.Result, error) {
	return r.setStatus(ctx, log, obj, common.StateInvalid, message, args...)
}

func (r *Reconciler) ready(ctx context.Context, log logr.Logger, obj *api.IPAMRange, message string, args ...interface{}) (ctrl.Result, error) {
	return r.setStatus(ctx, log, obj, common.StateReady, message, args...)
}

func (r *Reconciler) setStatus(ctx context.Context, log logr.Logger, obj *api.IPAMRange, state string, msg string, args ...interface{}) (ctrl.Result, error) {
	newIpamRange := obj.DeepCopy()
	newIpamRange.Status.State = state
	newIpamRange.Status.Message = fmt.Sprintf(msg, args...)
	log.Info("updating status", "state", state, "message", newIpamRange.Status.Message)
	if err := r.Status().Patch(ctx, newIpamRange, client.MergeFrom(obj)); err != nil {
		return utils.Requeue(err)
	}
	return utils.Succeeded()
}

func (r *Reconciler) HandleDeleted(ctx context.Context, log logr.Logger, key client.ObjectKey) (ctrl.Result, error) {
	log.Info("deleted", "key", key)
	r.cache.removeRange(key)
	objectId := utils.ObjectId{
		ObjectKey: key,
		GroupKind: api.IPAMRangeGK,
	}
	users := r.GetUsageCache().GetUsersFor(objectId)
	r.GetUsageCache().DeleteObject(objectId)
	r.TriggerAll(users)
	return utils.Succeeded()
}
