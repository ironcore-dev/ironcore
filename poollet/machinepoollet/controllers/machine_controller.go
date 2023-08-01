// Copyright 2022 OnMetal authors
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
//      http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.

package controllers

import (
	"context"
	"fmt"
	"strconv"

	"github.com/go-logr/logr"
	"github.com/onmetal/controller-utils/clientutils"
	commonv1alpha1 "github.com/onmetal/onmetal-api/api/common/v1alpha1"
	computev1alpha1 "github.com/onmetal/onmetal-api/api/compute/v1alpha1"
	networkingv1alpha1 "github.com/onmetal/onmetal-api/api/networking/v1alpha1"
	storagev1alpha1 "github.com/onmetal/onmetal-api/api/storage/v1alpha1"
	orimachine "github.com/onmetal/onmetal-api/ori/apis/machine"
	ori "github.com/onmetal/onmetal-api/ori/apis/machine/v1alpha1"
	orimeta "github.com/onmetal/onmetal-api/ori/apis/meta/v1alpha1"
	machinepoolletv1alpha1 "github.com/onmetal/onmetal-api/poollet/machinepoollet/api/v1alpha1"
	machinepoolletclient "github.com/onmetal/onmetal-api/poollet/machinepoollet/client"
	"github.com/onmetal/onmetal-api/poollet/machinepoollet/controllers/events"
	machinepoolletmachine "github.com/onmetal/onmetal-api/poollet/machinepoollet/machine"
	"github.com/onmetal/onmetal-api/poollet/machinepoollet/mcm"
	onmetalapiclient "github.com/onmetal/onmetal-api/utils/client"
	"github.com/onmetal/onmetal-api/utils/predicates"
	"golang.org/x/exp/maps"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
	corev1 "k8s.io/api/core/v1"
	apierrors "k8s.io/apimachinery/pkg/api/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/types"
	"k8s.io/client-go/tools/record"
	"k8s.io/kubectl/pkg/util/fieldpath"
	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/builder"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/controller/controllerutil"
	"sigs.k8s.io/controller-runtime/pkg/handler"
	"sigs.k8s.io/controller-runtime/pkg/predicate"
)

type MachineReconciler struct {
	record.EventRecorder
	client.Client

	MachineRuntime        orimachine.RuntimeService
	MachineRuntimeName    string
	MachineRuntimeVersion string

	MachineClassMapper mcm.MachineClassMapper

	MachinePoolName string

	DownwardAPILabels      map[string]string
	DownwardAPIAnnotations map[string]string

	WatchFilterValue string
}

func (r *MachineReconciler) machineKeyLabelSelector(machineKey client.ObjectKey) map[string]string {
	return map[string]string{
		machinepoolletv1alpha1.MachineNamespaceLabel: machineKey.Namespace,
		machinepoolletv1alpha1.MachineNameLabel:      machineKey.Name,
	}
}

func (r *MachineReconciler) machineUIDLabelSelector(machineUID types.UID) map[string]string {
	return map[string]string{
		machinepoolletv1alpha1.MachineUIDLabel: string(machineUID),
	}
}

//+kubebuilder:rbac:groups="",resources=events,verbs=create;patch
//+kubebuilder:rbac:groups="",resources=secrets,verbs=get;list;watch
//+kubebuilder:rbac:groups=compute.api.onmetal.de,resources=machines,verbs=get;list;watch;update;patch
//+kubebuilder:rbac:groups=compute.api.onmetal.de,resources=machines/status,verbs=get;update;patch
//+kubebuilder:rbac:groups=compute.api.onmetal.de,resources=machines/finalizers,verbs=update
//+kubebuilder:rbac:groups=storage.api.onmetal.de,resources=volumes,verbs=get;list;watch
//+kubebuilder:rbac:groups=networking.api.onmetal.de,resources=networkinterfaces,verbs=get;list;watch
//+kubebuilder:rbac:groups=networking.api.onmetal.de,resources=networks,verbs=get;list;watch
//+kubebuilder:rbac:groups=networking.api.onmetal.de,resources=virtualips,verbs=get;list;watch
//+kubebuilder:rbac:groups=networking.api.onmetal.de,resources=loadbalancers,verbs=get;list;watch
//+kubebuilder:rbac:groups=networking.api.onmetal.de,resources=loadbalancerroutings,verbs=get;list;watch
//+kubebuilder:rbac:groups=networking.api.onmetal.de,resources=natgateways,verbs=get;list;watch

