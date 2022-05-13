// Copyright 2022 OnMetal authors
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

package console

import (
	"context"
	"fmt"

	"github.com/onmetal/onmetal-api/api"
	"github.com/onmetal/onmetal-api/apis/compute"
	"github.com/onmetal/onmetal-api/apis/compute/validation"
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
	console, ok := obj.(*compute.Console)
	if !ok {
		return nil, nil, fmt.Errorf("given object is not a Console")
	}
	return console.Labels, SelectableFields(console), nil
}

func MatchConsole(label labels.Selector, field fields.Selector) apisrvstorage.SelectionPredicate {
	return apisrvstorage.SelectionPredicate{
		Label:    label,
		Field:    field,
		GetAttrs: GetAttrs,
	}
}

func SelectableFields(console *compute.Console) fields.Set {
	return generic.ObjectMetaFieldsSet(&console.ObjectMeta, true)
}

type consoleStrategy struct {
	runtime.ObjectTyper
	names.NameGenerator
}

var Strategy = consoleStrategy{api.Scheme, names.SimpleNameGenerator}

func (consoleStrategy) NamespaceScoped() bool {
	return true
}

func (consoleStrategy) PrepareForCreate(ctx context.Context, obj runtime.Object) {
}

func (consoleStrategy) PrepareForUpdate(ctx context.Context, obj, old runtime.Object) {
}

func (consoleStrategy) Validate(ctx context.Context, obj runtime.Object) field.ErrorList {
	console := obj.(*compute.Console)
	return validation.ValidateConsole(console)
}

func (consoleStrategy) WarningsOnCreate(ctx context.Context, obj runtime.Object) []string {
	return nil
}

func (consoleStrategy) AllowCreateOnUpdate() bool {
	return false
}

func (consoleStrategy) AllowUnconditionalUpdate() bool {
	return false
}

func (consoleStrategy) Canonicalize(obj runtime.Object) {
}

func (consoleStrategy) ValidateUpdate(ctx context.Context, obj, old runtime.Object) field.ErrorList {
	return field.ErrorList{}
}

func (consoleStrategy) WarningsOnUpdate(ctx context.Context, obj, old runtime.Object) []string {
	return nil
}

type consoleStatusStrategy struct {
	consoleStrategy
}

var StatusStrategy = consoleStatusStrategy{Strategy}

func (consoleStatusStrategy) GetResetFields() map[fieldpath.APIVersion]*fieldpath.Set {
	return map[fieldpath.APIVersion]*fieldpath.Set{
		"compute.api.onmetal.de/v1alpha1": fieldpath.NewSet(
			fieldpath.MakePathOrDie("spec"),
		),
	}
}

func (consoleStatusStrategy) PrepareForUpdate(ctx context.Context, obj, old runtime.Object) {
	newConsole := obj.(*compute.Console)
	oldConsole := old.(*compute.Console)
	newConsole.Spec = oldConsole.Spec
}

func (consoleStatusStrategy) ValidateUpdate(ctx context.Context, obj, old runtime.Object) field.ErrorList {
	newConsole := obj.(*compute.Console)
	oldConsole := old.(*compute.Console)
	return validation.ValidateConsoleUpdate(newConsole, oldConsole)
}

func (consoleStatusStrategy) WarningsOnUpdate(cxt context.Context, obj, old runtime.Object) []string {
	return nil
}
