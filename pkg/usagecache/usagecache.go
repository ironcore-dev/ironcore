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

package usagecache

import (
	"context"
	"fmt"
	"github.com/onmetal/onmetal-api/pkg/trigger"
	"sync"

	"k8s.io/apimachinery/pkg/runtime/schema"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/manager"

	"github.com/onmetal/onmetal-api/pkg/utils"
)

type usageCache struct {
	manager       manager.Manager
	trigger       Trigger
	client        client.Client
	registrations map[schema.GroupKind]*usageReconciler
	extractors    map[schema.GroupKind]*usageGKInfo
	lock          sync.RWMutex
	targets       map[utils.ObjectId]ObjectUsageInfo
	sources       map[utils.ObjectId]ObjectUsageInfo
	ready         trigger.Ready
}

type usageGKInfo struct {
	reconciler *usageReconciler
	relations  map[string]map[schema.GroupKind]Extractor
}

func (u *usageCache) Wait() {
	u.ready.Wait()
}

func (u *usageCache) RegisterExtractor(ctx context.Context, source schema.GroupKind, relation string, target schema.GroupKind, extractor Extractor) error {
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

func (u *usageCache) ReplaceObject(object client.Object) (utils.ObjectId, utils.ObjectIds) {
	id, usages := u.extractUsagesFor(object)
	u.ReplaceObjectUsageInfoForGKs(id, nil, usages)
	return id, u.getUsersFor(id)
}

func (u *usageCache) ReplaceObjectUsageInfo(id utils.ObjectId, usages ObjectUsageInfo) utils.ObjectIds {
	u.lock.Lock()
	defer u.lock.Unlock()
	u.replaceObjectUsageInfoForGKs(id, nil, usages)
	return u.getUsersFor(id)
}

func (u *usageCache) ReplaceObjectUsageInfoForGKs(id utils.ObjectId, gks utils.GroupKinds, usages ObjectUsageInfo) {
	u.lock.Lock()
	defer u.lock.Unlock()
	u.replaceObjectUsageInfoForGKs(id, gks, usages)
}

func (u *usageCache) DeleteObject(id utils.ObjectId) utils.ObjectIds {
	u.lock.Lock()
	defer u.lock.Unlock()
	u.cleanUp(id, u.sources, u.targets)
	return u.getUsersFor(id)
}

func (u *usageCache) cleanUp(id utils.ObjectId, src, dst map[utils.ObjectId]ObjectUsageInfo) utils.ObjectIds {
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

func (u *usageCache) addRelation(targets map[utils.ObjectId]ObjectUsageInfo, id utils.ObjectId, relation string, target utils.ObjectId) {
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

func (u *usageCache) removeRelation(targets map[utils.ObjectId]ObjectUsageInfo, id utils.ObjectId, relation string, target utils.ObjectId) {
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

func (u *usageCache) extractUsagesFor(obj client.Object) (utils.ObjectId, ObjectUsageInfo) {
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

func validForGKs(gks utils.GroupKinds, id utils.ObjectId) bool {
	return gks == nil || gks.Contains(id.GroupKind)
}

func (u *usageCache) replaceObjectUsageInfoForGKs(id utils.ObjectId, gks utils.GroupKinds, usages ObjectUsageInfo) {
	old := u.sources[id]
	for relation, targets := range usages {
		oldRel := old[relation]
		for target := range targets {
			if !validForGKs(gks, target) {
				continue
			}
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
			if !targets.Contains(target) && validForGKs(gks, target) {
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
				if validForGKs(gks, target) {
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
}

func (u *usageCache) GetUsersFor(id utils.ObjectId) utils.ObjectIds {
	u.lock.RLock()
	defer u.lock.RUnlock()
	return u.getUsersFor(id)
}

func (u *usageCache) getUsersFor(id utils.ObjectId) utils.ObjectIds {
	return u.targets[id].GetObjectIds()
}

func (u *usageCache) GetUsersForGK(id utils.ObjectId, gk schema.GroupKind) utils.ObjectIds {
	u.lock.RLock()
	defer u.lock.RUnlock()
	return u.targets[id].GetObjectIdsForGK(gk)
}

func (u *usageCache) GetUsersForRelationToGK(id utils.ObjectId, relation string, gk schema.GroupKind) utils.ObjectIds {
	u.lock.RLock()
	defer u.lock.RUnlock()
	return u.targets[id].GetObjectIdsForRelationToGK(relation, gk)
}

func (u *usageCache) GetUsersForRelation(id utils.ObjectId, relation string) utils.ObjectIds {
	u.lock.RLock()
	defer u.lock.RUnlock()
	return u.targets[id].GetObjectIdsForRelation(relation)
}

func (u *usageCache) GetUsedObjectsFor(id utils.ObjectId) utils.ObjectIds {
	u.lock.RLock()
	defer u.lock.RUnlock()
	return u.sources[id].GetObjectIds()
}

func (u *usageCache) GetUsedObjectsForGK(id utils.ObjectId, gk schema.GroupKind) utils.ObjectIds {
	u.lock.RLock()
	defer u.lock.RUnlock()
	return u.sources[id].GetObjectIdsForGK(gk)
}

func (u *usageCache) GetUsedObjectsForRelationToGK(id utils.ObjectId, relation string, gk schema.GroupKind) utils.ObjectIds {
	u.lock.RLock()
	defer u.lock.RUnlock()
	return u.sources[id].GetObjectIdsForRelationToGK(relation, gk)
}

func (u *usageCache) GetUsedObjectsForRelation(id utils.ObjectId, relation string) utils.ObjectIds {
	u.lock.RLock()
	defer u.lock.RUnlock()
	return u.sources[id].GetObjectIdsForRelation(relation)
}
