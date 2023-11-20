// SPDX-FileCopyrightText: 2023 SAP SE or an SAP affiliate company and IronCore contributors
// SPDX-License-Identifier: Apache-2.0

package compute

import (
	"context"
	"errors"
	"fmt"

	"github.com/ironcore-dev/ironcore/utils/annotations"
	"golang.org/x/exp/maps"

	"github.com/go-logr/logr"
	commonv1alpha1 "github.com/ironcore-dev/ironcore/api/common/v1alpha1"
	computev1alpha1 "github.com/ironcore-dev/ironcore/api/compute/v1alpha1"
	networkingv1alpha1 "github.com/ironcore-dev/ironcore/api/networking/v1alpha1"
	computeclient "github.com/ironcore-dev/ironcore/internal/client/compute"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/klog/v2"
	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/builder"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/handler"
	"sigs.k8s.io/controller-runtime/pkg/predicate"
)

type MachineEphemeralNetworkInterfaceReconciler struct {
	client.Client
}

//+kubebuilder:rbac:groups=compute.ironcore.dev,resources=machines,verbs=get;list;watch
//+kubebuilder:rbac:groups=networking.ironcore.dev,resources=networkinterfaces,verbs=get;list;watch;create;update;delete

func (r *MachineEphemeralNetworkInterfaceReconciler) Reconcile(ctx context.Context, req ctrl.Request) (ctrl.Result, error) {
	log := ctrl.LoggerFrom(ctx)
	machine := &computev1alpha1.Machine{}
	if err := r.Get(ctx, req.NamespacedName, machine); err != nil {
		return ctrl.Result{}, client.IgnoreNotFound(err)
	}

	return r.reconcileExists(ctx, log, machine)
}

func (r *MachineEphemeralNetworkInterfaceReconciler) reconcileExists(ctx context.Context, log logr.Logger, machine *computev1alpha1.Machine) (ctrl.Result, error) {
	if !machine.DeletionTimestamp.IsZero() {
		log.V(1).Info("Machine is deleting, nothing to do")
		return ctrl.Result{}, nil
	}

	return r.reconcile(ctx, log, machine)
}

func (r *MachineEphemeralNetworkInterfaceReconciler) ephemeralNetworkInterfaceByName(machine *computev1alpha1.Machine) map[string]*networkingv1alpha1.NetworkInterface {
	res := make(map[string]*networkingv1alpha1.NetworkInterface)
	for _, machineNic := range machine.Spec.NetworkInterfaces {
		ephemeral := machineNic.Ephemeral
		if ephemeral == nil {
			continue
		}

		nicName := computev1alpha1.MachineEphemeralNetworkInterfaceName(machine.Name, machineNic.Name)
		nic := &networkingv1alpha1.NetworkInterface{
			ObjectMeta: metav1.ObjectMeta{
				Namespace:   machine.Namespace,
				Name:        nicName,
				Labels:      ephemeral.NetworkInterfaceTemplate.Labels,
				Annotations: maps.Clone(ephemeral.NetworkInterfaceTemplate.Annotations),
			},
			Spec: ephemeral.NetworkInterfaceTemplate.Spec,
		}
		annotations.SetDefaultEphemeralManagedBy(nic)
		_ = ctrl.SetControllerReference(machine, nic, r.Scheme())
		nic.Spec.MachineRef = &commonv1alpha1.LocalUIDReference{
			Name: machine.Name,
			UID:  machine.UID,
		}
		res[nicName] = nic
	}
	return res
}

func (r *MachineEphemeralNetworkInterfaceReconciler) handleExistingNetworkInterface(ctx context.Context, log logr.Logger, machine *computev1alpha1.Machine, shouldManage bool, nic *networkingv1alpha1.NetworkInterface) error {
	if annotations.IsDefaultEphemeralControlledBy(nic, machine) {
		if shouldManage {
			log.V(1).Info("Ephemeral network interface is present and controlled by machine")
			return nil
		}

		if !nic.DeletionTimestamp.IsZero() {
			log.V(1).Info("Undesired ephemeral network interface is already deleting")
			return nil
		}

		log.V(1).Info("Deleting undesired ephemeral network interface")
		if err := r.Delete(ctx, nic); client.IgnoreNotFound(err) != nil {
			return fmt.Errorf("error deleting network interface %s: %w", nic.Name, err)
		}
		return nil
	}

	if shouldManage {
		log.V(1).Info("Won't adopt unmanaged network interface")
	}
	return nil
}

