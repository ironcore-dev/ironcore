// SPDX-FileCopyrightText: 2023 SAP SE or an SAP affiliate company and IronCore contributors
// SPDX-License-Identifier: Apache-2.0

package controllers

import (
	"context"
	"errors"
	"fmt"
	"github.com/go-logr/logr"
	"github.com/ironcore-dev/controller-utils/clientutils"
	commonv1alpha1 "github.com/ironcore-dev/ironcore/api/common/v1alpha1"
	computev1alpha1 "github.com/ironcore-dev/ironcore/api/compute/v1alpha1"
	networkingv1alpha1 "github.com/ironcore-dev/ironcore/api/networking/v1alpha1"
	storagev1alpha1 "github.com/ironcore-dev/ironcore/api/storage/v1alpha1"
	irimachine "github.com/ironcore-dev/ironcore/iri/apis/machine"
	iri "github.com/ironcore-dev/ironcore/iri/apis/machine/v1alpha1"
	irimeta "github.com/ironcore-dev/ironcore/iri/apis/meta/v1alpha1"
	"github.com/ironcore-dev/ironcore/poollet/machinepoollet/api/v1alpha1"
	utilclient "github.com/ironcore-dev/ironcore/utils/client"
	"github.com/ironcore-dev/ironcore/utils/predicates"
	"golang.org/x/exp/maps"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
	apierrors "k8s.io/apimachinery/pkg/api/errors"
	"k8s.io/apimachinery/pkg/types"
	"k8s.io/client-go/tools/record"
	"k8s.io/kubectl/pkg/util/fieldpath"
	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/builder"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/controller/controllerutil"
	"sigs.k8s.io/controller-runtime/pkg/predicate"
	"strconv"
)

type ReservationReconciler struct {
	record.EventRecorder
	client.Client

	MachineRuntime        irimachine.RuntimeService
	MachineRuntimeName    string
	MachineRuntimeVersion string

	MachinePoolName string

	DownwardAPILabels      map[string]string
	DownwardAPIAnnotations map[string]string

	WatchFilterValue string
}

func (r *ReservationReconciler) reservationKeyLabelSelector(reservationKey client.ObjectKey) map[string]string {
	return map[string]string{
		v1alpha1.ReservationNamespaceLabel: reservationKey.Namespace,
		v1alpha1.ReservationNameLabel:      reservationKey.Name,
	}
}

func (r *ReservationReconciler) reservationUIDLabelSelector(reservationUID types.UID) map[string]string {
	return map[string]string{
		v1alpha1.ReservationUIDLabel: string(reservationUID),
	}
}

//+kubebuilder:rbac:groups="",resources=events,verbs=create;patch
//+kubebuilder:rbac:groups="",resources=secrets,verbs=get;list;watch
//+kubebuilder:rbac:groups=compute.ironcore.dev,resources=machines,verbs=get;list;watch;update;patch
//+kubebuilder:rbac:groups=compute.ironcore.dev,resources=machines/status,verbs=get;update;patch
//+kubebuilder:rbac:groups=compute.ironcore.dev,resources=machines/finalizers,verbs=update
//+kubebuilder:rbac:groups=storage.ironcore.dev,resources=volumes,verbs=get;list;watch;update;patch
//+kubebuilder:rbac:groups=networking.ironcore.dev,resources=networkinterfaces,verbs=get;list;watch;update;patch
//+kubebuilder:rbac:groups=networking.ironcore.dev,resources=networkinterfaces/status,verbs=get;update;patch
//+kubebuilder:rbac:groups=networking.ironcore.dev,resources=networks,verbs=get;list;watch
//+kubebuilder:rbac:groups=ipam.ironcore.dev,resources=prefixes,verbs=get;list;watch

func (r *ReservationReconciler) Reconcile(ctx context.Context, req ctrl.Request) (ctrl.Result, error) {
	log := ctrl.LoggerFrom(ctx)
	reservation := &computev1alpha1.Reservation{}
	if err := r.Get(ctx, req.NamespacedName, reservation); err != nil {
		if !apierrors.IsNotFound(err) {
			return ctrl.Result{}, fmt.Errorf("error getting reservation %s: %w", req.NamespacedName, err)
		}
		return r.deleteGone(ctx, log, req.NamespacedName)
	}
	return r.reconcileExists(ctx, log, reservation)
}

