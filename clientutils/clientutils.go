package clientutils

import (
	"context"
	"fmt"

	apierrors "k8s.io/apimachinery/pkg/api/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/controller/controllerutil"
)

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
		return fmt.Errorf("existing object %s is not controlled by %s",
			key, client.ObjectKeyFromObject(owner))
	}
	return nil
}
