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

	"k8s.io/apimachinery/pkg/runtime/schema"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/manager"

	"github.com/onmetal/onmetal-api/pkg/utils"
)

type UsageCache struct {
	manager       manager.Manager
	trigger       ReconcilationTrigger
	client        client.Client
	registrations map[schema.GroupKind]*usageReconciler
	extractors    map[schema.GroupKind]*usageGKInfo
	lock          sync.RWMutex
	targets       map[utils.ObjectId]ObjectUsageInfo
	sources       map[utils.ObjectId]ObjectUsageInfo
	ready         Ready
}

type ObjectUsageInfo map[string]utils.ObjectIds

func NewObjectUsageInfo(config ...interface{}) ObjectUsageInfo {
	relation := ""
	info := ObjectUsageInfo{}
	for _, c := range config {
		switch v := c.(type) {
		case string:
			relation = v
		case utils.ObjectId:
			if relation == "" {
				panic("invalid UsageInfo spec")
			}
			old := info[relation]
			if old == nil {
				old = utils.ObjectIds{}
				info[relation] = old
			}
			old.Add(v)
		}
	}
	return info
}

func (i ObjectUsageInfo) Add(relation string, ids utils.ObjectIds) {
	if len(ids) == 0 {
		return
	}
	old := i[relation]
	if old != nil {
		old = utils.ObjectIds{}
		i[relation] = old
	}
	old.AddAll(ids)
}

func (i ObjectUsageInfo) Equal(other ObjectUsageInfo) bool {
	if len(i) == 0 && len(other) == 0 {
		return true
	}
	if len(i) != len(other) {
		return false
	}
	for relation, ids := range other {
		if !ids.Equal(i[relation]) {
			return false
		}
	}
	for relation := range i {
		if _, ok := other[relation]; !ok {
			return false
		}

	}
	return true
}

func (i ObjectUsageInfo) GetObjectIdsForGK(gk schema.GroupKind) utils.ObjectIds {
	if len(i) == 0 {
		return nil
	}
	oids := utils.ObjectIds{}
	for _, ids := range i {
		for id := range ids {
			if id.GroupKind == gk {
				oids.Add(id)
			}
		}
	}
	if len(oids) == 0 {
		return nil
	}
	return oids
}

func (i ObjectUsageInfo) GetObjectIdsForRelationToGK(relation string, gk schema.GroupKind) utils.ObjectIds {
	if len(i) == 0 {
		return nil
	}
	oids := utils.ObjectIds{}
	ids := i[relation]
	if len(ids) == 0 {
		return nil
	}
	for id := range ids {
		if id.GroupKind == gk {
			oids.Add(id)
		}
	}
	return oids
}

func (i ObjectUsageInfo) GetObjectIdsForRelation(relation string) utils.ObjectIds {
	if len(i) == 0 {
		return nil
	}
	return i[relation].Copy()
}

func (i ObjectUsageInfo) GetObjectIds() utils.ObjectIds {
	if len(i) == 0 {
		return nil
	}
	oids := utils.ObjectIds{}
	for _, ids := range i {
		oids.AddAll(ids)
	}
	if len(oids) == 0 {
		return nil
	}
	return oids
}

type usageGKInfo struct {
	reconciler *usageReconciler
	relations  map[string]map[schema.GroupKind]Extractor
}

func NewUsageCache(manager manager.Manager, trig ReconcilationTrigger) *UsageCache {
	mgr := &UsageCache{
		manager:    manager,
		trigger:    trig,
		extractors: map[schema.GroupKind]*usageGKInfo{},
		targets:    map[utils.ObjectId]ObjectUsageInfo{},
		sources:    map[utils.ObjectId]ObjectUsageInfo{},
	}
	if mgr.manager != nil {
		mgr.client = manager.GetClient()
	}
	return mgr
}

func (u *UsageCache) Wait() {
	u.ready.Wait()
}

type Extractor interface {
	ExtractUsage(object client.Object) utils.ObjectIds
}

type ExtractorFunc func(object client.Object) utils.ObjectIds

func (e ExtractorFunc) ExtractUsage(object client.Object) utils.ObjectIds {
	return e(object)
}

func (u *UsageCache) RegisterExtractor(ctx context.Context, source schema.GroupKind, relation string, target schema.GroupKind, extractor Extractor) error {
	u.lock.Lock()
	defer u.lock.Unlock()

	info := u.extractors[source]
	if info == nil {
		info = &usageGKInfo{
			reconciler: newUsageReconciler(source),
			relations:  map[string]map[schema.GroupKind]Extractor{},
		}
		u.extractors[source] = info
		info.reconciler.SetupWithCache(ctx, u)
	}

	targets := info.relations[relation]
	if targets == nil {
		targets = map[schema.GroupKind]Extractor{}
		info.relations[relation] = targets
	}

	if targets[target] != nil {
		return fmt.Errorf("extractor already registered for %s -> %s -> %s", source, relation, target)
	}

	targets[target] = extractor
	return nil
}

func (u *UsageCache) ReplaceObject(object client.Object) (utils.ObjectId, utils.ObjectIds) {
	id, usages := u.extractUsagesFor(object)
	u.lock.Lock()
	defer u.lock.Unlock()
	return id, u.replaceObjectUsageInfo(id, usages)
}

func (u *UsageCache) ReplaceObjectUsageInfo(id utils.ObjectId, usages ObjectUsageInfo) utils.ObjectIds {
	u.lock.Lock()
	defer u.lock.Unlock()
	return u.replaceObjectUsageInfo(id, usages)
}

