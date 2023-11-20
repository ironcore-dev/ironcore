// SPDX-FileCopyrightText: 2023 SAP SE or an SAP affiliate company and IronCore contributors
// SPDX-License-Identifier: Apache-2.0

package resourceaccess

import (
	"context"
	"time"

	ironcoreutilruntime "github.com/ironcore-dev/ironcore/utils/runtime"
	apierrors "k8s.io/apimachinery/pkg/api/errors"
	"k8s.io/utils/lru"
	"sigs.k8s.io/controller-runtime/pkg/client"
)

type liveLookupEntry[T ironcoreutilruntime.DeepCopier[T]] struct {
	expiry time.Time
	item   T
}

type Getter[T ironcoreutilruntime.DeepCopier[T], K any] interface {
	Get(ctx context.Context, key K) (T, error)
}

type liveCachedGetter[T ironcoreutilruntime.DeepCopier[T], K any] struct {
	getLive   func(ctx context.Context, key K) (T, error)
	getCached func(ctx context.Context, key K) (T, error)

	liveLookupCache *lru.Cache
	liveTTL         time.Duration
}

func NewPrimeLRUGetter[T ironcoreutilruntime.DeepCopier[T], K any](
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
	ironcoreutilruntime.DeepCopier[T]
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
