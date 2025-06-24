// SPDX-FileCopyrightText: 2024 SAP SE or an SAP affiliate company and IronCore contributors
// SPDX-License-Identifier: Apache-2.0

package server

import (
	"context"
	"errors"
	"fmt"

	v1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/fields"
	"k8s.io/apimachinery/pkg/labels"
	"k8s.io/apimachinery/pkg/util/sets"
	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/client"

	computev1alpha1 "github.com/ironcore-dev/ironcore/api/compute/v1alpha1"
	"github.com/ironcore-dev/ironcore/broker/machinebroker/apiutils"
	irievent "github.com/ironcore-dev/ironcore/iri/apis/event/v1alpha1"
	iri "github.com/ironcore-dev/ironcore/iri/apis/machine/v1alpha1"
)

const (
	InvolvedObjectKind               = "Machine"
	InvolvedObjectKindSelector       = "involvedObject.kind"
	InvolvedObjectAPIVersionSelector = "involvedObject.apiVersion"
	LogKeyMachineName                = "MachineName"
)

func (s *Server) listEvents(ctx context.Context) ([]*irievent.Event, error) {
	log := ctrl.LoggerFrom(ctx)
	machineEventList := &v1.EventList{}
	selectorField := fields.Set{
		InvolvedObjectKindSelector:       InvolvedObjectKind,
		InvolvedObjectAPIVersionSelector: computev1alpha1.SchemeGroupVersion.String(),
	}
	if err := s.cluster.Client().List(ctx, machineEventList,
		client.InNamespace(s.cluster.Namespace()), client.MatchingFieldsSelector{Selector: selectorField.AsSelector()},
	); err != nil {
		return nil, err
	}

	unmanagedMachines := sets.Set[string]{}
	var iriEvents []*irievent.Event
	for _, machineEvent := range machineEventList.Items {
		if unmanagedMachines.Has(machineEvent.InvolvedObject.Name) {
			log.V(2).Info("skipping unmanaged machine", LogKeyMachineName, machineEvent.InvolvedObject.Name)
			continue
		}

		ironcoreMachine, err := s.getIronCoreMachine(ctx, machineEvent.InvolvedObject.Name)
		if err != nil {
			if errors.Is(err, ErrMachineIsntManaged) || errors.Is(err, ErrMachineNotFound) {
				unmanagedMachines.Insert(machineEvent.InvolvedObject.Name)
				continue
			}
			log.V(1).Info("Unable to get ironcore machine", LogKeyMachineName, machineEvent.InvolvedObject.Name, "error", err.Error())
			continue
		}
		machineObjectMetadata, err := apiutils.GetObjectMetadata(&ironcoreMachine.ObjectMeta)
		if err != nil {
			continue
		}
		iriEvent := &irievent.Event{
			Spec: &irievent.EventSpec{
				InvolvedObjectMeta: machineObjectMetadata,
				Reason:             machineEvent.Reason,
				Message:            machineEvent.Message,
				Type:               machineEvent.Type,
				EventTime:          machineEvent.LastTimestamp.Unix(),
			},
		}
		iriEvents = append(iriEvents, iriEvent)
	}
	return iriEvents, nil
}

func (s *Server) filterEvents(events []*irievent.Event, filter *iri.EventFilter) []*irievent.Event {
	if filter == nil {
		return events
	}

	var (
		res []*irievent.Event
		sel = labels.SelectorFromSet(filter.LabelSelector)
	)
	for _, iriEvent := range events {
		if !sel.Matches(labels.Set(iriEvent.Spec.InvolvedObjectMeta.Labels)) {
			continue
		}

		if filter.EventsFromTime > 0 && filter.EventsToTime > 0 {
			if iriEvent.Spec.EventTime < filter.EventsFromTime || iriEvent.Spec.EventTime > filter.EventsToTime {
				continue
			}
		}

		res = append(res, iriEvent)
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
