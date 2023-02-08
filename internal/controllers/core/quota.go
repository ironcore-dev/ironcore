// Copyright 2023 OnMetal authors
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

package core

import (
	"context"
	"fmt"
	"time"

	"github.com/onmetal/controller-utils/metautils"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/predicate"
)

const ReplenishResourceQuotaAnnotation = "core.api.onmetal.de/replenish-resourcequota"

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
