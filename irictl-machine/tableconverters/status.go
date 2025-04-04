// SPDX-FileCopyrightText: 2023 SAP SE or an SAP affiliate company and IronCore contributors
// SPDX-License-Identifier: Apache-2.0

package tableconverters

import (
	iri "github.com/ironcore-dev/ironcore/iri/apis/machine/v1alpha1"
	"github.com/ironcore-dev/ironcore/irictl/api"
	"github.com/ironcore-dev/ironcore/irictl/tableconverter"
	"k8s.io/apimachinery/pkg/api/resource"
)

var (
	machineClassHeaders = []api.Header{
		{Name: "Name"},
		{Name: "CPU"},
		{Name: "Memory"},
		{Name: "Quantity"},
	}

	MachineClassStatus = tableconverter.Funcs[*iri.MachineClassStatus]{
		Headers: tableconverter.Headers(machineClassHeaders),
		Rows: tableconverter.SingleRowFrom(func(status *iri.MachineClassStatus) (api.Row, error) {
			return api.Row{
				status.MachineClass.Name,
				resource.NewMilliQuantity(status.MachineClass.Capabilities.CpuMillis, resource.DecimalSI).String(),
				resource.NewQuantity(status.MachineClass.Capabilities.MemoryBytes, resource.DecimalSI).String(),
				resource.NewQuantity(status.Quantity, resource.DecimalSI).String(),
			}, nil
		}),
	}

	MachineClassStatusSlice = tableconverter.SliceFuncs[*iri.MachineClassStatus](MachineClassStatus)
)

func init() {
	RegistryBuilder.Register(
		tableconverter.ToTagAndTypedAny[*iri.MachineClassStatus](MachineClassStatus),
		tableconverter.ToTagAndTypedAny[[]*iri.MachineClassStatus](MachineClassStatusSlice),
	)
}
