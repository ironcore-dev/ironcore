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

package tableconverters

import (
	ori "github.com/onmetal/onmetal-api/ori/apis/runtime/v1alpha1"
	"github.com/onmetal/onmetal-api/orictl/table"
	"github.com/onmetal/onmetal-api/orictl/table/tableconverter"
)

var (
	volumeHeaders = []table.Header{
		{Name: "Machine ID"},
		{Name: "Name"},
		{Name: "Device"},
		{Name: "SizeLimitBytes"},
		{Name: "Driver"},
		{Name: "Handle"},
	}
)

var Volume, VolumeSlice = tableconverter.ForType[*ori.Volume]( //nolint:revive
	func() ([]table.Header, error) {
		return volumeHeaders, nil
	},
	func(volume *ori.Volume) ([]table.Row, error) {
		return []table.Row{
			{
				volume.MachineId,
				volume.Name,
				volume.Device,
				volume.GetEmptyDisk().GetSizeLimitBytes(),
				volume.GetAccess().GetDriver(),
				volume.GetAccess().GetHandle(),
			},
		}, nil
	},
)

func init() {
	RegistryBuilder.Register(
		tableconverter.ToTaggedAny(Volume),
		tableconverter.ToTaggedAny(VolumeSlice),
	)
}