func (r *MachineReconciler) Reconcile(ctx context.Context, req ctrl.Request) (ctrl.Result, error) {
	log := ctrl.LoggerFrom(ctx)
	machine := &computev1alpha1.Machine{}
	if err := r.Get(ctx, req.NamespacedName, machine); err != nil {
		if !apierrors.IsNotFound(err) {
			return ctrl.Result{}, fmt.Errorf("error getting machine %s: %w", req.NamespacedName, err)
		}
		return r.deleteGone(ctx, log, req.NamespacedName)
	}
	return r.reconcileExists(ctx, log, machine)
}

func (r *MachineReconciler) listMachinesByMachineUID(ctx context.Context, machineUID types.UID) ([]*ori.Machine, error) {
	res, err := r.MachineRuntime.ListMachines(ctx, &ori.ListMachinesRequest{
		Filter: &ori.MachineFilter{LabelSelector: r.machineUIDLabelSelector(machineUID)},
	})
	if err != nil {
		return nil, fmt.Errorf("error listing machines by machine uid: %w", err)
	}
	return res.Machines, nil
}

func (r *MachineReconciler) listMachinesByMachineKey(ctx context.Context, machineKey client.ObjectKey) ([]*ori.Machine, error) {
	res, err := r.MachineRuntime.ListMachines(ctx, &ori.ListMachinesRequest{
		Filter: &ori.MachineFilter{LabelSelector: r.machineKeyLabelSelector(machineKey)},
	})
	if err != nil {
		return nil, fmt.Errorf("error listing machines by machine key: %w", err)
	}
	return res.Machines, nil
}

