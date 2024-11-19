// SPDX-FileCopyrightText: 2024 SAP SE or an SAP affiliate company and IronCore contributors
// SPDX-License-Identifier: Apache-2.0

package server

import (
	"context"
	"fmt"

	storagev1alpha1 "github.com/ironcore-dev/ironcore/api/storage/v1alpha1"
	"github.com/ironcore-dev/ironcore/broker/bucketbroker/apiutils"
	iri "github.com/ironcore-dev/ironcore/iri/apis/bucket/v1alpha1"
	irievent "github.com/ironcore-dev/ironcore/iri/apis/event/v1alpha1"

	v1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/fields"
	"k8s.io/apimachinery/pkg/labels"
	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/client"
)

const (
	InvolvedObjectKind               = "Bucket"
	InvolvedObjectKindSelector       = "involvedObject.kind"
	InvolvedObjectAPIVersionSelector = "involvedObject.apiVersion"
)

func (s *Server) listEvents(ctx context.Context) ([]*irievent.Event, error) {
	log := ctrl.LoggerFrom(ctx)
	bucketEventList := &v1.EventList{}
	selectorField := fields.Set{
		InvolvedObjectKindSelector:       InvolvedObjectKind,
		InvolvedObjectAPIVersionSelector: storagev1alpha1.SchemeGroupVersion.String(),
	}
	if err := s.client.List(ctx, bucketEventList,
		client.InNamespace(s.namespace), client.MatchingFieldsSelector{Selector: selectorField.AsSelector()},
	); err != nil {
		return nil, err
	}

	var iriEvents []*irievent.Event
	for _, bucketEvent := range bucketEventList.Items {
		ironcoreBucket, err := s.getIronCoreBucket(ctx, bucketEvent.InvolvedObject.Name)
		if err != nil {
			log.Error(err, "Unable to get ironcore bucket", "BucketName", bucketEvent.InvolvedObject.Name)
			continue
		}
		bucketObjectMetadata, err := apiutils.GetObjectMetadata(&ironcoreBucket.ObjectMeta)
		if err != nil {
			continue
		}
		iriEvent := &irievent.Event{
			Spec: &irievent.EventSpec{
				InvolvedObjectMeta: bucketObjectMetadata,
				Reason:             bucketEvent.Reason,
				Message:            bucketEvent.Message,
				Type:               bucketEvent.Type,
				EventTime:          bucketEvent.LastTimestamp.Unix(),
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
		return nil, fmt.Errorf("error listing bucket events : %w", err)
	}

	iriEvents = s.filterEvents(iriEvents, req.Filter)

	return &iri.ListEventsResponse{
		Events: iriEvents,
	}, nil
}