func (r *ReservationReconciler) getIRIReservationsForReservation(ctx context.Context, reservation *computev1alpha1.Reservation) ([]*iri.Reservation, error) {
	res, err := r.MachineRuntime.ListReservations(ctx, &iri.ListReservationsRequest{
		Filter: &iri.ReservationFilter{LabelSelector: r.reservationUIDLabelSelector(reservation.UID)},
	})
	if err != nil {
		return nil, fmt.Errorf("error listing reservations by reservation uid: %w", err)
	}
	return res.Reservations, nil
}

func (r *ReservationReconciler) listReservationsByReservationKey(ctx context.Context, reservationKey client.ObjectKey) ([]*iri.Reservation, error) {
	res, err := r.MachineRuntime.ListReservations(ctx, &iri.ListReservationsRequest{
		Filter: &iri.ReservationFilter{LabelSelector: r.reservationKeyLabelSelector(reservationKey)},
	})
	if err != nil {
		return nil, fmt.Errorf("error listing reservations by reservation key: %w", err)
	}
	return res.Reservations, nil
}

func (r *ReservationReconciler) getMachineByID(ctx context.Context, id string) (*iri.Machine, error) {
	res, err := r.MachineRuntime.ListMachines(ctx, &iri.ListMachinesRequest{
		Filter: &iri.MachineFilter{Id: id},
	})
	if err != nil {
		return nil, fmt.Errorf("error listing machines filtering by id: %w", err)
	}

	switch len(res.Machines) {
	case 0:
		return nil, status.Errorf(codes.NotFound, "machine %s not found", id)
	case 1:
		return res.Machines[0], nil
	default:
		return nil, fmt.Errorf("multiple machines found for id %s", id)
	}
}

func (r *ReservationReconciler) deleteReservations(ctx context.Context, log logr.Logger, reservations []*iri.Reservation) (bool, error) {
	var (
		errs        []error
		deletingIDs []string
	)
	for _, reservation := range reservations {
		reservationID := reservation.Metadata.Id
		log := log.WithValues("ReservationID", reservationID)
		log.V(1).Info("Deleting matching reservation")
		if _, err := r.MachineRuntime.DeleteReservation(ctx, &iri.DeleteReservationRequest{
			ReservationId: reservationID,
		}); err != nil {
			if status.Code(err) != codes.NotFound {
				errs = append(errs, fmt.Errorf("error deleting reservation %s: %w", reservationID, err))
			} else {
				log.V(1).Info("Reservation is already gone")
			}
		} else {
			log.V(1).Info("Issued reservation deletion")
			deletingIDs = append(deletingIDs, reservationID)
		}
	}

	switch {
	case len(errs) > 0:
		return false, fmt.Errorf("error(s) deleting matching reservation(s): %v", errs)
	case len(deletingIDs) > 0:
		log.V(1).Info("Reservations are still deleting", "DeletingIDs", deletingIDs)
		return false, nil
	default:
		log.V(1).Info("No reservation present")
		return true, nil
	}
}

func (r *ReservationReconciler) deleteGone(ctx context.Context, log logr.Logger, reservationKey client.ObjectKey) (ctrl.Result, error) {
	log.V(1).Info("Delete gone")

	log.V(1).Info("Listing reservations by reservation key")
	reservations, err := r.listReservationsByReservationKey(ctx, reservationKey)
	if err != nil {
		return ctrl.Result{}, fmt.Errorf("error listing reservations: %w", err)
	}

	ok, err := r.deleteReservations(ctx, log, reservations)
	if err != nil {
		return ctrl.Result{}, fmt.Errorf("error deleting reservations: %w", err)
	}
	if !ok {
		log.V(1).Info("Not all reservations are gone yet, requeueing")
		return ctrl.Result{Requeue: true}, nil
	}
	log.V(1).Info("Deleted gone")
	return ctrl.Result{}, nil
}

func (r *ReservationReconciler) reconcileExists(ctx context.Context, log logr.Logger, reservation *computev1alpha1.Reservation) (ctrl.Result, error) {
	if !reservation.DeletionTimestamp.IsZero() {
		return r.delete(ctx, log, reservation)
	}
	return r.reconcile(ctx, log, reservation)
}

