// Copyright 2022 IronCore authors
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
//      http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.

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
