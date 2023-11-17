// Copyright 2021 IronCore authors
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

	"github.com/go-logr/logr"
	"github.com/ironcore-dev/ironcore/api/common/v1alpha1"
	computev1alpha1 "github.com/ironcore-dev/ironcore/api/compute/v1alpha1"
	corev1alpha1 "github.com/ironcore-dev/ironcore/api/core/v1alpha1"
	computeclient "github.com/ironcore-dev/ironcore/internal/client/compute"
	"github.com/ironcore-dev/ironcore/internal/controllers/compute/scheduler"
	corev1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/api/resource"
	"k8s.io/apimachinery/pkg/labels"
	"k8s.io/client-go/tools/record"
	"k8s.io/client-go/util/workqueue"
	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/builder"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/controller"
	"sigs.k8s.io/controller-runtime/pkg/event"
	"sigs.k8s.io/controller-runtime/pkg/handler"
	"sigs.k8s.io/controller-runtime/pkg/manager"
	"sigs.k8s.io/controller-runtime/pkg/predicate"
)

const (
	outOfCapacity = "OutOfCapacity"
)

type MachineScheduler struct {
	record.EventRecorder
	client.Client

	Cache    *scheduler.Cache
	snapshot *scheduler.Snapshot
}

//+kubebuilder:rbac:groups="",resources=events,verbs=create;patch
//+kubebuilder:rbac:groups=compute.ironcore.dev,resources=machines,verbs=get;list;watch;update;patch
//+kubebuilder:rbac:groups=compute.ironcore.dev,resources=machines/status,verbs=get;update;patch
//+kubebuilder:rbac:groups=compute.ironcore.dev,resources=machinepools,verbs=get;list;watch

// Reconcile reconciles the desired with the actual state.
func (s *MachineScheduler) Reconcile(ctx context.Context, req ctrl.Request) (ctrl.Result, error) {
	log := ctrl.LoggerFrom(ctx)

	machine := &computev1alpha1.Machine{}
	if err := s.Get(ctx, req.NamespacedName, machine); err != nil {
		return ctrl.Result{}, client.IgnoreNotFound(err)
	}

	if s.skipSchedule(log, machine) {
		log.V(1).Info("Skipping scheduling for instance")
		return ctrl.Result{}, nil
	}

	return s.reconcileExists(ctx, log, machine)
}

func (s *MachineScheduler) skipSchedule(log logr.Logger, machine *computev1alpha1.Machine) bool {
	if !machine.DeletionTimestamp.IsZero() {
		return true
	}

	isAssumed, err := s.Cache.IsAssumedInstance(machine)
	if err != nil {
		log.Error(err, "Error checking whether machine has been assumed")
		return false
	}
	return isAssumed
}

func (s *MachineScheduler) matchesLabels(ctx context.Context, pool *scheduler.ContainerInfo, machine *computev1alpha1.Machine) bool {
	nodeLabels := labels.Set(pool.Node().Labels)
	machinePoolSelector := labels.SelectorFromSet(machine.Spec.MachinePoolSelector)

	return machinePoolSelector.Matches(nodeLabels)
}

func (s *MachineScheduler) tolerateTaints(ctx context.Context, pool *scheduler.ContainerInfo, machine *computev1alpha1.Machine) bool {
	return v1alpha1.TolerateTaints(machine.Spec.Tolerations, pool.Node().Spec.Taints)
}

func (s *MachineScheduler) fitsPool(ctx context.Context, pool *scheduler.ContainerInfo, machine *computev1alpha1.Machine) bool {
	machineClassName := machine.Spec.MachineClassRef.Name

	allocatable, ok := pool.Node().Status.Allocatable[corev1alpha1.ClassCountFor(corev1alpha1.ClassTypeMachineClass, machineClassName)]
	if !ok {
		return false
	}

	return allocatable.Cmp(*resource.NewQuantity(1, resource.DecimalSI)) >= 0
}