func (r *ReservationReconciler) delete(ctx context.Context, log logr.Logger, reservation *computev1alpha1.Reservation) (ctrl.Result, error) {
	log.V(1).Info("Delete")

	if !controllerutil.ContainsFinalizer(reservation, v1alpha1.ReservationFinalizer(r.MachinePoolName)) {
		log.V(1).Info("No finalizer present, nothing to do")
		return ctrl.Result{}, nil
	}

	log.V(1).Info("Finalizer present")

	log.V(1).Info("Deleting reservations by UID")
	ok, err := r.deleteReservationsByReservationUID(ctx, log, reservation.UID)
	if err != nil {
		return ctrl.Result{}, fmt.Errorf("error deleting reservations: %w", err)
	}
	if !ok {
		log.V(1).Info("Not all reservations are gone, requeueing")
		return ctrl.Result{Requeue: true}, nil
	}

	log.V(1).Info("Deleted iri reservations by UID, removing finalizer")
	if err := clientutils.PatchRemoveFinalizer(ctx, r.Client, reservation, v1alpha1.ReservationFinalizer(r.MachinePoolName)); err != nil {
		return ctrl.Result{}, fmt.Errorf("error removing finalizer: %w", err)
	}

	log.V(1).Info("Deleted")
	return ctrl.Result{}, nil
}

func (r *ReservationReconciler) deleteReservationsByReservationUID(ctx context.Context, log logr.Logger, reservationUID types.UID) (bool, error) {
	log.V(1).Info("Listing machines")
	res, err := r.MachineRuntime.ListReservations(ctx, &iri.ListReservationsRequest{
		Filter: &iri.ReservationFilter{
			LabelSelector: map[string]string{
				v1alpha1.ReservationUIDLabel: string(reservationUID),
			},
		},
	})
	if err != nil {
		return false, fmt.Errorf("error listing reservations: %w", err)
	}

	log.V(1).Info("Listed reservations", "NoOfMachines", len(res.Reservations))
	var (
		errs                   []error
		deletingReservationIDs []string
	)
	for _, reservation := range res.Reservations {
		reservationID := reservation.Metadata.Id
		log := log.WithValues("MachineID", reservationID)
		log.V(1).Info("Deleting reservation")
		_, err := r.MachineRuntime.DeleteReservation(ctx, &iri.DeleteReservationRequest{
			ReservationId: reservationID,
		})
		if err != nil {
			if status.Code(err) != codes.NotFound {
				errs = append(errs, fmt.Errorf("error deleting reservation %s: %w", reservationID, err))
			} else {
				log.V(1).Info("Reservation is already gone")
			}
		} else {
			log.V(1).Info("Issued reservation deletion")
			deletingReservationIDs = append(deletingReservationIDs, reservationID)
		}
	}

	switch {
	case len(errs) > 0:
		return false, fmt.Errorf("error(s) deleting reservation(s): %v", errs)
	case len(deletingReservationIDs) > 0:
		log.V(1).Info("Reservations are in deletion", "DeletingReservationIDs", deletingReservationIDs)
		return false, nil
	default:
		log.V(1).Info("All reservations are gone")
		return true, nil
	}
}

func (r *ReservationReconciler) reconcile(ctx context.Context, log logr.Logger, reservation *computev1alpha1.Reservation) (ctrl.Result, error) {
	log.V(1).Info("Reconcile")

	log.V(1).Info("Ensuring finalizer")
	modified, err := clientutils.PatchEnsureFinalizer(ctx, r.Client, reservation, v1alpha1.ReservationFinalizer(r.MachinePoolName))
	if err != nil {
		return ctrl.Result{}, fmt.Errorf("error ensuring finalizer: %w", err)
	}
	if modified {
		log.V(1).Info("Added finalizer, requeueing")
		return ctrl.Result{Requeue: true}, nil
	}
	log.V(1).Info("Finalizer is present")

	log.V(1).Info("Ensuring no reconcile annotation")
	modified, err = utilclient.PatchEnsureNoReconcileAnnotation(ctx, r.Client, reservation)
	if err != nil {
		return ctrl.Result{}, fmt.Errorf("error ensuring no reconcile annotation: %w", err)
	}
	if modified {
		log.V(1).Info("Removed reconcile annotation, requeueing")
		return ctrl.Result{Requeue: true}, nil
	}

	iriReservations, err := r.getIRIReservationsForReservation(ctx, reservation)
	if err != nil {
		return ctrl.Result{}, fmt.Errorf("error getting IRI reservations for reservation: %w", err)
	}

	switch len(iriReservations) {
	case 0:
		return r.create(ctx, log, reservation)
	case 1:
		iriReservation := iriReservations[0]
		return r.update(ctx, log, reservation, iriReservation)
	default:
		panic("unhandled: multiple IRI reservations")
	}
}

