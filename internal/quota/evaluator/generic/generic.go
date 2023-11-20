// SPDX-FileCopyrightText: 2023 SAP SE or an SAP affiliate company and IronCore contributors
// SPDX-License-Identifier: Apache-2.0

package generic

import (
	"context"

	corev1alpha1 "github.com/ironcore-dev/ironcore/api/core/v1alpha1"
	"github.com/ironcore-dev/ironcore/utils/quota"
	"github.com/ironcore-dev/ironcore/utils/quota/resourceaccess"
	ironcoreutilruntime "github.com/ironcore-dev/ironcore/utils/runtime"
	"k8s.io/apimachinery/pkg/api/resource"
	"k8s.io/apimachinery/pkg/runtime/schema"
	"sigs.k8s.io/controller-runtime/pkg/client"
)

type countEvaluator struct {
	gr  schema.GroupResource
	typ client.Object
}

func NewCountEvaluator(gr schema.GroupResource, typ client.Object) quota.Evaluator {
	return &countEvaluator{
		gr:  gr,
		typ: typ,
	}
}

func (e *countEvaluator) countResourceName() corev1alpha1.ResourceName {
	return corev1alpha1.ObjectCountQuotaResourceNameFor(e.gr)
}

func (e *countEvaluator) Type() client.Object {
	return e.typ
}

func (e *countEvaluator) MatchesResourceName(name corev1alpha1.ResourceName) bool {
	return name == e.countResourceName()
}

func (e *countEvaluator) MatchesResourceScopeSelectorRequirement(_ client.Object, _ corev1alpha1.ResourceScopeSelectorRequirement) (bool, error) {
	return false, nil
}

func (e *countEvaluator) Usage(context.Context, client.Object) (corev1alpha1.ResourceList, error) {
	return corev1alpha1.ResourceList{
		e.countResourceName(): *resource.NewQuantity(1, resource.DecimalSI),
	}, nil
}

type CapabilitiesReader interface {
	Get(ctx context.Context, className string) (corev1alpha1.ResourceList, bool)
}

type getterCapabilitiesReader[T ironcoreutilruntime.DeepCopier[T], K any] struct {
	getter              resourceaccess.Getter[T, K]
	extractCapabilities func(T) corev1alpha1.ResourceList
	makeKey             func(string) K
}

func (g *getterCapabilitiesReader[T, K]) Get(ctx context.Context, className string) (corev1alpha1.ResourceList, bool) {
	key := g.makeKey(className)
	obj, err := g.getter.Get(ctx, key)
	if err != nil {
		return nil, false
	}
	return g.extractCapabilities(obj), true
}

func NewGetterCapabilitiesReader[T ironcoreutilruntime.DeepCopier[T], K any](
	getter resourceaccess.Getter[T, K],
	extractCapabilities func(T) corev1alpha1.ResourceList,
	makeKey func(className string) K,
) CapabilitiesReader {
	return &getterCapabilitiesReader[T, K]{
		getter:              getter,
		extractCapabilities: extractCapabilities,
		makeKey:             makeKey,
	}
}
