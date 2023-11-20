// SPDX-FileCopyrightText: 2023 SAP SE or an SAP affiliate company and IronCore contributors
// SPDX-License-Identifier: Apache-2.0

package renderers

import (
	"github.com/ironcore-dev/ironcore/irictl-bucket/tableconverters"
	"github.com/ironcore-dev/ironcore/irictl/renderer"
	"github.com/ironcore-dev/ironcore/irictl/tableconverter"
)

var (
	RegistryBuilder renderer.RegistryBuilder
	AddToRegistry   = RegistryBuilder.AddToRegistry
)

func init() {
	RegistryBuilder.Add(renderer.AddToRegistry)
	RegistryBuilder.Add(func(registry *renderer.Registry) error {
		tableConverter := tableconverter.NewRegistry()
		if err := tableconverters.AddToRegistry(tableConverter); err != nil {
			return err
		}
		return registry.Register("table", renderer.NewTable(tableConverter))
	})
}
