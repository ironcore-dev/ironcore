// SPDX-FileCopyrightText: 2023 SAP SE or an SAP affiliate company and IronCore contributors
// SPDX-License-Identifier: Apache-2.0

package core

import (
	"context"
	"fmt"
	"time"

	"github.com/ironcore-dev/controller-utils/metautils"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/predicate"
)

const ReplenishResourceQuotaAnnotation = "core.ironcore.dev/replenish-resourcequota"

func HasReplenishResourceQuotaAnnotation(obj client.Object) bool {
	return metautils.HasAnnotation(obj, ReplenishResourceQuotaAnnotation)
}

func SetReplenishResourceQuotaAnnotation(obj client.Object) {
	metautils.SetAnnotation(obj, ReplenishResourceQuotaAnnotation, time.Now().Format(time.RFC3339Nano))
}

func RemoveReplenishResourceQuotaAnnotation(obj client.Object) {
	metautils.DeleteAnnotation(obj, ReplenishResourceQuotaAnnotation)
}

func PatchAddReplenishResourceQuotaAnnotation(ctx context.Context, c client.Client, obj client.Object) (modified bool, err error) {
	if HasReplenishResourceQuotaAnnotation(obj) {
		return false, nil
	}

	base := obj.DeepCopyObject().(client.Object)

	SetReplenishResourceQuotaAnnotation(obj)

	if err := c.Patch(ctx, obj, client.MergeFrom(base)); err != nil {
		return false, fmt.Errorf("error adding replenish resource quota annotation: %w", err)
	}
	return true, nil
}

func PatchRemoveReplenishResourceQuotaAnnotation(ctx context.Context, c client.Client, obj client.Object) (modified bool, err error) {
	if !HasReplenishResourceQuotaAnnotation(obj) {
		return false, nil
	}

	base := obj.DeepCopyObject().(client.Object)

	RemoveReplenishResourceQuotaAnnotation(obj)

	if err := c.Patch(ctx, obj, client.MergeFrom(base)); err != nil {
		return false, fmt.Errorf("error adding replenish resource quota annotation: %w", err)
	}
	return true, nil
}

var HasReplenishResourceQuotaPredicate = predicate.NewPredicateFuncs(func(obj client.Object) bool {
	return HasReplenishResourceQuotaAnnotation(obj)
})