func (r *MachineEphemeralNetworkInterfaceReconciler) handleCreateNetworkInterface(
	ctx context.Context,
	log logr.Logger,
	machine *computev1alpha1.Machine,
	nic *networkingv1alpha1.NetworkInterface,
) error {
	log.V(1).Info("Creating network interface")
	nicKey := client.ObjectKeyFromObject(nic)
	err := r.Create(ctx, nic)
	if client.IgnoreAlreadyExists(err) != nil {
		return err
	}
	if err == nil {
		return nil
	}

	// Due to a fast resync, we might get an already exists error.
	// In this case, try to fetch the network interface again and, when successful, treat it as managing
	// an existing network interface.
	if err := r.Get(ctx, nicKey, nic); err != nil {
		return fmt.Errorf("error getting network interface %s after already exists: %w", nicKey.Name, err)
	}

	return r.handleExistingNetworkInterface(ctx, log, machine, true, nic)
}

func (r *MachineEphemeralNetworkInterfaceReconciler) reconcile(ctx context.Context, log logr.Logger, machine *computev1alpha1.Machine) (ctrl.Result, error) {
	log.V(1).Info("Reconcile")

	log.V(1).Info("Listing network interfaces")
	nicList := &networkingv1alpha1.NetworkInterfaceList{}
	if err := r.List(ctx, nicList,
		client.InNamespace(machine.Namespace),
	); err != nil {
		return ctrl.Result{}, fmt.Errorf("error listing network interfaces: %w", err)
	}

	var (
		ephemNicByName = r.ephemeralNetworkInterfaceByName(machine)
		errs           []error
	)
	for _, nic := range nicList.Items {
		nicName := nic.Name
		_, shouldManage := ephemNicByName[nicName]
		delete(ephemNicByName, nicName)
		log := log.WithValues("NetworkInterface", klog.KObj(&nic), "ShouldManage", shouldManage)
		if err := r.handleExistingNetworkInterface(ctx, log, machine, shouldManage, &nic); err != nil {
			errs = append(errs, err)
		}
	}
	for _, nic := range ephemNicByName {
		log := log.WithValues("NetworkInterface", klog.KObj(nic))
		if err := r.handleCreateNetworkInterface(ctx, log, machine, nic); err != nil {
			errs = append(errs, err)
		}
	}

	if err := errors.Join(errs...); err != nil {
		return ctrl.Result{}, fmt.Errorf("error managing ephemeral network interfaces: %w", err)
	}

	log.V(1).Info("Reconciled")
	return ctrl.Result{}, nil
}

func (r *MachineEphemeralNetworkInterfaceReconciler) machineNotDeletingPredicate() predicate.Predicate {
	return predicate.NewPredicateFuncs(func(obj client.Object) bool {
		machine := obj.(*computev1alpha1.Machine)
		return machine.DeletionTimestamp.IsZero()
	})
}

func (r *MachineEphemeralNetworkInterfaceReconciler) enqueueByNetworkInterface() handler.EventHandler {
	return handler.EnqueueRequestsFromMapFunc(func(ctx context.Context, obj client.Object) []ctrl.Request {
		nic := obj.(*networkingv1alpha1.NetworkInterface)
		log := ctrl.LoggerFrom(ctx)

		machineList := &computev1alpha1.MachineList{}
		if err := r.List(ctx, machineList,
			client.InNamespace(nic.Namespace),
			client.MatchingFields{
				computeclient.MachineSpecNetworkInterfaceNamesField: nic.Name,
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

func (r *MachineEphemeralNetworkInterfaceReconciler) SetupWithManager(mgr ctrl.Manager) error {
	return ctrl.NewControllerManagedBy(mgr).
		Named("machineephemeralnetworkinterface").
		For(
			&computev1alpha1.Machine{},
			builder.WithPredicates(
				r.machineNotDeletingPredicate(),
			),
		).
		Owns(
			&networkingv1alpha1.NetworkInterface{},
		).
		Watches(
			&networkingv1alpha1.NetworkInterface{},
			r.enqueueByNetworkInterface(),
		).
		Complete(r)
}
