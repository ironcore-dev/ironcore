// SPDX-FileCopyrightText: 2023 SAP SE or an SAP affiliate company and IronCore contributors
// SPDX-License-Identifier: Apache-2.0

package common

import (
	"errors"
	"fmt"
	"os"

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

// CleanupSocketIfExists deletes any leftover socket at the given address, if any.
//
// If the file at the given address is no socket, an error is returned.
func CleanupSocketIfExists(address string) error {
	stat, err := os.Stat(address)
	if err != nil {
		if !errors.Is(err, os.ErrNotExist) {
			return fmt.Errorf("error stat-ing socket %q: %w", address, err)
		}
		return nil
	}

	if stat.Mode().Type()&os.ModeSocket == 0 {
		return fmt.Errorf("file at %s is not a socket", address)
	}

	if err := os.Remove(address); err != nil && !errors.Is(err, os.ErrNotExist) {
		return fmt.Errorf("error removing socket: %w", err)
	}
	return nil
}
