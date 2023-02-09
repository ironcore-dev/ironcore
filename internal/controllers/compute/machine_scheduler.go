// Copyright 2021 OnMetal authors
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
	"fmt"
	"math/rand"

	"github.com/go-logr/logr"
	computeclient "github.com/onmetal/onmetal-api/internal/client/compute"
	corev1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/labels"
	"k8s.io/apimachinery/pkg/util/sets"
	"k8s.io/client-go/tools/record"
	"k8s.io/client-go/util/workqueue"
	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/builder"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/event"
	"sigs.k8s.io/controller-runtime/pkg/handler"
	"sigs.k8s.io/controller-runtime/pkg/manager"
	"sigs.k8s.io/controller-runtime/pkg/predicate"
	"sigs.k8s.io/controller-runtime/pkg/source"

	"github.com/onmetal/onmetal-api/api/common/v1alpha1"
	computev1alpha1 "github.com/onmetal/onmetal-api/api/compute/v1alpha1"
)

type MachineScheduler struct {
	record.EventRecorder
	client.Client
}

//+kubebuilder:rbac:groups="",resources=events,verbs=create;patch
//+kubebuilder:rbac:groups=compute.api.onmetal.de,resources=machines,verbs=get;list;watch;update;patch
//+kubebuilder:rbac:groups=compute.api.onmetal.de,resources=machines/status,verbs=get;update;patch
//+kubebuilder:rbac:groups=compute.api.onmetal.de,resources=machinepools,verbs=get;list;watch

// Reconcile reconciles the desired with the actual state.
func (s *MachineScheduler) Reconcile(ctx context.Context, req ctrl.Request) (ctrl.Result, error) {
	log := ctrl.LoggerFrom(ctx)

	machine := &computev1alpha1.Machine{}
	if err := s.Get(ctx, req.NamespacedName, machine); err != nil {
		return ctrl.Result{}, client.IgnoreNotFound(err)
	}

	if !machine.DeletionTimestamp.IsZero() {
		log.Info("Machine is already deleting")
		return ctrl.Result{}, nil
	}
	if machine.Spec.MachinePoolRef != nil {
		log.Info("Machine is already assigned", "MachinePoolRef", machine.Spec.MachinePoolRef)
		return ctrl.Result{}, nil
	}
	return s.schedule(ctx, log, machine)
}

func (s *MachineScheduler) schedule(ctx context.Context, log logr.Logger, machine *computev1alpha1.Machine) (ctrl.Result, error) {
	log.Info("Scheduling machine")
	if machine.Status.State != computev1alpha1.MachineStatePending {
		base := machine.DeepCopy()
		machine.Status.State = computev1alpha1.MachineStatePending
		if err := s.Status().Patch(ctx, machine, client.MergeFrom(base)); err != nil {
			return ctrl.Result{}, fmt.Errorf("error patching machine state to pending: %w", err)
		}
		return ctrl.Result{Requeue: true}, nil
	}

	list := &computev1alpha1.MachinePoolList{}
	if err := s.List(ctx, list,
		client.MatchingFields{computeclient.MachinePoolAvailableMachineClassesField: machine.Spec.MachineClassRef.Name},
		client.MatchingLabels(machine.Spec.MachinePoolSelector),
	); err != nil {
		return ctrl.Result{}, fmt.Errorf("error listing machine pools: %w", err)
	}

	var available []computev1alpha1.MachinePool
	for _, pool := range list.Items {
		if pool.DeletionTimestamp.IsZero() {
			available = append(available, pool)
		}
	}
	if len(available) == 0 {
		log.Info("No machine pool available for machine class", "MachineClassRef", machine.Spec.MachineClassRef.Name)
		s.Eventf(machine, corev1.EventTypeNormal, "CannotSchedule", "No MachinePoolRef found for MachineClassRef %s", machine.Spec.MachineClassRef.Name)
		return ctrl.Result{}, nil
	}

	// Filter machine pools by checking if the machine tolerates all the taints of a machine pool
	var filtered []computev1alpha1.MachinePool
	for _, pool := range available {
		if v1alpha1.TolerateTaints(machine.Spec.Tolerations, pool.Spec.Taints) {
			filtered = append(filtered, pool)
		}
	}
	if len(filtered) == 0 {
		log.Info("No machine pool tolerated by the machine", "Tolerations", machine.Spec.Tolerations)
		s.Eventf(machine, corev1.EventTypeNormal, "CannotSchedule", "No MachinePoolRef tolerated by tolerations: %s", &machine.Spec.Tolerations)
		return ctrl.Result{}, nil
	}
	available = filtered

	// Get a random pool to distribute evenly.
	// TODO: Instead of random distribution, try to come up w/ metrics that include usage of each pool to
	// avoid unfortunate random distribution of items.
	pool := available[rand.Intn(len(available))]
	log = log.WithValues("MachinePoolRef", pool.Name)
	base := machine.DeepCopy()
	machine.Spec.MachinePoolRef = &corev1.LocalObjectReference{Name: pool.Name}
	log.Info("Patching machine")
	if err := s.Patch(ctx, machine, client.MergeFrom(base)); err != nil {
		return ctrl.Result{}, fmt.Errorf("error scheduling machine on pool: %w", err)
	}

	log.Info("Successfully assigned machine")
	return ctrl.Result{}, nil
}

