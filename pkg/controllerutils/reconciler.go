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

package controllerutils

import (
	"context"
	"github.com/go-logr/logr"
	"github.com/onmetal/onmetal-api/pkg/cache/ownercache"
	"github.com/onmetal/onmetal-api/pkg/cache/usagecache"
	"github.com/onmetal/onmetal-api/pkg/manager"
	"github.com/onmetal/onmetal-api/pkg/scopes"
	"github.com/onmetal/onmetal-api/pkg/utils"
	"k8s.io/apimachinery/pkg/types"
	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/controller/controllerutil"
)

type Reconciler struct {
	client.Client
	name          string
	manager       *manager.Manager
	finalizerName string
	scopes.ScopeEvaluator
	log logr.Logger
}

func NewReconciler(name, finalizerName string) Reconciler {
	return Reconciler{
		finalizerName: finalizerName,
		name:          name,
		log:           ctrl.Log.WithName(name),
	}
}

func (r *Reconciler) SetupWithManager(mgr *manager.Manager) {
	r.manager = mgr
	r.Client = mgr.GetClient()
	r.ScopeEvaluator = mgr.GetScopeEvaluator()
}

func (r *Reconciler) GetOwnerCache() ownercache.OwnerCache {
	return r.manager.GetOwnerCache()
}

func (r *Reconciler) GetUsageCache() usagecache.UsageCache {
	return r.manager.GetUsageCache()
}

func (r *Reconciler) GetFinalizerName() string {
	return r.finalizerName
}

func (r *Reconciler) GetClient() client.Client {
	return r.Client
}

func (r *Reconciler) GetManager() *manager.Manager {
	return r.manager
}

func (r *Reconciler) GetLog(key types.NamespacedName) *utils.Logger {
	return utils.NewLogger(r.log, r.name, key)
}

func (r *Reconciler) TriggerAll(ids utils.ObjectIds) {
	r.manager.TriggerAll(ids)
}

func (r *Reconciler) Trigger(id utils.ObjectId) {
	r.manager.Trigger(id)
}

func (r *Reconciler) Wait() {
	r.manager.Wait()
}

func (r *Reconciler) HasOtherFinalizers(object client.Object) bool {
	finalizers := object.GetFinalizers()
	has := utils.ContainsString(finalizers, r.finalizerName)
	return has && len(finalizers) == 1 || !has && len(finalizers) == 0
}

func (r *Reconciler) HasFinalizer(object client.Object, finalizerName ...string) bool {
	if len(finalizerName) > 0 {
		return utils.ContainsString(object.GetFinalizers(), finalizerName[0])
	}
	return utils.ContainsString(object.GetFinalizers(), r.finalizerName)
}

// CheckAndAssureFinalizer ensures that a finalizer is on a given runtime object
// Returns false if the finalizer has been added.
func (r *Reconciler) CheckAndAssureFinalizer(ctx context.Context, log *utils.Logger, object client.Object, finalizerName ...string) (bool, error) {
	fn := r.finalizerName
	if len(finalizerName) > 0 {
		fn = finalizerName[0]
	}
	if !utils.ContainsString(object.GetFinalizers(), fn) {
		log.Infof("setting finalizer %s", fn)
		controllerutil.AddFinalizer(object, fn)
		return false, r.Client.Update(ctx, object)
	}
	return true, nil
}

// AssureFinalizer ensures that a finalizer is on a given runtime object
func (r *Reconciler) AssureFinalizer(ctx context.Context, log *utils.Logger, object client.Object, finalizerName ...string) error {
	_, err := r.CheckAndAssureFinalizer(ctx, log, object, finalizerName...)
	return err
}

// AssureFinalizerRemoved ensures that a finalizer does not exist anymore for a given runtime object
func (r *Reconciler) AssureFinalizerRemoved(ctx context.Context, log *utils.Logger, object client.Object, finalizerName ...string) error {
	fn := r.finalizerName
	if len(finalizerName) > 0 {
		fn = finalizerName[0]
	}
	if utils.ContainsString(object.GetFinalizers(), fn) {
		log.Infof("removing finalizer %s", fn)
		controllerutil.RemoveFinalizer(object, fn)
		return r.Client.Update(ctx, object)
	}
	return nil
}
