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

package storage

import (
	"context"

	corev1alpha1 "github.com/onmetal/onmetal-api/api/core/v1alpha1"
	storagev1alpha1 "github.com/onmetal/onmetal-api/api/storage/v1alpha1"
	"github.com/onmetal/onmetal-api/client-go/informers"
	"github.com/onmetal/onmetal-api/client-go/onmetalapi"
	"github.com/onmetal/onmetal-api/internal/quota/evaluator/generic"
	utilsgeneric "github.com/onmetal/onmetal-api/utils/generic"
	"github.com/onmetal/onmetal-api/utils/quota"
	"github.com/onmetal/onmetal-api/utils/quota/resourceaccess"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"sigs.k8s.io/controller-runtime/pkg/client"
)

func NewEvaluators(capabilities generic.CapabilitiesReader) []quota.Evaluator {
	return []quota.Evaluator{
		NewVolumeEvaluator(capabilities),
	}
}

func extractVolumeClassCapabilities(volumeClass *storagev1alpha1.VolumeClass) corev1alpha1.ResourceList {
	return quota.KubernetesResourceListToResourceList(volumeClass.Capabilities)
}

func NewClientVolumeCapabilitiesReader(c client.Client) generic.CapabilitiesReader {
	getter := resourceaccess.NewTypedClientGetter[storagev1alpha1.VolumeClass](c)
	return generic.NewGetterCapabilitiesReader(getter,
		extractVolumeClassCapabilities,
		func(s string) client.ObjectKey { return client.ObjectKey{Name: s} },
	)
}

func NewPrimeLRUVolumeClassCapabilitiesReader(c onmetalapi.Interface, f informers.SharedInformerFactory) generic.CapabilitiesReader {
	getter := resourceaccess.NewPrimeLRUGetter[*storagev1alpha1.VolumeClass, string](
		func(ctx context.Context, className string) (*storagev1alpha1.VolumeClass, error) {
			return c.StorageV1alpha1().VolumeClasses().Get(ctx, className, metav1.GetOptions{})
		},
		func(ctx context.Context, className string) (*storagev1alpha1.VolumeClass, error) {
			return f.Storage().V1alpha1().VolumeClasses().Lister().Get(className)
		},
	)
	return generic.NewGetterCapabilitiesReader(getter, extractVolumeClassCapabilities, utilsgeneric.Identity[string])
}
