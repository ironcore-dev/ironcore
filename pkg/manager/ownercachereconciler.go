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
	"k8s.io/apimachinery/pkg/runtime/schema"
	"reflect"
	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/reconcile"
	"sync"
)

type Reconciler struct {
	client    client.Client
	gk        schema.GroupKind
	Log       logr.Logger
	cache     *OwnerCache
	setupOnce sync.Once
	objType   reflect.Type
}

func NewReconciler(gk schema.GroupKind) *Reconciler {
	return &Reconciler{
		gk:  gk,
		Log: ctrl.Log.WithName("controllers").WithName("Ownercache").WithName(gk.String()),
	}
}

func (r *Reconciler) setup(ctx context.Context) {
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

func (r *Reconciler) SetupWithCache(ctx context.Context, cache *OwnerCache) error {
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

func (r *Reconciler) Reconcile(ctx context.Context, request reconcile.Request) (reconcile.Result, error) {
	r.setupOnce.Do(func() { r.setup(ctx) })
	obj := reflect.New(r.objType).Interface().(client.Object)
	r.client.Get(ctx, client.ObjectKey{Namespace: request.Namespace, Name: request.Name}, obj)
	r.cache.ReplaceObject(obj)
	return ctrl.Result{}, nil
}
