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

package natgatewayrouting

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
	natGatewayRouting, ok := obj.(*networking.NATGatewayRouting)
	if !ok {
		return nil, nil, fmt.Errorf("given object is not a NATGatewayRouting")
	}
	return natGatewayRouting.Labels, SelectableFields(natGatewayRouting), nil
}

func MatchNATGatewayRouting(label labels.Selector, field fields.Selector) apisrvstorage.SelectionPredicate {
	return apisrvstorage.SelectionPredicate{
		Label:    label,
		Field:    field,
		GetAttrs: GetAttrs,
	}
}

func SelectableFields(natGatewayRouting *networking.NATGatewayRouting) fields.Set {
	return generic.ObjectMetaFieldsSet(&natGatewayRouting.ObjectMeta, true)
}

type natGatewayRoutingStrategy struct {
	runtime.ObjectTyper
	names.NameGenerator
}

var Strategy = natGatewayRoutingStrategy{api.Scheme, names.SimpleNameGenerator}

func (natGatewayRoutingStrategy) NamespaceScoped() bool {
	return true
}

func (natGatewayRoutingStrategy) PrepareForCreate(ctx context.Context, obj runtime.Object) {
}

func (natGatewayRoutingStrategy) PrepareForUpdate(ctx context.Context, obj, old runtime.Object) {
}

func (natGatewayRoutingStrategy) Validate(ctx context.Context, obj runtime.Object) field.ErrorList {
	natGatewayRouting := obj.(*networking.NATGatewayRouting)
	return validation.ValidateNATGatewayRouting(natGatewayRouting)
}

func (natGatewayRoutingStrategy) WarningsOnCreate(ctx context.Context, obj runtime.Object) []string {
	return nil
}

func (natGatewayRoutingStrategy) AllowCreateOnUpdate() bool {
	return false
}

func (natGatewayRoutingStrategy) AllowUnconditionalUpdate() bool {
	return false
}

func (natGatewayRoutingStrategy) Canonicalize(obj runtime.Object) {
}

func (natGatewayRoutingStrategy) ValidateUpdate(ctx context.Context, obj, old runtime.Object) field.ErrorList {
	newNATGatewayRouting := obj.(*networking.NATGatewayRouting)
	oldNATGatewayRouting := old.(*networking.NATGatewayRouting)
	return validation.ValidateNATGatewayRoutingUpdate(newNATGatewayRouting, oldNATGatewayRouting)
}

func (natGatewayRoutingStrategy) WarningsOnUpdate(ctx context.Context, obj, old runtime.Object) []string {
	return nil
}