func (r *ReservationReconciler) iriReservationLabels(reservation *computev1alpha1.Reservation) (map[string]string, error) {
	annotations := map[string]string{
		v1alpha1.ReservationUIDLabel:       string(reservation.UID),
		v1alpha1.ReservationNamespaceLabel: reservation.Namespace,
		v1alpha1.ReservationNameLabel:      reservation.Name,
	}

	for name, fieldPath := range r.DownwardAPILabels {
		value, err := fieldpath.ExtractFieldPathAsString(reservation, fieldPath)
		if err != nil {
			return nil, fmt.Errorf("error extracting downward api label %q: %w", name, err)
		}

		annotations[v1alpha1.DownwardAPILabel(name)] = value
	}
	return annotations, nil
}

func (r *ReservationReconciler) iriReservationAnnotations(
	reservation *computev1alpha1.Reservation,
	iriReservationGeneration int64,
) (map[string]string, error) {

	annotations := map[string]string{
		v1alpha1.ReservationGenerationAnnotation:    strconv.FormatInt(reservation.Generation, 10),
		v1alpha1.IRIReservationGenerationAnnotation: strconv.FormatInt(iriReservationGeneration, 10),
	}

	for name, fieldPath := range r.DownwardAPIAnnotations {
		value, err := fieldpath.ExtractFieldPathAsString(reservation, fieldPath)
		if err != nil {
			return nil, fmt.Errorf("error extracting downward api annotation %q: %w", name, err)
		}

		annotations[v1alpha1.DownwardAPIAnnotation(name)] = value
	}

	return annotations, nil
}

func (r *ReservationReconciler) create(
	ctx context.Context,
	log logr.Logger,
	reservation *computev1alpha1.Reservation,
) (ctrl.Result, error) {
	log.V(1).Info("Create")

	log.V(1).Info("Getting reservation config")
	iriReservation, ok, err := r.prepareIRIReservation(ctx, reservation)
	if err != nil {
		return ctrl.Result{}, fmt.Errorf("error preparing iri reservation: %w", err)
	}
	if !ok {
		log.V(1).Info("Reservation is not yet ready")
		return ctrl.Result{}, nil
	}

	log.V(1).Info("Creating reservation")
	res, err := r.MachineRuntime.CreateReservation(ctx, &iri.CreateReservationRequest{
		Reservation: iriReservation,
	})
	if err != nil {
		return ctrl.Result{}, fmt.Errorf("error creating reservation: %w", err)
	}
	log.V(1).Info("Created", "ReservationID", res.Reservation.Metadata.Id)

	log.V(1).Info("Updating status")
	if err := r.updateStatus(ctx, log, reservation, res.Reservation); err != nil {
		return ctrl.Result{}, fmt.Errorf("error updating reservation status: %w", err)
	}

	log.V(1).Info("Created")
	return ctrl.Result{}, nil
}

func (r *ReservationReconciler) getReservationGeneration(iriReservation *iri.Reservation) (int64, error) {
	return getAndParseFromStringMap(iriReservation.GetMetadata().GetAnnotations(),
		v1alpha1.ReservationGenerationAnnotation,
		parseInt64,
	)
}

func (r *ReservationReconciler) getIRIReservationGeneration(iriReservation *iri.Reservation) (int64, error) {
	return getAndParseFromStringMap(iriReservation.GetMetadata().GetAnnotations(),
		v1alpha1.IRIReservationGenerationAnnotation,
		parseInt64,
	)
}

