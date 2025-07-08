// SPDX-FileCopyrightText: 2023 SAP SE or an SAP affiliate company and IronCore contributors
// SPDX-License-Identifier: Apache-2.0

package server

import (
	"context"
	"fmt"

	iri "github.com/ironcore-dev/ironcore/iri/apis/machine/v1alpha1"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
	apierrors "k8s.io/apimachinery/pkg/api/errors"
)

func (s *Server) DeleteReservation(ctx context.Context, req *iri.DeleteReservationRequest) (*iri.DeleteReservationResponse, error) {
	reservationID := req.ReservationId
	log := s.loggerFrom(ctx, "ReservationID", reservationID)

	ironcoreReservation, err := s.getIronCoreReservation(ctx, reservationID)
	if err != nil {
		return nil, err
	}

	log.V(1).Info("Deleting ironcore reservation")
	if err := s.cluster.Client().Delete(ctx, ironcoreReservation); err != nil {
		if !apierrors.IsNotFound(err) {
			return nil, fmt.Errorf("error deleting ironcore reservation: %w", err)
		}
		return nil, status.Errorf(codes.NotFound, "reservation %s not found", reservationID)
	}

	return &iri.DeleteReservationResponse{}, nil
}
