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

package machinetableconverters

import (
	ori "github.com/onmetal/onmetal-api/ori/apis/machine/v1alpha1"
	"github.com/onmetal/onmetal-api/orictl/table"
	"github.com/onmetal/onmetal-api/orictl/table/tableconverter"
	"k8s.io/apimachinery/pkg/api/resource"
)

var (
	machineClassHeaders = []table.Header{
		{Name: "Name"},
		{Name: "CPU"},
		{Name: "Memory"},
	}
)

var MachineClass, MachineClassSlice = tableconverter.ForType[*ori.MachineClass]( //nolint:revive
	func() ([]table.Header, error) {
		return machineClassHeaders, nil
	},
	func(class *ori.MachineClass) ([]table.Row, error) {
		return []table.Row{
			{
				class.Name,
				resource.NewMilliQuantity(class.Capabilities.CpuMillis, resource.DecimalSI).String(),
				resource.NewQuantity(int64(class.Capabilities.MemoryBytes), resource.DecimalSI).String(),
			},
		}, nil
	},
)

func init() {
	RegistryBuilder.Register(
		tableconverter.ToTaggedAny(MachineClass),
		tableconverter.ToTaggedAny(MachineClassSlice),
	)
}
