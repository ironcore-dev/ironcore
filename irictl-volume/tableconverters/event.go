// SPDX-FileCopyrightText: 2024 SAP SE or an SAP affiliate company and IronCore contributors
// SPDX-License-Identifier: Apache-2.0

package tableconverters

import (
	iri "github.com/ironcore-dev/ironcore/iri/apis/event/v1alpha1"
	"github.com/ironcore-dev/ironcore/irictl/api"
	"github.com/ironcore-dev/ironcore/irictl/tableconverter"
	volumepoolletv1alpha1 "github.com/ironcore-dev/ironcore/poollet/volumepoollet/api/v1alpha1"
)

const (
	RootVolumeName      = "downward-api.volumepoollet.ironcore.dev/root-volume-name"
	RootVolumeNamespace = "downward-api.volumepoollet.ironcore.dev/root-volume-namespace"
)

var (
	eventHeaders = []api.Header{
		{Name: "InvolvedVolumeName"},
		{Name: "Type"},
		{Name: "Reason"},
		{Name: "Message"},
		{Name: "RootVolumeName"},
		{Name: "RootVolumeNamespace"},
	}

	Events = tableconverter.Funcs[*iri.Event]{
		Headers: tableconverter.Headers(eventHeaders),
		Rows: tableconverter.SingleRowFrom(func(event *iri.Event) (api.Row, error) {
			return api.Row{
				event.Spec.GetInvolvedObjectMeta().Id,
				event.Spec.Type,
				event.Spec.Reason,
				event.Spec.Message,
				getRootVolumeName(event.Spec.GetInvolvedObjectMeta().Labels),
				getRootVolumeNamespace(event.Spec.GetInvolvedObjectMeta().Labels),
			}, nil
		}),
	}

	EventsSlice = tableconverter.SliceFuncs[*iri.Event](Events)
)

func getRootVolumeName(labels map[string]string) string {
	var rootVolumeName string
	rootVolumeName, ok := labels[RootVolumeName]
	if !ok {
		return labels[volumepoolletv1alpha1.VolumeNameLabel]
	}
	return rootVolumeName
}
func getRootVolumeNamespace(labels map[string]string) string {
	var rootVolumeNamespace string
	rootVolumeNamespace, ok := labels[RootVolumeNamespace]
	if !ok {
		return labels[volumepoolletv1alpha1.VolumeNamespaceLabel]
	}
	return rootVolumeNamespace
}

func init() {
	RegistryBuilder.Register(
		tableconverter.ToTagAndTypedAny[*iri.Event](Events),
		tableconverter.ToTagAndTypedAny[[]*iri.Event](EventsSlice),
	)
}
