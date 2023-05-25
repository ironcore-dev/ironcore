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

package client

import (
	"context"
	"errors"
	"fmt"

	"github.com/onmetal/onmetal-api/utils/annotations"
	errors2 "k8s.io/apimachinery/pkg/api/errors"
	v1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	controllerruntime "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/controller/controllerutil"
	"sigs.k8s.io/controller-runtime/pkg/reconcile"
)

func PatchEnsureNoReconcileAnnotation(ctx context.Context, c client.Client, obj client.Object) (modified bool, err error) {
	if !annotations.HasReconcileAnnotation(obj) {
		return false, nil
	}

	if err := PatchRemoveReconcileAnnotation(ctx, c, obj); err != nil {
		return false, err
	}
	return true, nil
}

func PatchAddReconcileAnnotation(ctx context.Context, c client.Client, obj client.Object) error {
	base := obj.DeepCopyObject().(client.Object)

	annotations.SetReconcileAnnotation(obj)

	if err := c.Patch(ctx, obj, client.MergeFrom(base)); err != nil {
		return fmt.Errorf("error adding reconcile annotation: %w", err)
	}
	return nil
}

func PatchRemoveReconcileAnnotation(ctx context.Context, c client.Client, obj client.Object) error {
	base := obj.DeepCopyObject().(client.Object)

	annotations.RemoveReconcileAnnotation(obj)

	if err := c.Patch(ctx, obj, client.MergeFrom(base)); err != nil {
		return fmt.Errorf("error removing reconcile annotation: %w", err)
	}
	return nil
}

type ObjectPointer[O any] interface {
	client.Object
	*O
}

func ReconcileRequestsFromObjectSlice[S ~[]O, OP ObjectPointer[O], O any](objs S) []reconcile.Request {
	res := make([]reconcile.Request, len(objs))
	for i := range objs {
		obj := OP(&objs[i])
		res[i] = reconcile.Request{
			NamespacedName: client.ObjectKeyFromObject(obj),
		}
	}
	return res
}

func ByObjectCreationTimestamp[OP ObjectPointer[O], O any](obj1, obj2 O) bool {
	t1 := OP(&obj1).GetCreationTimestamp().Time
	t2 := OP(&obj2).GetCreationTimestamp().Time
	return t1.Before(t2)
}

func ObjectSliceOldestObjectIndex[S ~[]O, OP ObjectPointer[O], O any](objs []O) int {
	idx := -1
	for i := range objs {
		obj := OP(&objs[i])
		if idx < 0 || obj.GetCreationTimestamp().Time.Before(OP(&objs[idx]).GetCreationTimestamp().Time) {
			idx = i
		}
	}
	return idx
}

// ErrNotControlled is returned if the actual object is not controlled by the specified owner.
var ErrNotControlled = errors.New("not controlled")

// ControlledCreateOrGet gets an object if it is controlled by the owner or creates the object with the given owner.
// If the object exists but is not owned, an error is returned.
func ControlledCreateOrGet(ctx context.Context, c client.Client, owner client.Object, obj client.Object, mutate controllerutil.MutateFn) error {
	key := client.ObjectKeyFromObject(obj)
	if err := c.Get(ctx, key, obj); err != nil {
		if !errors2.IsNotFound(err) {
			return fmt.Errorf("error getting %s: %w", key, err)
		}

		if err := mutate(); err != nil {
			return fmt.Errorf("error transforming: %w", err)
		}
		if err := controllerruntime.SetControllerReference(owner, obj, c.Scheme()); err != nil {
			return fmt.Errorf("error setting controller reference: %w", err)
		}
		return c.Create(ctx, obj)
	}

	if !v1.IsControlledBy(obj, owner) {
		return fmt.Errorf("%w: existing object %s is not controlled by %s",
			ErrNotControlled, key, client.ObjectKeyFromObject(owner))
	}
	return nil
}
