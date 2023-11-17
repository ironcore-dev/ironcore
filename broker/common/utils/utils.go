// Copyright 2022 IronCore authors
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
