// SPDX-FileCopyrightText: 2023 SAP SE or an SAP affiliate company and IronCore contributors
// SPDX-License-Identifier: Apache-2.0

package tableconverters

import (
	"github.com/ironcore-dev/ironcore/irictl/tableconverter"
)

var (
	RegistryBuilder tableconverter.RegistryBuilder
	AddToRegistry   = RegistryBuilder.AddToRegistry
)
