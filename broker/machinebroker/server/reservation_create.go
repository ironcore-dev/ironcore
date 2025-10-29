// SPDX-FileCopyrightText: 2023 SAP SE or an SAP affiliate company and IronCore contributors
// SPDX-License-Identifier: Apache-2.0

package server

import (
	"context"
	"fmt"

	"github.com/go-logr/logr"
	computev1alpha1 "github.com/ironcore-dev/ironcore/api/compute/v1alpha1"
	"github.com/ironcore-dev/ironcore/api/core/v1alpha1"
	"github.com/ironcore-dev/ironcore/broker/common/cleaner"
	brokerutils "github.com/ironcore-dev/ironcore/broker/common/utils"
	machinebrokerv1alpha1 "github.com/ironcore-dev/ironcore/broker/machinebroker/api/v1alpha1"
	"github.com/ironcore-dev/ironcore/broker/machinebroker/apiutils"
	iri "github.com/ironcore-dev/ironcore/iri/apis/machine/v1alpha1"
	metav1alpha1 "github.com/ironcore-dev/ironcore/iri/apis/meta/v1alpha1"
	machinepoolletv1alpha1 "github.com/ironcore-dev/ironcore/poollet/machinepoollet/api/v1alpha1"
	"github.com/ironcore-dev/ironcore/utils/maps"
	"k8s.io/apimachinery/pkg/api/resource"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

func prepareIronCoreAnnotations[T metav1alpha1.Object](obj T) (map[string]string, error) {
	annotationsValue, err := apiutils.EncodeAnnotationsAnnotation(obj.GetMetadata().GetAnnotations())
	if err != nil {
		return nil, fmt.Errorf("error encoding annotations: %w", err)
	}

	labelsValue, err := apiutils.EncodeLabelsAnnotation(obj.GetMetadata().GetLabels())
	if err != nil {
		return nil, fmt.Errorf("error encoding labels: %w", err)
	}

	return map[string]string{
		machinebrokerv1alpha1.AnnotationsAnnotation: annotationsValue,
		machinebrokerv1alpha1.LabelsAnnotation:      labelsValue,
	}, nil
}

func (s *Server) createIronCoreReservation(
	ctx context.Context,
	log logr.Logger,
	iriReservation *iri.Reservation,
) (res *computev1alpha1.Reservation, retErr error) {

	labels := brokerutils.PrepareDownwardAPILabels(
		iriReservation.GetMetadata().GetLabels(),
		s.brokerDownwardAPILabels,
		machinepoolletv1alpha1.MachineDownwardAPIPrefix,
	)

	annotations, err := prepareIronCoreAnnotations(iriReservation)
	if err != nil {
		return nil, fmt.Errorf("error preparing ironcore reservation annotations: %w", err)
	}

	var resources = v1alpha1.ResourceList{}
	for name, quantity := range iriReservation.Spec.Resources {
		var q resource.Quantity
		if err := q.Unmarshal(quantity); err != nil {
			return nil, fmt.Errorf("error unmarshaling resource quantity: %w", err)
		}
		resources[v1alpha1.ResourceName(name)] = q
	}

	c, cleanup := s.setupCleaner(ctx, log, &retErr)
	defer cleanup()

	ironcoreReservation := &computev1alpha1.Reservation{
		ObjectMeta: metav1.ObjectMeta{
			Namespace:   s.cluster.Namespace(),
			Name:        s.cluster.IDGen().Generate(),
			Annotations: annotations,
			Labels: maps.AppendMap(labels, map[string]string{
				machinebrokerv1alpha1.ManagerLabel: machinebrokerv1alpha1.ReservationBrokerManager,
			}),
		},
		Spec: computev1alpha1.ReservationSpec{
			Resources: resources,
		},
	}
	log.V(1).Info("Creating ironcore reservation")
	if err := s.cluster.Client().Create(ctx, ironcoreReservation); err != nil {
		return nil, fmt.Errorf("error creating ironcore reservation: %w", err)
	}
	c.Add(cleaner.CleanupObject(s.cluster.Client(), ironcoreReservation))

	log.V(1).Info("Patching ironcore reservation as created")
	if err := apiutils.PatchCreated(ctx, s.cluster.Client(), ironcoreReservation); err != nil {
		return nil, fmt.Errorf("error patching ironcore reservation as created: %w", err)
	}

	return ironcoreReservation, nil
}

func (s *Server) CreateReservation(ctx context.Context, req *iri.CreateReservationRequest) (res *iri.CreateReservationResponse, retErr error) {
	log := s.loggerFrom(ctx)

	log.V(1).Info("Creating ironcore reservation")
	ironcoreReservation, err := s.createIronCoreReservation(ctx, log, req.Reservation)
	if err != nil {
		return nil, fmt.Errorf("error creating ironcore reservation: %w", err)
	}

	r, err := s.convertIronCoreReservation(ironcoreReservation)
	if err != nil {
		return nil, fmt.Errorf("error converting ironcore reservation: %w", err)
	}

	return &iri.CreateReservationResponse{
		Reservation: r,
	}, nil
}
