// SPDX-FileCopyrightText: 2023 SAP SE or an SAP affiliate company and IronCore contributors
// SPDX-License-Identifier: Apache-2.0

package vcm

import (
	"context"
	"errors"

	iri "github.com/ironcore-dev/ironcore/iri/apis/volume/v1alpha1"
	"github.com/ironcore-dev/ironcore/poollet/irievent"
	"k8s.io/apimachinery/pkg/api/resource"
	"sigs.k8s.io/controller-runtime/pkg/manager"
)

var (
	ErrNoMatchingVolumeClass        = errors.New("no matching volume class")
	ErrAmbiguousMatchingVolumeClass = errors.New("ambiguous matching volume classes")
)

type VolumeClassMapper interface {
	manager.Runnable
	GetVolumeClassFor(ctx context.Context, name string, capabilities *iri.VolumeClassCapabilities) (*iri.VolumeClass, *resource.Quantity, error)
	WaitForSync(ctx context.Context) error
	AddListener(listener irievent.Listener) (irievent.ListenerRegistration, error)
	RemoveListener(reg irievent.ListenerRegistration) error
}