func (r *MachineReconciler) getMachineByID(ctx context.Context, id string) (*ori.Machine, error) {
	res, err := r.MachineRuntime.ListMachines(ctx, &ori.ListMachinesRequest{
		Filter: &ori.MachineFilter{Id: id},
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

func (r *MachineReconciler) deleteMachines(ctx context.Context, log logr.Logger, machines []*ori.Machine) (bool, error) {
	var (
		errs        []error
		deletingIDs []string
	)
	for _, machine := range machines {
		machineID := machine.Metadata.Id
		log := log.WithValues("MachineID", machineID)
		log.V(1).Info("Deleting matching machine")
		if _, err := r.MachineRuntime.DeleteMachine(ctx, &ori.DeleteMachineRequest{
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

func (r *MachineReconciler) deleteGone(ctx context.Context, log logr.Logger, machineKey client.ObjectKey) (ctrl.Result, error) {
	log.V(1).Info("Delete gone")

	log.V(1).Info("Listing machines by machine key")
	machines, err := r.listMachinesByMachineKey(ctx, machineKey)
	if err != nil {
		return ctrl.Result{}, fmt.Errorf("error listing machines: %w", err)
	}

	ok, err := r.deleteMachines(ctx, log, machines)
	if err != nil {
		return ctrl.Result{}, fmt.Errorf("error deleting machines: %w", err)
	}
	if !ok {
		log.V(1).Info("Not all machines are gone yet, requeueing")
		return ctrl.Result{Requeue: true}, nil
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

	if !controllerutil.ContainsFinalizer(machine, machinepoolletv1alpha1.MachineFinalizer) {
		log.V(1).Info("No finalizer present, nothing to do")
		return ctrl.Result{}, nil
	}

	log.V(1).Info("Finalizer present")

	log.V(1).Info("Deleting machines by UID")
	ok, err := r.deleteMachinesByMachineUID(ctx, log, machine.UID)
	if err != nil {
		return ctrl.Result{}, fmt.Errorf("error deleting machines: %w", err)
	}
	if !ok {
		log.V(1).Info("Not all machines are gone, requeueing")
		return ctrl.Result{Requeue: true}, nil
	}

	log.V(1).Info("Deleted ori machines by UID, removing finalizer")
	if err := clientutils.PatchRemoveFinalizer(ctx, r.Client, machine, machinepoolletv1alpha1.MachineFinalizer); err != nil {
		return ctrl.Result{}, fmt.Errorf("error removing finalizer: %w", err)
	}

	log.V(1).Info("Deleted")
	return ctrl.Result{}, nil
}

func (r *MachineReconciler) deleteMachinesByMachineUID(ctx context.Context, log logr.Logger, machineUID types.UID) (bool, error) {
	log.V(1).Info("Listing machines")
	res, err := r.MachineRuntime.ListMachines(ctx, &ori.ListMachinesRequest{
		Filter: &ori.MachineFilter{
			LabelSelector: map[string]string{
				machinepoolletv1alpha1.MachineUIDLabel: string(machineUID),
			},
		},
	})
	if err != nil {
		return false, fmt.Errorf("error listing machines: %w", err)
	}

	log.V(1).Info("Listed machines", "NoOfMachines", len(res.Machines))
	var (
		errs               []error
		deletingMachineIDs []string
	)
	for _, machine := range res.Machines {
		machineID := machine.Metadata.Id
		log := log.WithValues("MachineID", machineID)
		log.V(1).Info("Deleting machine")
		_, err := r.MachineRuntime.DeleteMachine(ctx, &ori.DeleteMachineRequest{
			MachineId: machineID,
		})
		if err != nil {
			if status.Code(err) != codes.NotFound {
				errs = append(errs, fmt.Errorf("error deleting machine %s: %w", machineID, err))
			} else {
				log.V(1).Info("Machine is already gone")
			}
		} else {
			log.V(1).Info("Issued machine deletion")
			deletingMachineIDs = append(deletingMachineIDs, machineID)
		}
	}

	switch {
	case len(errs) > 0:
		return false, fmt.Errorf("error(s) deleting machine(s): %v", errs)
	case len(deletingMachineIDs) > 0:
		log.V(1).Info("Machines are in deletion", "DeletingMachineIDs", deletingMachineIDs)
		return false, nil
	default:
		log.V(1).Info("All machines are gone")
		return true, nil
	}
}

func (r *MachineReconciler) reconcile(ctx context.Context, log logr.Logger, machine *computev1alpha1.Machine) (ctrl.Result, error) {
	log.V(1).Info("Reconcile")

	log.V(1).Info("Ensuring finalizer")
	modified, err := clientutils.PatchEnsureFinalizer(ctx, r.Client, machine, machinepoolletv1alpha1.MachineFinalizer)
	if err != nil {
		return ctrl.Result{}, fmt.Errorf("error ensuring finalizer: %w", err)
	}
	if modified {
		log.V(1).Info("Added finalizer, requeueing")
		return ctrl.Result{Requeue: true}, nil
	}
	log.V(1).Info("Finalizer is present")

	log.V(1).Info("Ensuring no reconcile annotation")
	modified, err = onmetalapiclient.PatchEnsureNoReconcileAnnotation(ctx, r.Client, machine)
	if err != nil {
		return ctrl.Result{}, fmt.Errorf("error ensuring no reconcile annotation: %w", err)
	}
	if modified {
		log.V(1).Info("Removed reconcile annotation, requeueing")
		return ctrl.Result{Requeue: true}, nil
	}

	log.V(1).Info("Listing machines")
	machines, err := r.listMachinesByMachineUID(ctx, machine.UID)
	if err != nil {
		return ctrl.Result{}, fmt.Errorf("error listing machines: %w", err)
	}

	switch len(machines) {
	case 0:
		return r.create(ctx, log, machine)
	case 1:
		runtimeMachine := machines[0]
		return r.update(ctx, log, machine, runtimeMachine)
	default:
		panic("unhandled: multiple machines")
	}
}

func (r *MachineReconciler) oriMachineLabels(machine *computev1alpha1.Machine) (map[string]string, error) {
	annotations := map[string]string{
		machinepoolletv1alpha1.MachineUIDLabel:       string(machine.UID),
		machinepoolletv1alpha1.MachineNamespaceLabel: machine.Namespace,
		machinepoolletv1alpha1.MachineNameLabel:      machine.Name,
	}

	for name, fieldPath := range r.DownwardAPILabels {
		value, err := fieldpath.ExtractFieldPathAsString(machine, fieldPath)
		if err != nil {
			return nil, fmt.Errorf("error extracting downward api label %q: %w", name, err)
		}

		annotations[machinepoolletv1alpha1.DownwardAPILabel(name)] = value
	}
	return annotations, nil
}

func (r *MachineReconciler) oriMachineAnnotations(machine *computev1alpha1.Machine) (map[string]string, error) {
	annotations := map[string]string{
		machinepoolletv1alpha1.MachineGenerationAnnotation: strconv.FormatInt(machine.Generation, 10),
	}

	for name, fieldPath := range r.DownwardAPIAnnotations {
		value, err := fieldpath.ExtractFieldPathAsString(machine, fieldPath)
		if err != nil {
			return nil, fmt.Errorf("error extracting downward api annotation %q: %w", name, err)
		}

		annotations[machinepoolletv1alpha1.DownwardAPIAnnotation(name)] = value
	}

	return annotations, nil
}

func (r *MachineReconciler) create(ctx context.Context, log logr.Logger, machine *computev1alpha1.Machine) (ctrl.Result, error) {
	log.V(1).Info("Create")

	log.V(1).Info("Getting machine config")
	oriMachine, ok, err := r.prepareORIMachine(ctx, log, machine)
	if err != nil {
		return ctrl.Result{}, fmt.Errorf("error preparing ori machine: %w", err)
	}
	if !ok {
		log.V(1).Info("Machine is not yet ready")
		return ctrl.Result{}, nil
	}

	log.V(1).Info("Creating machine")
	res, err := r.MachineRuntime.CreateMachine(ctx, &ori.CreateMachineRequest{
		Machine: oriMachine,
	})
	if err != nil {
		return ctrl.Result{}, fmt.Errorf("error creating machine: %w", err)
	}
	log.V(1).Info("Created", "MachineID", res.Machine.Metadata.Id)

	log.V(1).Info("Updating status")
	if err := r.updateStatus(ctx, log, machine, res.Machine); err != nil {
		return ctrl.Result{}, fmt.Errorf("error updating machine status: %w", err)
	}

	log.V(1).Info("Created")
	return ctrl.Result{}, nil
}

func (r *MachineReconciler) getMachineGeneration(oriMachine *ori.Machine) (int64, error) {
	observedGenerationData, ok := oriMachine.Metadata.Annotations[machinepoolletv1alpha1.MachineGenerationAnnotation]
	if !ok {
		return 0, fmt.Errorf("ori machine has no machine generation data at %s", machinepoolletv1alpha1.MachineGenerationAnnotation)
	}

	generation, err := strconv.ParseInt(observedGenerationData, 10, 64)
	if err != nil {
		return 0, fmt.Errorf("error parsing machine generation annotation %s data %s: %w",
			machinepoolletv1alpha1.MachineGenerationAnnotation,
			observedGenerationData,
			err,
		)
	}

	return generation, nil
}

func (r *MachineReconciler) updateStatus(ctx context.Context, log logr.Logger, machine *computev1alpha1.Machine, oriMachine *ori.Machine) error {
	if actualORIGeneration, observedORIGeneration := oriMachine.Metadata.Generation, oriMachine.Status.ObservedGeneration; actualORIGeneration != observedORIGeneration {
		log.V(1).Info("ORI machine was not observed at the latest generation",
			"ActualGeneration", actualORIGeneration,
			"ObservedGeneration", observedORIGeneration,
		)
		return nil
	}

	base := machine.DeepCopy()
	now := metav1.Now()

	generation, err := r.getMachineGeneration(oriMachine)
	if err != nil {
		return err
	}

	machineID := machinepoolletmachine.MakeID(r.MachineRuntimeName, oriMachine.Metadata.Id)
	machine.Status.MachineID = machineID.String()
	machine.Status.ObservedGeneration = generation

	state, err := r.convertORIMachineState(oriMachine.Status.State)
	if err != nil {
		return err
	}

	machine.Status.State = state

	volumeStatuses, err := r.getVolumeStatusesForMachine(machine, oriMachine, now)
	if err != nil {
		return fmt.Errorf("error getting volume statuses: %w", err)
	}

	nicStatuses, err := r.getNetworkInterfaceStatusesForMachine(machine, oriMachine, now)
	if err != nil {
		return fmt.Errorf("error getting network interface statuses: %w", err)
	}

	machine.Status.Volumes = volumeStatuses
	machine.Status.NetworkInterfaces = nicStatuses

	if err := r.Status().Patch(ctx, machine, client.MergeFrom(base)); err != nil {
		return fmt.Errorf("error patching status: %w", err)
	}
	return nil
}

func (r *MachineReconciler) prepareORIPower(power computev1alpha1.Power) (ori.Power, error) {
	switch power {
	case computev1alpha1.PowerOn:
		return ori.Power_POWER_ON, nil
	case computev1alpha1.PowerOff:
		return ori.Power_POWER_OFF, nil
	default:
		return 0, fmt.Errorf("unknown power %q", power)
	}
}

func (r *MachineReconciler) updateORIPower(ctx context.Context, log logr.Logger, machine *computev1alpha1.Machine, oriMachine *ori.Machine) error {
	actualPower := oriMachine.Spec.Power
	desiredPower, err := r.prepareORIPower(machine.Spec.Power)
	if err != nil {
		return fmt.Errorf("error preparing ori power state: %w", err)
	}

	if actualPower == desiredPower {
		log.V(1).Info("Power is up-to-date", "Power", actualPower)
		return nil
	}

	if _, err := r.MachineRuntime.UpdateMachinePower(ctx, &ori.UpdateMachinePowerRequest{
		MachineId: oriMachine.Metadata.Id,
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
	oriMachine *ori.Machine,
) (ctrl.Result, error) {
	log.V(1).Info("Updating existing machine")

	var errs []error

	log.V(1).Info("Updating volumes")
	if err := r.updateORIVolumes(ctx, log, machine, oriMachine); err != nil {
		errs = append(errs, fmt.Errorf("error updating volumes: %w", err))
	}

	log.V(1).Info("Updating network interfaces")
	if err := r.updateORINetworkInterfaces(ctx, log, machine, oriMachine); err != nil {
		errs = append(errs, fmt.Errorf("error updating network interfaces: %w", err))
	}

	log.V(1).Info("Updating power state")
	if err := r.updateORIPower(ctx, log, machine, oriMachine); err != nil {
		errs = append(errs, fmt.Errorf("error updating power state: %w", err))
	}

	if len(errs) > 0 {
		return ctrl.Result{}, fmt.Errorf("error(s) updating machine: %v", errs)
	}

	log.V(1).Info("Updating annotations")
	if err := r.updateORIAnnotations(ctx, log, machine, oriMachine); err != nil {
		return ctrl.Result{}, fmt.Errorf("error updating annotations: %w", err)
	}

	log.V(1).Info("Getting ori machine")
	m, err := r.getMachineByID(ctx, oriMachine.Metadata.Id)
	if err != nil {
		return ctrl.Result{}, fmt.Errorf("error getting ori machine: %w", err)
	}
	oriMachine = m

	log.V(1).Info("Updating machine status")
	if err := r.updateStatus(ctx, log, machine, oriMachine); err != nil {
		return ctrl.Result{}, fmt.Errorf("error updating status")
	}

	log.V(1).Info("Updated existing machine")
	return ctrl.Result{}, nil
}

func (r *MachineReconciler) updateORIAnnotations(ctx context.Context, log logr.Logger, machine *computev1alpha1.Machine, oriMachine *ori.Machine) error {
	desiredAnnotations, err := r.oriMachineAnnotations(machine)
	if err != nil {
		return fmt.Errorf("error getting ori machine annotations: %w", err)
	}

	actualAnnotations := oriMachine.Metadata.Annotations

	if !maps.Equal(desiredAnnotations, actualAnnotations) {
		log.V(1).Info("Annotations are up-to-date", "Annotations", desiredAnnotations)
		return nil
	}

	if _, err := r.MachineRuntime.UpdateMachineAnnotations(ctx, &ori.UpdateMachineAnnotationsRequest{
		MachineId:   oriMachine.Metadata.Id,
		Annotations: desiredAnnotations,
	}); err != nil {
		return fmt.Errorf("error updating machine annotations: %w", err)
	}
	return nil
}

var oriMachineStateToMachineState = map[ori.MachineState]computev1alpha1.MachineState{
	ori.MachineState_MACHINE_PENDING:    computev1alpha1.MachineStatePending,
	ori.MachineState_MACHINE_RUNNING:    computev1alpha1.MachineStateRunning,
	ori.MachineState_MACHINE_SUSPENDED:  computev1alpha1.MachineStateShutdown,
	ori.MachineState_MACHINE_TERMINATED: computev1alpha1.MachineStateTerminated,
}

func (r *MachineReconciler) convertORIMachineState(state ori.MachineState) (computev1alpha1.MachineState, error) {
	if res, ok := oriMachineStateToMachineState[state]; ok {
		return res, nil
	}
	return "", fmt.Errorf("unknown machine state %v", state)
}

func (r *MachineReconciler) prepareORIMachineClass(ctx context.Context, machine *computev1alpha1.Machine, machineClassName string) (string, bool, error) {
	machineClass := &computev1alpha1.MachineClass{}
	machineClassKey := client.ObjectKey{Name: machineClassName}
	if err := r.Get(ctx, machineClassKey, machineClass); err != nil {
		if !apierrors.IsNotFound(err) {
			return "", false, fmt.Errorf("error getting machine class: %w", err)
		}

		r.Eventf(machine, corev1.EventTypeNormal, events.MachineClassNotReady, "Machine class %s is not ready: %v", machineClassName, err)
		return "", false, nil
	}

	caps, err := getORIMachineClassCapabilities(machineClass)
	if err != nil {
		return "", false, fmt.Errorf("error getting ori machine class capabilities: %w", err)
	}

	class, _, err := r.MachineClassMapper.GetMachineClassFor(ctx, machineClassName, caps)
	if err != nil {
		return "", false, fmt.Errorf("error getting matching machine class: %w", err)
	}
	return class.Name, true, nil
}

func getORIMachineClassCapabilities(machineClass *computev1alpha1.MachineClass) (*ori.MachineClassCapabilities, error) {
	cpu := machineClass.Capabilities.CPU()
	memory := machineClass.Capabilities.Memory()

	return &ori.MachineClassCapabilities{
		CpuMillis:   cpu.MilliValue(),
		MemoryBytes: uint64(memory.Value()),
	}, nil
}

func (r *MachineReconciler) prepareORIIgnitionData(ctx context.Context, machine *computev1alpha1.Machine, ignitionRef *commonv1alpha1.SecretKeySelector) ([]byte, bool, error) {
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

func (r *MachineReconciler) prepareORIMachine(ctx context.Context, log logr.Logger, machine *computev1alpha1.Machine) (*ori.Machine, bool, error) {
	var (
		ok   = true
		errs []error
	)

	class, classOK, err := r.prepareORIMachineClass(ctx, machine, machine.Spec.MachineClassRef.Name)
	switch {
	case err != nil:
		errs = append(errs, fmt.Errorf("error preparing ori machine class: %w", err))
	case !classOK:
		ok = false
	}

	var imageSpec *ori.ImageSpec
	if image := machine.Spec.Image; image != "" {
		imageSpec = &ori.ImageSpec{
			Image: image,
		}
	}

	var ignitionData []byte
	if ignitionRef := machine.Spec.IgnitionRef; ignitionRef != nil {
		data, ignitionSpecOK, err := r.prepareORIIgnitionData(ctx, machine, ignitionRef)
		switch {
		case err != nil:
			errs = append(errs, fmt.Errorf("error preparing ori ignition spec: %w", err))
		case !ignitionSpecOK:
			ok = false
		default:
			ignitionData = data
		}
	}

	machineNics, nonReadyNics, err := r.prepareORINetworkInterfaces(ctx, log, machine)
	switch {
	case err != nil:
		errs = append(errs, fmt.Errorf("error preparing ori machine network interfaces: %w", err))
	case len(nonReadyNics) > 0:
		ok = false
	}

	machineVolumes, machineVolumesOK, err := r.prepareORIVolumes(ctx, log, machine)
	switch {
	case err != nil:
		errs = append(errs, fmt.Errorf("error preparing ori machine volumes: %w", err))
	case !machineVolumesOK:
		ok = false
	}

	labels, err := r.oriMachineLabels(machine)
	if err != nil {
		errs = append(errs, fmt.Errorf("error preparing ori machine labels: %w", err))
	}

	annotations, err := r.oriMachineAnnotations(machine)
	if err != nil {
		errs = append(errs, fmt.Errorf("error preparing ori machine annotations: %w", err))
	}

	switch {
	case len(errs) > 0:
		return nil, false, fmt.Errorf("error(s) preparing machine: %v", errs)
	case !ok:
		return nil, false, nil
	default:
		return &ori.Machine{
			Metadata: &orimeta.ObjectMetadata{
				Labels:      labels,
				Annotations: annotations,
			},
			Spec: &ori.MachineSpec{
				Image:             imageSpec,
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

		return onmetalapiclient.ReconcileRequestsFromObjectSlice(machineList.Items)
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

		return onmetalapiclient.ReconcileRequestsFromObjectSlice(machineList.Items)
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

		return onmetalapiclient.ReconcileRequestsFromObjectSlice(machineList.Items)
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
		Complete(r)
}
