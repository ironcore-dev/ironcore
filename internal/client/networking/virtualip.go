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

package networking

import (
	"context"

	networkingv1alpha1 "github.com/onmetal/onmetal-api/api/networking/v1alpha1"
	"sigs.k8s.io/controller-runtime/pkg/client"
)

const (
	VirtualIPSpecTargetRefNameField = ".spec.targetRef.name"
)

func SetupVirtualIPSpecTargetRefNameFieldIndexer(ctx context.Context, indexer client.FieldIndexer) error {
	return indexer.IndexField(ctx, &networkingv1alpha1.VirtualIP{}, VirtualIPSpecTargetRefNameField, func(object client.Object) []string {
		virtualIP := object.(*networkingv1alpha1.VirtualIP)
		targetRef := virtualIP.Spec.TargetRef
		if targetRef == nil {
			return []string{""}
		}
		return []string{targetRef.Name}
	})
}
