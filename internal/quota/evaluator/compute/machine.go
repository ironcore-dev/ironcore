// Copyright 2023 OnMetal authors
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

package compute

import (
	"context"
	"fmt"

	computev1alpha1 "github.com/onmetal/onmetal-api/api/compute/v1alpha1"
	corev1alpha1 "github.com/onmetal/onmetal-api/api/core/v1alpha1"
	"github.com/onmetal/onmetal-api/internal/apis/compute"
	internalcomputev1alpha1 "github.com/onmetal/onmetal-api/internal/apis/compute/v1alpha1"
	"github.com/onmetal/onmetal-api/internal/quota/evaluator/generic"
	"github.com/onmetal/onmetal-api/utils/quota"
	"golang.org/x/exp/slices"
	apierrors "k8s.io/apimachinery/pkg/api/errors"
	"k8s.io/apimachinery/pkg/api/resource"
	"k8s.io/apimachinery/pkg/util/sets"
	"sigs.k8s.io/controller-runtime/pkg/client"
)

var (
	machineResource          = computev1alpha1.Resource("machines")
	machineCountResourceName = corev1alpha1.ObjectCountQuotaResourceNameFor(machineResource)

	MachineResourceNames = sets.New(
		machineCountResourceName,
		corev1alpha1.ResourceRequestsCPU,
		corev1alpha1.ResourceRequestsMemory,
	)
)

type machineEvaluator struct {
	capabilities generic.CapabilitiesReader
}

func NewMachineEvaluator(capabilities generic.CapabilitiesReader) quota.Evaluator {
	return &machineEvaluator{
		capabilities: capabilities,
	}
}

func (m *machineEvaluator) Type() client.Object {
	return &computev1alpha1.Machine{}
}

func (m *machineEvaluator) MatchesResourceName(name corev1alpha1.ResourceName) bool {
	return MachineResourceNames.Has(name)
}

func (m *machineEvaluator) MatchesResourceScopeSelectorRequirement(item client.Object, req corev1alpha1.ResourceScopeSelectorRequirement) (bool, error) {
	machine := item.(*computev1alpha1.Machine)

	switch req.ScopeName {
	case corev1alpha1.ResourceScopeMachineClass:
		return machineMatchesMachineClassScope(machine, req.Operator, req.Values), nil
	default:
		return false, nil
	}
}

func machineMatchesMachineClassScope(machine *computev1alpha1.Machine, op corev1alpha1.ResourceScopeSelectorOperator, values []string) bool {
	machineClassName := machine.Spec.MachineClassRef.Name

	switch op {
	case corev1alpha1.ResourceScopeSelectorOperatorExists:
		return true
	case corev1alpha1.ResourceScopeSelectorOperatorDoesNotExist:
		return false
	case corev1alpha1.ResourceScopeSelectorOperatorIn:
		return slices.Contains(values, machineClassName)
	case corev1alpha1.ResourceScopeSelectorOperatorNotIn:
		return !slices.Contains(values, machineClassName)
	default:
		return false
	}
}

func toExternalMachineOrError(obj client.Object) (*computev1alpha1.Machine, error) {
	switch t := obj.(type) {
	case *computev1alpha1.Machine:
		return t, nil
	case *compute.Machine:
		machine := &computev1alpha1.Machine{}
		if err := internalcomputev1alpha1.Convert_compute_Machine_To_v1alpha1_Machine(t, machine, nil); err != nil {
			return nil, err
		}
		return machine, nil
	default:
		return nil, fmt.Errorf("expect *compute.Machine or *computev1alpha1.Machine but got %v", t)
	}
}

func (m *machineEvaluator) Usage(ctx context.Context, item client.Object) (corev1alpha1.ResourceList, error) {
	machine, err := toExternalMachineOrError(item)
	if err != nil {
		return nil, err
	}

	machineClassName := machine.Spec.MachineClassRef.Name

	capabilities, ok := m.capabilities.Get(ctx, machineClassName)
	if !ok {
		return nil, apierrors.NewBadRequest(fmt.Sprintf("machine class %q not found", machineClassName))
	}

	return corev1alpha1.ResourceList{
		machineCountResourceName:            resource.MustParse("1"),
		corev1alpha1.ResourceRequestsCPU:    *capabilities.CPU(),
		corev1alpha1.ResourceRequestsMemory: *capabilities.Memory(),
	}, nil
}
