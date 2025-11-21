// SPDX-FileCopyrightText: 2023 SAP SE or an SAP affiliate company and IronCore contributors
// SPDX-License-Identifier: Apache-2.0

package controllers

import (
	"context"
	"errors"
	"fmt"
	"strconv"

	"github.com/go-logr/logr"
	"golang.org/x/exp/maps"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"

	"github.com/ironcore-dev/controller-utils/clientutils"
	commonv1alpha1 "github.com/ironcore-dev/ironcore/api/common/v1alpha1"
	computev1alpha1 "github.com/ironcore-dev/ironcore/api/compute/v1alpha1"
	networkingv1alpha1 "github.com/ironcore-dev/ironcore/api/networking/v1alpha1"
	storagev1alpha1 "github.com/ironcore-dev/ironcore/api/storage/v1alpha1"
	irimachine "github.com/ironcore-dev/ironcore/iri/apis/machine"
	iri "github.com/ironcore-dev/ironcore/iri/apis/machine/v1alpha1"
	irimeta "github.com/ironcore-dev/ironcore/iri/apis/meta/v1alpha1"
	poolletutils "github.com/ironcore-dev/ironcore/poollet/common/utils"
	"github.com/ironcore-dev/ironcore/poollet/machinepoollet/api/v1alpha1"
	machinepoolletclient "github.com/ironcore-dev/ironcore/poollet/machinepoollet/client"
	"github.com/ironcore-dev/ironcore/poollet/machinepoollet/controllers/events"
	"github.com/ironcore-dev/ironcore/poollet/machinepoollet/mcm"
	utilclient "github.com/ironcore-dev/ironcore/utils/client"
	utilsmaps "github.com/ironcore-dev/ironcore/utils/maps"
	"github.com/ironcore-dev/ironcore/utils/predicates"

	corev1 "k8s.io/api/core/v1"
	apierrors "k8s.io/apimachinery/pkg/api/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/types"
	"k8s.io/client-go/tools/record"
	"k8s.io/kubectl/pkg/util/fieldpath"
	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/builder"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/controller"
	"sigs.k8s.io/controller-runtime/pkg/controller/controllerutil"
	"sigs.k8s.io/controller-runtime/pkg/handler"
	"sigs.k8s.io/controller-runtime/pkg/predicate"
)

