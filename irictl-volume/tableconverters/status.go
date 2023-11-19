// Copyright 2022 IronCore authors
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

package tableconverters

import (
	iri "github.com/ironcore-dev/ironcore/iri/apis/volume/v1alpha1"
	"github.com/ironcore-dev/ironcore/irictl/api"
	"github.com/ironcore-dev/ironcore/irictl/tableconverter"
	"k8s.io/apimachinery/pkg/api/resource"
)

var (
	volumeClassHeaders = []api.Header{
		{Name: "Name"},
		{Name: "TPS"},
		{Name: "IOPS"},
		{Name: "Quantity"},
	}
)

var (
	VolumeClassStatus = tableconverter.Funcs[*iri.VolumeClassStatus]{
		Headers: tableconverter.Headers(volumeClassHeaders),
		Rows: tableconverter.SingleRowFrom(func(status *iri.VolumeClassStatus) (api.Row, error) {
			return api.Row{
				status.VolumeClass.Name,
				resource.NewQuantity(status.VolumeClass.Capabilities.Tps, resource.BinarySI).String(),
				resource.NewQuantity(status.VolumeClass.Capabilities.Iops, resource.DecimalSI).String(),
				resource.NewQuantity(status.Quantity, resource.BinarySI).String(),
			}, nil
		}),
	}
	VolumeClassStatusSlice = tableconverter.SliceFuncs[*iri.VolumeClassStatus](VolumeClassStatus)
)

func init() {
	RegistryBuilder.Register(
		tableconverter.ToTagAndTypedAny[*iri.VolumeClassStatus](VolumeClassStatus),
		tableconverter.ToTagAndTypedAny[[]*iri.VolumeClassStatus](VolumeClassStatusSlice),
	)
}
