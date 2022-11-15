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

	"github.com/go-logr/logr"
	"github.com/onmetal/controller-utils/clientutils"
	commonv1alpha1 "github.com/onmetal/onmetal-api/apis/common/v1alpha1"
	computev1alpha1 "github.com/onmetal/onmetal-api/apis/compute/v1alpha1"
	networkingv1alpha1 "github.com/onmetal/onmetal-api/apis/networking/v1alpha1"
	storagev1alpha1 "github.com/onmetal/onmetal-api/apis/storage/v1alpha1"
	onmetalapiclient "github.com/onmetal/onmetal-api/apiutils/client"
	"github.com/onmetal/onmetal-api/apiutils/predicates"
	machinepoolletv1alpha1 "github.com/onmetal/onmetal-api/machinepoollet/api/v1alpha1"
	machinepoolletclient "github.com/onmetal/onmetal-api/machinepoollet/client"
	"github.com/onmetal/onmetal-api/machinepoollet/controllers/events"
	"github.com/onmetal/onmetal-api/machinepoollet/mleg"
	ori "github.com/onmetal/onmetal-api/ori/apis/runtime/v1alpha1"
	utilslices "github.com/onmetal/onmetal-api/utils/slices"
	"golang.org/x/exp/slices"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
	corev1 "k8s.io/api/core/v1"
	apierrors "k8s.io/apimachinery/pkg/api/errors"
	"k8s.io/apimachinery/pkg/api/resource"
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

