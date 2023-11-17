// Copyright 2023 IronCore authors
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

package compute

import (
	"context"
	"errors"
	"fmt"
	"maps"

	"github.com/ironcore-dev/ironcore/utils/annotations"

	"github.com/go-logr/logr"
	computev1alpha1 "github.com/ironcore-dev/ironcore/api/compute/v1alpha1"
	storagev1alpha1 "github.com/ironcore-dev/ironcore/api/storage/v1alpha1"
	computeclient "github.com/ironcore-dev/ironcore/internal/client/compute"
	apierrors "k8s.io/apimachinery/pkg/api/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/klog/v2"
	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/builder"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/handler"
	"sigs.k8s.io/controller-runtime/pkg/predicate"
)

const (
	MachineEphemeralVolumeManager = "machine-ephemeral-volume"
)

type MachineEphemeralVolumeReconciler struct {
	client.Client
}

//+kubebuilder:rbac:groups=compute.ironcore.dev,resources=machines,verbs=get;list;watch
//+kubebuilder:rbac:groups=storage.ironcore.dev,resources=volumes,verbs=get;list;watch;create;update;delete

func (r *MachineEphemeralVolumeReconciler) Reconcile(ctx context.Context, req ctrl.Request) (ctrl.Result, error) {
	log := ctrl.LoggerFrom(ctx)
	machine := &computev1alpha1.Machine{}
	if err := r.Get(ctx, req.NamespacedName, machine); err != nil {
		return ctrl.Result{}, client.IgnoreNotFound(err)
	}

	return r.reconcileExists(ctx, log, machine)
}

func (r *MachineEphemeralVolumeReconciler) reconcileExists(ctx context.Context, log logr.Logger, machine *computev1alpha1.Machine) (ctrl.Result, error) {
	if !machine.DeletionTimestamp.IsZero() {
		log.V(1).Info("Machine is deleting, nothing to do")
		return ctrl.Result{}, nil
	}

	return r.reconcile(ctx, log, machine)
}

func (r *MachineEphemeralVolumeReconciler) ephemeralMachineVolumeByName(machine *computev1alpha1.Machine) map[string]*storagev1alpha1.Volume {
	res := make(map[string]*storagev1alpha1.Volume)
	for _, machineVolume := range machine.Spec.Volumes {
		ephemeral := machineVolume.Ephemeral
		if ephemeral == nil {
			continue
		}

		volumeName := computev1alpha1.MachineEphemeralVolumeName(machine.Name, machineVolume.Name)
		volume := &storagev1alpha1.Volume{
			ObjectMeta: metav1.ObjectMeta{
				Namespace:   machine.Namespace,
				Name:        volumeName,
				Labels:      ephemeral.VolumeTemplate.Labels,
				Annotations: maps.Clone(ephemeral.VolumeTemplate.Annotations),
			},
			Spec: ephemeral.VolumeTemplate.Spec,
		}
		annotations.SetDefaultEphemeralManagedBy(volume)
		_ = ctrl.SetControllerReference(machine, volume, r.Scheme())
		res[volumeName] = volume
	}
	return res
}

func (r *MachineEphemeralVolumeReconciler) handleExistingVolume(ctx context.Context, log logr.Logger, machine *computev1alpha1.Machine, shouldManage bool, volume *storagev1alpha1.Volume) error {
	if annotations.IsDefaultEphemeralControlledBy(volume, machine) {
		if shouldManage {
			log.V(1).Info("Ephemeral volume is present and controlled by machine")
			return nil
		}

		if !volume.DeletionTimestamp.IsZero() {
			log.V(1).Info("Undesired ephemeral volume is already deleting")
			return nil
		}

		log.V(1).Info("Deleting undesired ephemeral volume")
		if err := r.Delete(ctx, volume); client.IgnoreNotFound(err) != nil {
			return fmt.Errorf("error deleting volume %s: %w", volume.Name, err)
		}
		return nil
	}

	if shouldManage {
		log.V(1).Info("Won't adopt unmanaged volume")
	}
	return nil
}

