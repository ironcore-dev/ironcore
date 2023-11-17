// Copyright 2023 IronCore authors
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
