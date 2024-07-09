// SPDX-FileCopyrightText: 2024 SAP SE or an SAP affiliate company and IronCore contributors
// SPDX-License-Identifier: Apache-2.0

package mem

import (
	"context"
	"errors"

	"sigs.k8s.io/controller-runtime/pkg/manager"
)

var ErrNoMatchingMachineEvents = errors.New("no matching machine events")

type MachineEventMapper interface {
	manager.Runnable
	WaitForSync(ctx context.Context) error
}
