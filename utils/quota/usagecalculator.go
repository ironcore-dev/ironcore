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
	"fmt"

	"github.com/ironcore-dev/controller-utils/metautils"
	corev1alpha1 "github.com/ironcore-dev/ironcore/api/core/v1alpha1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/apis/meta/v1/unstructured"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apimachinery/pkg/runtime/schema"
	"k8s.io/apimachinery/pkg/util/sets"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/client/apiutil"
)

type usageCalculator struct {
	client   client.Client
	scheme   *runtime.Scheme
	registry Registry
}

func NewUsageCalculator(c client.Client, scheme *runtime.Scheme, registry Registry) UsageCalculator {
	return &usageCalculator{
		client:   c,
		scheme:   scheme,
		registry: registry,
	}
}

func (c *usageCalculator) getMatchingEvaluators(resourceNames sets.Set[corev1alpha1.ResourceName]) []Evaluator {
	var (
		evaluators         = c.registry.List()
		matchingEvaluators []Evaluator
	)
	for _, evaluator := range evaluators {
		if EvaluatorMatchesResourceNames(evaluator, resourceNames) {
			matchingEvaluators = append(matchingEvaluators, evaluator)
		}
	}
	return matchingEvaluators
}

func (c *usageCalculator) newList(obj client.Object) (schema.GroupVersionKind, client.ObjectList, error) {
	switch obj.(type) {
	case *unstructured.Unstructured:
		gvk := obj.GetObjectKind().GroupVersionKind()
		list := &unstructured.UnstructuredList{}
		list.SetGroupVersionKind(gvk)
		return gvk, list, nil
	case *metav1.PartialObjectMetadata:
		gvk := obj.GetObjectKind().GroupVersionKind()
		list := &metav1.PartialObjectMetadataList{}
		list.SetGroupVersionKind(gvk)
		return gvk, list, nil
	default:
		gvk, err := apiutil.GVKForObject(obj, c.scheme)
		if err != nil {
			return schema.GroupVersionKind{}, nil, fmt.Errorf("error getting gvk for %T: %w", obj, err)
		}
		list, err := metautils.NewListForGVK(c.scheme, gvk)
		if err != nil {
			return schema.GroupVersionKind{}, nil, fmt.Errorf("error creating list for %s: %w", gvk, err)
		}
		return gvk, list, nil
	}
}

func (c *usageCalculator) calculateUsage(
	ctx context.Context,
	evaluators []Evaluator,
	namespace string,
	scopeSelector *corev1alpha1.ResourceScopeSelector,
) (corev1alpha1.ResourceList, error) {
	usage := corev1alpha1.ResourceList{}
	for _, evaluator := range evaluators {
		gvk, list, err := c.newList(evaluator.Type())
		if err != nil {
			return nil, fmt.Errorf("error creating list for type %T: %w", evaluator.Type, err)
		}

		if err := c.client.List(ctx, list, client.InNamespace(namespace)); err != nil {
			return nil, fmt.Errorf("[gvk %s]: error listing objects: %w", gvk, err)
		}

		grUsage := corev1alpha1.ResourceList{}
		if err := metautils.EachListItem(list, func(obj client.Object) error {
			matches, err := EvaluatorMatchesResourceScopeSelector(evaluator, obj, scopeSelector)
			if err != nil {
				return fmt.Errorf("error matching resource scope selector: %w", err)
			}
			if !matches {
				return nil
			}

			itemUsage, err := evaluator.Usage(ctx, obj)
			if err != nil {
				return fmt.Errorf("error computing usage: %w", err)
			}

			grUsage = Add(grUsage, itemUsage)
			return nil
		}); err != nil {
			return nil, fmt.Errorf("[gvk %s]: error iterating list: %w", gvk, err)
		}

		usage = Add(usage, grUsage)
	}
	return usage, nil
}

func (c *usageCalculator) CalculateUsage(
	ctx context.Context,
	resourceQuota *corev1alpha1.ResourceQuota,
) (corev1alpha1.ResourceList, error) {
	evaluators := c.getMatchingEvaluators(ResourceNames(resourceQuota.Spec.Hard))
	return c.calculateUsage(ctx, evaluators, resourceQuota.Namespace, resourceQuota.Spec.ScopeSelector)
}
