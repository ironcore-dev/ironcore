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

	"k8s.io/apimachinery/pkg/util/sets"

	"github.com/go-logr/logr"
	"inet.af/netaddr"
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"
	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/builder"
	"sigs.k8s.io/controller-runtime/pkg/client"

	commonv1alpha1 "github.com/onmetal/onmetal-api/apis/common/v1alpha1"
	computev1alpha1 "github.com/onmetal/onmetal-api/apis/compute/v1alpha1"
	networkv1alpha1 "github.com/onmetal/onmetal-api/apis/network/v1alpha1"
	"github.com/onmetal/onmetal-api/predicates"
)

const (
	machineOwnerLabel          = "compute.onmetal.de/machine-owner"
	machineInterfaceFieldOwner = client.FieldOwner("compute.onmetal.de/machine-iface")
)

// MachineReconciler reconciles a Machine object
type MachineReconciler struct {
	client.Client
	Scheme *runtime.Scheme
}

//+kubebuilder:rbac:groups=compute.onmetal.de,resources=machines,verbs=get;list;watch;create;update;patch;delete
//+kubebuilder:rbac:groups=compute.onmetal.de,resources=machines/status,verbs=get;update;patch
//+kubebuilder:rbac:groups=compute.onmetal.de,resources=machines/finalizers,verbs=update
//+kubebuilder:rbac:groups=network.onmetal.de,resources=ipamranges,verbs=get;list;watch;create;update;patch

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

// SetupWithManager sets up the controller with the Manager.
func (r *MachineReconciler) SetupWithManager(mgr ctrl.Manager) error {
	if err := mgr.GetFieldIndexer().IndexField(
		context.Background(),
		&computev1alpha1.Machine{},
		machineClassNameField,
		func(object client.Object) []string {
			m := object.(*computev1alpha1.Machine)
			if m.Spec.MachineClass.Name == "" {
				return nil
			}
			return []string{m.Spec.MachineClass.Name}
		},
	); err != nil {
		return fmt.Errorf("indexing the field %s: %w", machineClassNameField, err)
	}

	return ctrl.NewControllerManagedBy(mgr).
		For(&computev1alpha1.Machine{}).
		Owns(&networkv1alpha1.IPAMRange{}, builder.WithPredicates(predicates.IPAMRangeAllocationsChangedPredicate{})).
		Complete(r)
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

func (r *MachineReconciler) reconcile(ctx context.Context, log logr.Logger, machine *computev1alpha1.Machine) (ctrl.Result, error) {
	var (
		interfaceStatuses  []computev1alpha1.InterfaceStatus
		existingIPAMRanges = &networkv1alpha1.IPAMRangeList{}
		ifaceCheckList     = sets.NewString()
	)

	// Delete unused IPAMRanges associated with Machine interfaces
	for _, iface := range machine.Spec.Interfaces {
		var (
			request  networkv1alpha1.IPAMRangeRequest
			ipamName = computev1alpha1.MachineInterfaceIPAMRangeName(machine.Name, iface.Name)
		)
		ifaceCheckList.Insert(ipamName)

		if iface.IP != nil {
			request.IPs = commonv1alpha1.NewIPRangePtr(netaddr.IPRangeFrom(iface.IP.IP, iface.IP.IP))
		} else {
			request.IPCount = 1
		}

		ifaceIPAMRange := &networkv1alpha1.IPAMRange{
			TypeMeta: metav1.TypeMeta{
				APIVersion: networkv1alpha1.GroupVersion.String(),
				Kind:       networkv1alpha1.IPAMRangeGK.Kind,
			},
			ObjectMeta: metav1.ObjectMeta{
				Namespace: machine.Namespace,
				Name:      ipamName,
				Labels: map[string]string{
					machineOwnerLabel: machine.Name,
				},
			},
			Spec: networkv1alpha1.IPAMRangeSpec{
				Parent: &corev1.LocalObjectReference{
					Name: networkv1alpha1.SubnetIPAMName(iface.Target.Name),
				},
				Requests: []networkv1alpha1.IPAMRangeRequest{request},
			},
		}
		if err := ctrl.SetControllerReference(machine, ifaceIPAMRange, r.Scheme); err != nil {
			return ctrl.Result{}, fmt.Errorf("could not own iface %s ipam range: %w", iface.Name, err)
		}
		if err := r.Patch(ctx, ifaceIPAMRange, client.Apply, machineInterfaceFieldOwner); err != nil {
			return ctrl.Result{}, fmt.Errorf("could not create iface %s ipam range: %w", iface.Name, err)
		}

		for _, allocation := range ifaceIPAMRange.Status.Allocations {
			if allocation.State != networkv1alpha1.IPAMRangeAllocationFree || allocation.IPs == nil {
				continue
			}

			ip := allocation.IPs.From
			interfaceStatuses = append(interfaceStatuses, computev1alpha1.InterfaceStatus{
				Name:     iface.Name,
				IP:       ip,
				Priority: iface.Priority,
			})
		}
	}

	// Delete IPAMRanges associated with interfaces deleted from machine
	if err := r.List(ctx, existingIPAMRanges, client.MatchingLabels{machineOwnerLabel: machine.Name}); err != nil {
		return ctrl.Result{}, fmt.Errorf("failed to list IPAMRanges: %w", err)
	}
	for _, i := range existingIPAMRanges.Items {
		if ifaceCheckList.Has(i.Name) {
			continue
		}

		ipamRange := &networkv1alpha1.IPAMRange{
			ObjectMeta: metav1.ObjectMeta{
				Namespace: machine.Namespace,
				Name:      i.Name,
			},
		}
		if err := r.Delete(ctx, ipamRange); err != nil {
			return ctrl.Result{}, fmt.Errorf("failed to delete IPAMRange: %w", err)
		}
	}

	outdatedStatusMachine := machine.DeepCopy()
	machine.Status.Interfaces = interfaceStatuses
	if err := r.Status().Patch(ctx, machine, client.MergeFrom(outdatedStatusMachine)); err != nil {
		return ctrl.Result{}, fmt.Errorf("could not update status: %w", err)
	}
	return ctrl.Result{}, nil
}

const machineClassNameField = ".spec.machineClass.name"
