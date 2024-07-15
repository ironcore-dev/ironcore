// SPDX-FileCopyrightText: 2024 SAP SE or an SAP affiliate company and IronCore contributors
// SPDX-License-Identifier: Apache-2.0

package tableconverters

import (
	iri "github.com/ironcore-dev/ironcore/iri/apis/machine/v1alpha1"
	"github.com/ironcore-dev/ironcore/irictl/api"
	"github.com/ironcore-dev/ironcore/irictl/tableconverter"
	machinepoolletv1alpha1 "github.com/ironcore-dev/ironcore/poollet/machinepoollet/api/v1alpha1"
)

const (
	RootMachineName      = "downward-api.machinepoollet.ironcore.dev/root-machine-name"
	RootMachineNamespace = "downward-api.machinepoollet.ironcore.dev/root-machine-namespace"
)

var (
	eventHeaders = []api.Header{
		{Name: "InvolvedMachineName"},
		{Name: "Type"},
		{Name: "Reason"},
		{Name: "Message"},
		{Name: "RootMachineName"},
		{Name: "RootMachineNamespace"},
	}

	Events = tableconverter.Funcs[*iri.Event]{
		Headers: tableconverter.Headers(eventHeaders),
		Rows: tableconverter.SingleRowFrom(func(event *iri.Event) (api.Row, error) {
			return api.Row{
				event.Spec.GetInvolvedObjectMeta().Id,
				event.Spec.Type,
				event.Spec.Reason,
				event.Spec.Message,
				getRootMachineName(event.Spec.GetInvolvedObjectMeta().Labels),
				getRootMachineNamespace(event.Spec.GetInvolvedObjectMeta().Labels),
			}, nil
		}),
	}

	EventsSlice = tableconverter.SliceFuncs[*iri.Event](Events)
)

func getRootMachineName(labels map[string]string) string {
	var rootMachineName string
	rootMachineName, ok := labels[RootMachineName]
	if !ok {
		return labels[machinepoolletv1alpha1.MachineNameLabel]
	}
	return rootMachineName
}
func getRootMachineNamespace(labels map[string]string) string {
	var rootMachineNamespace string
	rootMachineNamespace, ok := labels[RootMachineNamespace]
	if !ok {
		return labels[machinepoolletv1alpha1.MachineNamespaceLabel]
	}
	return rootMachineNamespace
}

func init() {
	RegistryBuilder.Register(
		tableconverter.ToTagAndTypedAny[*iri.Event](Events),
		tableconverter.ToTagAndTypedAny[[]*iri.Event](EventsSlice),
	)
}
