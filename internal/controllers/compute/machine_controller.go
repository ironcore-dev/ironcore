/*
 * Copyright (c) 2021 by the OnMetal authors.
 *
 * Licensed under the Apache License, Version 2.0 (the "License");
 * you may not use this file except in compliance with the License.
 * You may obtain a copy of the License at
 *
 *     http://www.apache.org/licenses/LICENSE-2.0
 *
 * Unless required by applicable law or agreed to in writing, software
 * distributed under the License is distributed on an "AS IS" BASIS,
 * WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
 * See the License for the specific language governing permissions and
 * limitations under the License.
 */

package compute

import (
	"context"
	"errors"
	"fmt"
	"reflect"

	"github.com/go-logr/logr"
	"github.com/onmetal/controller-utils/metautils"
	commonv1alpha1 "github.com/onmetal/onmetal-api/api/common/v1alpha1"
	computev1alpha1 "github.com/onmetal/onmetal-api/api/compute/v1alpha1"
	networkingv1alpha1 "github.com/onmetal/onmetal-api/api/networking/v1alpha1"
	storagev1alpha1 "github.com/onmetal/onmetal-api/api/storage/v1alpha1"
	"github.com/onmetal/onmetal-api/internal/client/compute"
	"github.com/onmetal/onmetal-api/internal/controllers/compute/events"
	client2 "github.com/onmetal/onmetal-api/utils/client"
	corev1 "k8s.io/api/core/v1"
	apierrors "k8s.io/apimachinery/pkg/api/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apimachinery/pkg/util/sets"
	"k8s.io/client-go/tools/record"
	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/handler"
	"sigs.k8s.io/controller-runtime/pkg/source"
)

const (
	ephemeralSourceMachineUIDLabel = "compute.api.onmetal.de/ephemeral-source-machine-uid"
)

// MachineReconciler reconciles a Machine object
type MachineReconciler struct {
	record.EventRecorder
	client.Client
	Scheme *runtime.Scheme
}

//+kubebuilder:rbac:groups="",resources=events,verbs=create;patch
//+kubebuilder:rbac:groups=compute.api.onmetal.de,resources=machines,verbs=get;list;watch;create;update;patch;delete
//+kubebuilder:rbac:groups=compute.api.onmetal.de,resources=machines/status,verbs=get;update;patch
//+kubebuilder:rbac:groups=compute.api.onmetal.de,resources=machines/finalizers,verbs=update
//+kubebuilder:rbac:groups=storage.api.onmetal.de,resources=volumes,verbs=get;list;watch;create;update;patch;delete
//+kubebuilder:rbac:groups=networking.api.onmetal.de,resources=networkinterfaces,verbs=get;list;watch;create;update;patch;delete

// Reconcile is part of the main kubernetes reconciliation loop which aims to
// move the current state of the cluster closer to the desired state.
func (r *MachineReconciler) Reconcile(ctx context.Context, req ctrl.Request) (ctrl.Result, error) {
	log := ctrl.LoggerFrom(ctx)
	machine := &computev1alpha1.Machine{}
	if err := r.Get(ctx, req.NamespacedName, machine); err != nil {
		return ctrl.Result{}, client.IgnoreNotFound(err)
	}

	return r.reconcileExists(ctx, log, machine)
}

func (r *MachineReconciler) reconcileExists(ctx context.Context, log logr.Logger, machine *computev1alpha1.Machine) (ctrl.Result, error) {
	if !machine.DeletionTimestamp.IsZero() {
		return r.delete(ctx, log, machine)
	}
	return r.reconcile(ctx, log, machine)
}

func (r *MachineReconciler) delete(ctx context.Context, log logr.Logger, machine *computev1alpha1.Machine) (ctrl.Result, error) {
	return ctrl.Result{}, nil
}

func (r *MachineReconciler) computeNetworkInterfaceStatusValues(machine *computev1alpha1.Machine, nic *networkingv1alpha1.NetworkInterface) (phase computev1alpha1.NetworkInterfacePhase, ips []commonv1alpha1.IP, virtualIP *commonv1alpha1.IP) {
	if !reflect.DeepEqual(nic.Spec.MachineRef, &commonv1alpha1.LocalUIDReference{Name: machine.Name, UID: machine.UID}) {
		return computev1alpha1.NetworkInterfacePhasePending, nil, nil
	}
	return computev1alpha1.NetworkInterfacePhaseBound, nic.Status.IPs, nic.Status.VirtualIP
}

