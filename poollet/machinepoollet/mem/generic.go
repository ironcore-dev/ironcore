// SPDX-FileCopyrightText: 2024 SAP SE or an SAP affiliate company and IronCore contributors
// SPDX-License-Identifier: Apache-2.0

package mem

import (
	"context"
	"fmt"
	"sync"
	"time"

	"github.com/go-logr/logr"
	"github.com/gogo/protobuf/proto"
	computev1alpha1 "github.com/ironcore-dev/ironcore/api/compute/v1alpha1"
	"github.com/ironcore-dev/ironcore/iri/apis/machine"
	iri "github.com/ironcore-dev/ironcore/iri/apis/machine/v1alpha1"
	"github.com/ironcore-dev/ironcore/poollet/machinepoollet/api/v1alpha1"
	"golang.org/x/exp/maps"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
	apierrors "k8s.io/apimachinery/pkg/api/errors"
	"k8s.io/apimachinery/pkg/util/wait"
	"k8s.io/client-go/tools/record"
	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/client"
)

type Generic struct {
	record.EventRecorder
	client.Client
	mu sync.RWMutex

	sync   bool
	synced chan struct{}

	machineEventsByID map[string]*iri.MachineEvents

	machineRuntime machine.RuntimeService

	relistPeriod time.Duration
}

func isNewEventsPresent(oldMachineEventsByID map[string]*iri.MachineEvents, machineEvent *iri.MachineEvents, machineID string) bool {
	oldMachineEvent, ok := oldMachineEventsByID[machineID]
	if !ok {
		return true
	}
	return !proto.Equal(machineEvent, oldMachineEvent)
}

func (g *Generic) relist(ctx context.Context, log logr.Logger) error {
	log.V(1).Info("Relisting machine cluster events")
	toEventFilterTime := time.Now()
	fromEventFilterTime := toEventFilterTime.Add(-1 * g.relistPeriod)
	res, err := g.machineRuntime.ListEvents(ctx, &iri.ListEventsRequest{
		Filter: &iri.EventFilter{EventsFromTime: fromEventFilterTime.Unix(), EventsToTime: toEventFilterTime.Add(-1 * g.relistPeriod).Unix()},
	})
	if err != nil {
		return fmt.Errorf("error listing machine cluster events: %w", err)
	}

	g.mu.Lock()
	defer g.mu.Unlock()

	oldMachineEventsByID := maps.Clone(g.machineEventsByID)

	maps.Clear(g.machineEventsByID)

	var shouldPublishEvents bool
	for _, machineEvent := range res.MachineEvents {
		if machine, err := g.getMachine(ctx, machineEvent.InvolvedObjectMeta.GetLabels()); err == nil {
			shouldPublishEvents = isNewEventsPresent(oldMachineEventsByID, machineEvent, string(machine.GetUID()))
			if shouldPublishEvents {
				for _, event := range machineEvent.Events {
					g.Eventf(machine, event.Spec.Type, event.Spec.Reason, event.Spec.Message)
				}
			}
			g.machineEventsByID[string(machine.GetUID())] = machineEvent
		}
	}

	if !g.sync {
		g.sync = true
		close(g.synced)
	}

	return nil
}

func (g *Generic) getMachine(ctx context.Context, labels map[string]string) (*computev1alpha1.Machine, error) {
	ironcoreMachine := &computev1alpha1.Machine{}
	ironcoreMachineKey := client.ObjectKey{Namespace: labels[v1alpha1.MachineNamespaceLabel], Name: labels[v1alpha1.MachineNameLabel]}
	if err := g.Client.Get(ctx, ironcoreMachineKey, ironcoreMachine); err != nil {
		if !apierrors.IsNotFound(err) {
			return nil, fmt.Errorf("error getting ironcore machine: %w", err)
		}
		return nil, status.Errorf(codes.NotFound, "machine %s not found", ironcoreMachineKey.Name)
	}
	return ironcoreMachine, nil
}

func (g *Generic) Start(ctx context.Context) error {
	log := ctrl.LoggerFrom(ctx).WithName("mem")
	wait.UntilWithContext(ctx, func(ctx context.Context) {
		if err := g.relist(ctx, log); err != nil {
			log.Error(err, "Error relisting")
		}
	}, g.relistPeriod)
	return nil
}

func (g *Generic) GetMachineEventFor(ctx context.Context, machineID string) ([]*iri.Event, error) {
	g.mu.RLock()
	defer g.mu.RUnlock()

	if byName, ok := g.machineEventsByID[machineID]; ok {
		return byName.Events, nil
	}
	return nil, ErrNoMatchingMachineEvents
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
		o.RelistPeriod = 1 * time.Minute
	}
}

func NewGeneric(client client.Client, runtime machine.RuntimeService, recorder record.EventRecorder, opts GenericOptions) MachineEventMapper {
	setGenericOptionsDefaults(&opts)
	return &Generic{
		synced:            make(chan struct{}),
		machineEventsByID: map[string]*iri.MachineEvents{},
		Client:            client,
		machineRuntime:    runtime,
		relistPeriod:      opts.RelistPeriod,
		EventRecorder:     recorder,
	}
}
