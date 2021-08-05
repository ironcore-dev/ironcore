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

package utils

import (
	"context"
	"fmt"
	common "github.com/onmetal/onmetal-api/apis/common/v1alpha1"
	api "github.com/onmetal/onmetal-api/apis/core/v1alpha1"
	"github.com/onmetal/onmetal-api/controllers/core"
	v1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/api/errors"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apimachinery/pkg/runtime/schema"
	"path"
	"regexp"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"strings"
)

type ScopeEvaluator struct {
	client.Client
}

func NewScopeEvaluator(client client.Client) *ScopeEvaluator {
	return &ScopeEvaluator{
		Client: client,
	}
}

var pathSpec *regexp.Regexp = regexp.MustCompile("/?[a-zA-Z0-9]+(/[a-zA-Z0-9]+)*")

// GetNamespace
//  /myaccount/myrootscope/mysubscope
// if error occurs:
// false = temporary error, needs requeue
// true = final error no requeue, propage error to status
func (s *ScopeEvaluator) GetNamespace(ctx context.Context, currentNamespace string, scope string) (string, bool, error) {
	scope = path.Clean(scope)
	if strings.HasPrefix(scope, "../") || scope == ".." {
		return "", true, fmt.Errorf("invalid scope path. no parent path allowed")
	}
	if !pathSpec.MatchString(scope) {
		return "", true, fmt.Errorf("scope name invalid. scope match %s", pathSpec)
	}
	comps := strings.Split(scope, "/")
	var accountNamespace v1.Namespace
	next := 0
	if path.IsAbs(scope) {
		var accountName string
		if comps[1] == api.MyAccountKey {
			if err := s.Get(ctx, client.ObjectKey{Name: currentNamespace}, &accountNamespace); err != nil {
				if errors.IsNotFound(err) {
					return "", true, fmt.Errorf("current namespace %q not found: %s", currentNamespace, err)
				} else {
					return "", false, err
				}
			}
			if accountNamespace.GetLabels() != nil {
				accountName = accountNamespace.GetLabels()[core.AccountLabel]
			}
		} else {
			accountName = comps[1]
		}
		if accountName == "" {
			return "", true, fmt.Errorf("namespace is not a scope")
		}
		var account api.Account
		if err := s.Get(ctx, client.ObjectKey{Name: accountName}, &account); err != nil {
			return "", errors.IsNotFound(err), fmt.Errorf("can not get account %q: %s", accountName, err)
		}
		currentNamespace = account.Status.Namespace
		next = 2
	}
	for next < len(comps) {
		var scope api.Scope
		if err := s.Get(ctx, client.ObjectKey{Name: comps[next], Namespace: currentNamespace}, &scope); err != nil {
			return "", errors.IsNotFound(err), fmt.Errorf("can not get scope %q/%q: %s", currentNamespace, comps[next], err)
		}
		currentNamespace = scope.Status.Namespace
		next++
	}
	return currentNamespace, false, nil
}

func (s *ScopeEvaluator) EvaluateScopedReference(ctx context.Context, currentNamespace string, scopeRef *common.ScopedReference, obj client.Object) (string, bool, error) {
	namespace := currentNamespace
	if scopeRef.Name == "" {
		return "", true, fmt.Errorf("name missing in scope reference")
	}
	if scopeRef.Scope != "" {
		var failed bool
		var err error
		namespace, failed, err = s.GetNamespace(ctx, currentNamespace, scopeRef.Scope)
		if err != nil || obj == nil {
			return namespace, failed, err
		}
	}
	if err := s.Get(ctx, client.ObjectKey{Name: scopeRef.Name, Namespace: namespace}, obj); err != nil {
		return namespace, errors.IsNotFound(err), fmt.Errorf("can not get %s/%s: %s", namespace, scopeRef.Name, err)
	}
	return namespace, false, nil
}

func (s *ScopeEvaluator) EvaluateScopedKindReference(ctx context.Context, currentNamespace string, kindRef *common.ScopedKindReference) (runtime.Object, string, bool, error) {
	namespace := currentNamespace
	if kindRef.Name == "" {
		return nil, "", true, fmt.Errorf("name missing in scoped kind reference")
	}
	if kindRef.Kind == "" {
		return nil, "", true, fmt.Errorf("kind missing in scoped kind reference")
	}
	if kindRef.APIGroup == "" {
		return nil, "", true, fmt.Errorf("apiGroup missing in scoped kind reference")
	}
	if kindRef.Scope != "" {
		var failed bool
		var err error
		namespace, failed, err = s.GetNamespace(ctx, currentNamespace, kindRef.Scope)
		if err != nil {
			return nil, namespace, failed, err
		}
	}
	gk := schema.GroupKind{Group: kindRef.APIGroup, Kind: kindRef.Kind}
	obj := GetObjectForGroupKind(s.Client, gk)
	if obj == nil {
		return nil, "", true, fmt.Errorf("invalid group kind for %s", gk)
	}
	if err := s.Get(ctx, client.ObjectKey{Name: kindRef.Name, Namespace: namespace}, obj); err != nil {
		return nil, namespace, errors.IsNotFound(err), fmt.Errorf("can not get %s/%s: %s", namespace, kindRef.Name, err)
	}
	return obj, namespace, false, nil
}
