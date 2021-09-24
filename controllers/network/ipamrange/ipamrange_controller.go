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
	"reflect"

	"github.com/go-logr/logr"
	common "github.com/onmetal/onmetal-api/apis/common/v1alpha1"
	"github.com/onmetal/onmetal-api/apis/core"
	api "github.com/onmetal/onmetal-api/apis/network/v1alpha1"
	"github.com/onmetal/onmetal-api/pkg/cache/usagecache"
	"github.com/onmetal/onmetal-api/pkg/controllerutils"
	"github.com/onmetal/onmetal-api/pkg/manager"
	"github.com/onmetal/onmetal-api/pkg/utils"
	"k8s.io/apimachinery/pkg/api/errors"
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
			log.Info("deleted")
			return r.HandleDeleted(ctx, log, client.ObjectKey{
				Namespace: req.Namespace,
				Name:      req.Name,
			})
		}
		return utils.Requeue(err)
	}
	if obj.ObjectMeta.DeletionTimestamp.IsZero() {
		log.Info("reconciling")
		return r.HandleReconcile(ctx, log, &obj)
	} else {
		log.Info("deleting")
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
	newObjectKey := utils.NewObjectKey(obj)
	current, err := r.cache.getRange(ctx, log, newObjectKey, obj)
	if err != nil {
		return utils.Requeue(err)
	}
	defer r.cache.release(log, newObjectKey)
	if current.error != "" {
		return r.invalid(ctx, log, current, current.error)
	}
	if obj.Spec.Parent != nil {
		// handle IPAMRange as request first
		result, err := r.reconcileRequest(ctx, log, current)
		if err != nil || !result.IsZero() {
			return result, err
		}
	} else {
		result, err := r.reconcileRootRequest(ctx, log, current)
		if err != nil || !result.IsZero() {
			return result, err
		}
	}
	if current.ipam != nil {
		return r.reconcileRange(ctx, log, current)
	}
	return ctrl.Result{}, nil
}

func (r *Reconciler) HandleDelete(ctx context.Context, log logr.Logger, obj *api.IPAMRange) (ctrl.Result, error) {
	newObjectKey := utils.NewObjectKey(obj)
	current, err := r.cache.getRange(ctx, log, newObjectKey, obj)
	if err != nil {
		return utils.Requeue(err)
	}
	defer r.cache.release(log, newObjectKey)
	return r.deleteRequest(ctx, log, current)
}

// TODO: generalize state handling
func (r *Reconciler) invalid(ctx context.Context, log logr.Logger, ipr *IPAM, message string, args ...interface{}) (ctrl.Result, error) {
	return r.setStatus(ctx, log, ipr, common.StateInvalid, message, args...)
}

func (r *Reconciler) setStatus(ctx context.Context, log logr.Logger, ipr *IPAM, state string, msg string, args ...interface{}) (ctrl.Result, error) {
	if err := updateStatus(ctx, log, r, ipr, func(obj *api.IPAMRange) {
		obj.Status.State = state
		obj.Status.Message = fmt.Sprintf(msg, args...)
	}, nil); err != nil {
		return utils.Requeue(err)
	}
	return utils.Succeeded()
}

func updateStatus(ctx context.Context, log logr.Logger, clt client.Client, ipr *IPAM, update func(ipamRange *api.IPAMRange), cache func()) error {
	newObj := ipr.object.DeepCopy()
	update(newObj)
	if !reflect.DeepEqual(newObj, ipr.object) {
		log.Info("updating status", "state", newObj.Status.State, "message", newObj.Status.Message)
		if err := clt.Status().Patch(ctx, newObj, client.MergeFrom(ipr.object)); err != nil {
			return err
		}
		// update self in cache
		if ipr.ipam == nil {
			ipr.updateFrom(log, newObj)
		} else {
			// attention; cache updated need to be done externally
			ipr.object = newObj
			if cache != nil {
				cache()
			}
		}
	}
	return nil
}

func (r *Reconciler) HandleDeleted(ctx context.Context, log logr.Logger, key client.ObjectKey) (ctrl.Result, error) {
	r.cache.removeRange(key)
	objectId := utils.ObjectId{
		ObjectKey: key,
		GroupKind: api.IPAMRangeGK,
	}
	r.GetUsageCache().DeleteObject(objectId)
	return utils.Succeeded()
}
