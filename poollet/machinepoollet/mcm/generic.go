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
	"fmt"
	"sync"
	"time"

	"github.com/go-logr/logr"
	"github.com/gogo/protobuf/proto"
	"github.com/ironcore-dev/ironcore/ori/apis/machine"
	ori "github.com/ironcore-dev/ironcore/ori/apis/machine/v1alpha1"
	"github.com/ironcore-dev/ironcore/poollet/orievent"
	"golang.org/x/exp/maps"
	"k8s.io/apimachinery/pkg/util/sets"
	"k8s.io/apimachinery/pkg/util/wait"
	ctrl "sigs.k8s.io/controller-runtime"
)

type capabilities struct {
	cpuMillis   int64
	memoryBytes int64
}

func getCapabilities(oriCaps *ori.MachineClassCapabilities) capabilities {
	return capabilities{
		cpuMillis:   oriCaps.CpuMillis,
		memoryBytes: oriCaps.MemoryBytes,
	}
}

type Generic struct {
	mu sync.RWMutex

	sync   bool
	synced chan struct{}

	listener sets.Set[*listener]

	machineClassByName         map[string]*ori.MachineClassStatus
	machineClassByCapabilities map[capabilities][]*ori.MachineClassStatus

	machineRuntime machine.RuntimeService

	relistPeriod time.Duration
}

func (g *Generic) AddListener(handler orievent.Listener) (orievent.ListenerRegistration, error) {
	g.mu.Lock()
	defer g.mu.Unlock()

	h := &listener{Listener: handler}

	g.listener.Insert(h)
	return &listenerRegistration{
		mcm:      g,
		listener: h,
	}, nil
}

func (g *Generic) RemoveListener(listener orievent.ListenerRegistration) error {
	g.mu.Lock()
	defer g.mu.Unlock()

	reg, ok := listener.(*listenerRegistration)
	if !ok {
		return fmt.Errorf("invalid listener registration")
	}

	g.listener.Delete(reg.listener)

	return nil
}

func shouldNotify(oldMachineClassByName map[string]*ori.MachineClassStatus, class *ori.MachineClassStatus) bool {
	oldMachineClass, ok := oldMachineClassByName[class.MachineClass.Name]
	if !ok {
		return true
	}

	return proto.Equal(class, oldMachineClass)
}

func (g *Generic) relist(ctx context.Context, log logr.Logger) error {
	log.V(1).Info("Relisting machine classes")
	res, err := g.machineRuntime.Status(ctx, &ori.StatusRequest{})
	if err != nil {
		return fmt.Errorf("error listing machine classes: %w", err)
	}

	g.mu.Lock()
	defer g.mu.Unlock()

	oldMachineClassByName := maps.Clone(g.machineClassByName)

	maps.Clear(g.machineClassByName)
	maps.Clear(g.machineClassByCapabilities)

	var notify bool
	for _, machineClassStatus := range res.MachineClassStatus {
		machineClass := machineClassStatus.GetMachineClass()
		notify = notify || shouldNotify(oldMachineClassByName, machineClassStatus)

		caps := capabilities{
			cpuMillis:   machineClass.Capabilities.CpuMillis,
			memoryBytes: machineClass.Capabilities.MemoryBytes,
		}
		g.machineClassByName[machineClass.Name] = machineClassStatus
		g.machineClassByCapabilities[caps] = append(g.machineClassByCapabilities[caps], machineClassStatus)
	}

	if notify {
		log.V(1).Info("Notify")
		for _, n := range g.listener.UnsortedList() {
			n.Enqueue()
		}
	}

	for _, machineClassStatus := range res.MachineClassStatus {
		machineClass := machineClassStatus.GetMachineClass()
		caps := capabilities{
			cpuMillis:   machineClass.Capabilities.CpuMillis,
			memoryBytes: machineClass.Capabilities.MemoryBytes,
		}
		g.machineClassByName[machineClass.Name] = machineClassStatus
		g.machineClassByCapabilities[caps] = append(g.machineClassByCapabilities[caps], machineClassStatus)
	}

	if !g.sync {
		g.sync = true
		close(g.synced)
	}

	return nil
}

func (g *Generic) Start(ctx context.Context) error {
	log := ctrl.LoggerFrom(ctx).WithName("mcm")
	wait.UntilWithContext(ctx, func(ctx context.Context) {
		if err := g.relist(ctx, log); err != nil {
			log.Error(err, "Error relisting")
		}
	}, g.relistPeriod)
	return nil
}

func (g *Generic) GetMachineClassFor(ctx context.Context, name string, caps *ori.MachineClassCapabilities) (*ori.MachineClass, int64, error) {
	g.mu.RLock()
	defer g.mu.RUnlock()

	expected := getCapabilities(caps)
	if byName, ok := g.machineClassByName[name]; ok && getCapabilities(byName.MachineClass.Capabilities) == expected {
		return byName.MachineClass, byName.Quantity, nil
	}

	if byCaps, ok := g.machineClassByCapabilities[expected]; ok {
		switch len(byCaps) {
		case 0:
			return nil, 0, ErrNoMatchingMachineClass
		case 1:
			classStatus := *byCaps[0]
			return classStatus.MachineClass, classStatus.Quantity, nil
		default:
			return nil, 0, ErrAmbiguousMatchingMachineClass
		}
	}

	return nil, 0, ErrNoMatchingMachineClass
}

func (g *Generic) WaitForSync(ctx context.Context) error {
	select {
	case <-g.synced:
		return nil
	case <-ctx.Done():
		return ctx.Err()
	}
}

type GenericOptions struct {
	RelistPeriod time.Duration
}

func setGenericOptionsDefaults(o *GenericOptions) {
	if o.RelistPeriod == 0 {
		o.RelistPeriod = 1 * time.Hour
	}
}

func NewGeneric(runtime machine.RuntimeService, opts GenericOptions) MachineClassMapper {
	setGenericOptionsDefaults(&opts)
	return &Generic{
		synced:                     make(chan struct{}),
		machineClassByName:         map[string]*ori.MachineClassStatus{},
		machineClassByCapabilities: map[capabilities][]*ori.MachineClassStatus{},
		listener:                   sets.New[*listener](),
		machineRuntime:             runtime,
		relistPeriod:               opts.RelistPeriod,
	}
}

type listener struct {
	orievent.Listener
}

type listenerRegistration struct {
	mcm      *Generic
	listener *listener
}
