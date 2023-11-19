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
	"time"

	iri "github.com/ironcore-dev/ironcore/iri/apis/machine/v1alpha1"
	"github.com/ironcore-dev/ironcore/irictl/api"
	"github.com/ironcore-dev/ironcore/irictl/tableconverter"
	"k8s.io/apimachinery/pkg/util/duration"
)

var (
	machineHeaders = []api.Header{
		{Name: "ID"},
		{Name: "Class"},
		{Name: "Image"},
		{Name: "State"},
		{Name: "Age"},
	}
)

var (
	Machine = tableconverter.Funcs[*iri.Machine]{
		Headers: tableconverter.Headers(machineHeaders),
		Rows: tableconverter.SingleRowFrom(func(machine *iri.Machine) (api.Row, error) {
			return api.Row{
				machine.Metadata.Id,
				machine.Spec.Class,
				machine.Spec.GetImage().GetImage(),
				machine.Status.State.String(),
				duration.HumanDuration(time.Since(time.Unix(0, machine.Metadata.CreatedAt))),
			}, nil
		}),
	}
	MachineSlice = tableconverter.SliceFuncs[*iri.Machine](Machine)
)

func init() {
	RegistryBuilder.Register(
		tableconverter.ToTagAndTypedAny[*iri.Machine](Machine),
		tableconverter.ToTagAndTypedAny[[]*iri.Machine](MachineSlice),
	)
}