func (r *ReservationReconciler) updateStatus(
	ctx context.Context,
	log logr.Logger,
	reservation *computev1alpha1.Reservation,
	iriReservation *iri.Reservation,
) error {
	requiredIRIGeneration, err := r.getIRIReservationGeneration(iriReservation)
	if err != nil {
		return err
	}

	iriGeneration := iriReservation.Metadata.Generation
	observedIRIGeneration := iriReservation.Status.ObservedGeneration

	if observedIRIGeneration < requiredIRIGeneration {
		log.V(1).Info("IRI reservation was not observed at the latest generation",
			"IRIGeneration", iriGeneration,
			"ObservedIRIGeneration", observedIRIGeneration,
			"RequiredIRIGeneration", requiredIRIGeneration,
		)
		return nil
	}

	var errs []error

	if err := r.updateReservationStatus(ctx, log, reservation, iriReservation); err != nil {
		errs = append(errs, err)
	}

	return errors.Join(errs...)
}

var iriReservationStateToReservationState = map[iri.ReservationState]computev1alpha1.ReservationState{
	iri.ReservationState_RESERVATION_STATE_PENDING:  computev1alpha1.ReservationStatePending,
	iri.ReservationState_RESERVATION_STATE_ACCEPTED: computev1alpha1.ReservationStateAccepted,
	iri.ReservationState_RESERVATION_STATE_REJECTED: computev1alpha1.ReservationStateRejected,
}

func (r *ReservationReconciler) convertIRIReservationState(state iri.ReservationState) (computev1alpha1.ReservationState, error) {
	if res, ok := iriReservationStateToReservationState[state]; ok {
		return res, nil
	}
	return "", fmt.Errorf("unknown reservation state %v", state)
}

func (r *ReservationReconciler) updateReservationStatus(ctx context.Context, log logr.Logger, reservation *computev1alpha1.Reservation, iriReservation *iri.Reservation) error {
	generation, err := r.getReservationGeneration(iriReservation)
	if err != nil {
		return err
	}

	state, err := r.convertIRIReservationState(iriReservation.Status.State)
	if err != nil {
		return err
	}

	var availablePools []computev1alpha1.ReservationPoolStatus
	for _, poolState := range reservation.Status.Pools {
		if poolState.Name == r.MachinePoolName {
			availablePools = append(availablePools, computev1alpha1.ReservationPoolStatus{
				Name:  r.MachinePoolName,
				State: state,
			})
		} else {
			availablePools = append(availablePools, poolState)
		}
	}

	reservation.Status.Pools = availablePools

	if err := r.Status().Patch(ctx, reservation, client.MergeFrom(base)); err != nil {
		return fmt.Errorf("error patching status: %w", err)
	}
	return nil
}

func (r *ReservationReconciler) update(
	ctx context.Context,
	log logr.Logger,
	machine *computev1alpha1.Machine,
	iriMachine *iri.Machine,
	nics []networkingv1alpha1.NetworkInterface,
	volumes []storagev1alpha1.Volume,
) (ctrl.Result, error) {
	log.V(1).Info("Updating existing machine")

	var errs []error

	log.V(1).Info("Updating network interfaces")
	iriNics, err := r.updateIRINetworkInterfaces(ctx, log, machine, iriMachine, nics)
	if err != nil {
		errs = append(errs, fmt.Errorf("error updating network interfaces: %w", err))
	}

	log.V(1).Info("Updating volumes")
	if err := r.updateIRIVolumes(ctx, log, machine, iriMachine, volumes); err != nil {
		errs = append(errs, fmt.Errorf("error updating volumes: %w", err))
	}

	log.V(1).Info("Updating power state")
	if err := r.updateIRIPower(ctx, log, machine, iriMachine); err != nil {
		errs = append(errs, fmt.Errorf("error updating power state: %w", err))
	}

	if len(errs) > 0 {
		return ctrl.Result{}, fmt.Errorf("error(s) updating machine: %v", errs)
	}

	log.V(1).Info("Updating annotations")
	nicMapping := r.computeNetworkInterfaceMapping(machine, nics, iriNics)
	if err := r.updateIRIAnnotations(ctx, log, machine, iriMachine, nicMapping); err != nil {
		return ctrl.Result{}, fmt.Errorf("error updating annotations: %w", err)
	}

	log.V(1).Info("Getting iri machine")
	iriMachine, err = r.getMachineByID(ctx, iriMachine.Metadata.Id)
	if err != nil {
		return ctrl.Result{}, fmt.Errorf("error getting iri machine: %w", err)
	}

	log.V(1).Info("Updating machine status")
	if err := r.updateStatus(ctx, log, machine, iriMachine, nics); err != nil {
		return ctrl.Result{}, fmt.Errorf("error updating status: %w", err)
	}

	log.V(1).Info("Updated existing machine")
	return ctrl.Result{}, nil
}

