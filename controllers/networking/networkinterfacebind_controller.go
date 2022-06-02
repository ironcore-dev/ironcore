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

package networking

import (
	"context"
	"fmt"
	"math/rand"
	"time"

	"github.com/go-logr/logr"
	commonv1alpha1 "github.com/onmetal/onmetal-api/apis/common/v1alpha1"
	computev1alpha1 "github.com/onmetal/onmetal-api/apis/compute/v1alpha1"
	networkingv1alpha1 "github.com/onmetal/onmetal-api/apis/networking/v1alpha1"
	"github.com/onmetal/onmetal-api/controllers/shared"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apimachinery/pkg/types"
	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/handler"
	"sigs.k8s.io/controller-runtime/pkg/source"
)

type NetworkInterfaceBindReconciler struct {
	client.Client
	APIReader   client.Reader
	Scheme      *runtime.Scheme
	BindTimeout time.Duration
}

//+kubebuilder:rbac:groups=networking.api.onmetal.de,resources=networkinterfaces,verbs=get;list;watch;create;update;patch;delete
//+kubebuilder:rbac:groups=networking.api.onmetal.de,resources=networkinterfaces/status,verbs=get;update;patch
//+kubebuilder:rbac:groups=networking.api.onmetal.de,resources=networkinterfaces/finalizers,verbs=update
//+kubebuilder:rbac:groups=compute.api.onmetal.de,resources=machines,verbs=get;list;watch

func (r *NetworkInterfaceBindReconciler) Reconcile(ctx context.Context, req ctrl.Request) (ctrl.Result, error) {
	log := ctrl.LoggerFrom(ctx)
	nic := &networkingv1alpha1.NetworkInterface{}
	if err := r.Get(ctx, req.NamespacedName, nic); err != nil {
		return ctrl.Result{}, client.IgnoreNotFound(err)
	}

	return r.reconcileExists(ctx, log, nic)
}

func (r *NetworkInterfaceBindReconciler) reconcileExists(ctx context.Context, log logr.Logger, nic *networkingv1alpha1.NetworkInterface) (ctrl.Result, error) {
	if !nic.DeletionTimestamp.IsZero() {
		return r.delete(ctx, log, nic)
	}
	return r.reconcile(ctx, log, nic)
}

func (r *NetworkInterfaceBindReconciler) delete(ctx context.Context, log logr.Logger, nic *networkingv1alpha1.NetworkInterface) (ctrl.Result, error) {
	return ctrl.Result{}, nil
}

func (r *NetworkInterfaceBindReconciler) phaseTransitionTimedOut(timestamp *metav1.Time) bool {
	if timestamp.IsZero() {
		return false
	}
	return timestamp.Add(r.BindTimeout).Before(time.Now())
}

func (r *NetworkInterfaceBindReconciler) reconcile(ctx context.Context, log logr.Logger, nic *networkingv1alpha1.NetworkInterface) (ctrl.Result, error) {
	if nic.Spec.MachineRef == nil {
		return r.reconcileUnbound(ctx, log, nic)
	}

	return r.reconcileBound(ctx, log, nic)
}

func (r *NetworkInterfaceBindReconciler) reconcileBound(ctx context.Context, log logr.Logger, nic *networkingv1alpha1.NetworkInterface) (ctrl.Result, error) {
	log.V(1).Info("Reconcile bound")

	machine := &computev1alpha1.Machine{}
	machineKey := client.ObjectKey{Namespace: nic.Namespace, Name: nic.Spec.MachineRef.Name}
	log = log.WithValues("MachineKey", client.ObjectKeyFromObject(machine))
	log.V(1).Info("Getting requester")
	err := r.APIReader.Get(ctx, machineKey, machine)
	if client.IgnoreNotFound(err) != nil {
		return ctrl.Result{}, fmt.Errorf("error getting requester %s: %w", machineKey, err)
	}

	machineExists := err == nil
	validReferences := machineExists && r.validReferences(nic, machine)
	phase := nic.Status.Phase
	phaseLastTransitionTime := nic.Status.LastPhaseTransitionTime

	log = log.WithValues(
		"MachineExists", machineExists,
		"ValidReferences", validReferences,
		"Phase", phase,
		"PhaseLastTransitionTime", phaseLastTransitionTime,
	)
	switch {
	case validReferences:
		log.V(1).Info("Setting to bound")
		if err := r.patchStatus(ctx, nic, networkingv1alpha1.NetworkInterfacePhaseBound); err != nil {
			return ctrl.Result{}, fmt.Errorf("error binding: %w", err)
		}

		log.V(1).Info("Successfully to bound.")
		return ctrl.Result{}, nil
	case !validReferences && phase == networkingv1alpha1.NetworkInterfacePhasePending && r.phaseTransitionTimedOut(phaseLastTransitionTime):
		log.V(1).Info("Bind is not ok and timed out, releasing")
		if err := r.release(ctx, nic); err != nil {
			return ctrl.Result{}, fmt.Errorf("error releasing: %w", err)
		}

		log.V(1).Info("Successfully released")
		return ctrl.Result{}, nil
	default:
		log.V(1).Info("Bind is not ok and not yet timed out, setting to pending")
		if err := r.patchStatus(ctx, nic, networkingv1alpha1.NetworkInterfacePhasePending); err != nil {
			return ctrl.Result{}, fmt.Errorf("error setting phase to pending: %w", err)
		}

		log.V(1).Info("Successfully set phase to pending")
		return r.requeueAfterBoundTimeout(nic), nil
	}
}

