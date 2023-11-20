// SPDX-FileCopyrightText: 2023 SAP SE or an SAP affiliate company and IronCore contributors
// SPDX-License-Identifier: Apache-2.0

package mcm

import (
	"context"
	"errors"

	iri "github.com/ironcore-dev/ironcore/iri/apis/machine/v1alpha1"
	"github.com/ironcore-dev/ironcore/poollet/irievent"
	"sigs.k8s.io/controller-runtime/pkg/manager"
)

var (
	ErrNoMatchingMachineClass        = errors.New("no matching machine class")
	ErrAmbiguousMatchingMachineClass = errors.New("ambiguous matching machine classes")
)

type MachineClassMapper interface {
	manager.Runnable
	GetMachineClassFor(ctx context.Context, name string, capabilities *iri.MachineClassCapabilities) (*iri.MachineClass, int64, error)
	WaitForSync(ctx context.Context) error
	AddListener(listener irievent.Listener) (irievent.ListenerRegistration, error)
	RemoveListener(reg irievent.ListenerRegistration) error
}
