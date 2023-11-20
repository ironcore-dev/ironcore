// SPDX-FileCopyrightText: 2023 SAP SE or an SAP affiliate company and IronCore contributors
// SPDX-License-Identifier: Apache-2.0

package utils

import (
	"context"

	apierrors "k8s.io/apimachinery/pkg/api/errors"
	"k8s.io/apimachinery/pkg/runtime/schema"
	"sigs.k8s.io/controller-runtime/pkg/client"
)

type ObjectPtr[E any] interface {
	*E
	client.Object
}

func ObjectSliceToMapByName[S ~[]E, ObjPtr ObjectPtr[E], E any](objs S) map[string]ObjPtr {
	res := make(map[string]ObjPtr)
	for i := range objs {
		obj := ObjPtr(&objs[i])
		res[obj.GetName()] = obj
	}
	return res
}

func ObjectByNameGetter[Obj client.Object](resource schema.GroupResource, objectByName map[string]Obj) func(name string) (Obj, error) {
	return func(name string) (Obj, error) {
		object, ok := objectByName[name]
		if !ok {
			var zero Obj
			return zero, apierrors.NewNotFound(resource, name)
		}

		return object, nil
	}
}

func ObjectSliceToByNameGetter[S ~[]E, ObjPtr ObjectPtr[E], E any](resource schema.GroupResource, objs S) func(name string) (ObjPtr, error) {
	objectByName := ObjectSliceToMapByName[S, ObjPtr](objs)
	return ObjectByNameGetter(resource, objectByName)
}

func ClientObjectGetter[ObjPtr ObjectPtr[Obj], Obj any](
	ctx context.Context,
	c client.Client,
	namespace string,
) func(name string) (ObjPtr, error) {
	return func(name string) (ObjPtr, error) {
		objPtr := ObjPtr(new(Obj))
		if err := c.Get(ctx, client.ObjectKey{Namespace: namespace, Name: name}, objPtr); err != nil {
			return nil, err
		}
		return objPtr, nil
	}
}
