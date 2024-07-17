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
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/types"
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

func (m *MachineEventMapper) relist(ctx context.Context, log logr.Logger) error {
	log.V(1).Info("Relisting machine cluster events")
	toEventFilterTime := time.Now()
	res, err := m.machineRuntime.ListEvents(ctx, &iri.ListEventsRequest{
		Filter: &iri.EventFilter{EventsFromTime: m.lastFetched.Unix(), EventsToTime: toEventFilterTime.Unix()},
	})
	if err != nil {
		return fmt.Errorf("error listing machine cluster events: %w", err)
	}

	m.lastFetched = toEventFilterTime
	for _, machineEvent := range res.Events {
		if machineEvent.Spec.InvolvedObjectMeta.Labels != nil {
			involvedMachine := &computev1alpha1.Machine{
				ObjectMeta: metav1.ObjectMeta{
					UID:       types.UID(machineEvent.Spec.InvolvedObjectMeta.Labels[v1alpha1.MachineUIDLabel]),
					Name:      machineEvent.Spec.InvolvedObjectMeta.Labels[v1alpha1.MachineNameLabel],
					Namespace: machineEvent.Spec.InvolvedObjectMeta.Labels[v1alpha1.MachineNamespaceLabel],
				},
			}
			m.Eventf(involvedMachine, machineEvent.Spec.Type, machineEvent.Spec.Reason, machineEvent.Spec.Message)
		}
	}

	return nil
}

func (m *MachineEventMapper) Start(ctx context.Context) error {
	log := ctrl.LoggerFrom(ctx).WithName("mem")
	m.lastFetched = time.Now()
	wait.UntilWithContext(ctx, func(ctx context.Context) {
		if err := m.relist(ctx, log); err != nil {
			log.Error(err, "Error relisting")
		}
	}, m.relistPeriod)
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
