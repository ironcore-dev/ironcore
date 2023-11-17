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

package vcm

import (
	"context"
	"errors"

	ori "github.com/ironcore-dev/ironcore/ori/apis/volume/v1alpha1"
	"github.com/ironcore-dev/ironcore/poollet/orievent"
	"k8s.io/apimachinery/pkg/api/resource"
	"sigs.k8s.io/controller-runtime/pkg/manager"
)

var (
	ErrNoMatchingVolumeClass        = errors.New("no matching volume class")
	ErrAmbiguousMatchingVolumeClass = errors.New("ambiguous matching volume classes")
)

type VolumeClassMapper interface {
	manager.Runnable
	GetVolumeClassFor(ctx context.Context, name string, capabilities *ori.VolumeClassCapabilities) (*ori.VolumeClass, *resource.Quantity, error)
	WaitForSync(ctx context.Context) error
	AddListener(listener orievent.Listener) (orievent.ListenerRegistration, error)
	RemoveListener(reg orievent.ListenerRegistration) error
}
