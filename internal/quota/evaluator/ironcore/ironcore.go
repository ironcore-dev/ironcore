// SPDX-FileCopyrightText: 2023 SAP SE or an SAP affiliate company and IronCore contributors
// SPDX-License-Identifier: Apache-2.0

package ironcore

import (
	informers "github.com/ironcore-dev/ironcore/client-go/informers/externalversions"
	ironcore "github.com/ironcore-dev/ironcore/client-go/ironcore/versioned"
	"github.com/ironcore-dev/ironcore/internal/quota/evaluator/compute"
	"github.com/ironcore-dev/ironcore/internal/quota/evaluator/generic"
	"github.com/ironcore-dev/ironcore/internal/quota/evaluator/storage"
	"github.com/ironcore-dev/ironcore/utils/quota"
	"sigs.k8s.io/controller-runtime/pkg/client"
)

func NewEvaluators(
	machineClassCapabilities,
	volumeClassCapabilities,
	bucketClassCapabilities generic.CapabilitiesReader,
) []quota.Evaluator {
	var evaluators []quota.Evaluator

	evaluators = append(evaluators, compute.NewEvaluators(machineClassCapabilities)...)
	evaluators = append(evaluators, storage.NewEvaluators(volumeClassCapabilities, bucketClassCapabilities)...)

	return evaluators
}

func NewEvaluatorsForAdmission(c ironcore.Interface, f informers.SharedInformerFactory) []quota.Evaluator {
	machineClassCapabilities := compute.NewPrimeLRUMachineClassCapabilitiesReader(c, f)
	volumeClassCapabilities := storage.NewPrimeLRUVolumeClassCapabilitiesReader(c, f)
	bucketClassCapabilities := storage.NewPrimeLRUBucketClassCapabilitiesReader(c, f)
	return NewEvaluators(machineClassCapabilities, volumeClassCapabilities, bucketClassCapabilities)
}

func NewEvaluatorsForControllers(c client.Client) []quota.Evaluator {
	machineClassCapabilities := compute.NewClientMachineCapabilitiesReader(c)
	volumeClassCapabilities := storage.NewClientVolumeCapabilitiesReader(c)
	bucketClassCapabilities := storage.NewClientBucketCapabilitiesReader(c)
	return NewEvaluators(machineClassCapabilities, volumeClassCapabilities, bucketClassCapabilities)
}
