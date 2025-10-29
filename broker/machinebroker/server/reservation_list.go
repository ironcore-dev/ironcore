// SPDX-FileCopyrightText: 2023 SAP SE or an SAP affiliate company and IronCore contributors
// SPDX-License-Identifier: Apache-2.0

package server

import (
	"context"
	"fmt"

	computev1alpha1 "github.com/ironcore-dev/ironcore/api/compute/v1alpha1"
	machinebrokerv1alpha1 "github.com/ironcore-dev/ironcore/broker/machinebroker/api/v1alpha1"
	"github.com/ironcore-dev/ironcore/broker/machinebroker/apiutils"
	iri "github.com/ironcore-dev/ironcore/iri/apis/machine/v1alpha1"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
	apierrors "k8s.io/apimachinery/pkg/api/errors"
	"k8s.io/apimachinery/pkg/labels"
	"sigs.k8s.io/controller-runtime/pkg/client"
)

func (s *Server) listIroncoreReservations(ctx context.Context) (*computev1alpha1.ReservationList, error) {
	ironcoreReservationList := &computev1alpha1.ReservationList{}
	if err := s.cluster.Client().List(ctx, ironcoreReservationList,
		client.InNamespace(s.cluster.Namespace()),
		client.MatchingLabels{
			machinebrokerv1alpha1.ManagerLabel: machinebrokerv1alpha1.ReservationBrokerManager,
			machinebrokerv1alpha1.CreatedLabel: "true",
		},
	); err != nil {
		return nil, err
	}
	return ironcoreReservationList, nil
}

func (s *Server) getIronCoreReservation(ctx context.Context, id string) (*computev1alpha1.Reservation, error) {
	ironcoreReservation := &computev1alpha1.Reservation{}
	ironcoreReservationKey := client.ObjectKey{Namespace: s.cluster.Namespace(), Name: id}
	if err := s.cluster.Client().Get(ctx, ironcoreReservationKey, ironcoreReservation); err != nil {
		if !apierrors.IsNotFound(err) {
			return nil, fmt.Errorf("error getting ironcore reservation %s: %w", id, err)
		}
		return nil, status.Errorf(codes.NotFound, "reservation %s not found", id)
	}
	if !apiutils.IsManagedBy(ironcoreReservation, machinebrokerv1alpha1.ReservationBrokerManager) || !apiutils.IsCreated(ironcoreReservation) {
		return nil, status.Errorf(codes.NotFound, "reservation %s not found", id)
	}
	return ironcoreReservation, nil
}

func (s *Server) listReservations(ctx context.Context) ([]*iri.Reservation, error) {
	ironcoreReservations, err := s.listIroncoreReservations(ctx)
	if err != nil {
		return nil, fmt.Errorf("error listing ironcore reservations: %w", err)
	}
	var res []*iri.Reservation
	for _, ironcoreReservation := range ironcoreReservations.Items {
		reservation, err := s.convertIronCoreReservation(&ironcoreReservation)
		if err != nil {
			return nil, err
		}

		res = append(res, reservation)
	}
	return res, nil
}

func (s *Server) filterReservations(reservations []*iri.Reservation, filter *iri.ReservationFilter) []*iri.Reservation {
	if filter == nil {
		return reservations
	}

	var (
		res []*iri.Reservation
		sel = labels.SelectorFromSet(filter.LabelSelector)
	)
	for _, iriReservation := range reservations {
		if !sel.Matches(labels.Set(iriReservation.Metadata.Labels)) {
			continue
		}

		res = append(res, iriReservation)
	}
	return res
}

func (s *Server) getReservation(ctx context.Context, id string) (*iri.Reservation, error) {
	ironCoreReservation, err := s.getIronCoreReservation(ctx, id)
	if err != nil {
		return nil, err
	}

	return s.convertIronCoreReservation(ironCoreReservation)
}

func (s *Server) ListReservations(ctx context.Context, req *iri.ListReservationsRequest) (*iri.ListReservationsResponse, error) {
	if filter := req.Filter; filter != nil && filter.Id != "" {
		reservation, err := s.getReservation(ctx, filter.Id)
		if err != nil {
			if status.Code(err) != codes.NotFound {
				return nil, err
			}
			return &iri.ListReservationsResponse{
				Reservations: []*iri.Reservation{},
			}, nil
		}

		return &iri.ListReservationsResponse{
			Reservations: []*iri.Reservation{reservation},
		}, nil
	}

	reservations, err := s.listReservations(ctx)
	if err != nil {
		return nil, err
	}

	reservations = s.filterReservations(reservations, req.Filter)

	return &iri.ListReservationsResponse{
		Reservations: reservations,
	}, nil
}
