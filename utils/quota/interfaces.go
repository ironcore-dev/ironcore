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

package quota

import (
	"context"

	corev1alpha1 "github.com/ironcore-dev/ironcore/api/core/v1alpha1"
	"sigs.k8s.io/controller-runtime/pkg/client"
)

type Evaluator interface {
	Type() client.Object
	MatchesResourceName(resourceName corev1alpha1.ResourceName) bool
	MatchesResourceScopeSelectorRequirement(item client.Object, req corev1alpha1.ResourceScopeSelectorRequirement) (bool, error)
	Usage(ctx context.Context, item client.Object) (corev1alpha1.ResourceList, error)
}

type Registry interface {
	// Add to registry
	Add(e Evaluator) error
	// Remove from registry
	Remove(obj client.Object) error
	// Get by group resource
	Get(obj client.Object) (Evaluator, error)
	// List from registry
	List() []Evaluator
}

type UsageCalculator interface {
	CalculateUsage(ctx context.Context, quota *corev1alpha1.ResourceQuota) (corev1alpha1.ResourceList, error)
}
