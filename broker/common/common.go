// Copyright 2022 OnMetal authors
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

package common

import (
	"fmt"

	apierrors "k8s.io/apimachinery/pkg/api/errors"
	"k8s.io/apimachinery/pkg/runtime/schema"
	"sigs.k8s.io/controller-runtime/pkg/client"
)

type ObjectPtr[E any] interface {
	client.Object
	*E
}

type ObjectKey interface {
	comparable
}

type ObjectGetter[O ObjectPtr[E], E any, K ObjectKey] struct {
	gr           schema.GroupResource
	keyFunc      func(O) K
	objectsByKey map[K]O
}

type ObjectGetterOptions[O ObjectPtr[E], E any, K ObjectKey] struct {
	objects []O
}

func (o *ObjectGetterOptions[O, E, K]) ApplyOptions(opts []ObjectGetterOption[O, E, K]) {
	for _, opt := range opts {
		opt.ApplyToObjectGetter(o)
	}
}

func ByObjectName[O ObjectPtr[E], E any]() func(O) string {
	return func(obj O) string {
		return obj.GetName()
	}
}

type objectSlice[K ObjectKey, O ObjectPtr[E], E any] []E

func (s objectSlice[K, O, E]) ApplyToObjectGetter(o *ObjectGetterOptions[O, E, K]) {
	for i := range s {
		objPtr := O(&s[i])
		o.objects = append(o.objects, objPtr)
	}
}

func ObjectSlice[K ObjectKey, O ObjectPtr[E], E any](objs []E) ObjectGetterOption[O, E, K] {
	return objectSlice[K, O, E](objs)
}

type ObjectGetterOption[O ObjectPtr[E], E any, K ObjectKey] interface {
	ApplyToObjectGetter(o *ObjectGetterOptions[O, E, K])
}

func NewObjectGetter[K ObjectKey, O ObjectPtr[E], E any](gr schema.GroupResource, keyFunc func(O) K, opts ...ObjectGetterOption[O, E, K]) (*ObjectGetter[O, E, K], error) {
	o := &ObjectGetterOptions[O, E, K]{}
	o.ApplyOptions(opts)

	objectsByKey := make(map[K]O)
	for _, obj := range o.objects {
		key := keyFunc(obj)
		if _, ok := objectsByKey[key]; ok {
			return nil, fmt.Errorf("duplicate key %v", key)
		}

		objectsByKey[key] = obj
	}
	return &ObjectGetter[O, E, K]{
		gr:           gr,
		keyFunc:      keyFunc,
		objectsByKey: objectsByKey,
	}, nil
}

func (r *ObjectGetter[O, E, K]) Get(key K) (O, error) {
	found, ok := r.objectsByKey[key]
	if !ok {
		return nil, apierrors.NewNotFound(r.gr, fmt.Sprintf("%v", key))
	}

	return found, nil
}
