/*
 * Copyright (c) 2022 by the OnMetal authors.
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

package aliasprefixrouting

import (
	"context"
	"fmt"

	"github.com/onmetal/onmetal-api/onmetal-apiserver/internal/api"
	"github.com/onmetal/onmetal-api/onmetal-apiserver/internal/apis/networking"
	"github.com/onmetal/onmetal-api/onmetal-apiserver/internal/apis/networking/validation"
	"k8s.io/apimachinery/pkg/fields"
	"k8s.io/apimachinery/pkg/labels"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apimachinery/pkg/util/validation/field"
	"k8s.io/apiserver/pkg/registry/generic"
	apisrvstorage "k8s.io/apiserver/pkg/storage"
	"k8s.io/apiserver/pkg/storage/names"
)

func GetAttrs(obj runtime.Object) (labels.Set, fields.Set, error) {
	aliasPrefixRouting, ok := obj.(*networking.AliasPrefixRouting)
	if !ok {
		return nil, nil, fmt.Errorf("given object is not a AliasPrefixRouting")
	}
	return aliasPrefixRouting.Labels, SelectableFields(aliasPrefixRouting), nil
}

func MatchAliasPrefixRouting(label labels.Selector, field fields.Selector) apisrvstorage.SelectionPredicate {
	return apisrvstorage.SelectionPredicate{
		Label:    label,
		Field:    field,
		GetAttrs: GetAttrs,
	}
}

func SelectableFields(aliasPrefixRouting *networking.AliasPrefixRouting) fields.Set {
	return generic.ObjectMetaFieldsSet(&aliasPrefixRouting.ObjectMeta, true)
}

type aliasPrefixRoutingStrategy struct {
	runtime.ObjectTyper
	names.NameGenerator
}

var Strategy = aliasPrefixRoutingStrategy{api.Scheme, names.SimpleNameGenerator}

func (aliasPrefixRoutingStrategy) NamespaceScoped() bool {
	return true
}

func (aliasPrefixRoutingStrategy) PrepareForCreate(ctx context.Context, obj runtime.Object) {
}

func (aliasPrefixRoutingStrategy) PrepareForUpdate(ctx context.Context, obj, old runtime.Object) {
}

func (aliasPrefixRoutingStrategy) Validate(ctx context.Context, obj runtime.Object) field.ErrorList {
	aliasPrefixRouting := obj.(*networking.AliasPrefixRouting)
	return validation.ValidateAliasPrefixRouting(aliasPrefixRouting)
}

func (aliasPrefixRoutingStrategy) WarningsOnCreate(ctx context.Context, obj runtime.Object) []string {
	return nil
}

func (aliasPrefixRoutingStrategy) AllowCreateOnUpdate() bool {
	return false
}

func (aliasPrefixRoutingStrategy) AllowUnconditionalUpdate() bool {
	return false
}

func (aliasPrefixRoutingStrategy) Canonicalize(obj runtime.Object) {
}

func (aliasPrefixRoutingStrategy) ValidateUpdate(ctx context.Context, obj, old runtime.Object) field.ErrorList {
	newAliasPrefixRouting := obj.(*networking.AliasPrefixRouting)
	oldAliasPrefixRouting := old.(*networking.AliasPrefixRouting)
	return validation.ValidateAliasPrefixRoutingUpdate(newAliasPrefixRouting, oldAliasPrefixRouting)
}

func (aliasPrefixRoutingStrategy) WarningsOnUpdate(ctx context.Context, obj, old runtime.Object) []string {
	return nil
}
