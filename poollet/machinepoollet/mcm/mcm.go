// Copyright 2022 OnMetal authors
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

	ori "github.com/onmetal/onmetal-api/ori/apis/machine/v1alpha1"
	"github.com/onmetal/onmetal-api/poollet/orievent"
	"sigs.k8s.io/controller-runtime/pkg/manager"
)

var (
	ErrNoMatchingMachineClass        = errors.New("no matching machine class")
	ErrAmbiguousMatchingMachineClass = errors.New("ambiguous matching machine classes")
)

type MachineClassMapper interface {
	manager.Runnable
	GetMachineClassFor(ctx context.Context, name string, capabilities *ori.MachineClassCapabilities) (*ori.MachineClass, error)
	WaitForSync(ctx context.Context) error
	AddListener(listener orievent.Listener) (orievent.ListenerRegistration, error)
	RemoveListener(reg orievent.ListenerRegistration) error
}
