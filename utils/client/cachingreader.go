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

package client

import (
	"context"

	"sigs.k8s.io/controller-runtime/pkg/client"
)

type CachingReaderBuilder struct {
	client client.Client
	funcs  []func(context.Context, *ReaderCache) error
}

func NewCachingReaderBuilder(c client.Client) *CachingReaderBuilder {
	return &CachingReaderBuilder{client: c}
}

func (b *CachingReaderBuilder) add(f func(context.Context, *ReaderCache) error) *CachingReaderBuilder {
	b.funcs = append(b.funcs, f)
	return b
}

func (b *CachingReaderBuilder) List(list client.ObjectList, opts ...client.ListOption) *CachingReaderBuilder {
	return b.add(func(ctx context.Context, cache *ReaderCache) error {
		if err := b.client.List(ctx, list, opts...); err != nil {
			return err
		}
		return cache.InsertList(list)
	})
}

func (b *CachingReaderBuilder) Get(key client.ObjectKey, obj client.Object, opts ...client.GetOption) *CachingReaderBuilder {
	return b.add(func(ctx context.Context, cache *ReaderCache) error {
		if err := b.client.Get(ctx, key, obj, opts...); err != nil {
			return err
		}
		return cache.Insert(obj)
	})
}

func (b *CachingReaderBuilder) Insert(obj client.Object) *CachingReaderBuilder {
	return b.add(func(ctx context.Context, cache *ReaderCache) error {
		return cache.Insert(obj)
	})
}

func (b *CachingReaderBuilder) InsertList(list client.ObjectList) *CachingReaderBuilder {
	return b.add(func(ctx context.Context, cache *ReaderCache) error {
		return cache.InsertList(list)
	})
}

func (b *CachingReaderBuilder) AddToCache(ctx context.Context, cache *ReaderCache) error {
	for _, f := range b.funcs {
		if err := f(ctx, cache); err != nil {
			return err
		}
	}
	return nil
}

func (b *CachingReaderBuilder) BuildCache(ctx context.Context) (*ReaderCache, error) {
	cache := NewReaderCache(b.client.Scheme())
	if err := b.AddToCache(ctx, cache); err != nil {
		return nil, err
	}
	return cache, nil
}

func (b *CachingReaderBuilder) Build(ctx context.Context) (*CachingReader, error) {
	cache, err := b.BuildCache(ctx)
	if err != nil {
		return nil, err
	}
	return NewCachingReader(b.client, cache), nil
}

func NewCachingReader(reader client.Reader, cache *ReaderCache) *CachingReader {
	return &CachingReader{
		reader: reader,
		cache:  cache,
	}
}

type CachingReader struct {
	reader client.Reader
	cache  *ReaderCache
}

func (r *CachingReader) Get(ctx context.Context, key client.ObjectKey, obj client.Object, opts ...client.GetOption) error {
	if err := r.cache.Get(ctx, key, obj, opts...); err == nil {
		return nil
	}

	if err := r.reader.Get(ctx, key, obj, opts...); err != nil {
		return err
	}

	_ = r.cache.Insert(obj)
	return nil
}

func (r *CachingReader) List(ctx context.Context, list client.ObjectList, opts ...client.ListOption) error {
	ok, err := r.cache.CanList(list)
	if err != nil {
		return err
	}
	if ok {
		return r.cache.List(ctx, list, opts...)
	}
	if err := r.reader.List(ctx, list, opts...); err != nil {
		return err
	}
	_ = r.cache.InsertList(list)
	return nil
}