func (r *MachineReconciler) applyNetworkInterface(
	ctx context.Context,
	log logr.Logger,
	machine *computev1alpha1.Machine,
	machineNic *computev1alpha1.NetworkInterface,
) (phase computev1alpha1.NetworkInterfacePhase, ips []commonv1alpha1.IP, virtualIP *commonv1alpha1.IP, bindWarning string, err error) {
	switch {
	case machineNic.NetworkInterfaceRef != nil:
		nic := &networkingv1alpha1.NetworkInterface{}
		nicKey := client.ObjectKey{Namespace: machine.Namespace, Name: machineNic.NetworkInterfaceRef.Name}
		log = log.WithValues("NetworkInterfaceKey", nicKey)
		log.V(1).Info("Getting referenced network interface")
		if err := r.Get(ctx, nicKey, nic); err != nil {
			if !apierrors.IsNotFound(err) {
				return "", nil, nil, "", fmt.Errorf("error getting network interface %s: %w", nicKey, err)
			}

			return computev1alpha1.NetworkInterfacePhasePending, nil, nil, fmt.Sprintf("network interface %s not found", nicKey.Name), nil
		}

		phase, ips, virtualIP = r.computeNetworkInterfaceStatusValues(machine, nic)
		return phase, ips, virtualIP, "", nil
	case machineNic.Ephemeral != nil:
		template := machineNic.Ephemeral.NetworkInterfaceTemplate
		nic := &networkingv1alpha1.NetworkInterface{
			ObjectMeta: metav1.ObjectMeta{
				Namespace: machine.Namespace,
				Name:      fmt.Sprintf("%s-%s", machine.Name, machineNic.Name),
			},
		}
		nicKey := client.ObjectKeyFromObject(nic)
		log = log.WithValues("NetworkInterfaceKey", nicKey)
		log.V(1).Info("Managing network interface")
		if err := client2.ControlledCreateOrGet(ctx, r.Client, machine, nic, func() error {
			nic.Labels = template.Labels
			nic.Annotations = template.Annotations
			metautils.SetLabel(nic, ephemeralSourceMachineUIDLabel, string(machine.UID))

			nic.Spec = template.Spec
			nic.Spec.MachineRef = &commonv1alpha1.LocalUIDReference{Name: machine.Name, UID: machine.UID}
			return nil
		}); err != nil {
			if !errors.Is(err, client2.ErrNotControlled) {
				return "", nil, nil, "", fmt.Errorf("error managing network interface %s: %w", nic.Name, err)
			}

			return computev1alpha1.NetworkInterfacePhasePending, nil, nil, fmt.Sprintf("network interface %s cannot be managed", nic.Name), nil
		}

		phase, ips, virtualIP = r.computeNetworkInterfaceStatusValues(machine, nic)
		return phase, ips, virtualIP, "", nil
	default:
		return "", nil, nil, "", fmt.Errorf("invalid interface %#v", machineNic)
	}
}

func (r *MachineReconciler) getOrInitNetworkInterfaceStatus(machine *computev1alpha1.Machine, machineNicName string) computev1alpha1.NetworkInterfaceStatus {
	for _, status := range machine.Status.NetworkInterfaces {
		if status.Name == machineNicName {
			return status
		}
	}
	return computev1alpha1.NetworkInterfaceStatus{
		Name: machineNicName,
	}
}

