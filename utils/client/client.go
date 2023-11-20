// SPDX-FileCopyrightText: 2023 SAP SE or an SAP affiliate company and IronCore contributors
// SPDX-License-Identifier: Apache-2.0

package client

import (
	"context"
	"errors"
	"fmt"

	"github.com/ironcore-dev/ironcore/utils/annotations"
	errors2 "k8s.io/apimachinery/pkg/api/errors"
	v1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/types"
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

type Object[O any] interface {
	client.Object
	*O
}

func ReconcileRequestsFromObjectStructSlice[O Object[OStruct], S ~[]OStruct, OStruct any](objs S) []reconcile.Request {
	res := make([]reconcile.Request, len(objs))
	for i := range objs {
		obj := O(&objs[i])
		res[i] = reconcile.Request{
			NamespacedName: client.ObjectKeyFromObject(obj),
		}
	}
	return res
}

func IterateObjectsInObjectStructSlice[O Object[OStruct], S ~[]OStruct, OStruct any](s S, yield func(O) bool) bool {
	for i := range s {
		obj := O(&s[i])
		if !yield(obj) {
			return false
		}
	}
	return true
}

func ForEachObjectInObjectStructSlice[O Object[OStruct], S ~[]OStruct, OStruct any](s S, f func(O)) {
	for i := range s {
		obj := O(&s[i])
		f(obj)
	}
}

func ObjectStructSliceToObjectByUIDMap[O Object[OStruct], S ~[]OStruct, OStruct any](s S) map[types.UID]O {
	m := make(map[types.UID]O)
	for i := range s {
		obj := O(&s[i])
		m[obj.GetUID()] = obj
	}
	return m
}

func ByObjectCreationTimestamp[OP Object[O], O any](obj1, obj2 O) bool {
	t1 := OP(&obj1).GetCreationTimestamp().Time
	t2 := OP(&obj2).GetCreationTimestamp().Time
	return t1.Before(t2)
}

func ObjectSliceOldestObjectIndex[S ~[]O, OP Object[O], O any](objs []O) int {
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