func (s *MachineScheduler) reconcileExists(ctx context.Context, log logr.Logger, machine *computev1alpha1.Machine) (ctrl.Result, error) {
	s.updateSnapshot()

	nodes := s.snapshot.ListNodes()
	if len(nodes) == 0 {
		s.EventRecorder.Event(machine, corev1.EventTypeNormal, outOfCapacity, "No nodes available to schedule machine on")
		return ctrl.Result{}, nil
	}

	var filteredNodes []*scheduler.ContainerInfo
	for _, node := range nodes {
		if !s.tolerateTaints(ctx, node, machine) {
			log.Info("node filtered", "reason", "taints do not match")
			continue
		}
		if !s.matchesLabels(ctx, node, machine) {
			log.Info("node filtered", "reason", "label do not match")
			continue
		}
		if !s.fitsPool(ctx, node, machine) {
			log.Info("node filtered", "reason", "resources do not match")
			continue
		}

		filteredNodes = append(filteredNodes, node)
	}

	if len(filteredNodes) == 0 {
		s.EventRecorder.Event(machine, corev1.EventTypeNormal, outOfCapacity, "No nodes available after filtering to schedule machine on")
		return ctrl.Result{}, nil
	}

	maxAllocatableNode := filteredNodes[0]
	for _, node := range filteredNodes[1:] {
		if node.MaxAllocatable(machine.Spec.MachineClassRef.Name) > maxAllocatableNode.MaxAllocatable(machine.Spec.MachineClassRef.Name) {
			maxAllocatableNode = node
		}
	}
	log.V(1).Info("Determined node to schedule on", "NodeName", maxAllocatableNode.Node().Name, "Instances", maxAllocatableNode.NumInstances(), "Allocatable", maxAllocatableNode.MaxAllocatable(machine.Spec.MachineClassRef.Name))

	log.V(1).Info("Assuming machine to be on node")
	if err := s.assume(machine, maxAllocatableNode.Node().Name); err != nil {
		return ctrl.Result{}, err
	}

	log.V(1).Info("Running binding asynchronously")
	go func() {
		if err := s.bindingCycle(ctx, log, machine); err != nil {
			if err := s.Cache.ForgetInstance(machine); err != nil {
				log.Error(err, "Error forgetting instance")
			}
		}
	}()
	return ctrl.Result{}, nil
}

func (s *MachineScheduler) updateSnapshot() {
	if s.snapshot == nil {
		s.snapshot = s.Cache.Snapshot()
	} else {
		s.snapshot.Update()
	}
}

func (s *MachineScheduler) assume(assumed *computev1alpha1.Machine, nodeName string) error {
	assumed.Spec.MachinePoolRef = &corev1.LocalObjectReference{Name: nodeName}
	if err := s.Cache.AssumeInstance(assumed); err != nil {
		return err
	}
	return nil
}

func (s *MachineScheduler) bindingCycle(ctx context.Context, log logr.Logger, assumedInstance *computev1alpha1.Machine) error {
	if err := s.bind(ctx, log, assumedInstance); err != nil {
		return fmt.Errorf("error binding: %w", err)
	}
	return nil
}

func (s *MachineScheduler) bind(ctx context.Context, log logr.Logger, assumed *computev1alpha1.Machine) error {
	defer func() {
		if err := s.Cache.FinishBinding(assumed); err != nil {
			log.Error(err, "Error finishing cache binding")
		}
	}()

	nonAssumed := assumed.DeepCopy()
	nonAssumed.Spec.MachinePoolRef = nil

	if err := s.Patch(ctx, assumed, client.MergeFrom(nonAssumed)); err != nil {
		return fmt.Errorf("error patching instance: %w", err)
	}
	return nil
}

func (s *MachineScheduler) enqueueUnscheduledMachines(ctx context.Context, queue workqueue.RateLimitingInterface) {
	log := ctrl.LoggerFrom(ctx)
	machineList := &computev1alpha1.MachineList{}
	if err := s.List(ctx, machineList, client.MatchingFields{computeclient.MachineSpecMachinePoolRefNameField: ""}); err != nil {
		log.Error(fmt.Errorf("could not list machines w/o machine pool: %w", err), "Error listing machine pools")
		return
	}

	for _, machine := range machineList.Items {
		if !machine.DeletionTimestamp.IsZero() {
			continue
		}
		if machine.Spec.MachinePoolRef != nil {
			continue
		}
		queue.Add(ctrl.Request{NamespacedName: client.ObjectKeyFromObject(&machine)})
	}
}

