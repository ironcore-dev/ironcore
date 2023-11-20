// SPDX-FileCopyrightText: 2023 SAP SE or an SAP affiliate company and IronCore contributors
// SPDX-License-Identifier: Apache-2.0

package renderer

var (
	LocalRegistryBuilder RegistryBuilder
	AddToRegistry        = LocalRegistryBuilder.AddToRegistry
)
