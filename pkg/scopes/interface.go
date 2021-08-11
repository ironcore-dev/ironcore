/*
 * Copyright (c) 2021 by the OnMetal authors.
 *
 * Licensed under the Apache License, Version 2.0 (the "License");
 * you may not use this file except in compliance with the License.
 * You may obtain a copy of the License at
 *
 *     http://www.apache.org/licenses/LICENSE-2.0
 *
 * Unless required by applicable law or agreed to in writing, software
 * distributed under the License is distributed on an "AS IS" BASIS,
 * WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
 * See the License for the specific language governing permissions and
 * limitations under the License.
 */

package scopes

import (
	"context"
	"github.com/onmetal/onmetal-api/apis/common/v1alpha1"
	"github.com/onmetal/onmetal-api/pkg/utils"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apimachinery/pkg/runtime/schema"
	"sigs.k8s.io/controller-runtime/pkg/client"
)

type ScopeEvaluator interface {
	GetNamespace(ctx context.Context, currentNamespace string, scope string) (string, bool, error)
	EvaluateScopedReferenceToObject(ctx context.Context, currentNamespace string, scopeRef *v1alpha1.ScopedReference, obj client.Object) (string, bool, error)
	EvaluateScopedReferenceToObjectId(ctx context.Context, currentNamespace string, gk schema.GroupKind, scopeRef *v1alpha1.ScopedReference) (utils.ObjectId, bool, error)
	EvaluateScopedKindReferenceToObject(ctx context.Context, currentNamespace string, kindRef *v1alpha1.ScopedKindReference) (runtime.Object, string, bool, error)
	EvaluateScopedKindReferenceToObjectId(ctx context.Context, currentNamespace string, kindRef *v1alpha1.ScopedKindReference) (utils.ObjectId, bool, error)
}

func NewScopeEvaluator(client client.Client) ScopeEvaluator {
	return &scopeEvaluator{
		Client: client,
	}
}
