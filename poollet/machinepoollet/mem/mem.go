// SPDX-FileCopyrightText: 2024 SAP SE or an SAP affiliate company and IronCore contributors
// SPDX-License-Identifier: Apache-2.0

package mem

import (
	"context"
	"fmt"
	"time"

	"github.com/go-logr/logr"
	computev1alpha1 "github.com/ironcore-dev/ironcore/api/compute/v1alpha1"
	"github.com/ironcore-dev/ironcore/iri/apis/machine"
	iri "github.com/ironcore-dev/ironcore/iri/apis/machine/v1alpha1"
	"github.com/ironcore-dev/ironcore/poollet/machinepoollet/api/v1alpha1"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
	apierrors "k8s.io/apimachinery/pkg/api/errors"
	"k8s.io/apimachinery/pkg/util/wait"
	"k8s.io/client-go/tools/record"
	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/manager"
)

type MachineEventMapper struct {
	manager.Runnable
	record.EventRecorder
	client.Client

	machineRuntime machine.RuntimeService

	relistPeriod time.Duration
	lastFetched  time.Time
}

func (g *MachineEventMapper) relist(ctx context.Context, log logr.Logger) error {
	log.V(1).Info("Relisting machine cluster events")
	toEventFilterTime := time.Now()
	res, err := g.machineRuntime.ListEvents(ctx, &iri.ListEventsRequest{
		Filter: &iri.EventFilter{EventsFromTime: g.lastFetched.Unix(), EventsToTime: toEventFilterTime.Unix()},
	})
	if err != nil {
		return fmt.Errorf("error listing machine cluster events: %w", err)
	}

	g.lastFetched = toEventFilterTime

	for _, machineEvent := range res.MachineEvents {
		if machine, err := g.getMachine(ctx, machineEvent.InvolvedObjectMeta.GetLabels()); err == nil {
			for _, event := range machineEvent.Events {
				g.Eventf(machine, event.Spec.Type, event.Spec.Reason, event.Spec.Message)
			}
		}
	}

	return nil
}

func (g *MachineEventMapper) getMachine(ctx context.Context, labels map[string]string) (*computev1alpha1.Machine, error) {
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

func (g *MachineEventMapper) Start(ctx context.Context) error {
	log := ctrl.LoggerFrom(ctx).WithName("mem")
	g.lastFetched = time.Now()
	wait.UntilWithContext(ctx, func(ctx context.Context) {
		if err := g.relist(ctx, log); err != nil {
			log.Error(err, "Error relisting")
		}
	}, g.relistPeriod)
	return nil
}

type MachineEventMapperOptions struct {
	RelistPeriod time.Duration
}

func setMachineEventMapperOptionsDefaults(o *MachineEventMapperOptions) {
	if o.RelistPeriod == 0 {
		o.RelistPeriod = 1 * time.Minute
	}
}

func NewMachineEventMapper(client client.Client, runtime machine.RuntimeService, recorder record.EventRecorder, opts MachineEventMapperOptions) *MachineEventMapper {
	setMachineEventMapperOptionsDefaults(&opts)
	return &MachineEventMapper{
		Client:         client,
		machineRuntime: runtime,
		relistPeriod:   opts.RelistPeriod,
		EventRecorder:  recorder,
	}
}