func (r *MachineReconciler) pruneEphemeralNetworkInterfaces(ctx context.Context, machine *computev1alpha1.Machine) error {
	nicList := &networkingv1alpha1.NetworkInterfaceList{}
	if err := r.List(ctx, nicList,
		client.InNamespace(machine.Namespace),
		client.MatchingLabels{
			ephemeralSourceMachineUIDLabel: string(machine.UID),
		},
	); err != nil {
		return fmt.Errorf("error listing network interfaces: %w", err)
	}

	var (
		activeNames = sets.New(computev1alpha1.MachineNetworkInterfaceNames(machine)...)
		errs        []error
	)

	for _, nic := range nicList.Items {
		if activeNames.Has(nic.Name) {
			continue
		}
		if !metav1.IsControlledBy(&nic, machine) {
			continue
		}

		if err := r.Delete(ctx, &nic); client.IgnoreNotFound(err) != nil {
			errs = append(errs, fmt.Errorf("error pruning ephemeral network interface %s: %w", client.ObjectKeyFromObject(&nic), err))
		}
	}
	if len(errs) > 0 {
		return fmt.Errorf("error(s) pruning ephemeral network interfaces: %v", errs)
	}
	return nil
}

func (r *MachineReconciler) bindNetworkInterfaces(ctx context.Context, log logr.Logger, machine *computev1alpha1.Machine) ([]computev1alpha1.NetworkInterfaceStatus, error) {
	defer func() {
		if err := r.pruneEphemeralNetworkInterfaces(ctx, machine); err != nil {
			log.Error(err, "Error pruning ephemeral network interfaces")
		}
	}()

	var (
		res          []computev1alpha1.NetworkInterfaceStatus
		bindWarnings []string
		now          = metav1.Now()
	)
	for _, machineNic := range machine.Spec.NetworkInterfaces {
		phase, ips, virtualIP, bindError, err := r.applyNetworkInterface(ctx, log, machine, &machineNic)
		if err != nil {
			return nil, fmt.Errorf("[interface %s]: %w", machineNic.Name, err)
		}
		if bindError != "" {
			bindWarnings = append(bindWarnings, bindError)
		}

		status := r.getOrInitNetworkInterfaceStatus(machine, machineNic.Name)
		if phase != status.Phase {
			status.LastPhaseTransitionTime = &now
		}
		status.Phase = phase
		status.IPs = ips
		status.VirtualIP = virtualIP

		res = append(res, status)
	}
	if len(bindWarnings) > 0 {
		r.Eventf(machine, corev1.EventTypeWarning, events.FailedBindingNetworkInterfaces, "Failed binding network interface(s): %v", bindWarnings)
	}

	return res, nil
}

func (r *MachineReconciler) getOrInitVolumeStatus(machine *computev1alpha1.Machine, name string) computev1alpha1.VolumeStatus {
	for _, status := range machine.Status.Volumes {
		if status.Name == name {
			return status
		}
	}
	return computev1alpha1.VolumeStatus{
		Name: name,
	}
}

func (r *MachineReconciler) pruneEphemeralVolumes(ctx context.Context, machine *computev1alpha1.Machine) error {
	volumeList := &storagev1alpha1.VolumeList{}
	if err := r.List(ctx, volumeList,
		client.InNamespace(machine.Namespace),
		client.MatchingLabels{
			ephemeralSourceMachineUIDLabel: string(machine.UID),
		},
	); err != nil {
		return fmt.Errorf("error listing volumes: %w", err)
	}

	var (
		activeNames = sets.New(computev1alpha1.MachineVolumeNames(machine)...)
		errs        []error
	)

	for _, volume := range volumeList.Items {
		if activeNames.Has(volume.Name) {
			continue
		}
		if !metav1.IsControlledBy(&volume, machine) {
			continue
		}

		if err := r.Delete(ctx, &volume); client.IgnoreNotFound(err) != nil {
			errs = append(errs, fmt.Errorf("error pruning ephemeral volume %s: %w", client.ObjectKeyFromObject(&volume), err))
		}
	}
	if len(errs) > 0 {
		return fmt.Errorf("error(s) pruning ephemeral volumes: %v", errs)
	}
	return nil
}

