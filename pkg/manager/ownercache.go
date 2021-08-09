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
	"sync"

	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime/schema"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/manager"

	"github.com/onmetal/onmetal-api/pkg/utils"
)

type OwnerCache struct {
	manager       manager.Manager
	trigger       ReconcilationTrigger
	client        client.Client
	registrations map[schema.GroupKind]*Reconciler
	lock          sync.RWMutex
	owners        map[utils.ObjectId]utils.ObjectIds
	serfs         map[utils.ObjectId]utils.ObjectIds
	ready         Ready
}

func NewOwnerCache(manager manager.Manager, trig ReconcilationTrigger) *OwnerCache {
	return &OwnerCache{
		manager:       manager,
		trigger:       trig,
		client:        manager.GetClient(),
		registrations: map[schema.GroupKind]*Reconciler{},
		owners:        map[utils.ObjectId]utils.ObjectIds{},
		serfs:         map[utils.ObjectId]utils.ObjectIds{},
	}
}

func (o *OwnerCache) Wait() {
	o.ready.Wait()
}

func (o *OwnerCache) RegisterGroupKind(ctx context.Context, gk schema.GroupKind) error {
	o.lock.Lock()
	defer o.lock.Unlock()

	if o.registrations[gk] != nil {
		return nil
	}
	r := NewReconciler(gk)
	r.SetupWithCache(ctx, o)
	return nil
}

func (o *OwnerCache) ReplaceObject(object client.Object) (utils.ObjectId, utils.ObjectIds) {
	o.lock.Lock()
	defer o.lock.Unlock()
	return o.replaceObject(object)
}

func (o *OwnerCache) DeleteObject(id utils.ObjectId) utils.ObjectIds {
	o.lock.Lock()
	defer o.lock.Unlock()
	old := o.serfs[id]
	if old != nil {
		for owner := range old {
			delete(o.owners[owner], id)
			if len(o.owners[owner]) == 0 {
				delete(o.owners, owner)
			}
		}
		delete(o.serfs, id)
	}
	return old
}

func (o *OwnerCache) replaceObject(object client.Object) (utils.ObjectId, utils.ObjectIds) {
	id := utils.NewObjectId(object)
	oids := utils.GetOwnerIdsFor(object)
	old := o.serfs[id]
	if !oids.Equal(old) {
		for owner := range old {
			if _, ok := oids[owner]; !ok {
				delete(o.owners[owner], id)
				if len(o.owners[owner]) == 0 {
					delete(o.owners, owner)
				}
			}
		}
		for owner := range oids {
			if _, ok := old[owner]; !ok {
				m := o.owners[owner]
				if m == nil {
					m = utils.ObjectIds{}
					o.owners[owner] = m
				}
				m[id] = struct{}{}
			}
		}
		fmt.Printf("owners of: %s:%s\n", id, oids)
		if len(oids) > 0 {
			o.serfs[id] = oids
		} else {
			delete(o.serfs, id)
		}
	}
	return id, oids.Join(old)
}

func (o *OwnerCache) GetOwnersFor(id utils.ObjectId) utils.ObjectIds {
	o.lock.RLock()
	defer o.lock.RUnlock()
	return o.serfs[id].Copy()
}

func (o *OwnerCache) GetOwnersByTypeFor(id utils.ObjectId, gk schema.GroupKind) utils.ObjectIds {
	o.lock.RLock()
	defer o.lock.RUnlock()
	ids := utils.ObjectIds{}
	for i := range o.serfs[id] {
		if i.GroupKind == gk {
			ids.Add(i)
		}
	}
	return ids
}

func (o *OwnerCache) GetSerfsFor(id utils.ObjectId) utils.ObjectIds {
	o.lock.RLock()
	defer o.lock.RUnlock()
	return o.owners[id].Copy()
}

func (o *OwnerCache) GetSerfsWithTypeFor(id utils.ObjectId, gk schema.GroupKind) utils.ObjectIds {
	o.lock.RLock()
	defer o.lock.RUnlock()
	ids := utils.ObjectIds{}
	for i := range o.owners[id] {
		if i.GroupKind == gk {
			ids.Add(i)
		}
	}
	return ids
}

func (o *OwnerCache) GetSerfsForObject(obj client.Object) utils.ObjectIds {
	return o.GetSerfsFor(utils.NewObjectId(obj))
}

func (o *OwnerCache) GetSerfsWithTypeForObject(obj client.Object, gk schema.GroupKind) utils.ObjectIds {
	return o.GetSerfsWithTypeFor(utils.NewObjectId(obj), gk)
}

func (o *OwnerCache) CreateSerf(ctx context.Context, owner, serf client.Object, opts ...client.CreateOption) error {
	serf.SetOwnerReferences([]metav1.OwnerReference{OwnerRefForObject(owner)})
	o.lock.Lock()
	defer o.lock.Unlock()
	if err := o.client.Create(ctx, serf, opts...); err != nil {
		return err
	}
	o.replaceObject(serf)
	return nil
}

func OwnerRefForObject(obj client.Object) metav1.OwnerReference {
	return metav1.OwnerReference{
		APIVersion:         obj.GetObjectKind().GroupVersionKind().GroupVersion().String(),
		Kind:               obj.GetObjectKind().GroupVersionKind().Kind,
		Name:               obj.GetName(),
		UID:                obj.GetUID(),
		Controller:         nil,
		BlockOwnerDeletion: nil,
	}
}
