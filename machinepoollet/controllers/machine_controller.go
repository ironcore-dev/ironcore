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
	"net/netip"

	"github.com/go-logr/logr"
	"github.com/onmetal/controller-utils/clientutils"
	commonv1alpha1 "github.com/onmetal/onmetal-api/apis/common/v1alpha1"
	computev1alpha1 "github.com/onmetal/onmetal-api/apis/compute/v1alpha1"
	networkingv1alpha1 "github.com/onmetal/onmetal-api/apis/networking/v1alpha1"
	storagev1alpha1 "github.com/onmetal/onmetal-api/apis/storage/v1alpha1"
	onmetalapiclient "github.com/onmetal/onmetal-api/apiutils/client"
	machinepoolletclient "github.com/onmetal/onmetal-api/machinepoollet/client"
	"github.com/onmetal/onmetal-api/machinepoollet/controllers/events"
	"github.com/onmetal/onmetal-api/machinepoollet/mleg"
	ori "github.com/onmetal/onmetal-api/ori/apis/runtime/v1alpha1"
	"golang.org/x/exp/slices"
	corev1 "k8s.io/api/core/v1"
	apierrors "k8s.io/apimachinery/pkg/api/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/client-go/tools/record"
	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/builder"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/controller/controllerutil"
	"sigs.k8s.io/controller-runtime/pkg/handler"
	"sigs.k8s.io/controller-runtime/pkg/manager"
	"sigs.k8s.io/controller-runtime/pkg/predicate"
	"sigs.k8s.io/controller-runtime/pkg/source"
)

const (
	MachineUIDLabel       = "machinepoollet.compute.api.onmetal.de/machine-uid"
	MachineNamespaceLabel = "machinepoollet.compute.api.onmetal.de/machine-namespace"
	MachineNameLabel      = "machinepoollet.compute.api.onmetal.de/machine-name"
)

type MachineReconciler struct {
	record.EventRecorder
	client.Client

	MachineRuntime ori.MachineRuntimeClient

	MachineFinalizer string
	MachinePoolName  string
}

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