func (s *MachineScheduler) enqueueMatchingUnscheduledMachines(ctx context.Context, pool *computev1alpha1.MachinePool, queue workqueue.RateLimitingInterface) {
	log := ctrl.LoggerFrom(ctx)
	list := &computev1alpha1.MachineList{}
	if err := s.List(ctx, list, client.MatchingFields{computeclient.MachineSpecMachinePoolRefNameField: ""}); err != nil {
		log.Error(fmt.Errorf("could not list machines w/o machine pool: %w", err), "Error listing machine pools")
		return
	}

	availableClassNames := sets.NewString()
	for _, availableMachineClass := range pool.Status.AvailableMachineClasses {
		availableClassNames.Insert(availableMachineClass.Name)
	}

	for _, machine := range list.Items {
		machinePoolSelector := labels.SelectorFromSet(machine.Spec.MachinePoolSelector)
		if availableClassNames.Has(machine.Spec.MachineClassRef.Name) && machinePoolSelector.Matches(labels.Set(pool.Labels)) {
			queue.Add(ctrl.Request{NamespacedName: client.ObjectKeyFromObject(&machine)})
		}
	}
}

func (s *MachineScheduler) SetupWithManager(mgr manager.Manager) error {
	ctx := context.Background()
	ctx = ctrl.LoggerInto(ctx, ctrl.Log.WithName("machine-scheduler").WithName("setup"))

	return ctrl.NewControllerManagedBy(mgr).
		Named("machine-scheduler").
		// Enqueue unscheduled machines.
		For(&computev1alpha1.Machine{},
			builder.WithPredicates(
				predicate.NewPredicateFuncs(func(object client.Object) bool {
					machine := object.(*computev1alpha1.Machine)
					return machine.DeletionTimestamp.IsZero() && machine.Spec.MachinePoolRef == nil
				}),
			),
		).
		// Enqueue unscheduled machines if a machine pool w/ required machine classes becomes available.
		Watches(
			&source.Kind{Type: &computev1alpha1.MachinePool{}},
			handler.Funcs{
				CreateFunc: func(event event.CreateEvent, queue workqueue.RateLimitingInterface) {
					pool := event.Object.(*computev1alpha1.MachinePool)
					s.enqueueMatchingUnscheduledMachines(ctx, pool, queue)
				},
				UpdateFunc: func(event event.UpdateEvent, queue workqueue.RateLimitingInterface) {
					pool := event.ObjectNew.(*computev1alpha1.MachinePool)
					s.enqueueMatchingUnscheduledMachines(ctx, pool, queue)
				},
				GenericFunc: func(event event.GenericEvent, queue workqueue.RateLimitingInterface) {
					pool := event.Object.(*computev1alpha1.MachinePool)
					s.enqueueMatchingUnscheduledMachines(ctx, pool, queue)
				},
			},
		).
		Complete(s)
}