type MachineReconciler struct {
	record.EventRecorder
	client.Client

	MachineRuntime ori.MachineRuntimeClient

	MachinePoolName string

	WatchFilterValue string
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
				machinepoolletv1alpha1.MachineNamespaceLabel: machineKey.Namespace,
				machinepoolletv1alpha1.MachineNameLabel:      machineKey.Name,
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
			if status.Code(err) != codes.NotFound {
				errs = append(errs, fmt.Errorf("error deleting machine %s: %w", machine.Id, err))
			} else {
				log.V(1).Info("Machine is already gone")
			}
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

	if !controllerutil.ContainsFinalizer(machine, machinepoolletv1alpha1.MachineFinalizer) {
		log.V(1).Info("No finalizer present, nothing to do")
		return ctrl.Result{}, nil
	}

	log.V(1).Info("Finalizer present")

	log.V(1).Info("Listing machines")
	res, err := r.MachineRuntime.ListMachines(ctx, &ori.ListMachinesRequest{
		Filter: &ori.MachineFilter{
			LabelSelector: map[string]string{
				machinepoolletv1alpha1.MachineUIDLabel: string(machine.UID),
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
			if status.Code(err) != codes.NotFound {
				errs = append(errs, fmt.Errorf("error deleting machine %s: %w", machine.Id, err))
			} else {
				log.V(1).Info("Machine is already gone")
			}
		}
	}

	if len(errs) > 0 {
		return ctrl.Result{}, fmt.Errorf("error(s) deleting machines: %v", errs)
	}

	log.V(1).Info("Deleted all runtime machines, removing finalizer")
	if err := clientutils.PatchRemoveFinalizer(ctx, r.Client, machine, machinepoolletv1alpha1.MachineFinalizer); err != nil {
		return ctrl.Result{}, fmt.Errorf("error removing finalizer: %w", err)
	}

	log.V(1).Info("Deleted")
	return ctrl.Result{}, nil
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
				machinepoolletv1alpha1.MachineUIDLabel: string(machine.UID),
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

	log.V(1).Info("Updating status")
	if err := r.updateStatus(ctx, log, machine, res.Machine.Id); err != nil {
		return ctrl.Result{}, fmt.Errorf("error updating machine status: %w", err)
	}

	log.V(1).Info("Created")
	return ctrl.Result{Requeue: true}, nil
}

func (r *MachineReconciler) updateStatus(
	ctx context.Context,
	log logr.Logger,
	machine *computev1alpha1.Machine,
	machineID string,
) error {
	log.V(1).Info("Getting runtime status")
	res, err := r.MachineRuntime.MachineStatus(ctx, &ori.MachineStatusRequest{
		MachineId: machineID,
	})
	if err != nil {
		return fmt.Errorf("error getting machine status: %w", err)
	}

	base := machine.DeepCopy()
	now := metav1.Now()
	runtimeStatus := res.Status

	runtimeVolumeStatusByName := utilslices.ToMap(runtimeStatus.Volumes, func(volumeStatus *ori.VolumeStatus) string { return volumeStatus.Name })
	for i := range machine.Status.Volumes {
		volumeStatus := &machine.Status.Volumes[i]
		runtimeVolumeStatus := runtimeVolumeStatusByName[volumeStatus.Name]
		r.updateVolumeStatus(volumeStatus, runtimeVolumeStatus, now)
		delete(runtimeVolumeStatusByName, volumeStatus.Name)
	}
	for name, runtimeVolumeStatus := range runtimeVolumeStatusByName {
		volumeStatus := &computev1alpha1.VolumeStatus{Name: name}
		r.updateVolumeStatus(volumeStatus, runtimeVolumeStatus, now)
		machine.Status.Volumes = append(machine.Status.Volumes, *volumeStatus)
	}

	runtimeNetworkInterfaceStatusByName := utilslices.ToMap(runtimeStatus.NetworkInterfaces, func(networkInterfaceStatus *ori.NetworkInterfaceStatus) string { return networkInterfaceStatus.Name })
	for i := range machine.Status.NetworkInterfaces {
		networkInterfaceStatus := &machine.Status.NetworkInterfaces[i]
		runtimeNetworkInterfaceStatus := runtimeNetworkInterfaceStatusByName[networkInterfaceStatus.Name]
		if err := r.updateNetworkInterfaceStatus(networkInterfaceStatus, runtimeNetworkInterfaceStatus, now); err != nil {
			return fmt.Errorf("error updating network interface %s status: %w", networkInterfaceStatus.Name, err)
		}
		delete(runtimeNetworkInterfaceStatusByName, networkInterfaceStatus.Name)
	}
	for name, runtimeNetworkInterfaceStatus := range runtimeNetworkInterfaceStatusByName {
		networkInterfaceStatus := &computev1alpha1.NetworkInterfaceStatus{Name: name}
		if err := r.updateNetworkInterfaceStatus(networkInterfaceStatus, runtimeNetworkInterfaceStatus, now); err != nil {
			return fmt.Errorf("error updating network interface %s status: %w", networkInterfaceStatus.Name, err)
		}
		machine.Status.NetworkInterfaces = append(machine.Status.NetworkInterfaces, *networkInterfaceStatus)
	}

	machine.Status.State = r.oriMachineStateToComputeV1Alpha1MachineState(runtimeStatus.State)

	if err := r.Status().Patch(ctx, machine, client.MergeFrom(base)); err != nil {
		return fmt.Errorf("error patching status: %w", err)
	}
	return nil
}

func (r *MachineReconciler) updateVolumeStatus(volumeStatus *computev1alpha1.VolumeStatus, runtimeVolumeStatus *ori.VolumeStatus, now metav1.Time) {
	var (
		newState   computev1alpha1.VolumeState
		device     string
		emptyDisk  *computev1alpha1.EmptyDiskVolumeStatus
		referenced *computev1alpha1.ReferencedVolumeStatus
	)

	if runtimeVolumeStatus != nil {
		newState = r.oriVolumeStateToComputeV1Alpha1VolumeState(runtimeVolumeStatus.State)
		device = runtimeVolumeStatus.Device
		if runtimeEmptyDisk := runtimeVolumeStatus.EmptyDisk; runtimeEmptyDisk != nil {
			emptyDisk = &computev1alpha1.EmptyDiskVolumeStatus{
				Size: resource.NewQuantity(int64(runtimeEmptyDisk.SizeBytes), resource.DecimalSI),
			}
		}
		if runtimeAccess := runtimeVolumeStatus.Access; runtimeAccess != nil {
			referenced = &computev1alpha1.ReferencedVolumeStatus{
				Driver: runtimeAccess.Driver,
				Handle: runtimeAccess.Handle,
			}
		}
	} else {
		newState = computev1alpha1.VolumeStateDetached
	}

	if volumeStatus.State != newState {
		volumeStatus.LastStateTransitionTime = &now
	}

	volumeStatus.State = newState
	volumeStatus.Device = device
	volumeStatus.EmptyDisk = emptyDisk
	volumeStatus.Referenced = referenced
}

func (r *MachineReconciler) updateNetworkInterfaceStatus(
	networkInterfaceStatus *computev1alpha1.NetworkInterfaceStatus,
	runtimeNetworkInterfaceStatus *ori.NetworkInterfaceStatus,
	now metav1.Time,
) error {
	var (
		newState      computev1alpha1.NetworkInterfaceState
		networkHandle string
		ips           []commonv1alpha1.IP
		virtualIP     *commonv1alpha1.IP
	)

	if runtimeNetworkInterfaceStatus != nil {
		networkHandle = runtimeNetworkInterfaceStatus.Network.Handle

		for _, ipString := range runtimeNetworkInterfaceStatus.Ips {
			ip, err := commonv1alpha1.ParseIP(ipString)
			if err != nil {
				return fmt.Errorf("error parsing ip %s: %w", ipString, err)
			}

			ips = append(ips, ip)
		}

		if runtimeVirtualIP := runtimeNetworkInterfaceStatus.VirtualIp; runtimeVirtualIP != nil {
			ip, err := commonv1alpha1.ParseIP(runtimeVirtualIP.Ip)
			if err != nil {
				return fmt.Errorf("error parsing virtual ip %s: %w", runtimeVirtualIP.Ip, err)
			}

			virtualIP = &ip
		}
	} else {
		newState = computev1alpha1.NetworkInterfaceStateDetached
	}

	if networkInterfaceStatus.State != newState {
		networkInterfaceStatus.LastStateTransitionTime = &now
	}

	networkInterfaceStatus.State = newState
	networkInterfaceStatus.NetworkHandle = networkHandle
	networkInterfaceStatus.IPs = ips
	networkInterfaceStatus.VirtualIP = virtualIP
	return nil
}

func (r *MachineReconciler) update(
	ctx context.Context,
	log logr.Logger,
	machine *computev1alpha1.Machine,
	runtimeMachine *ori.Machine,
) (ctrl.Result, error) {
	log.V(1).Info("Updating existing machine")

	var errs []error

	log.V(1).Info("Reconciling network interfaces")
	err := r.reconcileNetworkInterfaces(ctx, log, machine, runtimeMachine.Id)
	if err != nil {
		errs = append(errs, fmt.Errorf("error reconciling network interfaces: %w", err))
	}

	log.V(1).Info("Reconciling volumes")
	if err := r.reconcileVolumes(ctx, log, machine, runtimeMachine.Id); err != nil {
		errs = append(errs, fmt.Errorf("error reconciling volumes: %w", err))
	}

	log.V(1).Info("Updating status")
	if err := r.updateStatus(ctx, log, machine, runtimeMachine.Id); err != nil {
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

func (r *MachineReconciler) reconcileNetworkInterfaces(
	ctx context.Context,
	log logr.Logger,
	machine *computev1alpha1.Machine,
	machineID string,
) error {
	res, err := r.MachineRuntime.ListNetworkInterfaces(ctx, &ori.ListNetworkInterfacesRequest{
		Filter: &ori.NetworkInterfaceFilter{MachineId: machineID},
	})
	if err != nil {
		return fmt.Errorf("error listing network interfaces: %w", err)
	}

	specNetworkInterfaceByName := utilslices.ToMap(machine.Spec.NetworkInterfaces, func(v computev1alpha1.NetworkInterface) string { return v.Name })
	existingNetworkInterfaceByName := utilslices.ToMap(res.NetworkInterfaces, func(v *ori.NetworkInterface) string { return v.Name })

	var errs []error

	for name, specNetworkInterface := range specNetworkInterfaceByName {
		log := log.WithValues("NetworkInterfaceName", name)

		existingNetworkInterface := existingNetworkInterfaceByName[name]

		if err := r.applyNetworkInterface(ctx, log, machine, specNetworkInterface, machineID, existingNetworkInterface); err != nil {
			if !IsDependencyNotReadyError(err) {
				r.EventRecorder.Eventf(machine, corev1.EventTypeWarning, events.ErrorApplyingNetworkInterface, "Error applying network interface %s: %v", name, err)
				errs = append(errs, fmt.Errorf("[network interface %s] %w", name, err))
			} else {
				r.EventRecorder.Eventf(machine, corev1.EventTypeNormal, events.NetworkInterfaceNotReady, "Network interface %s not ready: %v", name, err)
			}
		}
	}

	for name := range existingNetworkInterfaceByName {
		log := log.WithValues("NetworkInterfaceName", name)

		if _, ok := specNetworkInterfaceByName[name]; ok {
			continue
		}

		if err := r.deleteNetworkInterface(ctx, log, machineID, name); err != nil {
			r.EventRecorder.Eventf(machine, corev1.EventTypeWarning, events.ErrorDetachingNetworkInterface, "Error detaching network interface %s: %v", name, err)
			errs = append(errs, fmt.Errorf("error detaching network interface %s: %w", name, err))
		}
	}

	if len(errs) > 0 {
		return fmt.Errorf("error(s) reconciling network interfaces: %v", errs)
	}
	return nil
}

func (r *MachineReconciler) isNetworkInterfaceUpToDate(config *ori.NetworkInterfaceConfig, networkInterface *ori.NetworkInterface) bool {
	if !r.networkInterfaceNetworkCompatible(config.Network, networkInterface.Network) {
		return false
	}

	if !slices.Equal(config.Ips, networkInterface.Ips) {
		return false
	}

	if config.VirtualIp != nil {
		if networkInterface.VirtualIp == nil {
			return false
		}

		if config.VirtualIp.Ip != networkInterface.VirtualIp.Ip {
			return false
		}
	}

	return true
}

func (r *MachineReconciler) networkInterfaceNetworkCompatible(config, current *ori.NetworkConfig) bool {
	return config.Handle == current.Handle
}

func (r *MachineReconciler) deleteNetworkInterface(
	ctx context.Context,
	log logr.Logger,
	machineID string,
	networkInterfaceName string,
) error {
	log.V(1).Info("Detaching network interface")
	if _, err := r.MachineRuntime.DeleteNetworkInterface(ctx, &ori.DeleteNetworkInterfaceRequest{
		MachineId:            machineID,
		NetworkInterfaceName: networkInterfaceName,
	}); err != nil {
		if status.Code(err) != codes.NotFound {
			return fmt.Errorf("error detaching network interface: %w", err)
		}
		log.V(1).Info("Network interface is already gone")
	}
	return nil
}

func (r *MachineReconciler) applyNetworkInterface(
	ctx context.Context,
	log logr.Logger,
	machine *computev1alpha1.Machine,
	networkInterface computev1alpha1.NetworkInterface,
	machineID string,
	existingNetworkInterface *ori.NetworkInterface,
) error {
	log.V(1).Info("Getting network interface config")
	config, err := r.getORINetworkInterfaceConfig(ctx, machine, &networkInterface)
	if err != nil {
		return fmt.Errorf("error getting machine network interface config: %w", err)
	}

	if existingNetworkInterface != nil {
		if r.isNetworkInterfaceUpToDate(config, existingNetworkInterface) {
			log.V(1).Info("Network interface is up-to-date")
			return nil
		}

		if r.networkInterfaceNetworkCompatible(config.Network, existingNetworkInterface.Network) {
			log.V(1).Info("Network interface is not up-to-date but network is compatible, updating")
			_, err := r.MachineRuntime.UpdateNetworkInterface(ctx, &ori.UpdateNetworkInterfaceRequest{
				MachineId:            machineID,
				NetworkInterfaceName: networkInterface.Name,
				Ips:                  config.Ips,
				VirtualIp:            config.VirtualIp,
			})
			if err != nil {
				return fmt.Errorf("error updating network interface: %w", err)
			}

			return nil
		}

		log.V(1).Info("Network interface is not up-to-date and network is incompatible, deleting")
		if _, err := r.MachineRuntime.DeleteNetworkInterface(ctx, &ori.DeleteNetworkInterfaceRequest{
			MachineId:            machineID,
			NetworkInterfaceName: networkInterface.Name,
		}); err != nil {
			err = fmt.Errorf("error deleting network interface: %w", err)
			return err
		}
	}

	log.V(1).Info("Creating network interface")
	if _, err := r.MachineRuntime.CreateNetworkInterface(ctx, &ori.CreateNetworkInterfaceRequest{
		MachineId: machineID,
		Config:    config,
	}); err != nil {
		return fmt.Errorf("error attaching network interface: %w", err)
	}
	return nil
}

func (r *MachineReconciler) isVolumeUpToDate(config *ori.VolumeConfig, volume *ori.Volume) bool {
	switch {
	case config.EmptyDisk != nil && volume.EmptyDisk != nil:
		return config.EmptyDisk.SizeLimitBytes == volume.EmptyDisk.SizeLimitBytes
	case config.Access != nil && volume.Access != nil:
		return config.Access.Driver == volume.Access.Driver && config.Access.Handle == volume.Access.Handle
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
	existingVolume *ori.Volume,
) error {
	log.V(1).Info("Getting volume config")
	config, err := r.getORIVolumeConfig(ctx, machine, &volume)
	if err != nil {
		return fmt.Errorf("error getting machine volume config: %w", err)
	}

	if existingVolume != nil {
		if r.isVolumeUpToDate(config, existingVolume) {
			log.V(1).Info("Existing volume is up-to-date")
			return nil
		}

		log.V(1).Info("Existing volume is not up-to-date, deleting")
		if _, err := r.MachineRuntime.DeleteVolume(ctx, &ori.DeleteVolumeRequest{
			MachineId:  machineID,
			VolumeName: volume.Name,
		}); err != nil {
			if status.Code(err) != codes.NotFound {
				return fmt.Errorf("error deleting volume: %w", err)
			} else {
				log.V(1).Info("Volume is already gone")
			}
		}
	}

	log.V(1).Info("Creating volume")
	if _, err := r.MachineRuntime.CreateVolume(ctx, &ori.CreateVolumeRequest{
		MachineId: machineID,
		Config:    config,
	}); err != nil {
		return fmt.Errorf("error creating volume: %w", err)
	}
	return nil
}

func (r *MachineReconciler) deleteVolume(
	ctx context.Context,
	log logr.Logger,
	machineID string,
	volumeName string,
) error {
	log.V(1).Info("Detaching volume")
	if _, err := r.MachineRuntime.DeleteVolume(ctx, &ori.DeleteVolumeRequest{
		MachineId:  machineID,
		VolumeName: volumeName,
	}); err != nil {
		if status.Code(err) != codes.NotFound {
			return fmt.Errorf("error detaching volume: %w", err)
		}
		log.V(1).Info("Volume is already gone")
	}
	return nil
}

func (r *MachineReconciler) reconcileVolumes(
	ctx context.Context,
	log logr.Logger,
	machine *computev1alpha1.Machine,
	machineID string,
) error {
	res, err := r.MachineRuntime.ListVolumes(ctx, &ori.ListVolumesRequest{
		Filter: &ori.VolumeFilter{MachineId: machineID},
	})
	if err != nil {
		return fmt.Errorf("error listing volumes for machine: %w", err)
	}

	specVolumeByName := utilslices.ToMap(machine.Spec.Volumes, func(v computev1alpha1.Volume) string { return v.Name })
	existingVolumeByName := utilslices.ToMap(res.Volumes, func(v *ori.Volume) string { return v.Name })

	var errs []error

	for name, specVolume := range specVolumeByName {
		log := log.WithValues("VolumeName", name)

		existingVolume := existingVolumeByName[name]
		if err := r.applyVolume(ctx, log, machine, specVolume, machineID, existingVolume); err != nil {
			if !IsDependencyNotReadyError(err) {
				r.EventRecorder.Eventf(machine, corev1.EventTypeWarning, events.ErrorApplyingVolume, "Error applying volume %s: %w", name, err)
				errs = append(errs, fmt.Errorf("error applying volume %s: %w", name, err))
			} else {
				r.EventRecorder.Eventf(machine, corev1.EventTypeNormal, events.VolumeNotReady, "Volume %s not ready: %w", name, err)
			}
		}
	}

	for name := range existingVolumeByName {
		log := log.WithValues("VolumeName", name)

		if _, ok := specVolumeByName[name]; ok {
			continue
		}

		if err := r.deleteVolume(ctx, log, machineID, name); err != nil {
			r.Eventf(machine, corev1.EventTypeWarning, events.ErrorDetachingVolume, "Error detaching volume %s: %w", name, err)
			errs = append(errs, fmt.Errorf("error deleting volume %s: %w", name, err))
		}
	}

	if len(errs) > 0 {
		return fmt.Errorf("error(s) reconciling volumes: %v", errs)
	}
	return nil
}

func (r *MachineReconciler) getMachineConfig(ctx context.Context, log logr.Logger, machine *computev1alpha1.Machine) (*ori.MachineConfig, error) {
	log.V(1).Info("Getting machine resources")
	machineResources, err := r.getORIMachineResources(ctx, machine)
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
		ignitionConfig, err = r.getORIIgnitionConfig(ctx, machine, ignitionRef)
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
		networkInterfaceConfig, err := r.getORINetworkInterfaceConfig(ctx, machine, &networkInterface)
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
		volumeConfig, err := r.getORIVolumeConfig(ctx, machine, &volume)
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

	machineMetadata := r.getORIMachineMetadata(machine)

	return &ori.MachineConfig{
		Metadata:          machineMetadata,
		Image:             machine.Spec.Image,
		Resources:         machineResources,
		Ignition:          ignitionConfig,
		Volumes:           volumeConfigs,
		NetworkInterfaces: networkInterfaceConfigs,
		Annotations:       map[string]string{},
		Labels: map[string]string{
			machinepoolletv1alpha1.MachineUIDLabel:       string(machine.UID),
			machinepoolletv1alpha1.MachineNamespaceLabel: machine.Namespace,
			machinepoolletv1alpha1.MachineNameLabel:      machine.Name,
		},
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

func (r *MachineReconciler) SetupWithManager(mgr ctrl.Manager) error {
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
				MachineRunsInMachinePoolPredicate(r.MachinePoolName),
				predicates.ResourceHasFilterLabel(log, r.WatchFilterValue),
				predicates.ResourceIsNotExternallyManaged(log),
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
