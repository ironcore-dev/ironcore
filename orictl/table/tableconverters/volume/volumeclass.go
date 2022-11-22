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

package volume

import (
	ori "github.com/onmetal/onmetal-api/ori/apis/volume/v1alpha1"
	"github.com/onmetal/onmetal-api/orictl/table"
	"github.com/onmetal/onmetal-api/orictl/table/tableconverter"
)

var (
	volumeClassHeaders = []table.Header{
		{Name: "Name"},
		{Name: "TPS"},
		{Name: "IOPS"},
	}
)

var VolumeClass, VolumeClassSlice = tableconverter.ForType[*ori.VolumeClass]( //nolint:revive
	func() ([]table.Header, error) {
		return volumeClassHeaders, nil
	},
	func(class *ori.VolumeClass) ([]table.Row, error) {
		return []table.Row{
			{
				class.Name,
				class.Capabilities.Tps,
				class.Capabilities.Iops,
			},
		}, nil
	},
)

func init() {
	RegistryBuilder.Register(
		tableconverter.ToTaggedAny(VolumeClass),
		tableconverter.ToTaggedAny(VolumeClassSlice),
	)
}
