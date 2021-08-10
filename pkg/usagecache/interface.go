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
	"github.com/onmetal/onmetal-api/pkg/utils"
	"k8s.io/apimachinery/pkg/runtime/schema"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/manager"
)

type Extractor interface {
	ExtractUsage(object client.Object) utils.ObjectIds
}

type ExtractorFunc func(object client.Object) utils.ObjectIds

func (e ExtractorFunc) ExtractUsage(object client.Object) utils.ObjectIds {
	return e(object)
}

type Trigger interface {
	Trigger(id utils.ObjectId)
}

type UsageCache interface {
	Wait()

	RegisterExtractor(ctx context.Context, source schema.GroupKind, relation string, target schema.GroupKind, extractor Extractor) error
	ReplaceObject(object client.Object) (utils.ObjectId, utils.ObjectIds)
	ReplaceObjectUsageInfo(id utils.ObjectId, usages ObjectUsageInfo) utils.ObjectIds

	DeleteObject(id utils.ObjectId) utils.ObjectIds

	GetUsersFor(id utils.ObjectId) utils.ObjectIds
	GetUsersForGK(id utils.ObjectId, gk schema.GroupKind) utils.ObjectIds
	GetUsersForRelationToGK(id utils.ObjectId, relation string, gk schema.GroupKind) utils.ObjectIds
	GetUsersForRelation(id utils.ObjectId, relation string) utils.ObjectIds
	GetUsedObjectsFor(id utils.ObjectId) utils.ObjectIds
	GetUsedObjectsForGK(id utils.ObjectId, gk schema.GroupKind) utils.ObjectIds
	GetUsedObjectsForRelationToGK(id utils.ObjectId, relation string, gk schema.GroupKind) utils.ObjectIds
	GetUsedObjectsForRelation(id utils.ObjectId, relation string) utils.ObjectIds
}

func NewUsageCache(manager manager.Manager, trigger Trigger) *usageCache {
	mgr := &usageCache{
		manager:    manager,
		trigger:    trigger,
		extractors: map[schema.GroupKind]*usageGKInfo{},
		targets:    map[utils.ObjectId]ObjectUsageInfo{},
		sources:    map[utils.ObjectId]ObjectUsageInfo{},
	}
	if mgr.manager != nil {
		mgr.client = manager.GetClient()
	}
	return mgr
}
