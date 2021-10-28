package compute

import (
	"context"
	"fmt"
	"github.com/go-logr/logr"
	computev1alpha1 "github.com/onmetal/onmetal-api/apis/compute/v1alpha1"
	corev1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/util/sets"
	"k8s.io/client-go/tools/record"
	"k8s.io/client-go/util/workqueue"
	"math/rand"
	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/builder"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/event"
	"sigs.k8s.io/controller-runtime/pkg/handler"
	"sigs.k8s.io/controller-runtime/pkg/manager"
	"sigs.k8s.io/controller-runtime/pkg/predicate"
	"sigs.k8s.io/controller-runtime/pkg/source"
)

const (
	machinePoolStatusAvailableMachineClassesNameField = ".status.availableMachineClasses[*].name"
	machineSpecMachinePoolNameField                   = ".spec.machinePool.name"
)

type MachineScheduler struct {
	client.Client
	Events record.EventRecorder
}

//+kubebuilder:rbac:groups=compute.onmetal.de,resources=machines,verbs=get;list;watch;update;patch
//+kubebuilder:rbac:groups=compute.onmetal.de,resources=machines/status,verbs=get;update;patch
//+kubebuilder:rbac:groups=compute.onmetal.de,resources=machinepools,verbs=get;list;watch

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
	if machine.Spec.MachinePool.Name != "" {
		log.Info("Machine is already assigned")
		return ctrl.Result{}, nil
	}
	return s.schedule(ctx, log, machine)
}

func (s *MachineScheduler) schedule(ctx context.Context, log logr.Logger, machine *computev1alpha1.Machine) (ctrl.Result, error) {
	log.Info("Scheduling machine")
	list := &computev1alpha1.MachinePoolList{}
	if err := s.List(ctx, list, client.MatchingFields{machinePoolStatusAvailableMachineClassesNameField: machine.Spec.MachineClass.Name}); err != nil {
		return ctrl.Result{}, fmt.Errorf("error listing machine pools: %w", err)
	}

	if machine.Status.State != computev1alpha1.MachineStatePending {
		base := machine.DeepCopy()
		machine.Status.State = computev1alpha1.MachineStatePending
		if err := s.Status().Patch(ctx, machine, client.MergeFrom(base)); err != nil {
			return ctrl.Result{}, fmt.Errorf("error patching machine state to pending: %w", err)
		}
		return ctrl.Result{Requeue: true}, nil
	}

	if len(list.Items) == 0 {
		log.Info("No machine pool available for machine class", "MachineClass", machine.Spec.MachineClass.Name)
		s.Events.Eventf(machine, corev1.EventTypeNormal, "CannotSchedule", "No MachinePool found for MachineClass %s", machine.Spec.MachineClass.Name)
		return ctrl.Result{}, nil
	}

	// Get a random pool to distribute evenly.
	// TODO: Instead of random distribution, try to come up w/ metrics that include usage of each pool to
	// avoid unfortunate random distribution of items.
	pool := list.Items[rand.Intn(len(list.Items))]
	log = log.WithValues("MachinePool", pool.Name)
	base := machine.DeepCopy()
	machine.Spec.MachinePool.Name = pool.Name
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
	if err := s.List(ctx, list, client.MatchingFields{machineSpecMachinePoolNameField: ""}); err != nil {
		log.Error(fmt.Errorf("could not list machines w/o machine pool: %w", err), "Error listing machine pools")
		return
	}

	availableClassNames := sets.NewString()
	for _, availableMachineClass := range pool.Status.AvailableMachineClasses {
		availableClassNames.Insert(availableMachineClass.Name)
	}

	for _, machine := range list.Items {
		if availableClassNames.Has(machine.Spec.MachineClass.Name) {
			queue.Add(ctrl.Request{NamespacedName: client.ObjectKeyFromObject(&machine)})
		}
	}
}

func (s *MachineScheduler) SetupWithManager(mgr manager.Manager) error {
	ctx := context.Background()
	ctx = ctrl.LoggerInto(ctx, ctrl.Log.WithName("machine-scheduler").WithName("setup"))

	if err := mgr.GetFieldIndexer().IndexField(ctx, &computev1alpha1.MachinePool{}, machinePoolStatusAvailableMachineClassesNameField, func(object client.Object) []string {
		machinePool := object.(*computev1alpha1.MachinePool)
		names := make([]string, 0, len(machinePool.Status.AvailableMachineClasses))
		for _, availableMachineClass := range machinePool.Status.AvailableMachineClasses {
			names = append(names, availableMachineClass.Name)
		}
		return names
	}); err != nil {
		return fmt.Errorf("could not setup field indexer for %s: %w", machinePoolStatusAvailableMachineClassesNameField, err)
	}

	if err := mgr.GetFieldIndexer().IndexField(ctx, &computev1alpha1.Machine{}, machineSpecMachinePoolNameField, func(object client.Object) []string {
		machine := object.(*computev1alpha1.Machine)
		return []string{machine.Spec.MachinePool.Name}
	}); err != nil {
		return fmt.Errorf("could not setup field indexer for %s: %w", machineSpecMachinePoolNameField, err)
	}

	return ctrl.NewControllerManagedBy(mgr).
		Named("machine-scheduler").
		// Enqueue unscheduled machines.
		For(&computev1alpha1.Machine{},
			builder.WithPredicates(
				predicate.NewPredicateFuncs(func(object client.Object) bool {
					machine := object.(*computev1alpha1.Machine)
					return machine.DeletionTimestamp.IsZero() && machine.Spec.MachinePool.Name == ""
				}),
			),
		).
		// Enqueue unscheduled machines if a machine pool w/ required machine classes becomes available.
		Watches(&source.Kind{Type: &computev1alpha1.MachinePool{}},
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
