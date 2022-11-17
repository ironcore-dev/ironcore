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

package aliasprefix

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
	"sigs.k8s.io/structured-merge-diff/v4/fieldpath"
)

func GetAttrs(obj runtime.Object) (labels.Set, fields.Set, error) {
	aliasPrefix, ok := obj.(*networking.AliasPrefix)
	if !ok {
		return nil, nil, fmt.Errorf("given object is not a AliasPrefix")
	}
	return aliasPrefix.Labels, SelectableFields(aliasPrefix), nil
}

func MatchAliasPrefix(label labels.Selector, field fields.Selector) apisrvstorage.SelectionPredicate {
	return apisrvstorage.SelectionPredicate{
		Label:    label,
		Field:    field,
		GetAttrs: GetAttrs,
	}
}

func SelectableFields(aliasPrefix *networking.AliasPrefix) fields.Set {
	return generic.ObjectMetaFieldsSet(&aliasPrefix.ObjectMeta, true)
}

type aliasPrefixStrategy struct {
	runtime.ObjectTyper
	names.NameGenerator
}

var Strategy = aliasPrefixStrategy{api.Scheme, names.SimpleNameGenerator}

func (aliasPrefixStrategy) NamespaceScoped() bool {
	return true
}

func (aliasPrefixStrategy) PrepareForCreate(ctx context.Context, obj runtime.Object) {
}

func (aliasPrefixStrategy) PrepareForUpdate(ctx context.Context, obj, old runtime.Object) {
}

func (aliasPrefixStrategy) Validate(ctx context.Context, obj runtime.Object) field.ErrorList {
	aliasPrefix := obj.(*networking.AliasPrefix)
	return validation.ValidateAliasPrefix(aliasPrefix)
}

func (aliasPrefixStrategy) WarningsOnCreate(ctx context.Context, obj runtime.Object) []string {
	return nil
}

func (aliasPrefixStrategy) AllowCreateOnUpdate() bool {
	return false
}

func (aliasPrefixStrategy) AllowUnconditionalUpdate() bool {
	return false
}

func (aliasPrefixStrategy) Canonicalize(obj runtime.Object) {
}

func (aliasPrefixStrategy) ValidateUpdate(ctx context.Context, obj, old runtime.Object) field.ErrorList {
	newAliasPrefix := obj.(*networking.AliasPrefix)
	oldAliasPrefix := old.(*networking.AliasPrefix)
	return validation.ValidateAliasPrefixUpdate(newAliasPrefix, oldAliasPrefix)
}

func (aliasPrefixStrategy) WarningsOnUpdate(ctx context.Context, obj, old runtime.Object) []string {
	return nil
}

type aliasPrefixStatusStrategy struct {
	aliasPrefixStrategy
}

var StatusStrategy = aliasPrefixStatusStrategy{Strategy}

func (aliasPrefixStatusStrategy) GetResetFields() map[fieldpath.APIVersion]*fieldpath.Set {
	return map[fieldpath.APIVersion]*fieldpath.Set{
		"networking.api.onmetal.de/v1alpha1": fieldpath.NewSet(
			fieldpath.MakePathOrDie("spec"),
		),
	}
}

func (aliasPrefixStatusStrategy) PrepareForUpdate(ctx context.Context, obj, old runtime.Object) {
}

func (aliasPrefixStatusStrategy) ValidateUpdate(ctx context.Context, obj, old runtime.Object) field.ErrorList {
	newAliasPrefix := obj.(*networking.AliasPrefix)
	oldAliasPrefix := old.(*networking.AliasPrefix)
	return validation.ValidateAliasPrefixUpdate(newAliasPrefix, oldAliasPrefix)
}

func (aliasPrefixStatusStrategy) WarningsOnUpdate(cxt context.Context, obj, old runtime.Object) []string {
	return nil
}
