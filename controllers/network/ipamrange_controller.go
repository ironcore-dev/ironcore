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
	"github.com/go-logr/logr"
	common "github.com/onmetal/onmetal-api/apis/common/v1alpha1"
	"github.com/onmetal/onmetal-api/pkg/cache/usagecache"
	"github.com/onmetal/onmetal-api/pkg/manager"
	"github.com/onmetal/onmetal-api/pkg/scopes"
	"github.com/onmetal/onmetal-api/pkg/utils"
	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/log"

	api "github.com/onmetal/onmetal-api/apis/network/v1alpha1"
)

// IPAMRangeReconciler reconciles a IPAMRange object
type IPAMRangeReconciler struct {
	manager *manager.Manager
	scopes.ScopeEvaluator
	client.Client
	Log logr.Logger
}

//+kubebuilder:rbac:groups=network.onmetal.de,resources=ipamranges,verbs=get;list;watch;create;update;patch;delete
//+kubebuilder:rbac:groups=network.onmetal.de,resources=ipamranges/status,verbs=get;update;patch
//+kubebuilder:rbac:groups=network.onmetal.de,resources=ipamranges/finalizers,verbs=update

// Reconcile
// if parent -> handle request part and set range
// handle range
func (r *IPAMRangeReconciler) Reconcile(ctx context.Context, req ctrl.Request) (ctrl.Result, error) {
	r.manager.Wait() // wait until all caches are initialized
	_ = log.FromContext(ctx)
	log := r.Log.WithValues("ipamrange", req.NamespacedName)

	var ipamrange api.IPAMRange
	if err := r.Get(ctx, req.NamespacedName, &ipamrange); err != nil {
		return utils.SucceededIfNotFound(err)
	}

	//log.Info(fmt.Sprintf("reconcile %s/%s/%s/%s", ipamrange.APIVersion, ipamrange.Kind, ipamrange.Namespace, ipamrange.Name))

	if ipamrange.ObjectMeta.DeletionTimestamp.IsZero() {
		return r.HandleReconcile(ctx, log, &ipamrange)
	} else {
		return r.HandleDelete(ctx, log, &ipamrange)
	}
}

// SetupWithManager sets up the controller with the Manager.
func (r *IPAMRangeReconciler) SetupWithManager(mgr *manager.Manager) error {
	r.manager = mgr
	r.Client = mgr.GetClient()
	r.ScopeEvaluator = mgr.GetScopeEvaluator()
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

type IPAMUsageExtractor struct {
	scopeEvaluator scopes.ScopeEvaluator
}

func (e *IPAMUsageExtractor) ExtractUsage(object client.Object) utils.ObjectIds {
	ipam := object.(*api.IPAMRange)
	id, _, err := e.scopeEvaluator.EvaluateScopedReferenceToObjectId(context.Background(), object.GetNamespace(), api.IPAMRangeGK, ipam.Spec.Parent)
	if err != nil {
		return nil
	}
	return utils.NewObjectIds(id)
}

func (r *IPAMRangeReconciler) HandleReconcile(ctx context.Context, log logr.Logger, ipamrange *api.IPAMRange) (ctrl.Result, error) {
	log.Info("handle reconcile")
	if ipamrange.Spec.Parent != nil {
		// handle IPAMRange as request first
		var parent api.IPAMRange
		namespace, fail, err := r.EvaluateScopedReferenceToObject(ctx, ipamrange.Namespace, ipamrange.Spec.Parent, &parent)
		if err != nil {
			log.Error(err, "")
			if fail {
				newipamrange := ipamrange.DeepCopy()
				newipamrange.Status.Message = err.Error()
				newipamrange.Status.State = common.StateInvalid
				if err := r.Status().Patch(ctx, newipamrange, client.MergeFrom(ipamrange)); err != nil {
					return utils.Requeue(err)
				} else {
					return ctrl.Result{}, err
				}
			} else {
				return utils.Requeue(err)
			}
		}
		log.Info("found namespace for scope", "namespace", namespace, "scope", ipamrange.Spec.Parent.Scope)
	}
	return ctrl.Result{}, nil
}

func (r *IPAMRangeReconciler) HandleDelete(ctx context.Context, log logr.Logger, ipamrange *api.IPAMRange) (ctrl.Result, error) {
	return ctrl.Result{}, nil
}