func (r *MachineReconciler) bindVolumes(ctx context.Context, log logr.Logger, machine *computev1alpha1.Machine) ([]computev1alpha1.VolumeStatus, error) {
	defer func() {
		if err := r.pruneEphemeralVolumes(ctx, machine); err != nil {
			log.Error(err, "Error pruning ephemeral volumes")
		}
	}()

	var (
		res          []computev1alpha1.VolumeStatus
		bindWarnings []string
		now          = metav1.Now()
	)
	for _, machineVolume := range machine.Spec.Volumes {
		name := machineVolume.Name
		phase, bindWarning, err := r.bindVolume(ctx, log, machine, &machineVolume)
		if err != nil {
			return nil, fmt.Errorf("[volume %s] %w", name, err)
		}
		if bindWarning != "" {
			bindWarnings = append(bindWarnings, bindWarning)
		}

		status := r.getOrInitVolumeStatus(machine, name)
		if phase != status.Phase {
			status.LastPhaseTransitionTime = &now
		}
		status.Phase = phase
		res = append(res, status)
	}

	if len(bindWarnings) > 0 {
		r.Eventf(machine, corev1.EventTypeWarning, events.FailedBindingVolumes, "Failed binding volume(s): %v", bindWarnings)
	}

	return res, nil
}

func (r *MachineReconciler) patchStatus(
	ctx context.Context,
	machine *computev1alpha1.Machine,
	nicStates []computev1alpha1.NetworkInterfaceStatus,
	volumeStates []computev1alpha1.VolumeStatus,
) error {
	base := machine.DeepCopy()
	machine.Status.NetworkInterfaces = nicStates
	machine.Status.Volumes = volumeStates
	if err := r.Patch(ctx, machine, client.MergeFrom(base)); err != nil {
		return fmt.Errorf("error patching machine: %w", err)
	}
	return nil
}

func (r *MachineReconciler) reconcile(ctx context.Context, log logr.Logger, machine *computev1alpha1.Machine) (ctrl.Result, error) {
	log.V(1).Info("Reconciling")

	log.V(1).Info("Applying network interfaces")
	nics, err := r.bindNetworkInterfaces(ctx, log, machine)
	if err != nil {
		return ctrl.Result{}, fmt.Errorf("error applying network interfaces: %w", err)
	}

	log.V(1).Info("Applying  volumes")
	volumes, err := r.bindVolumes(ctx, log, machine)
	if err != nil {
		return ctrl.Result{}, fmt.Errorf("error applying volumes: %w", err)
	}

	log.V(1).Info("Patching status")
	if err := r.patchStatus(ctx, machine, nics, volumes); err != nil {
		return ctrl.Result{}, fmt.Errorf("error patching machine status: %w", err)
	}

	log.V(1).Info("Successfully reconciled")
	return ctrl.Result{}, nil
}

func (r *MachineReconciler) computeVolumePhase(machine *computev1alpha1.Machine, volume *storagev1alpha1.Volume) computev1alpha1.VolumePhase {
	if !reflect.DeepEqual(volume.Spec.ClaimRef, &commonv1alpha1.LocalUIDReference{Name: machine.Name, UID: machine.UID}) {
		return computev1alpha1.VolumePhasePending
	}
	return computev1alpha1.VolumePhaseBound
}