func (r *MachineReconciler) deleteGone(ctx context.Context, log logr.Logger, machineKey client.ObjectKey) (ctrl.Result, error) {
	log.V(1).Info("Delete gone")

	log.V(1).Info("Listing machines matching key")
	res, err := r.MachineRuntime.ListMachines(ctx, &ori.ListMachinesRequest{
		Filter: &ori.MachineFilter{
			LabelSelector: map[string]string{
				MachineNamespaceLabel: machineKey.Namespace,
				MachineNameLabel:      machineKey.Name,
			},
		},
	})
	if err != nil {
		return ctrl.Result{}, fmt.Errorf("error listing machines matching key: %w", err)
	}

	log.V(1).Info("Listed machines matching key", "NoOfMachines", len(res.Machines))
	var errs []error
	for _, machine := range res.Machines {
		log := log.WithValues("MachineID", machine.Id)
		log.V(1).Info("Deleting matching machine")
		if _, err := r.MachineRuntime.DeleteMachine(ctx, &ori.DeleteMachineRequest{
			MachineId: machine.Id,
		}); err != nil {
			errs = append(errs, fmt.Errorf("error deleting machine %s: %w", machine.Id, err))
		}
	}

	if len(errs) > 0 {
		return ctrl.Result{}, fmt.Errorf("error(s) deleting matching machine(s): %v", errs)
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

	if !controllerutil.ContainsFinalizer(machine, r.MachineFinalizer) {
		log.V(1).Info("No finalizer present, nothing to do")
		return ctrl.Result{}, nil
	}

	log.V(1).Info("Finalizer present")

	log.V(1).Info("Listing machines")
	res, err := r.MachineRuntime.ListMachines(ctx, &ori.ListMachinesRequest{
		Filter: &ori.MachineFilter{
			LabelSelector: map[string]string{
				MachineUIDLabel: string(machine.UID),
			},
		},
	})
	if err != nil {
		return ctrl.Result{}, fmt.Errorf("error listing machines: %w", err)
	}

	log.V(1).Info("Listed machines", "NoOfMachines", len(res.Machines))
	var errs []error
	for _, machine := range res.Machines {
		log := log.WithValues("MachineID", machine.Id)
		log.V(1).Info("Deleting machine")
		_, err := r.MachineRuntime.DeleteMachine(ctx, &ori.DeleteMachineRequest{
			MachineId: machine.Id,
		})
		if err != nil {
			errs = append(errs, fmt.Errorf("error deleting machine %s: %w", machine.Id, err))
		}
	}

	if len(errs) > 0 {
		return ctrl.Result{}, fmt.Errorf("error(s) deleting machines: %v", errs)
	}

	log.V(1).Info("Deleted all runtime machines, removing finalizer")
	if err := clientutils.PatchRemoveFinalizer(ctx, r.Client, machine, r.MachineFinalizer); err != nil {
		return ctrl.Result{}, fmt.Errorf("error removing finalizer: %w", err)
	}

	log.V(1).Info("Deleted")
	return ctrl.Result{}, nil
}

func (r *MachineReconciler) reconcile(ctx context.Context, log logr.Logger, machine *computev1alpha1.Machine) (ctrl.Result, error) {
	log.V(1).Info("Reconcile")

	log.V(1).Info("Ensuring finalizer")
	modified, err := clientutils.PatchEnsureFinalizer(ctx, r.Client, machine, r.MachineFinalizer)
	if err != nil {
		return ctrl.Result{}, fmt.Errorf("error ensuring finalizer: %w", err)
	}
	if modified {
		log.V(1).Info("Added finalizer, requeueing")
		return ctrl.Result{}, nil
	}
	log.V(1).Info("Finalizer is present")

	log.V(1).Info("Ensuring no reconcile annotation")
	modified, err = onmetalapiclient.PatchEnsureNoReconcileAnnotation(ctx, r.Client, machine)
	if err != nil {
		return ctrl.Result{}, fmt.Errorf("error ensuring no reconcile annotation: %w", err)
	}
	if modified {
		log.V(1).Info("Removed reconcile annotation, requeueing")
		return ctrl.Result{}, nil
	}

	log.V(1).Info("Listing machines")
	res, err := r.MachineRuntime.ListMachines(ctx, &ori.ListMachinesRequest{
		Filter: &ori.MachineFilter{
			LabelSelector: map[string]string{
				MachineUIDLabel: string(machine.UID),
			},
		},
	})
	if err != nil {
		return ctrl.Result{}, fmt.Errorf("error listing machines: %w", err)
	}

	switch len(res.Machines) {
	case 0:
		return r.create(ctx, log, machine)
	case 1:
		runtimeMachine := res.Machines[0]
		return r.update(ctx, log, machine, runtimeMachine)
	default:
		panic("unhandled: multiple machines")
	}
}

func (r *MachineReconciler) create(ctx context.Context, log logr.Logger, machine *computev1alpha1.Machine) (ctrl.Result, error) {
	log.V(1).Info("Create")

	log.V(1).Info("Getting machine config")
	machineConfig, err := r.getMachineConfig(ctx, log, machine)
	if err != nil {
		err = fmt.Errorf("error getting machine config: %w", err)
		return ctrl.Result{}, IgnoreDependencyNotReadyError(err)
	}

	log.V(1).Info("Creating machine")
	res, err := r.MachineRuntime.CreateMachine(ctx, &ori.CreateMachineRequest{
		Config: machineConfig,
	})
	if err != nil {
		return ctrl.Result{}, fmt.Errorf("error creating machine: %w", err)
	}
	log.V(1).Info("Created", "MachineID", res.Machine.Id)

	log.V(1).Info("Patching status")
	if err := r.patchStatus(
		ctx,
		machine,
		computev1alpha1.MachineStatePending,
		machine.Status.NetworkInterfaces,
		machine.Status.Volumes,
	); err != nil {
		return ctrl.Result{}, fmt.Errorf("error patching machine status: %w", err)
	}

	log.V(1).Info("Created")
	return ctrl.Result{Requeue: true}, nil
}

func (r *MachineReconciler) update(
	ctx context.Context,
	log logr.Logger,
	machine *computev1alpha1.Machine,
	runtimeMachine *ori.Machine,
) (ctrl.Result, error) {
	log.V(1).Info("Updating existing machine")

	log.V(1).Info("Getting runtime machine status")
	res, err := r.MachineRuntime.MachineStatus(ctx, &ori.MachineStatusRequest{
		MachineId: runtimeMachine.Id,
	})
	if err != nil {
		return ctrl.Result{}, fmt.Errorf("error getting runtime machine status: %w", err)
	}

	var errs []error

	log.V(1).Info("Reconciling network interfaces")
	networkInterfaceStates, err := r.reconcileNetworkInterfaces(ctx, log, machine, runtimeMachine.Id, res.Status)
	if err != nil {
		errs = append(errs, err)
	}

	log.V(1).Info("Reconciling volumes")
	volumeStates, err := r.reconcileVolumes(ctx, log, machine, runtimeMachine.Id, res.Status)
	if err != nil {
		errs = append(errs, err)
	}

	log.V(1).Info("Patching status")
	if err := r.patchStatus(
		ctx,
		machine,
		ORIMachineStateToComputeV1Alpha1MachineState(res.Status.State),
		networkInterfaceStates,
		volumeStates,
	); err != nil {
		if len(errs) > 0 {
			log.V(1).Info("Error(s) reconciling machine", "Errors", errs)
		}
		return ctrl.Result{}, fmt.Errorf("error patching status: %w", err)
	}

	if len(errs) > 0 {
		return ctrl.Result{}, fmt.Errorf("error(s) reconciling machine: %v", errs)
	}

	log.V(1).Info("Updated existing machine")
	return ctrl.Result{}, nil
}

func (r *MachineReconciler) patchStatus(
	ctx context.Context,
	machine *computev1alpha1.Machine,
	state computev1alpha1.MachineState,
	networkInterfaceStates []computev1alpha1.NetworkInterfaceStatus,
	volumeStates []computev1alpha1.VolumeStatus,
) error {
	base := machine.DeepCopy()
	machine.Status.State = state
	machine.Status.NetworkInterfaces = networkInterfaceStates
	machine.Status.Volumes = volumeStates
	if err := r.Status().Patch(ctx, machine, client.MergeFrom(base)); err != nil {
		return fmt.Errorf("error patching machine status: %w", err)
	}
	return nil
}

func (r *MachineReconciler) reconcileNetworkInterfaces(
	ctx context.Context,
	log logr.Logger,
	machine *computev1alpha1.Machine,
	machineID string,
	status *ori.MachineStatus,
) ([]computev1alpha1.NetworkInterfaceStatus, error) {
	specNetworkInterfaceByName := GroupBy(machine.Spec.NetworkInterfaces, func(v computev1alpha1.NetworkInterface) string { return v.Name })
	statusNetworkInterfaceByName := GroupBy(status.NetworkInterfaces, func(v *ori.NetworkInterfaceStatus) string { return v.Name })
	runtimeNetworkInterfaceStateByName := make(map[string]networkInterfaceStatus)
	var errs []error

	for name, specNetworkInterface := range specNetworkInterfaceByName {
		log := log.WithValues("NetworkInterfaceName", name)

		var status *ori.NetworkInterfaceStatus
		if statusNetworkInterface, ok := statusNetworkInterfaceByName[name]; ok {
			status = statusNetworkInterface
		}

		newStatus, err := r.applyNetworkInterface(ctx, log, machine, specNetworkInterface, machineID, status)
		if err != nil {
			if !IsDependencyNotReadyError(err) {
				r.EventRecorder.Eventf(machine, corev1.EventTypeWarning, events.ErrorApplyingNetworkInterface, "Error applying network interface %s: %v", name, err)
				errs = append(errs, fmt.Errorf("[network interface %s] %w", name, err))
			} else {
				r.EventRecorder.Eventf(machine, corev1.EventTypeNormal, events.NetworkInterfaceNotReady, "Network interface %s not ready: %v", name, err)
			}
		}

		runtimeNetworkInterfaceStateByName[name] = newStatus
	}

	for name, statusNetworkInterface := range statusNetworkInterfaceByName {
		log := log.WithValues("NetworkInterfaceName", name)

		if _, ok := specNetworkInterfaceByName[name]; ok {
			continue
		}

		if err := r.detachNetworkInterface(ctx, log, machineID, name); err != nil {
			r.EventRecorder.Eventf(machine, corev1.EventTypeWarning, events.ErrorDetachingNetworkInterface, "Error detaching network interface %s: %v", name, err)
			errs = append(errs, fmt.Errorf("error detaching network interface %s: %w", name, err))
			runtimeNetworkInterfaceStateByName[name] = networkInterfaceStatusFromStatus(statusNetworkInterface)
		}
	}

	states := make([]computev1alpha1.NetworkInterfaceStatus, 0, len(machine.Status.NetworkInterfaces))
	for _, state := range machine.Status.NetworkInterfaces {
		status, ok := runtimeNetworkInterfaceStateByName[state.Name]
		if ok {
			state.IPs = NetIPAddrsToCommonV1Alpha1IPs(status.IPs)
			if status.VirtualIP != nil {
				state.VirtualIP = &commonv1alpha1.IP{Addr: *status.VirtualIP}
			}
		}

		states = append(states, state)
	}
	if len(errs) > 0 {
		return states, fmt.Errorf("error(s) reconciling network interfaces: %v", errs)
	}
	return states, nil
}

type networkInterfaceStatus struct {
	IPs       []netip.Addr
	VirtualIP *netip.Addr
}

func networkInterfaceStatusFromStatus(status *ori.NetworkInterfaceStatus) networkInterfaceStatus {
	var virtualIP *netip.Addr
	if status.VirtualIp != nil {
		ip := netip.MustParseAddr(status.VirtualIp.Ip) // TODO: Remove Must
		virtualIP = &ip
	}

	var ips []netip.Addr
	for _, ip := range status.Ips {
		ips = append(ips, netip.MustParseAddr(ip)) // TODO: Remove Must
	}

	return networkInterfaceStatus{
		IPs:       ips,
		VirtualIP: virtualIP,
	}
}

func networkInterfaceStatusFromConfig(config *ori.NetworkInterfaceConfig) networkInterfaceStatus {
	var virtualIP *netip.Addr
	if config.VirtualIp != nil {
		ip := netip.MustParseAddr(config.VirtualIp.Ip) // TODO: Remove Must
		virtualIP = &ip
	}

	var ips []netip.Addr
	for _, ip := range config.Ips {
		ips = append(ips, netip.MustParseAddr(ip)) // TODO: Remove must
	}

	return networkInterfaceStatus{
		IPs:       ips,
		VirtualIP: virtualIP,
	}
}

func (r *MachineReconciler) networkInterfaceUpToDate(config ori.NetworkInterfaceConfig, status ori.NetworkInterfaceStatus) bool {
	if !r.networkInterfaceNetworkCompatible(config, status.Network) {
		return false
	}

	if !slices.Equal(config.Ips, status.Ips) {
		return false
	}

	if config.VirtualIp != nil {
		if status.VirtualIp == nil {
			return false
		}

		if config.VirtualIp.Ip != status.VirtualIp.Ip {
			return false
		}
	}

	return true
}

func (r *MachineReconciler) networkInterfaceNetworkCompatible(config ori.NetworkInterfaceConfig, status *ori.NetworkStatus) bool {
	return config.Network.Name == status.Name &&
		config.Network.Uid == status.Uid
}

func (r *MachineReconciler) detachNetworkInterface(
	ctx context.Context,
	log logr.Logger,
	machineID string,
	networkInterfaceName string,
) error {
	log.V(1).Info("Detaching network interface")
	_, err := r.MachineRuntime.DetachNetworkInterface(ctx, &ori.DetachNetworkInterfaceRequest{
		MachineId:            machineID,
		NetworkInterfaceName: networkInterfaceName,
	})
	if err != nil {
		return fmt.Errorf("error detaching network interface: %w", err)
	}
	return nil
}

func (r *MachineReconciler) applyNetworkInterface(
	ctx context.Context,
	log logr.Logger,
	machine *computev1alpha1.Machine,
	networkInterface computev1alpha1.NetworkInterface,
	machineID string,
	status *ori.NetworkInterfaceStatus,
) (networkInterfaceStatus, error) {
	log.V(1).Info("Getting network interface config")
	config, err := GetORIMachineNetworkInterfaceConfig(ctx, r.Client, machine, &networkInterface)
	if err != nil {
		err = fmt.Errorf("error getting machine network interface config: %w", err)
		if status != nil {
			return networkInterfaceStatusFromStatus(status), err
		}

		return networkInterfaceStatus{}, err
	}

	if status != nil {
		if r.networkInterfaceUpToDate(*config, *status) {
			log.V(1).Info("Network interface is up-to-date")
			return networkInterfaceStatusFromStatus(status), nil
		}

		if r.networkInterfaceNetworkCompatible(*config, status.Network) {
			log.V(1).Info("Network interface is not up-to-date but network is compatible, updating")
			_, err := r.MachineRuntime.UpdateNetworkInterface(ctx, &ori.UpdateNetworkInterfaceRequest{
				MachineId:            machineID,
				NetworkInterfaceName: networkInterface.Name,
				Ips:                  config.Ips,
				VirtualIp:            config.VirtualIp,
			})
			if err != nil {
				err = fmt.Errorf("error updating network interface: %w", err)
				return networkInterfaceStatusFromStatus(status), err
			}

			return networkInterfaceStatusFromConfig(config), nil
		}

		log.V(1).Info("Network interface is not up-to-date and network is incompatible, detaching")
		if _, err := r.MachineRuntime.DetachNetworkInterface(ctx, &ori.DetachNetworkInterfaceRequest{
			MachineId:            machineID,
			NetworkInterfaceName: networkInterface.Name,
		}); err != nil {
			err = fmt.Errorf("error detaching network interface: %w", err)
			return networkInterfaceStatusFromStatus(status), err
		}
	}

	log.V(1).Info("Attaching network interface")
	_, err = r.MachineRuntime.AttachNetworkInterface(ctx, &ori.AttachNetworkInterfaceRequest{
		Config: config,
	})
	if err != nil {
		err = fmt.Errorf("error attaching network interface: %w", err)
		return networkInterfaceStatus{}, err
	}
	return networkInterfaceStatusFromConfig(config), nil
}

func (r *MachineReconciler) canUpdateVolume(config ori.VolumeConfig, status ori.VolumeStatus) bool {
	if config.Device != status.Device {
		return false
	}

	switch {
	case config.Access != nil && status.Access != nil:
		return config.Access.Driver == status.Access.Driver
	case config.EmptyDisk != nil && status.EmptyDisk != nil:
		return true
	default:
		return false
	}
}

func (r *MachineReconciler) applyVolume(
	ctx context.Context,
	log logr.Logger,
	machine *computev1alpha1.Machine,
	volume computev1alpha1.Volume,
	machineID string,
	status *ori.VolumeStatus,
) (string, error) {
	log.V(1).Info("Getting volume config")
	config, err := GetORIMachineVolumeConfig(ctx, r.Client, machine, &volume)
	if err != nil {
		err = fmt.Errorf("error getting machine volume config: %w", err)
		if status != nil {
			return status.DeviceId, err
		}

		return "", err
	}

	if status != nil {
		if r.canUpdateVolume(*config, *status) {
			log.V(1).Info("Confirming volume")
			confirmRes, err := r.MachineRuntime.ConfirmVolume(ctx, &ori.ConfirmVolumeRequest{
				MachineId:  machineID,
				VolumeName: volume.Name,
				Access:     config.Access,
				EmptyDisk:  config.EmptyDisk,
			})
			if err != nil {
				err = fmt.Errorf("error confirming volume")
				return status.DeviceId, err
			}

			if confirmRes.Confirmation != nil {
				log.V(1).Info("Existing volume is valid")
				return confirmRes.Confirmation.DeviceId, nil
			}

			log.V(1).Info("Updating existing volume")
			res, err := r.MachineRuntime.UpdateVolume(ctx, &ori.UpdateVolumeRequest{
				MachineId:  machineID,
				VolumeName: volume.Name,
				Access:     config.Access,
				EmptyDisk:  config.EmptyDisk,
			})
			if err != nil {
				err = fmt.Errorf("error updating volume: %w", err)
				return status.DeviceId, err
			}
			return res.DeviceId, nil
		}

		log.V(1).Info("Detaching existing volume")
		if _, err := r.MachineRuntime.DetachVolume(ctx, &ori.DetachVolumeRequest{
			MachineId:  machineID,
			VolumeName: volume.Name,
		}); err != nil {
			err = fmt.Errorf("error detaching volume: %w", err)
			return status.DeviceId, err
		}
	}

	log.V(1).Info("Attaching volume")
	res, err := r.MachineRuntime.AttachVolume(ctx, &ori.AttachVolumeRequest{
		MachineId: machineID,
		Config:    config,
	})
	if err != nil {
		err = fmt.Errorf("error attaching volume: %w", err)
		return "", err
	}
	return res.DeviceId, nil
}

func (r *MachineReconciler) detachVolume(
	ctx context.Context,
	log logr.Logger,
	machineID string,
	volumeName string,
) error {
	log.V(1).Info("Detaching volume")
	_, err := r.MachineRuntime.DetachVolume(ctx, &ori.DetachVolumeRequest{
		MachineId:  machineID,
		VolumeName: volumeName,
	})
	if err != nil {
		return fmt.Errorf("error detaching volume: %w", err)
	}
	return nil
}

func (r *MachineReconciler) reconcileVolumes(
	ctx context.Context,
	log logr.Logger,
	machine *computev1alpha1.Machine,
	machineID string,
	status *ori.MachineStatus,
) ([]computev1alpha1.VolumeStatus, error) {
	specVolumeByName := GroupBy(machine.Spec.Volumes, func(v computev1alpha1.Volume) string { return v.Name })
	statusVolumeByName := GroupBy(status.Volumes, func(v *ori.VolumeStatus) string { return v.Name })
	deviceIDByName := make(map[string]string)
	var errs []error

	for name, specVolume := range specVolumeByName {
		log := log.WithValues("VolumeName", name)

		var status *ori.VolumeStatus
		if statusVolume, ok := statusVolumeByName[name]; ok {
			status = statusVolume
		}

		deviceID, err := r.applyVolume(ctx, log, machine, specVolume, machineID, status)
		if err != nil {
			if !IsDependencyNotReadyError(err) {
				r.EventRecorder.Eventf(machine, corev1.EventTypeWarning, events.ErrorApplyingVolume, "Error applying volume %s: %w", name, err)
				errs = append(errs, fmt.Errorf("error applying volume %s: %w", name, err))
			} else {
				r.EventRecorder.Eventf(machine, corev1.EventTypeNormal, events.VolumeNotReady, "Volume %s not ready: %w", name, err)
			}
		}

		deviceIDByName[name] = deviceID
	}

	for name, statusVolume := range statusVolumeByName {
		log := log.WithValues("VolumeName", name)

		if _, ok := specVolumeByName[name]; ok {
			continue
		}

		if err := r.detachVolume(ctx, log, machineID, name); err != nil {
			r.Eventf(machine, corev1.EventTypeWarning, events.ErrorDetachingVolume, "Error detaching volume %s: %w", name, err)
			errs = append(errs, fmt.Errorf("error detaching volume %s: %w", name, err))
			deviceIDByName[name] = statusVolume.Name
		}
	}

	states := make([]computev1alpha1.VolumeStatus, 0, len(machine.Status.Volumes))
	for _, state := range machine.Status.Volumes {
		deviceID, ok := deviceIDByName[state.Name]
		if ok {
			state.DeviceID = deviceID
		}

		states = append(states, state)
	}
	if len(errs) > 0 {
		return states, fmt.Errorf("error(s) reconciling volumes: %v", errs)
	}
	return states, nil
}

func (r *MachineReconciler) getMachineConfig(ctx context.Context, log logr.Logger, machine *computev1alpha1.Machine) (*ori.MachineConfig, error) {
	log.V(1).Info("Getting machine resources")
	machineResources, err := GetORIMachineResources(ctx, r.Client, machine)
	if err != nil {
		if !IsDependencyNotReadyError(err) {
			r.EventRecorder.Eventf(machine, corev1.EventTypeWarning, events.ErrorGettingMachineResources, "Error getting machine resources: %v", err)
			return nil, fmt.Errorf("error getting machine resources: %w", err)
		}

		r.EventRecorder.Eventf(machine, corev1.EventTypeNormal, events.MachineResourcesNotReady, "Machine resources not ready: %v", err)
		return nil, fmt.Errorf("machine resources not ready: %w", err)
	}

	var ignitionConfig *ori.IgnitionConfig
	if ignitionRef := machine.Spec.IgnitionRef; ignitionRef != nil {
		log.V(1).Info("Getting machine ignition config")
		ignitionConfig, err = GetORIIgnitionConfig(ctx, r.Client, machine, ignitionRef)
		if err != nil {
			if !IsDependencyNotReadyError(err) {
				r.EventRecorder.Eventf(machine, corev1.EventTypeWarning, events.ErrorGettingIgnitionConfig, "Error getting ignition config: %v", err)
				return nil, fmt.Errorf("error getting machine ignition config: %w", err)
			}

			r.EventRecorder.Eventf(machine, corev1.EventTypeNormal, events.IgnitionNotReady, "Ignition is not ready: %v", err)
			return nil, fmt.Errorf("ignition is not ready: %w", err)
		}
	}

	var networkInterfaceConfigs []*ori.NetworkInterfaceConfig
	for _, networkInterface := range machine.Spec.NetworkInterfaces {
		log.V(1).Info("Getting network interface config", "MachineNetworkInterfaceName", networkInterface.Name)
		networkInterfaceConfig, err := GetORIMachineNetworkInterfaceConfig(ctx, r.Client, machine, &networkInterface)
		if err != nil {
			if !IsDependencyNotReadyError(err) {
				r.EventRecorder.Eventf(machine, corev1.EventTypeWarning, events.ErrorGettingNetworkInterfaceConfig, "Error getting network interface %s config: %v", networkInterface.Name, err)
				return nil, fmt.Errorf("error getting machine network interface %s config: %w", networkInterface.Name, err)
			}

			r.EventRecorder.Eventf(machine, corev1.EventTypeNormal, events.NetworkInterfaceNotReady, "Network interface %s not ready: %v", networkInterface.Name, err)
			return nil, fmt.Errorf("network interface %s not ready: %w", networkInterface.Name, err)
		}

		networkInterfaceConfigs = append(networkInterfaceConfigs, networkInterfaceConfig)
	}

	var volumeConfigs []*ori.VolumeConfig
	for _, volume := range machine.Spec.Volumes {
		log.V(1).Info("Getting machine volume config", "MachineVolumeName", volume.Name)
		volumeConfig, err := GetORIMachineVolumeConfig(ctx, r.Client, machine, &volume)
		if err != nil {
			if !IsDependencyNotReadyError(err) {
				r.EventRecorder.Eventf(machine, corev1.EventTypeWarning, events.ErrorGettingVolumeConfig, "Error getting volume %s config: %v", volume.Name, err)
				return nil, fmt.Errorf("error getting machine volume %s config: %w", volume.Name, err)
			}

			r.EventRecorder.Eventf(machine, corev1.EventTypeNormal, events.VolumeNotReady, "Volume %s not ready: %v", volume.Name, err)
			return nil, fmt.Errorf("volume %s not ready: %w", volume.Name, err)
		}

		volumeConfigs = append(volumeConfigs, volumeConfig)
	}

	machineMetadata := ORIMachineMetadata(machine)

	return &ori.MachineConfig{
		Metadata:          machineMetadata,
		Resources:         machineResources,
		Image:             machine.Spec.Image,
		Ignition:          ignitionConfig,
		Volumes:           volumeConfigs,
		NetworkInterfaces: networkInterfaceConfigs,
	}, nil
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

func (r *MachineReconciler) enqueueMachinesReferencingVolume(ctx context.Context, log logr.Logger) handler.EventHandler {
	return handler.EnqueueRequestsFromMapFunc(func(obj client.Object) []ctrl.Request {
		volume := obj.(*storagev1alpha1.Volume)
		machineList := &computev1alpha1.MachineList{}
		if err := r.List(ctx, machineList,
			client.InNamespace(volume.Namespace),
			client.MatchingFields{machinepoolletclient.MachineSpecVolumeNamesField: volume.Name},
		); err != nil {
			log.Error(err, "Error listing machines using volume", "VolumeKey", client.ObjectKeyFromObject(volume))
			return nil
		}

		return r.makeRequestsForMachinesRunningInMachinePool(machineList)
	})
}

func (r *MachineReconciler) enqueueMachinesReferencingSecret(ctx context.Context, log logr.Logger) handler.EventHandler {
	return handler.EnqueueRequestsFromMapFunc(func(obj client.Object) []ctrl.Request {
		secret := obj.(*corev1.Secret)
		machineList := &computev1alpha1.MachineList{}
		if err := r.List(ctx, machineList,
			client.InNamespace(secret.Namespace),
			client.MatchingFields{machinepoolletclient.MachineSpecSecretNamesField: secret.Name},
		); err != nil {
			log.Error(err, "Error listing machines using secret", "SecretKey", client.ObjectKeyFromObject(secret))
			return nil
		}

		return r.makeRequestsForMachinesRunningInMachinePool(machineList)
	})
}

func (r *MachineReconciler) enqueueMachinesReferencingNetworkInterface(ctx context.Context, log logr.Logger) handler.EventHandler {
	return handler.EnqueueRequestsFromMapFunc(func(obj client.Object) []ctrl.Request {
		nic := obj.(*networkingv1alpha1.NetworkInterface)
		machineList := &computev1alpha1.MachineList{}
		if err := r.List(ctx, machineList,
			client.InNamespace(nic.Namespace),
			client.MatchingFields{machinepoolletclient.MachineSpecNetworkInterfaceNamesField: nic.Name},
		); err != nil {
			log.Error(err, "Error listing machines using secret", "NetworkInterfaceKey", client.ObjectKeyFromObject(nic))
			return nil
		}

		return r.makeRequestsForMachinesRunningInMachinePool(machineList)
	})
}

func (r *MachineReconciler) makeRequestsForMachinesRunningInMachinePool(machineList *computev1alpha1.MachineList) []ctrl.Request {
	var res []ctrl.Request
	for _, machine := range machineList.Items {
		if MachineRunsInMachinePool(&machine, r.MachinePoolName) {
			res = append(res, ctrl.Request{NamespacedName: client.ObjectKeyFromObject(&machine)})
		}
	}
	return res
}

func (r *MachineReconciler) SetupWithManager(mgr ctrl.Manager, prct ...predicate.Predicate) error {
	log := ctrl.Log.WithName("machinepoollet")
	ctx := ctrl.LoggerInto(context.TODO(), log)

	gen := mleg.NewGeneric(r.MachineRuntime, mleg.GenericOptions{})

	if err := mgr.Add(gen); err != nil {
		return fmt.Errorf("error adding machine lifecycle event generator: %w", err)
	}

	if err := mgr.Add(manager.RunnableFunc(func(ctx context.Context) error {
		log := log.WithName("machine-lifecycle-event")
		for {
			select {
			case <-ctx.Done():
				return nil
			case evt := <-gen.Watch():
				log = log.WithValues(
					"EventType", evt.Type,
					"MachineID", evt.ID,
					"Namespace", evt.Metadata.Namespace,
					"Name", evt.Metadata.Name,
					"UID", evt.Metadata.UID,
				)
				log.V(5).Info("Received lifecycle event")

				machine := &computev1alpha1.Machine{
					ObjectMeta: metav1.ObjectMeta{
						Namespace: evt.Metadata.Namespace,
						Name:      evt.Metadata.Name,
					},
				}
				if err := onmetalapiclient.PatchAddReconcileAnnotation(ctx, r.Client, machine); client.IgnoreNotFound(err) != nil {
					log.Error(err, "Error adding reconcile annotation")
				}
			}
		}
	})); err != nil {
		return fmt.Errorf("error adding machine lifecycle event handler: %w", err)
	}

	return ctrl.NewControllerManagedBy(mgr).
		For(
			&computev1alpha1.Machine{},
			builder.WithPredicates(
				append([]predicate.Predicate{
					MachineRunsInMachinePoolPredicate(r.MachinePoolName),
				}, prct...)...,
			),
		).
		Watches(
			&source.Kind{Type: &corev1.Secret{}},
			r.enqueueMachinesReferencingSecret(ctx, log),
		).
		Watches(
			&source.Kind{Type: &networkingv1alpha1.NetworkInterface{}},
			r.enqueueMachinesReferencingNetworkInterface(ctx, log),
		).
		Watches(
			&source.Kind{Type: &storagev1alpha1.Volume{}},
			r.enqueueMachinesReferencingVolume(ctx, log),
		).
		Complete(r)
}
