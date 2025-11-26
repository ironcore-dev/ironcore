// SPDX-FileCopyrightText: 2023 SAP SE or an SAP affiliate company and IronCore contributors
// SPDX-License-Identifier: Apache-2.0

package tableconverters

import (
	iri "github.com/ironcore-dev/ironcore/iri/apis/volume/v1alpha1"
	"github.com/ironcore-dev/ironcore/irictl/api"
	"github.com/ironcore-dev/ironcore/irictl/tableconverter"
)

var (
	versionsHeaders = []api.Header{
		{Name: "Name"},
		{Name: "Version"},
	}

	VersionResponse = tableconverter.Funcs[*iri.VersionResponse]{
		Headers: tableconverter.Headers(versionsHeaders),
		Rows: tableconverter.SingleRowFrom(func(versionInfo *iri.VersionResponse) (api.Row, error) {
			return api.Row{
				versionInfo.RuntimeName,
				versionInfo.RuntimeVersion,
			}, nil
		}),
	}

	VersionResponseSlice = tableconverter.SliceFuncs[*iri.VersionResponse](VersionResponse)
)

func init() {
	RegistryBuilder.Register(
		tableconverter.ToTagAndTypedAny[*iri.VersionResponse](VersionResponse),
		tableconverter.ToTagAndTypedAny[[]*iri.VersionResponse](VersionResponseSlice),
	)
}
