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

	computev1alpha1 "github.com/ironcore-dev/ironcore/api/compute/v1alpha1"
	"github.com/ironcore-dev/ironcore/broker/machinebroker/apiutils"
	iri "github.com/ironcore-dev/ironcore/iri/apis/machine/v1alpha1"
)

const (
	InvolvedObjectKind               = "Machine"
	InvolvedObjectKindSelector       = "involvedObject.kind"
	InvolvedObjectAPIversionSelector = "involvedObject.apiVersion"
)

func (s *Server) listEvents(ctx context.Context) ([]*iri.Event, error) {
	machineEventList := &v1.EventList{}
	selectorField := fields.Set{}
	selectorField[InvolvedObjectKindSelector] = InvolvedObjectKind
	selectorField[InvolvedObjectAPIversionSelector] = computev1alpha1.SchemeGroupVersion.String()

	if err := s.cluster.Client().List(ctx, machineEventList,
		client.InNamespace(s.cluster.Namespace()), client.MatchingFieldsSelector{Selector: selectorField.AsSelector()},
	); err != nil {
		return nil, err
	}

	var iriEvents []*iri.Event
	var eventTime int64
	for _, machineEvent := range machineEventList.Items {
		ironcoreMachine, err := s.getIronCoreMachine(ctx, machineEvent.InvolvedObject.Name)
		if err != nil {
			continue
		} else {
			machineObjectMetadata, err := apiutils.GetObjectMetadata(&ironcoreMachine.ObjectMeta)
			if err != nil {
				continue
			} else {
				eventTime = machineEvent.LastTimestamp.Unix()
				if eventTime < 0 {
					eventTime = machineEvent.FirstTimestamp.Unix()
				}
				iriEvent := &iri.Event{
					Spec: &iri.EventSpec{
						InvolvedObjectMeta: machineObjectMetadata,
						Reason:             machineEvent.Reason,
						Message:            machineEvent.Message,
						Type:               machineEvent.Type,
						EventTime:          eventTime,
					},
				}
				iriEvents = append(iriEvents, iriEvent)
			}
		}
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
		if !sel.Matches(labels.Set(iriEvent.Spec.InvolvedObjectMeta.Labels)) {
			continue
		}
		if filter.EventsFromTime > 0 && filter.EventsToTime > 0 {
			if iriEvent.Spec.EventTime >= filter.EventsFromTime && iriEvent.Spec.EventTime <= filter.EventsToTime {
				res = append(res, iriEvent)
			}
		} else {
			res = append(res, iriEvent)
		}

	}
	return res
}

func (s *Server) ListEvents(ctx context.Context, req *iri.ListEventsRequest) (*iri.ListEventsResponse, error) {
	iriEvents, err := s.listEvents(ctx)
	if err != nil {
		return nil, fmt.Errorf("error listing machine events : %w", err)
	}

	iriEvents = s.filterEvents(iriEvents, req.Filter)

	return &iri.ListEventsResponse{
		Events: iriEvents,
	}, nil
}