func (r *MachineReconciler) bindVolume(
	ctx context.Context,
	log logr.Logger,
	machine *computev1alpha1.Machine,
	machineVolume *computev1alpha1.Volume,
) (phase computev1alpha1.VolumePhase, bindWarning string, err error) {
	switch {
	case machineVolume.EmptyDisk != nil:
		return computev1alpha1.VolumePhaseBound, "", nil
	case machineVolume.VolumeRef != nil:
		volume := &storagev1alpha1.Volume{}
		volumeKey := client.ObjectKey{Namespace: machine.Namespace, Name: machineVolume.VolumeRef.Name}
		log.V(1).Info("Getting volume", "VolumeKey", volumeKey)
		if err := r.Get(ctx, volumeKey, volume); err != nil {
			if !apierrors.IsNotFound(err) {
				return "", "", fmt.Errorf("error getting volume %s: %w", volumeKey, err)
			}

			return computev1alpha1.VolumePhasePending, fmt.Sprintf("volume %s not found", volumeKey.Name), nil
		}
		return r.computeVolumePhase(machine, volume), "", nil
	case machineVolume.Ephemeral != nil:
		template := machineVolume.Ephemeral.VolumeTemplate
		volume := &storagev1alpha1.Volume{
			ObjectMeta: metav1.ObjectMeta{
				Namespace: machine.Namespace,
				Name:      computev1alpha1.MachineEphemeralVolumeName(machine.Name, machineVolume.Name),
			},
		}
		volumeKey := client.ObjectKeyFromObject(volume)
		log = log.WithValues("VolumeKey", volumeKey)
		log.V(1).Info("Managing volume")
		if err := client2.ControlledCreateOrGet(ctx, r.Client, machine, volume, func() error {
			volume.Labels = template.Labels
			volume.Annotations = template.Annotations
			metav1.SetMetaDataLabel(&volume.ObjectMeta, ephemeralSourceMachineUIDLabel, string(machine.UID))

			volume.Spec = template.Spec
			volume.Spec.ClaimRef = &commonv1alpha1.LocalUIDReference{Name: machine.Name, UID: machine.UID}
			return nil
		}); err != nil {
			if !errors.Is(err, client2.ErrNotControlled) {
				return "", "", fmt.Errorf("error managing volume %s: %w", volume.Name, err)
			}

			return computev1alpha1.VolumePhasePending, fmt.Sprintf("volume %s cannot be managed", volume.Name), nil
		}

		return r.computeVolumePhase(machine, volume), "", nil
	default:
		return "", "", fmt.Errorf("invalid volume %#v", machineVolume)
	}
}

// SetupWithManager sets up the controller with the Manager.
func (r *MachineReconciler) SetupWithManager(mgr ctrl.Manager) error {
	ctx := context.Background()
	log := ctrl.Log.WithName("machine").WithName("setup")

	return ctrl.NewControllerManagedBy(mgr).
		For(&computev1alpha1.Machine{}).
		Owns(&networkingv1alpha1.NetworkInterface{}).
		Owns(&storagev1alpha1.Volume{}).
		Watches(
			&source.Kind{Type: &networkingv1alpha1.NetworkInterface{}},
			r.enqueueByMachineNetworkInterfaceReferences(ctx, log),
		).
		Watches(
			&source.Kind{Type: &storagev1alpha1.Volume{}},
			r.enqueueByMachineVolumeReferences(ctx, log),
		).
		Complete(r)
}

func (r *MachineReconciler) enqueueByMachineNetworkInterfaceReferences(ctx context.Context, log logr.Logger) handler.EventHandler {
	return handler.EnqueueRequestsFromMapFunc(func(obj client.Object) []ctrl.Request {
		nic := obj.(*networkingv1alpha1.NetworkInterface)
		log = log.WithValues("NetworkInterfaceKey", client.ObjectKeyFromObject(nic))

		machineList := &computev1alpha1.MachineList{}
		if err := r.List(ctx, machineList,
			client.InNamespace(nic.Namespace),
			client.MatchingFields{
				compute.MachineSpecNetworkInterfaceNamesField: nic.Name,
			},
		); err != nil {
			log.Error(err, "Error listing machines using network interface")
			return nil
		}

		res := make([]ctrl.Request, 0, len(machineList.Items))
		for _, machine := range machineList.Items {
			res = append(res, ctrl.Request{NamespacedName: client.ObjectKeyFromObject(&machine)})
		}
		return res
	})
}

func (r *MachineReconciler) enqueueByMachineVolumeReferences(ctx context.Context, log logr.Logger) handler.EventHandler {
	return handler.EnqueueRequestsFromMapFunc(func(obj client.Object) []ctrl.Request {
		volume := obj.(*storagev1alpha1.Volume)
		log = log.WithValues("VolumeKey", client.ObjectKeyFromObject(volume))

		machineList := &computev1alpha1.MachineList{}
		if err := r.List(ctx, machineList,
			client.InNamespace(volume.Namespace),
			client.MatchingFields{
				compute.MachineSpecVolumeNamesField: volume.Name,
			},
		); err != nil {
			log.Error(err, "Error listing machines using volume")
			return nil
		}

		res := make([]ctrl.Request, 0, len(machineList.Items))
		for _, machine := range machineList.Items {
			res = append(res, ctrl.Request{NamespacedName: client.ObjectKeyFromObject(&machine)})
		}
		return res
	})
}
