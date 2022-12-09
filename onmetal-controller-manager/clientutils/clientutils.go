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

package clientutils

import (
	"context"
	"errors"
	"fmt"

	apierrors "k8s.io/apimachinery/pkg/api/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/controller/controllerutil"
	"sigs.k8s.io/controller-runtime/pkg/reconcile"
)

// ErrNotControlled is returned if the actual object is not controlled by the specified owner.
var ErrNotControlled = errors.New("not controlled")

// ControlledCreateOrGet gets an object if it is controlled by the owner or creates the object with the given owner.
// If the object exists but is not owned, an error is returned.
func ControlledCreateOrGet(ctx context.Context, c client.Client, owner client.Object, obj client.Object, mutate controllerutil.MutateFn) error {
	key := client.ObjectKeyFromObject(obj)
	if err := c.Get(ctx, key, obj); err != nil {
		if !apierrors.IsNotFound(err) {
			return fmt.Errorf("error getting %s: %w", key, err)
		}

		if err := mutate(); err != nil {
			return fmt.Errorf("error transforming: %w", err)
		}
		if err := ctrl.SetControllerReference(owner, obj, c.Scheme()); err != nil {
			return fmt.Errorf("error setting controller reference: %w", err)
		}
		return c.Create(ctx, obj)
	}

	if !metav1.IsControlledBy(obj, owner) {
		return fmt.Errorf("%w: existing object %s is not controlled by %s",
			ErrNotControlled, key, client.ObjectKeyFromObject(owner))
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