type MachineReconciler struct {
	record.EventRecorder
	client.Client

	MachineRuntime        irimachine.RuntimeService
	MachineRuntimeName    string
	MachineRuntimeVersion string

	MachineClassMapper mcm.MachineClassMapper

	MachinePoolName string

	MachineDownwardAPILabels      map[string]string
	MachineDownwardAPIAnnotations map[string]string

	NicDownwardAPILabels      map[string]string
	NicDownwardAPIAnnotations map[string]string

	VolumeDownwardAPILabels      map[string]string
	VolumeDownwardAPIAnnotations map[string]string

	NetworkDownwardAPILabels      map[string]string
	NetworkDownwardAPIAnnotations map[string]string

	WatchFilterValue string

	MaxConcurrentReconciles int
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

func (r *MachineReconciler) Reconcile(ctx context.Context, req ctrl.Request) (ctrl.Result, error) {
	log := ctrl.LoggerFrom(ctx)
	machine := &computev1alpha1.Machine{}
	if err := r.Get(ctx, req.NamespacedName, machine); err != nil {
		if !apierrors.IsNotFound(err) {
			return ctrl.Result{}, fmt.Errorf("error getting machine %s: %w", req.NamespacedName, err)
		}
		return r.deleteGone(ctx, log, machine)
	}
	return r.reconcileExists(ctx, log, machine)
}

func (r *MachineReconciler) machineUIDLabelSelector(machineUID types.UID) map[string]string {
	return map[string]string{
		v1alpha1.MachineUIDLabel: string(machineUID),
	}
}

func (r *MachineReconciler) getIRIMachinesForMachine(ctx context.Context, machine *computev1alpha1.Machine) ([]*iri.Machine, error) {
	res, err := r.MachineRuntime.ListMachines(ctx, &iri.ListMachinesRequest{
		Filter: &iri.MachineFilter{LabelSelector: r.machineUIDLabelSelector(machine.UID)},
	})
	if err != nil {
		return nil, fmt.Errorf("error listing machines by machine uid label filter: %w", err)
	}

	return res.Machines, nil
}

func (r *MachineReconciler) getMachineByID(ctx context.Context, id string) (*iri.Machine, error) {
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

func (r *MachineReconciler) deleteMachines(ctx context.Context, log logr.Logger, machines []*iri.Machine) (bool, error) {
	var (
		errs        []error
		deletingIDs []string
	)
	for _, machine := range machines {
		machineID := machine.Metadata.Id
		log := log.WithValues("MachineID", machineID)
		log.V(1).Info("Deleting matching machine")
		if _, err := r.MachineRuntime.DeleteMachine(ctx, &iri.DeleteMachineRequest{
			MachineId: machineID,
		}); err != nil {
			if status.Code(err) != codes.NotFound {
				errs = append(errs, fmt.Errorf("error deleting machine %s: %w", machineID, err))
			} else {
				log.V(1).Info("Machine is already gone")
			}
		} else {
			log.V(1).Info("Issued machine deletion")
			deletingIDs = append(deletingIDs, machineID)
		}
	}

	switch {
	case len(errs) > 0:
		return false, fmt.Errorf("error(s) deleting matching machine(s): %v", errs)
	case len(deletingIDs) > 0:
		log.V(1).Info("Machines are still deleting", "DeletingIDs", deletingIDs)
		return false, nil
	default:
		log.V(1).Info("No machine present")
		return true, nil
	}
}

func (r *MachineReconciler) deleteGone(ctx context.Context, log logr.Logger, machine *computev1alpha1.Machine) (ctrl.Result, error) {
	log.V(1).Info("Delete gone")

	log.V(1).Info("Listing IRI machines by machine")
	machines, err := r.getIRIMachinesForMachine(ctx, machine)
	if err != nil {
		return ctrl.Result{}, fmt.Errorf("error listing machines: %w", err)
	}

	ok, err := r.deleteMachines(ctx, log, machines)
	if err != nil {
		return ctrl.Result{}, fmt.Errorf("error deleting machines: %w", err)
	}
	if !ok {
		log.V(1).Info("Not all machines are gone yet, requeueing")
		return ctrl.Result{RequeueAfter: 1}, nil
	}
	log.V(1).Info("Deleted gone")
	return ctrl.Result{}, nil
}

func (r *MachineReconciler) reconcileExists(ctx context.Context, log logr.Logger, machine *computev1alpha1.Machine) (ctrl.Result, error) {
	if !machine.DeletionTimestamp.IsZero() {
		return r.delete(ctx, log, machine)
	}
	return r.reconcile(ctx, log, machine)
}

func (r *MachineReconciler) delete(ctx context.Context, log logr.Logger, machine *computev1alpha1.Machine) (ctrl.Result, error) {
	log.V(1).Info("Delete")

	if !controllerutil.ContainsFinalizer(machine, v1alpha1.MachineFinalizer) {
		log.V(1).Info("No finalizer present, nothing to do")
		return ctrl.Result{}, nil
	}

	log.V(1).Info("Finalizer present")

	log.V(1).Info("Deleting IRI machine for machine")
	res, err := r.deleteGone(ctx, log, machine)
	if err != nil {
		return ctrl.Result{}, fmt.Errorf("error deleting machines: %w", err)
	}
	if !res.IsZero() {
		log.V(1).Info("Not all machines are gone, requeueing")
		return res, nil
	}
	log.V(1).Info("Deleted iri machines for machine, removing finalizer")
	if err := clientutils.PatchRemoveFinalizer(ctx, r.Client, machine, v1alpha1.MachineFinalizer); err != nil {
		return ctrl.Result{}, fmt.Errorf("error removing finalizer: %w", err)
	}

	log.V(1).Info("Deleted")
	return ctrl.Result{}, nil
}

func (r *MachineReconciler) reconcile(ctx context.Context, log logr.Logger, machine *computev1alpha1.Machine) (ctrl.Result, error) {
	log.V(1).Info("Reconcile")

	log.V(1).Info("Ensuring finalizer")
	modified, err := clientutils.PatchEnsureFinalizer(ctx, r.Client, machine, v1alpha1.MachineFinalizer)
	if err != nil {
		return ctrl.Result{}, fmt.Errorf("error ensuring finalizer: %w", err)
	}
	if modified {
		log.V(1).Info("Added finalizer, requeueing")
		return ctrl.Result{RequeueAfter: 1}, nil
	}
	log.V(1).Info("Finalizer is present")

	log.V(1).Info("Ensuring no reconcile annotation")
	modified, err = utilclient.PatchEnsureNoReconcileAnnotation(ctx, r.Client, machine)
	if err != nil {
		return ctrl.Result{}, fmt.Errorf("error ensuring no reconcile annotation: %w", err)
	}
	if modified {
		log.V(1).Info("Removed reconcile annotation, requeueing")
		return ctrl.Result{RequeueAfter: 1}, nil
	}

	nics, err := r.getNetworkInterfacesForMachine(ctx, machine)
	if err != nil {
		return ctrl.Result{}, fmt.Errorf("error getting network interfaces for machine: %w", err)
	}

	volumes, err := r.getVolumesForMachine(ctx, machine)
	if err != nil {
		return ctrl.Result{}, fmt.Errorf("error getting volumes for machine: %w", err)
	}

	iriMachines, err := r.getIRIMachinesForMachine(ctx, machine)
	if err != nil {
		return ctrl.Result{}, fmt.Errorf("error getting IRI machines for machine: %w", err)
	}

	switch len(iriMachines) {
	case 0:
		return r.create(ctx, log, machine, nics, volumes)
	case 1:
		iriMachine := iriMachines[0]
		return r.update(ctx, log, machine, iriMachine, nics, volumes)
	default:
		panic("unhandled: multiple IRI machines")
	}
}

func (r *MachineReconciler) iriMachineLabels(machine *computev1alpha1.Machine) (map[string]string, error) {
	labels := map[string]string{
		v1alpha1.MachineUIDLabel:       string(machine.UID),
		v1alpha1.MachineNamespaceLabel: machine.Namespace,
		v1alpha1.MachineNameLabel:      machine.Name,
	}

	apiLabels, err := poolletutils.PrepareDownwardAPILabels(machine, r.MachineDownwardAPILabels, v1alpha1.MachineDownwardAPIPrefix)
	if err != nil {
		return nil, err
	}
	labels = utilsmaps.AppendMap(labels, apiLabels)
	return labels, nil
}

func (r *MachineReconciler) iriMachineAnnotations(
	machine *computev1alpha1.Machine,
	iriMachineGeneration int64,
	nicMappings map[string]v1alpha1.ObjectUIDRef,
) (map[string]string, error) {
	nicMappingString, err := v1alpha1.EncodeNetworkInterfaceMapping(nicMappings)
	if err != nil {
		return nil, err
	}

	annotations := map[string]string{
		v1alpha1.MachineGenerationAnnotation:       strconv.FormatInt(machine.Generation, 10),
		v1alpha1.IRIMachineGenerationAnnotation:    strconv.FormatInt(iriMachineGeneration, 10),
		v1alpha1.NetworkInterfaceMappingAnnotation: nicMappingString,
	}

	for name, fieldPath := range r.MachineDownwardAPIAnnotations {
		value, err := fieldpath.ExtractFieldPathAsString(machine, fieldPath)
		if err != nil {
			return nil, fmt.Errorf("error extracting downward api annotation %q: %w", name, err)
		}

		annotations[poolletutils.DownwardAPIAnnotation(v1alpha1.MachineDownwardAPIPrefix, name)] = value
	}

	return annotations, nil
}

func (r *MachineReconciler) create(
	ctx context.Context,
	log logr.Logger,
	machine *computev1alpha1.Machine,
	nics []networkingv1alpha1.NetworkInterface,
	volumes []storagev1alpha1.Volume,
) (ctrl.Result, error) {
	log.V(1).Info("Create")

	log.V(1).Info("Getting machine config")
	iriMachine, ok, err := r.prepareIRIMachine(ctx, log, machine, nics, volumes)
	if err != nil {
		return ctrl.Result{}, fmt.Errorf("error preparing iri machine: %w", err)
	}
	if !ok {
		log.V(1).Info("Machine is not yet ready")
		return ctrl.Result{}, nil
	}

	if machine.Spec.Image != "" { //nolint:staticcheck
		r.Eventf(machine, corev1.EventTypeWarning, "ImageRefDeprecated", "Image reference in spec.image is deprecated")
	}

	log.V(1).Info("Creating machine")
	res, err := r.MachineRuntime.CreateMachine(ctx, &iri.CreateMachineRequest{
		Machine: iriMachine,
	})
	if err != nil {
		return ctrl.Result{}, fmt.Errorf("error creating machine: %w", err)
	}
	log.V(1).Info("Created", "MachineID", res.Machine.Metadata.Id)

	log.V(1).Info("Updating status")
	if err := r.updateStatus(ctx, log, machine, res.Machine, nics); err != nil {
		return ctrl.Result{}, fmt.Errorf("error updating machine status: %w", err)
	}

	log.V(1).Info("Created")
	return ctrl.Result{}, nil
}

func (r *MachineReconciler) getMachineGeneration(iriMachine *iri.Machine) (int64, error) {
	return getAndParseFromStringMap(iriMachine.GetMetadata().GetAnnotations(),
		v1alpha1.MachineGenerationAnnotation,
		parseInt64,
	)
}

func (r *MachineReconciler) getIRIMachineGeneration(iriMachine *iri.Machine) (int64, error) {
	return getAndParseFromStringMap(iriMachine.GetMetadata().GetAnnotations(),
		v1alpha1.IRIMachineGenerationAnnotation,
		parseInt64,
	)
}

func (r *MachineReconciler) getNetworkInterfaceMapping(iriMachine *iri.Machine) (map[string]v1alpha1.ObjectUIDRef, error) {
	return getAndParseFromStringMap(iriMachine.GetMetadata().GetAnnotations(),
		v1alpha1.NetworkInterfaceMappingAnnotation,
		v1alpha1.DecodeNetworkInterfaceMapping,
	)
}

func (r *MachineReconciler) updateStatus(
	ctx context.Context,
	log logr.Logger,
	machine *computev1alpha1.Machine,
	iriMachine *iri.Machine,
	nics []networkingv1alpha1.NetworkInterface,
) error {
	requiredIRIGeneration, err := r.getIRIMachineGeneration(iriMachine)
	if err != nil {
		return err
	}

	iriGeneration := iriMachine.Metadata.Generation
	observedIRIGeneration := iriMachine.Status.ObservedGeneration

	if observedIRIGeneration < requiredIRIGeneration {
		log.V(1).Info("IRI machine was not observed at the latest generation",
			"IRIGeneration", iriGeneration,
			"ObservedIRIGeneration", observedIRIGeneration,
			"RequiredIRIGeneration", requiredIRIGeneration,
		)
		return nil
	}

	var errs []error

	if err := r.updateMachineStatus(ctx, machine, iriMachine); err != nil {
		errs = append(errs, err)
	}
	if err := r.updateNetworkInterfaceStatus(ctx, iriMachine, nics); err != nil {
		errs = append(errs, err)
	}

	return errors.Join(errs...)
}

func (r *MachineReconciler) updateNetworkInterfaceProviderID(
	ctx context.Context,
	nic *networkingv1alpha1.NetworkInterface,
	providerID string,
) error {
	base := nic.DeepCopy()
	nic.Spec.ProviderID = providerID
	if err := r.Patch(ctx, nic, client.StrategicMergeFrom(base)); err != nil {
		return err
	}
	return nil
}

func (r *MachineReconciler) updateNetworkInterfaceStatus(
	ctx context.Context,
	iriMachine *iri.Machine,
	nics []networkingv1alpha1.NetworkInterface,
) error {
	nicMapping, err := r.getNetworkInterfaceMapping(iriMachine)
	if err != nil {
		return err
	}

	var (
		unhandledNicByUID = utilclient.ObjectStructSliceToObjectByUIDMap[*networkingv1alpha1.NetworkInterface](nics)
		errs              []error
	)

	for _, iriNicStatus := range iriMachine.GetStatus().GetNetworkInterfaces() {
		ref, ok := nicMapping[iriNicStatus.Name]
		if !ok {
			continue
		}

		nic, ok := utilsmaps.Pop(unhandledNicByUID, ref.UID)
		if !ok {
			continue
		}

		if nic.Spec.ProviderID != iriNicStatus.Handle {
			if err := r.updateNetworkInterfaceProviderID(ctx, nic, iriNicStatus.Handle); err != nil {
				errs = append(errs, err)
			}
		}
	}

	for _, nic := range unhandledNicByUID {
		if nic.Spec.ProviderID != "" {
			if err := r.updateNetworkInterfaceProviderID(ctx, nic, ""); err != nil {
				errs = append(errs, err)
			}
		}
	}

	return errors.Join(errs...)
}

func (r *MachineReconciler) updateMachineStatus(ctx context.Context, machine *computev1alpha1.Machine, iriMachine *iri.Machine) error {
	now := metav1.Now()

	generation, err := r.getMachineGeneration(iriMachine)
	if err != nil {
		return err
	}

	machineID := poolletutils.MakeID(r.MachineRuntimeName, iriMachine.Metadata.Id)

	state, err := r.convertIRIMachineState(iriMachine.Status.State)
	if err != nil {
		return err
	}

	volumeStatuses, err := r.getVolumeStatusesForMachine(machine, iriMachine, now)
	if err != nil {
		return fmt.Errorf("error getting volume statuses: %w", err)
	}

	nicStatuses, err := r.getNetworkInterfaceStatusesForMachine(machine, iriMachine, now)
	if err != nil {
		return fmt.Errorf("error getting network interface statuses: %w", err)
	}

	base := machine.DeepCopy()

	machine.Status.State = state
	machine.Status.MachineID = machineID.String()
	machine.Status.ObservedGeneration = generation
	machine.Status.Volumes = volumeStatuses
	machine.Status.NetworkInterfaces = nicStatuses
	machine.Status.Conditions = r.computeMachineConditions(state, volumeStatuses, nicStatuses, now)

	if err := r.Status().Patch(ctx, machine, client.MergeFrom(base)); err != nil {
		return fmt.Errorf("error patching status: %w", err)
	}
	return nil
}

// computeMachineConditions computes the conditions for the machine based on its current state.
func (r *MachineReconciler) computeMachineConditions(
	state computev1alpha1.MachineState,
	volumeStatuses []computev1alpha1.VolumeStatus,
	nicStatuses []computev1alpha1.NetworkInterfaceStatus,
	now metav1.Time,
) []computev1alpha1.MachineCondition {
	var conditions []computev1alpha1.MachineCondition

	conditions = append(conditions, r.computeMachineReadyCondition(state, now))

	if len(volumeStatuses) > 0 {
		if c := r.computeVolumesReadyCondition(volumeStatuses, now); c.Type != "" {
			conditions = append(conditions, c)
		}
	}

	if len(nicStatuses) > 0 {
		if c := r.computeNetworkInterfacesReadyCondition(nicStatuses, now); c.Type != "" {
			conditions = append(conditions, c)
		}
	}

	return conditions
}

func (r *MachineReconciler) computeMachineReadyCondition(state computev1alpha1.MachineState, now metav1.Time) computev1alpha1.MachineCondition {
	status, reason, message := corev1.ConditionFalse, "NotReady", "Machine is not ready"

	switch state {
	case computev1alpha1.MachineStateRunning:
		status, reason, message = corev1.ConditionTrue, "Running", "Machine is running"
	case computev1alpha1.MachineStatePending:
		status, reason, message = corev1.ConditionFalse, "Pending", "Machine is pending"
	case computev1alpha1.MachineStateTerminating, computev1alpha1.MachineStateTerminated:
		status, reason, message = corev1.ConditionFalse, "Terminating", "Machine is terminating or terminated"
	}

	return computev1alpha1.MachineCondition{
		Type:               "Ready",
		Status:             status,
		Reason:             reason,
		Message:            message,
		LastTransitionTime: now,
	}
}

func (r *MachineReconciler) computeVolumesReadyCondition(volumeStatuses []computev1alpha1.VolumeStatus, now metav1.Time) computev1alpha1.MachineCondition {
	if len(volumeStatuses) == 0 {
		return computev1alpha1.MachineCondition{}
	}

	status, reason, message := corev1.ConditionTrue, "VolumesReady", "All volumes are ready"

	for _, vs := range volumeStatuses {
		if vs.State != computev1alpha1.VolumeStateAttached {
			status = corev1.ConditionFalse
			reason = fmt.Sprintf("VolumeNotReady: %s", vs.Name)
			message = fmt.Sprintf("Volume %s is not attached (state: %s)", vs.Name, vs.State)
			break
		}
	}

	return computev1alpha1.MachineCondition{
		Type:               computev1alpha1.MachineConditionType("VolumesReady"),
		Status:             status,
		Reason:             reason,
		Message:            message,
		LastTransitionTime: now,
	}
}

func (r *MachineReconciler) computeNetworkInterfacesReadyCondition(nicStatuses []computev1alpha1.NetworkInterfaceStatus, now metav1.Time) computev1alpha1.MachineCondition {
	if len(nicStatuses) == 0 {
		return computev1alpha1.MachineCondition{}
	}

	status, reason, message := corev1.ConditionTrue, "NetworkInterfacesReady", "All network interfaces are ready"

	for _, nicStatus := range nicStatuses {
		if nicStatus.State != computev1alpha1.NetworkInterfaceStateAttached {
			status = corev1.ConditionFalse
			reason = fmt.Sprintf("NetworkInterfaceNotReady: %s", nicStatus.Name)
			message = fmt.Sprintf("Network interface %s is not attached (state: %s)", nicStatus.Name, nicStatus.State)
			break
		}
	}

	return computev1alpha1.MachineCondition{
		Type:               computev1alpha1.MachineConditionType("NetworkInterfacesReady"),
		Status:             status,
		Reason:             reason,
		Message:            message,
		LastTransitionTime: now,
	}
}

func (r *MachineReconciler) prepareIRIPower(power computev1alpha1.Power) (iri.Power, error) {
	switch power {
	case computev1alpha1.PowerOn:
		return iri.Power_POWER_ON, nil
	case computev1alpha1.PowerOff:
		return iri.Power_POWER_OFF, nil
	default:
		return 0, fmt.Errorf("unknown power %q", power)
	}
}

func (r *MachineReconciler) updateIRIPower(ctx context.Context, log logr.Logger, machine *computev1alpha1.Machine, iriMachine *iri.Machine) error {
	actualPower := iriMachine.Spec.Power
	desiredPower, err := r.prepareIRIPower(machine.Spec.Power)
	if err != nil {
		return fmt.Errorf("error preparing iri power state: %w", err)
	}

	if actualPower == desiredPower {
		log.V(1).Info("Power is up-to-date", "Power", actualPower)
		return nil
	}

	if _, err := r.MachineRuntime.UpdateMachinePower(ctx, &iri.UpdateMachinePowerRequest{
		MachineId: iriMachine.Metadata.Id,
		Power:     desiredPower,
	}); err != nil {
		return fmt.Errorf("error updating machine power state: %w", err)
	}
	return nil
}

func (r *MachineReconciler) update(
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

func (r *MachineReconciler) updateIRIAnnotations(
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

var iriMachineStateToMachineState = map[iri.MachineState]computev1alpha1.MachineState{
	iri.MachineState_MACHINE_PENDING:     computev1alpha1.MachineStatePending,
	iri.MachineState_MACHINE_RUNNING:     computev1alpha1.MachineStateRunning,
	iri.MachineState_MACHINE_SUSPENDED:   computev1alpha1.MachineStateShutdown,
	iri.MachineState_MACHINE_TERMINATED:  computev1alpha1.MachineStateTerminated,
	iri.MachineState_MACHINE_TERMINATING: computev1alpha1.MachineStateTerminating,
}

func (r *MachineReconciler) convertIRIMachineState(state iri.MachineState) (computev1alpha1.MachineState, error) {
	if res, ok := iriMachineStateToMachineState[state]; ok {
		return res, nil
	}
	return "", fmt.Errorf("unknown machine state %v", state)
}

func (r *MachineReconciler) prepareIRIMachineClass(ctx context.Context, machine *computev1alpha1.Machine, machineClassName string) (string, bool, error) {
	machineClass := &computev1alpha1.MachineClass{}
	machineClassKey := client.ObjectKey{Name: machineClassName}
	if err := r.Get(ctx, machineClassKey, machineClass); err != nil {
		if !apierrors.IsNotFound(err) {
			return "", false, fmt.Errorf("error getting machine class: %w", err)
		}

		r.Eventf(machine, corev1.EventTypeNormal, events.MachineClassNotReady, "Machine class %s is not ready: %v", machineClassName, err)
		return "", false, nil
	}

	caps := getIRIMachineClassCapabilities(machineClass)

	class, _, err := r.MachineClassMapper.GetMachineClassFor(ctx, machineClassName, caps)
	if err != nil {
		return "", false, fmt.Errorf("error getting matching machine class: %w", err)
	}
	return class.Name, true, nil
}

func getIRIMachineClassCapabilities(machineClass *computev1alpha1.MachineClass) *iri.MachineClassCapabilities {
	resources := map[string]int64{}
	resourceList := machineClass.Capabilities

	for resource, quantity := range resourceList {
		resources[string(resource)] = quantity.Value()
	}

	return &iri.MachineClassCapabilities{
		Resources: resources,
	}
}

func (r *MachineReconciler) prepareIRIIgnitionData(ctx context.Context, machine *computev1alpha1.Machine, ignitionRef *commonv1alpha1.SecretKeySelector) ([]byte, bool, error) {
	ignitionSecret := &corev1.Secret{}
	ignitionSecretKey := client.ObjectKey{Namespace: machine.Namespace, Name: ignitionRef.Name}
	if err := r.Get(ctx, ignitionSecretKey, ignitionSecret); err != nil {
		if !apierrors.IsNotFound(err) {
			return nil, false, err
		}

		r.Eventf(machine, corev1.EventTypeNormal, events.IgnitionNotReady, "Ignition not ready: %v", err)
		return nil, false, nil
	}

	ignitionKey := ignitionRef.Key
	if ignitionKey == "" {
		ignitionKey = computev1alpha1.DefaultIgnitionKey
	}

	data, ok := ignitionSecret.Data[ignitionKey]
	if !ok {
		r.Eventf(machine, corev1.EventTypeNormal, events.IgnitionNotReady, "Ignition has no data at key %s", ignitionKey)
		return nil, false, nil
	}

	return data, true, nil
}

func (r *MachineReconciler) prepareIRIMachine(
	ctx context.Context,
	log logr.Logger,
	machine *computev1alpha1.Machine,
	nics []networkingv1alpha1.NetworkInterface,
	volumes []storagev1alpha1.Volume,
) (*iri.Machine, bool, error) {
	var (
		ok   = true
		errs []error
	)

	class, classOK, err := r.prepareIRIMachineClass(ctx, machine, machine.Spec.MachineClassRef.Name)
	switch {
	case err != nil:
		errs = append(errs, fmt.Errorf("error preparing iri machine class: %w", err))
	case !classOK:
		ok = false
	}

	var ignitionData []byte
	if ignitionRef := machine.Spec.IgnitionRef; ignitionRef != nil {
		data, ignitionSpecOK, err := r.prepareIRIIgnitionData(ctx, machine, ignitionRef)
		switch {
		case err != nil:
			errs = append(errs, fmt.Errorf("error preparing iri ignition spec: %w", err))
		case !ignitionSpecOK:
			ok = false
		default:
			ignitionData = data
		}
	}

	machineNics, machineNicMappings, machineNicsOK, err := r.prepareIRINetworkInterfacesForMachine(ctx, machine, nics)
	switch {
	case err != nil:
		errs = append(errs, fmt.Errorf("error preparing iri machine network interfaces: %w", err))
	case !machineNicsOK:
		ok = false
	}

	machineVolumes, machineVolumesOK, err := r.prepareIRIVolumes(ctx, log, machine, volumes)
	switch {
	case err != nil:
		errs = append(errs, fmt.Errorf("error preparing iri machine volumes: %w", err))
	case !machineVolumesOK:
		ok = false
	}

	labels, err := r.iriMachineLabels(machine)
	if err != nil {
		errs = append(errs, fmt.Errorf("error preparing iri machine labels: %w", err))
	}

	annotations, err := r.iriMachineAnnotations(machine, 0, machineNicMappings)
	if err != nil {
		errs = append(errs, fmt.Errorf("error preparing iri machine annotations: %w", err))
	}

	switch {
	case len(errs) > 0:
		return nil, false, fmt.Errorf("error(s) preparing machine: %v", errs)
	case !ok:
		return nil, false, nil
	default:
		return &iri.Machine{
			Metadata: &irimeta.ObjectMetadata{
				Labels:      labels,
				Annotations: annotations,
			},
			Spec: &iri.MachineSpec{
				Class:             class,
				IgnitionData:      ignitionData,
				Volumes:           machineVolumes,
				NetworkInterfaces: machineNics,
			},
		}, true, nil
	}
}

func MachineRunsInMachinePool(machine *computev1alpha1.Machine, machinePoolName string) bool {
	machinePoolRef := machine.Spec.MachinePoolRef
	if machinePoolRef == nil {
		return false
	}

	return machinePoolRef.Name == machinePoolName
}

func MachineRunsInMachinePoolPredicate(machinePoolName string) predicate.Predicate {
	return predicate.NewPredicateFuncs(func(object client.Object) bool {
		machine := object.(*computev1alpha1.Machine)
		return MachineRunsInMachinePool(machine, machinePoolName)
	})
}

func (r *MachineReconciler) matchingWatchLabel() client.ListOption {
	var labels map[string]string
	if r.WatchFilterValue != "" {
		labels = map[string]string{
			commonv1alpha1.WatchLabel: r.WatchFilterValue,
		}
	}
	return client.MatchingLabels(labels)
}

func (r *MachineReconciler) enqueueMachinesReferencingVolume() handler.EventHandler {
	return handler.EnqueueRequestsFromMapFunc(func(ctx context.Context, obj client.Object) []ctrl.Request {
		volume := obj.(*storagev1alpha1.Volume)
		log := ctrl.LoggerFrom(ctx)

		machineList := &computev1alpha1.MachineList{}
		if err := r.List(ctx, machineList,
			client.InNamespace(volume.Namespace),
			client.MatchingFields{
				machinepoolletclient.MachineSpecVolumeNamesField: volume.Name,
			},
			r.matchingWatchLabel(),
		); err != nil {
			log.Error(err, "Error listing machines using volume", "VolumeKey", client.ObjectKeyFromObject(volume))
			return nil
		}

		return utilclient.ReconcileRequestsFromObjectStructSlice[*computev1alpha1.Machine](machineList.Items)
	})
}

func (r *MachineReconciler) enqueueMachinesReferencingSecret() handler.EventHandler {
	return handler.EnqueueRequestsFromMapFunc(func(ctx context.Context, obj client.Object) []ctrl.Request {
		secret := obj.(*corev1.Secret)
		log := ctrl.LoggerFrom(ctx)

		machineList := &computev1alpha1.MachineList{}
		if err := r.List(ctx, machineList,
			client.InNamespace(secret.Namespace),
			client.MatchingFields{
				machinepoolletclient.MachineSpecSecretNamesField: secret.Name,
			},
			r.matchingWatchLabel(),
		); err != nil {
			log.Error(err, "Error listing machines using secret", "SecretKey", client.ObjectKeyFromObject(secret))
			return nil
		}

		return utilclient.ReconcileRequestsFromObjectStructSlice[*computev1alpha1.Machine](machineList.Items)
	})
}

func (r *MachineReconciler) enqueueMachinesReferencingNetworkInterface() handler.EventHandler {
	return handler.EnqueueRequestsFromMapFunc(func(ctx context.Context, obj client.Object) []ctrl.Request {
		nic := obj.(*networkingv1alpha1.NetworkInterface)
		log := ctrl.LoggerFrom(ctx)

		machineList := &computev1alpha1.MachineList{}
		if err := r.List(ctx, machineList,
			client.InNamespace(nic.Namespace),
			client.MatchingFields{
				machinepoolletclient.MachineSpecNetworkInterfaceNamesField: nic.Name,
			},
			r.matchingWatchLabel(),
		); err != nil {
			log.Error(err, "Error listing machines using secret", "NetworkInterfaceKey", client.ObjectKeyFromObject(nic))
			return nil
		}

		return utilclient.ReconcileRequestsFromObjectStructSlice[*computev1alpha1.Machine](machineList.Items)
	})
}

func (r *MachineReconciler) SetupWithManager(mgr ctrl.Manager) error {
	log := ctrl.Log.WithName("machinepoollet")

	return ctrl.NewControllerManagedBy(mgr).
		For(
			&computev1alpha1.Machine{},
			builder.WithPredicates(
				MachineRunsInMachinePoolPredicate(r.MachinePoolName),
				predicates.ResourceHasFilterLabel(log, r.WatchFilterValue),
				predicates.ResourceIsNotExternallyManaged(log),
			),
		).
		Watches(
			&corev1.Secret{},
			r.enqueueMachinesReferencingSecret(),
		).
		Watches(
			&networkingv1alpha1.NetworkInterface{},
			r.enqueueMachinesReferencingNetworkInterface(),
		).
		Watches(
			&storagev1alpha1.Volume{},
			r.enqueueMachinesReferencingVolume(),
		).
		WithOptions(
			controller.Options{
				MaxConcurrentReconciles: r.MaxConcurrentReconciles,
			}).
		Complete(r)
}
