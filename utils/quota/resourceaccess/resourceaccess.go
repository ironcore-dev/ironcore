// Copyright 2023 OnMetal authors
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

package resourceaccess

import (
	"context"
	"time"

	onmetalutilruntime "github.com/onmetal/onmetal-api/utils/runtime"
	apierrors "k8s.io/apimachinery/pkg/api/errors"
	"k8s.io/utils/lru"
	"sigs.k8s.io/controller-runtime/pkg/client"
)

type liveLookupEntry[T onmetalutilruntime.DeepCopier[T]] struct {
	expiry time.Time
	item   T
}

type Getter[T onmetalutilruntime.DeepCopier[T], K any] interface {
	Get(ctx context.Context, key K) (T, error)
}

type liveCachedGetter[T onmetalutilruntime.DeepCopier[T], K any] struct {
	getLive   func(ctx context.Context, key K) (T, error)
	getCached func(ctx context.Context, key K) (T, error)

	liveLookupCache *lru.Cache
	liveTTL         time.Duration
}

func NewPrimeLRUGetter[T onmetalutilruntime.DeepCopier[T], K any](
	getLive func(ctx context.Context, key K) (T, error),
	getCached func(ctx context.Context, key K) (T, error),
) Getter[T, K] {
	return &liveCachedGetter[T, K]{
		getLive:   getLive,
		getCached: getCached,

		liveLookupCache: lru.New(100),
		liveTTL:         30 * time.Second,
	}
}

func (g *liveCachedGetter[T, K]) Get(ctx context.Context, key K) (T, error) {
	obj, err := g.getCached(ctx, key)
	if err != nil && !apierrors.IsNotFound(err) {
		var zero T
		return zero, err
	}

	if apierrors.IsNotFound(err) {
		lruItemObj, ok := g.liveLookupCache.Get(key)
		if !ok || lruItemObj.(liveLookupEntry[T]).expiry.Before(time.Now()) {
			liveObj, err := g.getLive(ctx, key)
			if err != nil {
				var zero T
				return zero, err
			}

			newEntry := liveLookupEntry[T]{
				expiry: time.Now().Add(g.liveTTL),
				item:   liveObj,
			}
			g.liveLookupCache.Add(key, newEntry)
			lruItemObj = newEntry
		}

		lruEntry := lruItemObj.(liveLookupEntry[T])
		obj = lruEntry.item
	}

	// always return a copy to avoid cache mutation.
	return obj.DeepCopy(), nil
}

type Object[T any] interface {
	client.Object
	onmetalutilruntime.DeepCopier[T]
}

type clientGetter[T Object[T]] struct {
	newObject func() T
	client    client.Client
}

func NewClientGetter[T Object[T]](newObject func() T, c client.Client) Getter[T, client.ObjectKey] {
	return &clientGetter[T]{
		newObject: newObject,
		client:    c,
	}
}

func NewTypedClientGetter[T any, TObj interface {
	Object[TObj]
	*T
}](c client.Client) Getter[TObj, client.ObjectKey] {
	return NewClientGetter[TObj](func() TObj {
		return TObj(new(T))
	}, c)
}

func (c *clientGetter[T]) Get(ctx context.Context, key client.ObjectKey) (T, error) {
	obj := c.newObject()
	return obj, c.client.Get(ctx, key, obj)
}
