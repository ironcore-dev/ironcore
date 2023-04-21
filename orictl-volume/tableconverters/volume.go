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
	"time"

	ori "github.com/onmetal/onmetal-api/ori/apis/volume/v1alpha1"
	"github.com/onmetal/onmetal-api/orictl/api"
	"github.com/onmetal/onmetal-api/orictl/tableconverter"
	"k8s.io/apimachinery/pkg/util/duration"
)

var (
	volumeHeaders = []api.Header{
		{Name: "ID"},
		{Name: "Class"},
		{Name: "Image"},
		{Name: "State"},
		{Name: "Age"},
	}
)

var (
	Volume = tableconverter.Funcs[*ori.Volume]{
		Headers: tableconverter.Headers(volumeHeaders),
		Rows: tableconverter.SingleRowFrom(func(volume *ori.Volume) (api.Row, error) {
			return api.Row{
				volume.Metadata.Id,
				volume.Spec.Class,
				volume.Spec.Image,
				volume.Status.State.String(),
				duration.HumanDuration(time.Since(time.Unix(0, volume.Metadata.CreatedAt))),
			}, nil
		}),
	}
	VolumeSlice = tableconverter.SliceFuncs[*ori.Volume](Volume)
)

func init() {
	RegistryBuilder.Register(
		tableconverter.ToTagAndTypedAny[*ori.Volume](Volume),
		tableconverter.ToTagAndTypedAny[[]*ori.Volume](VolumeSlice),
	)
}
