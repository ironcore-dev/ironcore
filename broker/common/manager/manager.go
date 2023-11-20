// SPDX-FileCopyrightText: 2023 SAP SE or an SAP affiliate company and IronCore contributors
// SPDX-License-Identifier: Apache-2.0

package manager

import "sigs.k8s.io/controller-runtime/pkg/manager"

type Manager interface {
	Add(runnable manager.Runnable) error
}