func (r *MachineEphemeralVolumeReconciler) handleCreateVolume(ctx context.Context, log logr.Logger, machine *computev1alpha1.Machine, volume *storagev1alpha1.Volume) error {
	volumeKey := client.ObjectKeyFromObject(volume)
	err := r.Create(ctx, volume)
	if err == nil {
		return nil
	}
	if !apierrors.IsAlreadyExists(err) {
		return err
	}

	// Due to a fast resync, we might get an already exists error.
	// In this case, try to fetch the volume again and, when successful, treat it as managing
	// an existing volume.
	if err := r.Get(ctx, volumeKey, volume); err != nil {
		return fmt.Errorf("error getting volume %s after already exists: %w", volumeKey.Name, err)
	}

	// Treat a retrieved volume as an existing we should manage.
	return r.handleExistingVolume(ctx, log, machine, true, volume)
}

func (r *MachineEphemeralVolumeReconciler) reconcile(ctx context.Context, log logr.Logger, machine *computev1alpha1.Machine) (ctrl.Result, error) {
	log.V(1).Info("Reconcile")

	log.V(1).Info("Listing volumes")
	volumeList := &storagev1alpha1.VolumeList{}
	if err := r.List(ctx, volumeList,
		client.InNamespace(machine.Namespace),
	); err != nil {
		return ctrl.Result{}, fmt.Errorf("error listing volumes: %w", err)
	}

	var (
		ephemVolumeByName = r.ephemeralMachineVolumeByName(machine)
		errs              []error
	)
	for _, volume := range volumeList.Items {
		volumeName := volume.Name
		_, shouldManage := ephemVolumeByName[volumeName]
		delete(ephemVolumeByName, volumeName)
		log := log.WithValues("Volume", klog.KObj(&volume), "ShouldManage", shouldManage)
		if err := r.handleExistingVolume(ctx, log, machine, shouldManage, &volume); err != nil {
			errs = append(errs, err)
		}
	}

	for _, volume := range ephemVolumeByName {
		log := log.WithValues("Volume", klog.KObj(volume))
		if err := r.handleCreateVolume(ctx, log, machine, volume); err != nil {
			errs = append(errs, err)
		}
	}

	if err := errors.Join(errs...); err != nil {
		return ctrl.Result{}, fmt.Errorf("error managing ephemeral volumes: %w", err)
	}

	log.V(1).Info("Reconciled")
	return ctrl.Result{}, nil
}

func (r *MachineEphemeralVolumeReconciler) machineNotDeletingPredicate() predicate.Predicate {
	return predicate.NewPredicateFuncs(func(obj client.Object) bool {
		machine := obj.(*computev1alpha1.Machine)
		return machine.DeletionTimestamp.IsZero()
	})
}

func (r *MachineEphemeralVolumeReconciler) enqueueByVolume() handler.EventHandler {
	return handler.EnqueueRequestsFromMapFunc(func(ctx context.Context, obj client.Object) []ctrl.Request {
		volume := obj.(*storagev1alpha1.Volume)
		log := ctrl.LoggerFrom(ctx)

		machineList := &computev1alpha1.MachineList{}
		if err := r.List(ctx, machineList,
			client.InNamespace(volume.Namespace),
			client.MatchingFields{
				computeclient.MachineSpecVolumeNamesField: volume.Name,
			},
		); err != nil {
			log.Error(err, "Error listing machines")
			return nil
		}

		var reqs []ctrl.Request
		for _, machine := range machineList.Items {
			if !machine.DeletionTimestamp.IsZero() {
				continue
			}

			reqs = append(reqs, ctrl.Request{NamespacedName: client.ObjectKeyFromObject(&machine)})
		}
		return reqs
	})
}

func (r *MachineEphemeralVolumeReconciler) SetupWithManager(mgr ctrl.Manager) error {
	return ctrl.NewControllerManagedBy(mgr).
		Named("machineephemeralvolume").
		For(
			&computev1alpha1.Machine{},
			builder.WithPredicates(
				r.machineNotDeletingPredicate(),
			),
		).
		Owns(
			&storagev1alpha1.Volume{},
		).
		Watches(
			&storagev1alpha1.Volume{},
			r.enqueueByVolume(),
		).
		Complete(r)
}
