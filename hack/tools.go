// SPDX-FileCopyrightText: 2023 SAP SE or an SAP affiliate company and IronCore contributors
// SPDX-License-Identifier: Apache-2.0

// Package tools

//go:build tools
// +build tools

package hack

import (
	_ "k8s.io/code-generator"
)
