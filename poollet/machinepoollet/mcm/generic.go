// SPDX-FileCopyrightText: 2023 SAP SE or an SAP affiliate company and IronCore contributors
// SPDX-License-Identifier: Apache-2.0

package mcm

import (
	"context"
	"fmt"
	"reflect"
	"sync"
	"time"

	"github.com/go-logr/logr"
	"github.com/ironcore-dev/ironcore/iri/apis/machine"
	iri "github.com/ironcore-dev/ironcore/iri/apis/machine/v1alpha1"
	"github.com/ironcore-dev/ironcore/poollet/irievent"
	"golang.org/x/exp/maps"
	"google.golang.org/protobuf/proto"
	"k8s.io/apimachinery/pkg/util/sets"
	"k8s.io/apimachinery/pkg/util/wait"
	ctrl "sigs.k8s.io/controller-runtime"
)

const (
	CpuMillis   = "cpuMillis"
	MemoryBytes = "memoryBytes"
)

type capabilities map[string]int64

func getAllCapabilities(iriCaps *iri.MachineClassCapabilities) capabilities {

	capabilities := capabilities{
		CpuMillis:   iriCaps.CpuMillis,
		MemoryBytes: iriCaps.MemoryBytes,
	}

	maps.Copy(capabilities, iriCaps.AdditionalResources)

	return capabilities
}

func getMachineClassByCapabilities(machineClassByCapabilities map[*iri.MachineClassStatus]capabilities, capabilities capabilities) []*iri.MachineClassStatus {
	matchingMachineClasses := []*iri.MachineClassStatus{}
	for machineClass, machineClassCapabilities := range machineClassByCapabilities {
		if reflect.DeepEqual(machineClassCapabilities, capabilities) {
			matchingMachineClasses = append(matchingMachineClasses, machineClass)
		}
	}
	return matchingMachineClasses
}

type Generic struct {
	mu sync.RWMutex

	sync   bool
	synced chan struct{}

	listener sets.Set[*listener]

	machineClassByName         map[string]*iri.MachineClassStatus
	machineClassByCapabilities map[*iri.MachineClassStatus]capabilities

	machineRuntime machine.RuntimeService

	relistPeriod time.Duration
}

func (g *Generic) AddListener(handler irievent.Listener) (irievent.ListenerRegistration, error) {
	g.mu.Lock()
	defer g.mu.Unlock()

	h := &listener{Listener: handler}

	g.listener.Insert(h)
	return &listenerRegistration{
		mcm:      g,
		listener: h,
	}, nil
}

func (g *Generic) RemoveListener(listener irievent.ListenerRegistration) error {
	g.mu.Lock()
	defer g.mu.Unlock()

	reg, ok := listener.(*listenerRegistration)
	if !ok {
		return fmt.Errorf("invalid listener registration")
	}

	g.listener.Delete(reg.listener)

	return nil
}

func shouldNotify(oldMachineClassByName map[string]*iri.MachineClassStatus, class *iri.MachineClassStatus) bool {
	oldMachineClass, ok := oldMachineClassByName[class.MachineClass.Name]
	if !ok {
		return true
	}

	return proto.Equal(class, oldMachineClass)
}

func (g *Generic) relist(ctx context.Context, log logr.Logger) error {
	log.V(1).Info("Relisting machine classes")
	res, err := g.machineRuntime.Status(ctx, &iri.StatusRequest{})
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

		g.machineClassByName[machineClass.Name] = machineClassStatus
		g.machineClassByCapabilities[machineClassStatus] = getAllCapabilities(machineClass.Capabilities)
	}

	if notify {
		log.V(1).Info("Notify")
		for _, n := range g.listener.UnsortedList() {
			n.Enqueue()
		}
	}

	for _, machineClassStatus := range res.MachineClassStatus {
		machineClass := machineClassStatus.GetMachineClass()
		g.machineClassByName[machineClass.Name] = machineClassStatus
		g.machineClassByCapabilities[machineClassStatus] = getAllCapabilities(machineClass.Capabilities)
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

func (g *Generic) GetMachineClassFor(ctx context.Context, name string, caps *iri.MachineClassCapabilities) (*iri.MachineClass, int64, error) {
	g.mu.RLock()
	defer g.mu.RUnlock()

	expected := getAllCapabilities(caps)
	if byName, ok := g.machineClassByName[name]; ok && reflect.DeepEqual(getAllCapabilities(byName.MachineClass.Capabilities), expected) {
		return byName.MachineClass, byName.Quantity, nil
	}

	byCaps := getMachineClassByCapabilities(g.machineClassByCapabilities, expected)
	switch len(byCaps) {
	case 0:
		return nil, 0, ErrNoMatchingMachineClass
	case 1:
		return byCaps[0].MachineClass, byCaps[0].Quantity, nil
	default:
		return nil, 0, ErrAmbiguousMatchingMachineClass
	}

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
		machineClassByName:         map[string]*iri.MachineClassStatus{},
		machineClassByCapabilities: map[*iri.MachineClassStatus]capabilities{},
		listener:                   sets.New[*listener](),
		machineRuntime:             runtime,
		relistPeriod:               opts.RelistPeriod,
	}
}

type listener struct {
	irievent.Listener
}

type listenerRegistration struct {
	mcm      *Generic
	listener *listener
}
