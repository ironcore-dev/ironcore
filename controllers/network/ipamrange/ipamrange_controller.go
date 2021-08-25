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
	"github.com/onmetal/onmetal-api/pkg/ipam"
	"github.com/onmetal/onmetal-api/pkg/manager"
	"github.com/onmetal/onmetal-api/pkg/scopes"
	"github.com/onmetal/onmetal-api/pkg/utils"
	"k8s.io/apimachinery/pkg/runtime"
	"net"
	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/log"
)

const (
	finalizerName = core.LabelDomain + "/ipamrange"
)

// Reconciler reconciles a IPAMRange object
type Reconciler struct {
	manager *manager.Manager
	scopes.ScopeEvaluator
	client.Client
	Log    logr.Logger
	Scheme *runtime.Scheme
	cache  *IPAMCache
}

//+kubebuilder:rbac:groups=network.onmetal.de,resources=ipamranges,verbs=get;list;watch;create;update;patch;delete
//+kubebuilder:rbac:groups=network.onmetal.de,resources=ipamranges/status,verbs=get;update;patch
//+kubebuilder:rbac:groups=network.onmetal.de,resources=ipamranges/finalizers,verbs=update

// Reconcile
// if parent -> handle request part and set range
func (r *Reconciler) Reconcile(ctx context.Context, req ctrl.Request) (ctrl.Result, error) {
	r.manager.Wait() // wait until all caches are initialized
	_ = log.FromContext(ctx)
	log := utils.NewLogger(r.Log, "ipamrange", req.NamespacedName)
	var obj api.IPAMRange
	if err := r.Get(ctx, req.NamespacedName, &obj); err != nil {
		return utils.SucceededIfNotFound(err)
	}
	if obj.ObjectMeta.DeletionTimestamp.IsZero() {
		return r.HandleReconcile(ctx, log, &obj)
	} else {
		return r.HandleDelete(ctx, log, &obj)
	}
}

// SetupWithManager sets up the controller with the Manager.
func (r *Reconciler) SetupWithManager(mgr *manager.Manager) error {
	r.manager = mgr
	r.Client = mgr.GetClient()
	r.ScopeEvaluator = mgr.GetScopeEvaluator()
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

func (r *Reconciler) HandleReconcile(ctx context.Context, log *utils.Logger, obj *api.IPAMRange) (ctrl.Result, error) {
	log.Infof("handle reconcile")
	newObjectKey := utils.NewObjectKey(obj)
	current, err := r.cache.getRange(ctx, newObjectKey, obj)
	if current.error != "" {
		return r.invalid(ctx, log, obj, current.error)
	}
	if err != nil {
		return utils.Requeue(err)
	}
	defer r.cache.release(newObjectKey)
	if obj.Spec.Parent != nil {
		// handle IPAMRange as request first
		result, err := r.reconcileRequest(ctx, log, current)
		if err != nil || !result.IsZero() {
			return result, err
		}
	} else {
		if current.ipam == nil {
			if len(current.requestSpecs) == 0 {
				return r.invalid(ctx, log, obj, "at least one cidr must be specified for root (no parent) ipam range")
			}
			cidrs := []string{}
			for i, c := range current.requestSpecs {
				if !c.IsCIDR() {
					return r.invalid(ctx, log, obj, "request spec %d does is not a valid cidr %s for a root ipam range", i, c)
				}
				_, cidr, _ := net.ParseCIDR(c.String())
				if ipam.CIDRHostMaskSize(cidr) == 0 {
					return r.invalid(ctx, log, obj, "root cidr must have more than one ip address")
				}
				cidrs = append(cidrs, c.String())
			}
			newObj := current.object.DeepCopy()
			newObj.Status.CIDRs = cidrs
			if err := r.Status().Patch(ctx, newObj, client.MergeFrom(obj)); err != nil {
				return utils.Requeue(err)
			}
			return utils.Succeeded()
		}
	}
	if current.ipam != nil {
		return r.reconcileRange(ctx, log, current)
	}
	return ctrl.Result{}, nil
}

func (r *Reconciler) HandleDelete(ctx context.Context, log *utils.Logger, obj *api.IPAMRange) (ctrl.Result, error) {
	log.Infof("handle deletion")
	newObjectKey := utils.NewObjectKey(obj)
	current, err := r.cache.getRange(ctx, newObjectKey, obj)
	if err != nil {
		return utils.Requeue(err)
	}
	defer r.cache.release(newObjectKey)
	return r.deleteRequest(ctx, log, current)
}

// TODO: generalize state handling
func (r *Reconciler) invalid(ctx context.Context, log *utils.Logger, obj *api.IPAMRange, message string, args ...interface{}) (ctrl.Result, error) {
	return r.setStatus(ctx, log, obj, common.StateInvalid, message, args...)
}

func (r *Reconciler) ready(ctx context.Context, log *utils.Logger, obj *api.IPAMRange, message string, args ...interface{}) (ctrl.Result, error) {
	return r.setStatus(ctx, log, obj, common.StateReady, message, args...)
}

func (r *Reconciler) setStatus(ctx context.Context, log *utils.Logger, obj *api.IPAMRange, state string, msg string, args ...interface{}) (ctrl.Result, error) {
	newIpamRange := obj.DeepCopy()
	newIpamRange.Status.State = state
	newIpamRange.Status.Message = fmt.Sprintf(msg, args...)
	log.Infof("updating status %s:%s", state, newIpamRange.Status.Message)
	if err := r.Status().Patch(ctx, newIpamRange, client.MergeFrom(obj)); err != nil {
		return utils.Requeue(err)
	}
	return utils.Succeeded()
}
