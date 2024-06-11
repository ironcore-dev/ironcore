// SPDX-FileCopyrightText: 2024 SAP SE or an SAP affiliate company and IronCore contributors
// SPDX-License-Identifier: Apache-2.0

package mem

import (
	"context"
	"errors"

	iri "github.com/ironcore-dev/ironcore/iri/apis/machine/v1alpha1"
	"sigs.k8s.io/controller-runtime/pkg/manager"
)

var ErrNoMatchingMachineEvents = errors.New("no matching machine events")

type MachineEventMapper interface {
	manager.Runnable
	GetMachineEventFor(ctx context.Context, machineID string) ([]*iri.Event, error)
	WaitForSync(ctx context.Context) error
}