func (s *MachineScheduler) isMachineAssigned() predicate.Predicate {
	return predicate.NewPredicateFuncs(func(obj client.Object) bool {
		machine := obj.(*computev1alpha1.Machine)
		return machine.Spec.MachinePoolRef != nil
	})
}

func (s *MachineScheduler) isMachineNotAssigned() predicate.Predicate {
	return predicate.NewPredicateFuncs(func(obj client.Object) bool {
		machine := obj.(*computev1alpha1.Machine)
		return machine.Spec.MachinePoolRef == nil
	})
}

func (s *MachineScheduler) handleMachine() handler.EventHandler {
	return handler.Funcs{
		CreateFunc: func(ctx context.Context, evt event.CreateEvent, queue workqueue.RateLimitingInterface) {
			machine := evt.Object.(*computev1alpha1.Machine)
			log := ctrl.LoggerFrom(ctx)

			if err := s.Cache.AddInstance(machine); err != nil {
				log.Error(err, "Error adding machine to cache")
			}
		},
		UpdateFunc: func(ctx context.Context, evt event.UpdateEvent, queue workqueue.RateLimitingInterface) {
			log := ctrl.LoggerFrom(ctx)

			oldInstance := evt.ObjectOld.(*computev1alpha1.Machine)
			newInstance := evt.ObjectNew.(*computev1alpha1.Machine)
			if err := s.Cache.UpdateInstance(oldInstance, newInstance); err != nil {
				log.Error(err, "Error updating machine in cache")
			}
		},
		DeleteFunc: func(ctx context.Context, evt event.DeleteEvent, queue workqueue.RateLimitingInterface) {
			log := ctrl.LoggerFrom(ctx)

			instance := evt.Object.(*computev1alpha1.Machine)
			if err := s.Cache.RemoveInstance(instance); err != nil {
				log.Error(err, "Error adding machine to cache")
			}
		},
	}
}

func (s *MachineScheduler) handleMachinePool() handler.EventHandler {
	return handler.Funcs{
		CreateFunc: func(ctx context.Context, evt event.CreateEvent, queue workqueue.RateLimitingInterface) {
			pool := evt.Object.(*computev1alpha1.MachinePool)
			s.Cache.AddContainer(pool)
			s.enqueueUnscheduledMachines(ctx, queue)
		},
		UpdateFunc: func(ctx context.Context, evt event.UpdateEvent, queue workqueue.RateLimitingInterface) {
			oldPool := evt.ObjectOld.(*computev1alpha1.MachinePool)
			newPool := evt.ObjectNew.(*computev1alpha1.MachinePool)
			s.Cache.UpdateContainer(oldPool, newPool)
			s.enqueueUnscheduledMachines(ctx, queue)
		},
		DeleteFunc: func(ctx context.Context, evt event.DeleteEvent, queue workqueue.RateLimitingInterface) {
			log := ctrl.LoggerFrom(ctx)

			pool := evt.Object.(*computev1alpha1.MachinePool)
			if err := s.Cache.RemoveContainer(pool); err != nil {
				log.Error(err, "Error removing machine pool from cache")
			}
		},
	}
}

func (s *MachineScheduler) SetupWithManager(mgr manager.Manager) error {
	return ctrl.NewControllerManagedBy(mgr).
		Named("machine-scheduler").
		WithOptions(controller.Options{
			// Only a single concurrent reconcile since it is serialized on the scheduling algorithm's node fitting.
			MaxConcurrentReconciles: 1,
		}).
		// Enqueue unscheduled machines.
		For(&computev1alpha1.Machine{},
			builder.WithPredicates(
				s.isMachineNotAssigned(),
			),
		).
		Watches(
			&computev1alpha1.Machine{},
			s.handleMachine(),
			builder.WithPredicates(
				s.isMachineAssigned(),
			),
		).
		// Enqueue unscheduled machines if a machine pool w/ required machine classes becomes available.
		Watches(
			&computev1alpha1.MachinePool{},
			s.handleMachinePool(),
		).
		Complete(s)
}
