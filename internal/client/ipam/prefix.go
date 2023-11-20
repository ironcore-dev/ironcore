// SPDX-FileCopyrightText: 2023 SAP SE or an SAP affiliate company and IronCore contributors
// SPDX-License-Identifier: Apache-2.0

package ipam

import (
	"context"

	ipamv1alpha1 "github.com/ironcore-dev/ironcore/api/ipam/v1alpha1"
	"sigs.k8s.io/controller-runtime/pkg/client"
)

const (
	PrefixSpecIPFamilyField      = "spec.ipFamily"
	PrefixSpecParentRefNameField = "spec.parentRef.name"
)

func SetupPrefixSpecIPFamilyFieldIndexer(ctx context.Context, indexer client.FieldIndexer) error {
	return indexer.IndexField(ctx, &ipamv1alpha1.Prefix{}, PrefixSpecIPFamilyField, func(obj client.Object) []string {
		prefix := obj.(*ipamv1alpha1.Prefix)
		return []string{string(prefix.Spec.IPFamily)}
	})
}

func SetupPrefixSpecParentRefFieldIndexer(ctx context.Context, indexer client.FieldIndexer) error {
	return indexer.IndexField(ctx, &ipamv1alpha1.Prefix{}, PrefixSpecParentRefNameField, func(obj client.Object) []string {
		prefix := obj.(*ipamv1alpha1.Prefix)
		parentRef := prefix.Spec.ParentRef
		if parentRef == nil {
			return []string{""}
		}
		return []string{parentRef.Name}
	})
}
