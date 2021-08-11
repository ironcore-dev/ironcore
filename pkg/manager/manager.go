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
	"github.com/onmetal/onmetal-api/pkg/cache/ownercache"
	"github.com/onmetal/onmetal-api/pkg/cache/usagecache"
	"github.com/onmetal/onmetal-api/pkg/scopes"
	"github.com/onmetal/onmetal-api/pkg/trigger"
	"k8s.io/apimachinery/pkg/runtime/schema"
	"k8s.io/client-go/rest"
	"sigs.k8s.io/controller-runtime/pkg/controller"
	"sigs.k8s.io/controller-runtime/pkg/manager"
)

type Manager struct {
	manager.Manager
	triggers       trigger.ReconcilationTrigger
	ownerCache     ownercache.OwnerCache
	usageCache     usagecache.UsageCache
	scopeEvaluator scopes.ScopeEvaluator
}

func NewManager(config *rest.Config, options manager.Options) (*Manager, error) {
	mgr, err := manager.New(config, options)
	if err != nil {
		return nil, err
	}
	trig := trigger.NewReconcilationTrigger()
	oc := ownercache.NewOwnerCache(mgr, trig)
	uc := usagecache.NewUsageCache(mgr, trig)
	se := scopes.NewScopeEvaluator(mgr.GetClient())
	return &Manager{
		Manager:        mgr,
		ownerCache:     oc,
		usageCache:     uc,
		scopeEvaluator: se,
		triggers:       trig,
	}, nil
}

func (m *Manager) GetOwnerCache() ownercache.OwnerCache {
	return m.ownerCache
}

func (m *Manager) GetUsageCache() usagecache.UsageCache {
	return m.usageCache
}

func (m *Manager) GetScopeEvaluator() scopes.ScopeEvaluator {
	return m.scopeEvaluator
}

func (m *Manager) RegisterControllerFor(gk schema.GroupKind, controller controller.Controller) {
	m.triggers.RegisterControllerFor(gk, controller)
}

func (m *Manager) Wait() {
	m.usageCache.Wait()
	m.ownerCache.Wait()
}
