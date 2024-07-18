// SPDX-FileCopyrightText: 2024 SAP SE or an SAP affiliate company and IronCore contributors
// SPDX-License-Identifier: Apache-2.0

package tableconverters

import (
	iri "github.com/ironcore-dev/ironcore/iri/apis/event/v1alpha1"
	"github.com/ironcore-dev/ironcore/irictl/api"
	"github.com/ironcore-dev/ironcore/irictl/tableconverter"
	bucketpoolletv1alpha1 "github.com/ironcore-dev/ironcore/poollet/bucketpoollet/api/v1alpha1"
)

const (
	RootBucketName      = "downward-api.bucketpoollet.ironcore.dev/root-bucket-name"
	RootBucketNamespace = "downward-api.bucketpoollet.ironcore.dev/root-bucket-namespace"
)

var (
	eventHeaders = []api.Header{
		{Name: "InvolvedBucketName"},
		{Name: "Type"},
		{Name: "Reason"},
		{Name: "Message"},
		{Name: "RootBucketName"},
		{Name: "RootBucketNamespace"},
	}

	Events = tableconverter.Funcs[*iri.Event]{
		Headers: tableconverter.Headers(eventHeaders),
		Rows: tableconverter.SingleRowFrom(func(event *iri.Event) (api.Row, error) {
			return api.Row{
				event.Spec.GetInvolvedObjectMeta().Id,
				event.Spec.Type,
				event.Spec.Reason,
				event.Spec.Message,
				getRootBucketName(event.Spec.GetInvolvedObjectMeta().Labels),
				getRootBucketNamespace(event.Spec.GetInvolvedObjectMeta().Labels),
			}, nil
		}),
	}

	EventsSlice = tableconverter.SliceFuncs[*iri.Event](Events)
)

func getRootBucketName(labels map[string]string) string {
	var rootBucketName string
	rootBucketName, ok := labels[RootBucketName]
	if !ok {
		return labels[bucketpoolletv1alpha1.BucketNameLabel]
	}
	return rootBucketName
}
func getRootBucketNamespace(labels map[string]string) string {
	var rootBucketNamespace string
	rootBucketNamespace, ok := labels[RootBucketNamespace]
	if !ok {
		return labels[bucketpoolletv1alpha1.BucketNamespaceLabel]
	}
	return rootBucketNamespace
}

func init() {
	RegistryBuilder.Register(
		tableconverter.ToTagAndTypedAny[*iri.Event](Events),
		tableconverter.ToTagAndTypedAny[[]*iri.Event](EventsSlice),
	)
}
