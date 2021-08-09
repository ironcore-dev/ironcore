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

package manager

import (
	"context"
	"fmt"
	"github.com/go-logr/logr"
	"github.com/onmetal/onmetal-api/pkg/utils"
	"k8s.io/apimachinery/pkg/api/errors"
	"k8s.io/apimachinery/pkg/runtime/schema"
	"reflect"
	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/reconcile"
	"sync"
)

type usageReconciler struct {
	client    client.Client
	gk        schema.GroupKind
	Log       logr.Logger
	cache     *UsageCache
	setupOnce sync.Once
	objType   reflect.Type
}

func newUsageReconciler(gk schema.GroupKind) *usageReconciler {
	return &usageReconciler{
		gk:  gk,
		Log: ctrl.Log.WithName("controllers").WithName("Usagecache").WithName(gk.String()),
	}
}

func (r *usageReconciler) setup(ctx context.Context) {
	list := utils.GetObjectListForGroupKind(r.client, r.gk)
	defer r.cache.ready.Remove()
	err := r.client.List(ctx, list)
	list.GetContinue()
	if err == nil {
		it := utils.MustItemListIterator(list)
		for it.HasNext() {
			o := it.Next()
			r.cache.ReplaceObject(o)
		}
	} else {
		panic(fmt.Sprintf("failed to setup %s ownercache", r.gk))
	}
}

func (r *usageReconciler) SetupWithCache(ctx context.Context, cache *UsageCache) error {
	r.client = cache.client
	r.cache = cache
	obj := utils.GetObjectForGroupKind(r.client, r.gk)
	if obj == nil {
		return fmt.Errorf("no version found for %s", r.gk)
	}
	r.objType = reflect.TypeOf(obj).Elem()
	err := ctrl.NewControllerManagedBy(cache.manager).
		For(obj).
		Complete(r)
	if err == nil {
		list := utils.GetObjectListForGroupKind(r.client, r.gk)
		if list == nil {
			return fmt.Errorf("no list object found for %s", r.gk)
		}
		cache.ready.Add()
	}
	return nil
}

func (r *usageReconciler) Reconcile(ctx context.Context, request reconcile.Request) (reconcile.Result, error) {
	r.setupOnce.Do(func() { r.setup(ctx) })
	obj := reflect.New(r.objType).Interface().(client.Object)
	key := client.ObjectKey{Namespace: request.Namespace, Name: request.Name}
	var oids utils.ObjectIds
	id := utils.ObjectId{
		ObjectKey: key,
		GroupKind: r.gk,
	}
	if err := r.client.Get(ctx, key, obj); err != nil {
		if errors.IsNotFound(err) {
			r.Log.Info(fmt.Sprintf("deleting %s from ownercache", id))
			oids = r.cache.DeleteObject(id)
		} else {
			return utils.Requeue(err)
		}
	} else {
		r.Log.Info(fmt.Sprintf("updating %s in ownercache", id))
		_, oids = r.cache.ReplaceObject(obj)
	}
	if r.cache.trigger != nil {
		for id := range oids {
			r.Log.Info(fmt.Sprintf("enqueue %s", id))
			r.cache.trigger.Trigger(id)
		}
	}
	return ctrl.Result{}, nil
}
