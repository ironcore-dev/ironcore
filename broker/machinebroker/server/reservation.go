// SPDX-FileCopyrightText: 2023 SAP SE or an SAP affiliate company and IronCore contributors
// SPDX-License-Identifier: Apache-2.0

package server

import (
	"fmt"

	computev1alpha1 "github.com/ironcore-dev/ironcore/api/compute/v1alpha1"
	"github.com/ironcore-dev/ironcore/broker/machinebroker/apiutils"
	iri "github.com/ironcore-dev/ironcore/iri/apis/machine/v1alpha1"
)

var ironcoreReservationStateToReservationState = map[computev1alpha1.ReservationState]iri.ReservationState{
	computev1alpha1.ReservationStatePending:  iri.ReservationState_RESERVATION_STATE_PENDING,
	computev1alpha1.ReservationStateAccepted: iri.ReservationState_RESERVATION_STATE_ACCEPTED,
	computev1alpha1.ReservationStateRejected: iri.ReservationState_RESERVATION_STATE_REJECTED,
}

func (s *Server) convertIronCoreReservationState(state computev1alpha1.ReservationState) (iri.ReservationState, error) {
	if res, ok := ironcoreReservationStateToReservationState[state]; ok {
		return res, nil
	}
	return 0, fmt.Errorf("unknown ironcore reservation state %q", state)
}

func (s *Server) convertIronCoreReservationStatus(ironCoreReservation *computev1alpha1.Reservation) (iri.ReservationState, error) {
	if ironCoreReservation.Spec.Pools == nil || ironCoreReservation.Status.Pools == nil {
		return iri.ReservationState_RESERVATION_STATE_PENDING, nil
	}

	//TODO make configurable
	for _, pool := range ironCoreReservation.Status.Pools {
		state, err := s.convertIronCoreReservationState(pool.State)
		if err != nil {
			return iri.ReservationState_RESERVATION_STATE_PENDING, err
		}

		switch state {
		case iri.ReservationState_RESERVATION_STATE_REJECTED:
			return state, nil
		case iri.ReservationState_RESERVATION_STATE_PENDING:
			return state, nil

		}
	}

	return iri.ReservationState_RESERVATION_STATE_REJECTED, nil
}

func (s *Server) convertIronCoreReservation(ironCoreReservation *computev1alpha1.Reservation) (*iri.Reservation, error) {
	metadata, err := apiutils.GetObjectMetadata(ironCoreReservation)
	if err != nil {
		return nil, err
	}

	state, err := s.convertIronCoreReservationStatus(ironCoreReservation)
	if err != nil {
		return nil, err
	}

	return &iri.Reservation{
		Metadata: metadata,
		Spec: &iri.ReservationSpec{
			Resources: nil,
		},
		Status: &iri.ReservationStatus{
			State: state,
		},
	}, nil
}
