// Copyright 2023 IronCore authors
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
//      http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.

package generic

import (
	"fmt"

	"github.com/ironcore-dev/ironcore/utils/quota"
	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/client"
)

func NewReplenishReconcilerForRegistryAndType(
	c client.Client,
	registry quota.Registry,
	typ client.Object,
) (*ReplenishReconciler, error) {
	evaluator, err := registry.Get(typ)
	if err != nil {
		return nil, fmt.Errorf("error getting evaluator: %w", err)
	}
	if evaluator == nil {
		return nil, fmt.Errorf("no evaluator for type %T", typ)
	}

	return NewReplenishReconciler(ReplenishReconcilerOptions{
		Client:    c,
		Type:      typ,
		Evaluator: evaluator,
	})
}

type ReplenishReconcilersBuilder []func(c client.Client, registry quota.Registry) ([]*ReplenishReconciler, error)

func (r *ReplenishReconcilersBuilder) Register(objs ...client.Object) *ReplenishReconcilersBuilder {
	return r.Add(func(c client.Client, registry quota.Registry) ([]*ReplenishReconciler, error) {
		var res []*ReplenishReconciler
		for _, typ := range objs {
			reconciler, err := NewReplenishReconcilerForRegistryAndType(c, registry, typ)
			if err != nil {
				return nil, fmt.Errorf("error creating reconciler for %T: %w", typ, err)
			}

			res = append(res, reconciler)
		}
		return res, nil
	})
}

func (r *ReplenishReconcilersBuilder) Add(funcs ...func(c client.Client, registry quota.Registry) ([]*ReplenishReconciler, error)) *ReplenishReconcilersBuilder {
	*r = append(*r, funcs...)
	return r
}

func (r *ReplenishReconcilersBuilder) NewReplenishReconcilers(c client.Client, registry quota.Registry) ([]*ReplenishReconciler, error) {
	var res []*ReplenishReconciler
	for _, f := range *r {
		reconcilers, err := f(c, registry)
		if err != nil {
			return nil, err
		}

		res = append(res, reconcilers...)
	}
	return res, nil
}

func SetupReplenishReconcilersWithManager(mgr ctrl.Manager, reconcilers []*ReplenishReconciler) error {
	for _, reconciler := range reconcilers {
		if err := reconciler.SetupWithManager(mgr); err != nil {
			return err
		}
	}
	return nil
}
