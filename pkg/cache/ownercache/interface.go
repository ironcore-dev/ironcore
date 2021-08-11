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
	"github.com/onmetal/onmetal-api/pkg/utils"
	"k8s.io/apimachinery/pkg/runtime/schema"
	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/manager"
)

type OwnerCache interface {
	Wait()
	RegisterGroupKind(ctx context.Context, gk schema.GroupKind) error
	ReplaceObject(object client.Object) (utils.ObjectId, utils.ObjectIds)

	DeleteObject(id utils.ObjectId) utils.ObjectIds

	GetOwnersFor(id utils.ObjectId) utils.ObjectIds
	GetOwnersByTypeFor(id utils.ObjectId, gk schema.GroupKind) utils.ObjectIds
	GetSerfsFor(id utils.ObjectId) utils.ObjectIds
	GetSerfsWithTypeFor(id utils.ObjectId, gk schema.GroupKind) utils.ObjectIds
	GetSerfsForObject(obj client.Object) utils.ObjectIds
	GetSerfsWithTypeForObject(obj client.Object, gk schema.GroupKind) utils.ObjectIds

	CreateSerf(ctx context.Context, owner, serf client.Object, opts ...client.CreateOption) error
}

type Trigger interface {
	Trigger(id utils.ObjectId)
}

func NewOwnerCache(manager manager.Manager, trigger Trigger) *ownerCache {
	mgr := &ownerCache{
		manager:       manager,
		trigger:       trigger,
		client:        manager.GetClient(),
		registrations: map[schema.GroupKind]*ownerReconciler{},
		owners:        map[utils.ObjectId]utils.ObjectIds{},
		serfs:         map[utils.ObjectId]utils.ObjectIds{},
		ready:         utils.NewReady(ctrl.Log.WithName("Ownercache"), "setup ownercache"),
	}
	if mgr.manager != nil {
		mgr.client = manager.GetClient()
	}
	return mgr
}