func (r *NetworkInterfaceBindReconciler) reconcileUnbound(ctx context.Context, log logr.Logger, nic *networkingv1alpha1.NetworkInterface) (ctrl.Result, error) {
	log.V(1).Info("Reconcile unbound")

	log.V(1).Info("Searching for suitable requester")
	machine, err := r.getRequestingMachine(ctx, nic)
	if err != nil {
		return ctrl.Result{}, fmt.Errorf("error getting requesting requester: %w", err)
	}

	if machine == nil {
		log.V(1).Info("No requester found, setting phase unbound")
		if err := r.patchStatus(ctx, nic, networkingv1alpha1.NetworkInterfacePhaseUnbound); err != nil {
			return ctrl.Result{}, err
		}
		log.V(1).Info("Successfully set phase to unbound")
		return ctrl.Result{}, nil
	}

	log = log.WithValues("MachineKey", client.ObjectKeyFromObject(machine))
	log.V(1).Info("Requester found, assigning")
	if err := r.assign(ctx, nic, machine); err != nil {
		return ctrl.Result{}, err
	}

	log.V(1).Info("Successfully assigned")
	return ctrl.Result{}, nil
}

func (r *NetworkInterfaceBindReconciler) assign(ctx context.Context, nic *networkingv1alpha1.NetworkInterface, machine *computev1alpha1.Machine) error {
	base := nic.DeepCopy()
	nic.Spec.MachineRef = &commonv1alpha1.LocalUIDReference{Name: machine.Name, UID: machine.UID}
	if err := r.Patch(ctx, nic, client.MergeFrom(base)); err != nil {
		return fmt.Errorf("error assigning: %w", err)
	}
	return nil
}

func (r *NetworkInterfaceBindReconciler) getRequestingMachine(ctx context.Context, nic *networkingv1alpha1.NetworkInterface) (*computev1alpha1.Machine, error) {
	machineList := &computev1alpha1.MachineList{}
	if err := r.List(ctx, machineList,
		client.InNamespace(nic.Namespace),
		client.MatchingFields{
			shared.MachineNetworkInterfaceNamesField: nic.Name,
		},
	); err != nil {
		return nil, fmt.Errorf("error listing matching machines: %w", err)
	}

	var matches []computev1alpha1.Machine
	for _, machine := range machineList.Items {
		if !machine.DeletionTimestamp.IsZero() {
			continue
		}

		matches = append(matches, machine)
	}
	if len(matches) == 0 {
		return nil, nil
	}
	match := matches[rand.Intn(len(matches))]
	return &match, nil
}

func (r *NetworkInterfaceBindReconciler) release(ctx context.Context, nic *networkingv1alpha1.NetworkInterface) error {
	base := nic.DeepCopy()
	nic.Spec.MachineRef = nil
	if err := r.Patch(ctx, nic, client.MergeFrom(base)); err != nil {
		return fmt.Errorf("error releasing network interface: %w", err)
	}
	return nil
}