func (u *UsageCache) DeleteObject(id utils.ObjectId) utils.ObjectIds {
	u.lock.Lock()
	defer u.lock.Unlock()
	u.cleanUp(id, u.sources, u.targets)
	return u.getUsersFor(id)
}

func (u *UsageCache) cleanUp(id utils.ObjectId, src, dst map[utils.ObjectId]ObjectUsageInfo) utils.ObjectIds {
	oids := utils.ObjectIds{}
	old := src[id]
	if old != nil {
		for relation, targets := range old {
			oids.AddAll(targets)
			for target := range targets {
				u.removeRelation(dst, target, relation, id)
			}
		}
		delete(src, id)
	}
	return oids
}

func (u *UsageCache) addRelation(targets map[utils.ObjectId]ObjectUsageInfo, id utils.ObjectId, relation string, target utils.ObjectId) {
	old := targets[id]
	if old == nil {
		old = ObjectUsageInfo{}
		targets[id] = old
	}
	oldRelation := old[relation]
	if oldRelation == nil {
		oldRelation = utils.ObjectIds{}
		old[relation] = oldRelation
	}
	oldRelation.Add(target)
}

func (u *UsageCache) removeRelation(targets map[utils.ObjectId]ObjectUsageInfo, id utils.ObjectId, relation string, target utils.ObjectId) {
	old := targets[id]
	if len(old) == 0 {
		return
	}
	oldRelation := old[relation]
	if len(oldRelation) == 0 {
		return
	}
	delete(oldRelation, target)
	if len(oldRelation) == 0 {
		delete(old, relation)
		if len(old) == 0 {
			delete(targets, id)
		}
	}
}

func (u *UsageCache) extractUsagesFor(obj client.Object) (utils.ObjectId, ObjectUsageInfo) {
	usages := ObjectUsageInfo{}
	oid := utils.NewObjectId(obj)
	gkInfo := u.extractors[oid.GroupKind]
	if gkInfo != nil {
		for relation, targets := range gkInfo.relations {
			for _, extractor := range targets {
				usages.Add(relation, extractor.ExtractUsage(obj))
			}
		}
	}
	return oid, usages
}

func (u *UsageCache) replaceObjectUsageInfo(id utils.ObjectId, usages ObjectUsageInfo) utils.ObjectIds {
	old := u.sources[id]
	if !old.Equal(usages) {
		for relation, targets := range usages {
			oldRel := old[relation]
			for target := range targets {
				if !oldRel.Contains(target) {
					if old == nil {
						old = ObjectUsageInfo{}
						u.sources[id] = old
					}
					if oldRel == nil {
						oldRel = utils.ObjectIds{}
						old[relation] = oldRel
					}
					oldRel.Add(target)
					u.addRelation(u.targets, target, relation, id)
				}
			}
			for target := range oldRel {
				if !targets.Contains(target) {
					delete(oldRel, target)
					if len(oldRel) == 0 {
						delete(old, relation)
						if len(old) == 0 {
							delete(u.sources, id)
						}
					}
					u.removeRelation(u.targets, target, relation, id)
				}
			}
		}
		for relation, targets := range old {
			newRel := usages[relation]
			if len(newRel) == 0 {
				for target := range targets {
					delete(targets, target)
					if len(targets) == 0 {
						delete(old, relation)
						if len(old) == 0 {
							delete(u.sources, id)
						}
					}
					u.removeRelation(u.targets, target, relation, id)
				}
			}
		}
	}
	return u.targets[id].GetObjectIds()
}

func (u *UsageCache) GetUsersFor(id utils.ObjectId) utils.ObjectIds {
	u.lock.RLock()
	defer u.lock.RUnlock()
	return u.getUsersFor(id)
}

func (u *UsageCache) getUsersFor(id utils.ObjectId) utils.ObjectIds {
	return u.targets[id].GetObjectIds()
}

func (u *UsageCache) GetUsersForGK(id utils.ObjectId, gk schema.GroupKind) utils.ObjectIds {
	u.lock.RLock()
	defer u.lock.RUnlock()
	return u.targets[id].GetObjectIdsForGK(gk)
}

func (u *UsageCache) GetUsersForRelationToGK(id utils.ObjectId, relation string, gk schema.GroupKind) utils.ObjectIds {
	u.lock.RLock()
	defer u.lock.RUnlock()
	return u.targets[id].GetObjectIdsForRelationToGK(relation, gk)
}

func (u *UsageCache) GetUsersForRelation(id utils.ObjectId, relation string) utils.ObjectIds {
	u.lock.RLock()
	defer u.lock.RUnlock()
	return u.targets[id].GetObjectIdsForRelation(relation)
}

func (u *UsageCache) GetUsedObjectsFor(id utils.ObjectId) utils.ObjectIds {
	u.lock.RLock()
	defer u.lock.RUnlock()
	return u.sources[id].GetObjectIds()
}

func (u *UsageCache) GetUsedObjectsForGK(id utils.ObjectId, gk schema.GroupKind) utils.ObjectIds {
	u.lock.RLock()
	defer u.lock.RUnlock()
	return u.sources[id].GetObjectIdsForGK(gk)
}

func (u *UsageCache) GetUsedObjectsForRelationToGK(id utils.ObjectId, relation string, gk schema.GroupKind) utils.ObjectIds {
	u.lock.RLock()
	defer u.lock.RUnlock()
	return u.sources[id].GetObjectIdsForRelationToGK(relation, gk)
}

func (u *UsageCache) GetUsedObjectsForRelation(id utils.ObjectId, relation string) utils.ObjectIds {
	u.lock.RLock()
	defer u.lock.RUnlock()
	return u.sources[id].GetObjectIdsForRelation(relation)
}