func (r *ReservationReconciler) updateIRIAnnotations(
	ctx context.Context,
	log logr.Logger,
	machine *computev1alpha1.Machine,
	iriMachine *iri.Machine,
	nicMapping map[string]v1alpha1.ObjectUIDRef,
) error {
	desiredAnnotations, err := r.iriMachineAnnotations(machine, iriMachine.GetMetadata().GetGeneration(), nicMapping)
	if err != nil {
		return fmt.Errorf("error getting iri machine annotations: %w", err)
	}

	actualAnnotations := iriMachine.Metadata.Annotations

	if maps.Equal(desiredAnnotations, actualAnnotations) {
		log.V(1).Info("Annotations are up-to-date", "Annotations", desiredAnnotations)
		return nil
	}

	if _, err := r.MachineRuntime.UpdateMachineAnnotations(ctx, &iri.UpdateMachineAnnotationsRequest{
		MachineId:   iriMachine.Metadata.Id,
		Annotations: desiredAnnotations,
	}); err != nil {
		return fmt.Errorf("error updating machine annotations: %w", err)
	}
	return nil
}

func (r *ReservationReconciler) prepareIRIReservation(
	ctx context.Context,
	reservation *computev1alpha1.Reservation,
) (*iri.Reservation, bool, error) {
	var (
		errs []error
	)

	labels, err := r.iriReservationLabels(reservation)
	if err != nil {
		errs = append(errs, fmt.Errorf("error preparing iri reservation labels: %w", err))
	}

	annotations, err := r.iriReservationAnnotations(reservation, 1)
	if err != nil {
		errs = append(errs, fmt.Errorf("error preparing iri reservation annotations: %w", err))
	}

	var resources = map[string][]byte{}
	for resource, quantity := range reservation.Spec.Resources {
		if data, err := quantity.Marshal(); err != nil {
			resources[string(resource)] = data
		} else {
			errs = append(errs, fmt.Errorf("error marshaling quantity (%s): %w", resource, err))
		}

	}

	switch {
	case len(errs) > 0:
		return nil, false, fmt.Errorf("error(s) preparing reservation: %v", errs)
	default:
		return &iri.Reservation{
			Metadata: &irimeta.ObjectMetadata{
				Labels:      labels,
				Annotations: annotations,
			},
			Spec: &iri.ReservationSpec{
				Resources: resources,
			},
		}, true, nil
	}
}

func ReservationAssignedToMachinePool(reservation *computev1alpha1.Reservation, machinePoolName string) bool {
	for _, pool := range reservation.Spec.Pools {
		if pool.Name == machinePoolName {
			return true
		}
	}

	return false
}

func ReservationAssignedToMachinePoolPredicate(machinePoolName string) predicate.Predicate {
	return predicate.NewPredicateFuncs(func(object client.Object) bool {
		reservation := object.(*computev1alpha1.Reservation)
		return ReservationAssignedToMachinePool(reservation, machinePoolName)
	})
}

func (r *ReservationReconciler) matchingWatchLabel() client.ListOption {
	var labels map[string]string
	if r.WatchFilterValue != "" {
		labels = map[string]string{
			commonv1alpha1.WatchLabel: r.WatchFilterValue,
		}
	}
	return client.MatchingLabels(labels)
}

func (r *ReservationReconciler) SetupWithManager(mgr ctrl.Manager) error {
	log := ctrl.Log.WithName("machinepoollet")

	return ctrl.NewControllerManagedBy(mgr).
		For(
			&computev1alpha1.Reservation{},
			builder.WithPredicates(
				ReservationAssignedToMachinePoolPredicate(r.MachinePoolName),
				predicates.ResourceHasFilterLabel(log, r.WatchFilterValue),
				predicates.ResourceIsNotExternallyManaged(log),
			),
		).
		Complete(r)
}