func (r *NetworkInterfaceBindReconciler) requeueAfterBoundTimeout(nic *networkingv1alpha1.NetworkInterface) ctrl.Result {
	if nic.Status.LastPhaseTransitionTime != nil {
		boundTimeoutExpirationDuration := time.Until(nic.Status.LastPhaseTransitionTime.Add(r.BindTimeout)).Round(time.Second)
		if boundTimeoutExpirationDuration <= 0 {
			return ctrl.Result{Requeue: true}
		}
		return ctrl.Result{RequeueAfter: boundTimeoutExpirationDuration}
	}
	return ctrl.Result{}
}

func (r *NetworkInterfaceBindReconciler) validReferences(nic *networkingv1alpha1.NetworkInterface, machine *computev1alpha1.Machine) bool {
	machineRef := nic.Spec.MachineRef
	if machineRef.UID != machine.UID {
		return false
	}

	for _, name := range shared.MachineNetworkInterfaceNames(machine) {
		if name == nic.Name {
			return true
		}
	}
	return false
}

func (r *NetworkInterfaceBindReconciler) patchStatus(ctx context.Context, nic *networkingv1alpha1.NetworkInterface, phase networkingv1alpha1.NetworkInterfacePhase) error {
	now := metav1.Now()
	base := nic.DeepCopy()

	if phase != nic.Status.Phase {
		nic.Status.LastPhaseTransitionTime = &now
	}
	nic.Status.Phase = phase

	if err := r.Status().Patch(ctx, nic, client.MergeFrom(base)); err != nil {
		return fmt.Errorf("error patching status: %w", err)
	}
	return nil
}

const networkInterfaceSpecMachineRefNameField = ".spec.machineRef.name"

func (r *NetworkInterfaceBindReconciler) SetupWithManager(mgr ctrl.Manager) error {
	ctx := context.Background()
	log := ctrl.Log.WithName("networkinterfacebind").WithName("setup")

	if err := mgr.GetFieldIndexer().IndexField(ctx, &networkingv1alpha1.NetworkInterface{}, networkInterfaceSpecMachineRefNameField, func(obj client.Object) []string {
		nic := obj.(*networkingv1alpha1.NetworkInterface)

		machineRef := nic.Spec.MachineRef
		if machineRef == nil {
			return []string{""}
		}

		return []string{machineRef.Name}
	}); err != nil {
		return err
	}

	return ctrl.NewControllerManagedBy(mgr).
		Named("networkinterfacebind").
		For(&networkingv1alpha1.NetworkInterface{}).
		Watches(
			&source.Kind{Type: &computev1alpha1.Machine{}},
			r.enqueueByMachineNetworkInterfaceReference(),
		).
		Watches(
			&source.Kind{Type: &computev1alpha1.Machine{}},
			r.enqueueByMachineNameEqualNetworkInterfaceMachineRefName(ctx, log),
		).
		Complete(r)
}

func (r *NetworkInterfaceBindReconciler) enqueueByMachineNetworkInterfaceReference() handler.EventHandler {
	return handler.EnqueueRequestsFromMapFunc(func(obj client.Object) []ctrl.Request {
		machine := obj.(*computev1alpha1.Machine)

		var reqs []ctrl.Request
		for _, name := range shared.MachineNetworkInterfaceNames(machine) {
			reqs = append(reqs, ctrl.Request{NamespacedName: client.ObjectKey{Namespace: machine.Namespace, Name: name}})
		}

		return reqs
	})
}

func (r *NetworkInterfaceBindReconciler) enqueueByMachineNameEqualNetworkInterfaceMachineRefName(ctx context.Context, log logr.Logger) handler.EventHandler {
	return handler.EnqueueRequestsFromMapFunc(func(obj client.Object) []ctrl.Request {
		machine := obj.(*computev1alpha1.Machine)

		nicList := &networkingv1alpha1.NetworkInterfaceList{}
		if err := r.List(ctx, nicList,
			client.InNamespace(machine.Namespace),
			client.MatchingFields{
				networkInterfaceSpecMachineRefNameField: machine.Name,
			},
		); err != nil {
			log.Error(err, "Error listing network interfaces targeting machine")
			return nil
		}

		res := make([]ctrl.Request, 0, len(nicList.Items))
		for _, item := range nicList.Items {
			res = append(res, ctrl.Request{
				NamespacedName: types.NamespacedName{
					Name:      item.GetName(),
					Namespace: item.GetNamespace(),
				},
			})
		}
		return res
	})
}
