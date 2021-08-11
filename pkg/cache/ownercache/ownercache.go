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

package ownercache

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

type ownerCache struct {
	manager       manager.Manager
	trigger       Trigger
	client        client.Client
	registrations map[schema.GroupKind]*ownerReconciler
	lock          sync.RWMutex
	owners        map[utils.ObjectId]utils.ObjectIds
	serfs         map[utils.ObjectId]utils.ObjectIds
	ready         *utils.Ready
}

func (o *ownerCache) Wait() {
	o.ready.Wait()
}

func (o *ownerCache) RegisterGroupKind(ctx context.Context, gk schema.GroupKind) error {
	o.lock.Lock()
	defer o.lock.Unlock()

	if o.registrations[gk] != nil {
		return nil
	}
	r := newOwnerReconciler(gk)
	o.registrations[gk] = r
	r.SetupWithCache(ctx, o)
	return nil
}

func (o *ownerCache) ReplaceObject(object client.Object) (utils.ObjectId, utils.ObjectIds) {
	o.lock.Lock()
	defer o.lock.Unlock()
	return o.replaceObject(object)
}

func (o *ownerCache) DeleteObject(id utils.ObjectId) utils.ObjectIds {
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

func (o *ownerCache) replaceObject(object client.Object) (utils.ObjectId, utils.ObjectIds) {
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

func (o *ownerCache) GetOwnersFor(id utils.ObjectId) utils.ObjectIds {
	o.lock.RLock()
	defer o.lock.RUnlock()
	return o.serfs[id].Copy()
}

func (o *ownerCache) GetOwnersByTypeFor(id utils.ObjectId, gk schema.GroupKind) utils.ObjectIds {
	o.lock.RLock()
	defer o.lock.RUnlock()
	ids := utils.ObjectIds{}
	for i := range o.serfs[id] {
		if i.GroupKind == gk {
			ids.Add(i)
		}
	}
	if len(ids) == 0 {
		return nil
	}
	return ids
}

func (o *ownerCache) GetSerfsFor(id utils.ObjectId) utils.ObjectIds {
	o.lock.RLock()
	defer o.lock.RUnlock()
	return o.owners[id].Copy()
}

func (o *ownerCache) GetSerfsWithTypeFor(id utils.ObjectId, gk schema.GroupKind) utils.ObjectIds {
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

func (o *ownerCache) GetSerfsForObject(obj client.Object) utils.ObjectIds {
	return o.GetSerfsFor(utils.NewObjectId(obj))
}

func (o *ownerCache) GetSerfsWithTypeForObject(obj client.Object, gk schema.GroupKind) utils.ObjectIds {
	return o.GetSerfsWithTypeFor(utils.NewObjectId(obj), gk)
}

func (o *ownerCache) CreateSerf(ctx context.Context, owner, serf client.Object, opts ...client.CreateOption) error {
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
