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
	"fmt"
	"reflect"

	"github.com/go-logr/logr"
	commonv1alpha1 "github.com/onmetal/onmetal-api/apis/common/v1alpha1"
	networkingv1alpha1 "github.com/onmetal/onmetal-api/apis/networking/v1alpha1"
	onmetalapiclientutils "github.com/onmetal/onmetal-api/clientutils"
	"github.com/onmetal/onmetal-api/controllers/shared"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"
	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/handler"
	"sigs.k8s.io/controller-runtime/pkg/source"

	computev1alpha1 "github.com/onmetal/onmetal-api/apis/compute/v1alpha1"
)

// MachineReconciler reconciles a Machine object
type MachineReconciler struct {
	client.Client
	Scheme *runtime.Scheme
}

//+kubebuilder:rbac:groups=compute.api.onmetal.de,resources=machines,verbs=get;list;watch;create;update;patch;delete
//+kubebuilder:rbac:groups=compute.api.onmetal.de,resources=machines/status,verbs=get;update;patch
//+kubebuilder:rbac:groups=compute.api.onmetal.de,resources=machines/finalizers,verbs=update

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

func (r *MachineReconciler) getOrManageNetworkInterface(ctx context.Context, log logr.Logger, machine *computev1alpha1.Machine, name string, iface *computev1alpha1.NetworkInterface) (*networkingv1alpha1.NetworkInterface, error) {
	switch {
	case iface.NetworkInterfaceRef != nil:
		nic := &networkingv1alpha1.NetworkInterface{}
		nicKey := client.ObjectKey{Namespace: machine.Namespace, Name: iface.NetworkInterfaceRef.Name}
		log = log.WithValues("NetworkInterfaceKey", nicKey)
		log.V(1).Info("Getting referenced network interface")
		if err := r.Get(ctx, nicKey, nic); err != nil {
			return nil, fmt.Errorf("error getting network interface %s: %w", nicKey, err)
		}

		return nic, nil
	case iface.Ephemeral != nil:
		template := iface.Ephemeral.NetworkInterfaceTemplate
		nic := &networkingv1alpha1.NetworkInterface{
			ObjectMeta: metav1.ObjectMeta{
				Namespace: machine.Namespace,
				Name:      fmt.Sprintf("%s-%s", machine.Name, name),
			},
		}
		nicKey := client.ObjectKeyFromObject(nic)
		log = log.WithValues("NetworkInterfaceKey", nicKey)
		log.V(1).Info("Managing network interface")
		if err := onmetalapiclientutils.ControlledCreateOrGet(ctx, r.Client, machine, nic, func() error {
			nic.Labels = template.Labels
			nic.Annotations = template.Annotations
			nic.Spec = template.Spec
			nic.Spec.MachineRef = &commonv1alpha1.LocalUIDReference{Name: machine.Name, UID: machine.UID}
			return nil
		}); err != nil {
			return nil, fmt.Errorf("error managing network interface %s: %w", nic.Name, err)
		}

		return nic, nil
	default:
		return nil, fmt.Errorf("invalid interface %#v", iface)
	}
}

func (r *MachineReconciler) applyNetworkInterfaces(ctx context.Context, log logr.Logger, machine *computev1alpha1.Machine) ([]computev1alpha1.NetworkInterfaceStatus, error) {
	var res []computev1alpha1.NetworkInterfaceStatus
	for _, iface := range machine.Spec.NetworkInterfaces {
		nic, err := r.getOrManageNetworkInterface(ctx, log, machine, iface.Name, &iface)
		if err != nil {
			return nil, fmt.Errorf("[interface %s]: %w", iface.Name, err)
		}

		if !reflect.DeepEqual(nic.Spec.MachineRef, &commonv1alpha1.LocalUIDReference{Name: machine.Name, UID: machine.UID}) {
			log.V(1).Info("Network interface does not yet bind machine",
				"NetworkInterfaceKey", client.ObjectKeyFromObject(nic),
			)
			continue
		}

		res = append(res, computev1alpha1.NetworkInterfaceStatus{
			Name:      iface.Name,
			IPs:       nic.Status.IPs,
			VirtualIP: nic.Status.VirtualIP,
		})
	}
	return res, nil
}

func (r *MachineReconciler) patchStatus(ctx context.Context, machine *computev1alpha1.Machine, ifaceStates []computev1alpha1.NetworkInterfaceStatus) error {
	base := machine.DeepCopy()
	machine.Status.NetworkInterfaces = ifaceStates
	if err := r.Patch(ctx, machine, client.MergeFrom(base)); err != nil {
		return fmt.Errorf("error patching machine: %w", err)
	}
	return nil
}

func (r *MachineReconciler) reconcile(ctx context.Context, log logr.Logger, machine *computev1alpha1.Machine) (ctrl.Result, error) {
	log.V(1).Info("Reconciling")

	log.V(1).Info("Applying network interfaces")
	ifaceStates, err := r.applyNetworkInterfaces(ctx, log, machine)
	if err != nil {
		return ctrl.Result{}, fmt.Errorf("error applying network interfaces: %w", err)
	}

	log.V(1).Info("Patching status")
	if err := r.patchStatus(ctx, machine, ifaceStates); err != nil {
		return ctrl.Result{}, fmt.Errorf("error patching machine status: %w", err)
	}

	log.V(1).Info("Successfully reconciled")
	return ctrl.Result{}, nil
}

// SetupWithManager sets up the controller with the Manager.
func (r *MachineReconciler) SetupWithManager(mgr ctrl.Manager) error {
	ctx := context.Background()
	log := ctrl.Log.WithName("machine").WithName("setup")

	return ctrl.NewControllerManagedBy(mgr).
		For(&computev1alpha1.Machine{}).
		Owns(&networkingv1alpha1.NetworkInterface{}).
		Watches(
			&source.Kind{Type: &networkingv1alpha1.NetworkInterface{}},
			r.enqueueByMachineNetworkInterfaceReferences(log, ctx),
		).
		Complete(r)
}

func (r *MachineReconciler) enqueueByMachineNetworkInterfaceReferences(log logr.Logger, ctx context.Context) handler.EventHandler {
	return handler.EnqueueRequestsFromMapFunc(func(obj client.Object) []ctrl.Request {
		nic := obj.(*networkingv1alpha1.NetworkInterface)
		log = log.WithValues("NetworkInterfaceKey", client.ObjectKeyFromObject(nic))

		machineList := &computev1alpha1.MachineList{}
		if err := r.List(ctx, machineList,
			client.InNamespace(nic.Namespace),
			client.MatchingFields{
				shared.MachineNetworkInterfaceNamesField: nic.Name,
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
