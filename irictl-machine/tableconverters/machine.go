// SPDX-FileCopyrightText: 2023 SAP SE or an SAP affiliate company and IronCore contributors
// SPDX-License-Identifier: Apache-2.0

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
