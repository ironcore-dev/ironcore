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
	"fmt"

	"github.com/onmetal/onmetal-api/utils/annotations"
	"sigs.k8s.io/controller-runtime/pkg/client"
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

func ReconcileRequestsFromObjectSlice[S ~[]O, ObjPtr interface {
	client.Object
	*O
}, O any](objs S) []reconcile.Request {
	res := make([]reconcile.Request, len(objs))
	for i := range objs {
		obj := ObjPtr(&objs[i])
		res[i] = reconcile.Request{
			NamespacedName: client.ObjectKeyFromObject(obj),
		}
	}
	return res
}
