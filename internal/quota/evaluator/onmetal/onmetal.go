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

package onmetal

import (
	"github.com/onmetal/onmetal-api/client-go/informers"
	"github.com/onmetal/onmetal-api/client-go/onmetalapi"
	"github.com/onmetal/onmetal-api/internal/quota/evaluator/compute"
	"github.com/onmetal/onmetal-api/internal/quota/evaluator/generic"
	"github.com/onmetal/onmetal-api/internal/quota/evaluator/storage"
	"github.com/onmetal/onmetal-api/utils/quota"
	"sigs.k8s.io/controller-runtime/pkg/client"
)

func NewEvaluators(
	machineClassCapabilities,
	volumeClassCapabilities generic.CapabilitiesReader,
) []quota.Evaluator {
	var evaluators []quota.Evaluator

	evaluators = append(evaluators, compute.NewEvaluators(machineClassCapabilities)...)
	evaluators = append(evaluators, storage.NewEvaluators(volumeClassCapabilities)...)

	return evaluators
}

func NewEvaluatorsForAdmission(c onmetalapi.Interface, f informers.SharedInformerFactory) []quota.Evaluator {
	machineClassCapabilities := compute.NewPrimeLRUMachineClassCapabilitiesReader(c, f)
	volumeClassCapabilities := storage.NewPrimeLRUVolumeClassCapabilitiesReader(c, f)
	return NewEvaluators(machineClassCapabilities, volumeClassCapabilities)
}

func NewEvaluatorsForControllers(c client.Client) []quota.Evaluator {
	machineClassCapabilities := compute.NewClientMachineCapabilitiesReader(c)
	volumeClassCapabilities := storage.NewClientVolumeCapabilitiesReader(c)
	return NewEvaluators(machineClassCapabilities, volumeClassCapabilities)
}
