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

package ipamrange

import (
	"context"
	api "github.com/onmetal/onmetal-api/apis/network/v1alpha1"
	"github.com/onmetal/onmetal-api/pkg/scopes"
	"github.com/onmetal/onmetal-api/pkg/utils"
	"sigs.k8s.io/controller-runtime/pkg/client"
)

type IPAMUsageExtractor struct {
	scopeEvaluator scopes.ScopeEvaluator
}

func (e *IPAMUsageExtractor) ExtractUsage(object client.Object) utils.ObjectIds {
	ipam := object.(*api.IPAMRange)
	id, _, err := e.scopeEvaluator.EvaluateScopedReferenceToObjectId(context.Background(), object.GetNamespace(), api.IPAMRangeGK, ipam.Spec.Parent)
	if err != nil {
		return nil
	}
	return utils.NewObjectIds(id)
}
