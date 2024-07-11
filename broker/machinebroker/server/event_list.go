// SPDX-FileCopyrightText: 2024 SAP SE or an SAP affiliate company and IronCore contributors
// SPDX-License-Identifier: Apache-2.0

package server

import (
	"context"
	"fmt"

	v1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/fields"
	"k8s.io/apimachinery/pkg/labels"
	"sigs.k8s.io/controller-runtime/pkg/client"

	"github.com/ironcore-dev/ironcore/broker/machinebroker/apiutils"
	iri "github.com/ironcore-dev/ironcore/iri/apis/machine/v1alpha1"
)

const InvolvedObjectName = "involvedObject.name"

func (s *Server) listEvents(ctx context.Context, machineID string) ([]*iri.Event, error) {
	machineEventList := &v1.EventList{}
	selectorField := fields.Set{}
	selectorField[InvolvedObjectName] = machineID

	if err := s.cluster.Client().List(ctx, machineEventList,
		client.InNamespace(s.cluster.Namespace()), client.MatchingFieldsSelector{Selector: selectorField.AsSelector()},
	); err != nil {
		return nil, fmt.Errorf("error listing machine events: %w", err)
	}

	var iriEvents []*iri.Event
	for _, machineEvent := range machineEventList.Items {
		iriEvent := &iri.Event{
			Metadata: apiutils.GetIRIObjectMetadata(&machineEvent.ObjectMeta),
			Spec: &iri.EventSpec{
				Reason:    machineEvent.Reason,
				Message:   machineEvent.Message,
				Type:      machineEvent.Type,
				EventTime: machineEvent.FirstTimestamp.Unix(),
			},
		}
		iriEvents = append(iriEvents, iriEvent)
	}
	return iriEvents, nil
}

func (s *Server) filterEvents(events []*iri.Event, filter *iri.EventFilter) []*iri.Event {
	if filter == nil {
		return events
	}

	var (
		res []*iri.Event
		sel = labels.SelectorFromSet(filter.LabelSelector)
	)
	for _, iriEvent := range events {
		if !sel.Matches(labels.Set(iriEvent.Metadata.Labels)) {
			continue
		}
		if (filter.EventsFromTime >= 0 && iriEvent.Spec.EventTime >= filter.EventsFromTime) && (filter.EventsToTime >= 0 && iriEvent.Spec.EventTime <= filter.EventsToTime) {
			res = append(res, iriEvent)
		}

	}
	return res
}

func (s *Server) ListEvents(ctx context.Context, req *iri.ListEventsRequest) (*iri.ListEventsResponse, error) {
	ironcoreMachineList, err := s.listIroncoreMachines(ctx)
	if err != nil {
		return nil, err
	}
	var machineEvents []*iri.MachineEvents
	for i := range ironcoreMachineList.Items {
		ironcoreMachine := &ironcoreMachineList.Items[i]
		iriEvents, err := s.listEvents(ctx, ironcoreMachine.Name)
		if err != nil {
			return nil, err
		}

		iriEvents = s.filterEvents(iriEvents, req.Filter)
		ironcoreMachineMeta, err := apiutils.GetObjectMetadata(&ironcoreMachine.ObjectMeta)
		if err != nil {
			return nil, fmt.Errorf("error listing machine events: %w", err)
		}
		machineEvents = append(machineEvents, &iri.MachineEvents{InvolvedObjectMeta: ironcoreMachineMeta, Events: iriEvents})
	}

	return &iri.ListEventsResponse{
		MachineEvents: machineEvents,
	}, nil
}
