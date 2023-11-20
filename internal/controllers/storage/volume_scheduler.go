// SPDX-FileCopyrightText: 2023 SAP SE or an SAP affiliate company and IronCore contributors
// SPDX-License-Identifier: Apache-2.0

package storage

import (
	"context"
	"fmt"

	"github.com/go-logr/logr"
	"github.com/ironcore-dev/ironcore/api/common/v1alpha1"
	corev1alpha1 "github.com/ironcore-dev/ironcore/api/core/v1alpha1"
	storagev1alpha1 "github.com/ironcore-dev/ironcore/api/storage/v1alpha1"
	storageclient "github.com/ironcore-dev/ironcore/internal/client/storage"
	"github.com/ironcore-dev/ironcore/internal/controllers/storage/scheduler"
	corev1 "k8s.io/api/core/v1"
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

type VolumeScheduler struct {
	record.EventRecorder
	client.Client

	Cache    *scheduler.Cache
	snapshot *scheduler.Snapshot
}

//+kubebuilder:rbac:groups="",resources=events,verbs=create;patch
//+kubebuilder:rbac:groups=storage.ironcore.dev,resources=volumes,verbs=get;list;watch;update;patch
//+kubebuilder:rbac:groups=storage.ironcore.dev,resources=volumes/status,verbs=get;update;patch
//+kubebuilder:rbac:groups=storage.ironcore.dev,resources=volumepools,verbs=get;list;watch

// Reconcile reconciles the desired with the actual state.
func (s *VolumeScheduler) Reconcile(ctx context.Context, req ctrl.Request) (ctrl.Result, error) {
	log := ctrl.LoggerFrom(ctx)

	volume := &storagev1alpha1.Volume{}
	if err := s.Get(ctx, req.NamespacedName, volume); err != nil {
		return ctrl.Result{}, client.IgnoreNotFound(err)
	}

	if s.skipSchedule(log, volume) {
		return ctrl.Result{}, nil
	}

	return s.reconcileExists(ctx, log, volume)
}

func (s *VolumeScheduler) skipSchedule(log logr.Logger, volume *storagev1alpha1.Volume) bool {
	if !volume.DeletionTimestamp.IsZero() {
		log.V(1).Info("Skipping scheduling for instance", "Reason", "Deleting")
		return true
	}

	if volume.Spec.VolumeClassRef == nil {
		log.V(1).Info("Skipping scheduling for instance", "Reason", "No VolumeClassRef")
		return true
	}

	isAssumed, err := s.Cache.IsAssumedInstance(volume)
	if err != nil {
		log.Error(err, "Error checking whether volume has been assumed")
		return false
	}

	log.V(1).Info("Skipping scheduling for instance", "Reason", "Assumed")
	return isAssumed
}

func (s *VolumeScheduler) matchesLabels(ctx context.Context, pool *scheduler.ContainerInfo, volume *storagev1alpha1.Volume) bool {
	nodeLabels := labels.Set(pool.Node().Labels)
	volumePoolSelector := labels.SelectorFromSet(volume.Spec.VolumePoolSelector)

	return volumePoolSelector.Matches(nodeLabels)
}

func (s *VolumeScheduler) tolerateTaints(ctx context.Context, pool *scheduler.ContainerInfo, volume *storagev1alpha1.Volume) bool {
	return v1alpha1.TolerateTaints(volume.Spec.Tolerations, pool.Node().Spec.Taints)
}

func (s *VolumeScheduler) fitsPool(ctx context.Context, pool *scheduler.ContainerInfo, volume *storagev1alpha1.Volume) bool {
	volumeClassName := volume.Spec.VolumeClassRef.Name

	allocatable, ok := pool.Node().Status.Allocatable[corev1alpha1.ClassCountFor(corev1alpha1.ClassTypeVolumeClass, volumeClassName)]
	if !ok {
		return false
	}

	return allocatable.Cmp(*volume.Spec.Resources.Storage()) >= 0
}

func (s *VolumeScheduler) updateSnapshot() {
	if s.snapshot == nil {
		s.snapshot = s.Cache.Snapshot()
	} else {
		s.snapshot.Update()
	}
}

func (s *VolumeScheduler) assume(assumed *storagev1alpha1.Volume, nodeName string) error {
	assumed.Spec.VolumePoolRef = &corev1.LocalObjectReference{Name: nodeName}
	if err := s.Cache.AssumeInstance(assumed); err != nil {
		return err
	}
	return nil
}

func (s *VolumeScheduler) bindingCycle(ctx context.Context, log logr.Logger, assumedInstance *storagev1alpha1.Volume) error {
	if err := s.bind(ctx, log, assumedInstance); err != nil {
		return fmt.Errorf("error binding: %w", err)
	}
	return nil
}

func (s *VolumeScheduler) bind(ctx context.Context, log logr.Logger, assumed *storagev1alpha1.Volume) error {
	defer func() {
		if err := s.Cache.FinishBinding(assumed); err != nil {
			log.Error(err, "Error finishing cache binding")
		}
	}()

	nonAssumed := assumed.DeepCopy()
	nonAssumed.Spec.VolumePoolRef = nil

	if err := s.Patch(ctx, assumed, client.MergeFrom(nonAssumed)); err != nil {
		return fmt.Errorf("error patching instance: %w", err)
	}
	return nil
}

func (s *VolumeScheduler) reconcileExists(ctx context.Context, log logr.Logger, volume *storagev1alpha1.Volume) (ctrl.Result, error) {
	s.updateSnapshot()

	nodes := s.snapshot.ListNodes()
	if len(nodes) == 0 {
		s.EventRecorder.Event(volume, corev1.EventTypeNormal, outOfCapacity, "No nodes available to schedule volume on")
		return ctrl.Result{}, nil
	}

	var filteredNodes []*scheduler.ContainerInfo
	for _, node := range nodes {
		if !s.tolerateTaints(ctx, node, volume) {
			log.Info("node filtered", "reason", "taints do not match")
			continue
		}
		if !s.matchesLabels(ctx, node, volume) {
			log.Info("node filtered", "reason", "label do not match")
			continue
		}
		if !s.fitsPool(ctx, node, volume) {
			log.Info("node filtered", "reason", "resources do not match")
			continue
		}

		filteredNodes = append(filteredNodes, node)
	}

	if len(filteredNodes) == 0 {
		s.EventRecorder.Event(volume, corev1.EventTypeNormal, outOfCapacity, "No nodes available after filtering to schedule volume on")
		return ctrl.Result{}, nil
	}

	maxAllocatableNode := filteredNodes[0]
	for _, node := range filteredNodes[1:] {
		current := node.MaxAllocatable(volume.Spec.VolumeClassRef.Name)
		if current.Cmp(maxAllocatableNode.MaxAllocatable(volume.Spec.VolumeClassRef.Name)) == 1 {
			maxAllocatableNode = node
		}
	}
	log.V(1).Info("Determined node to schedule on", "NodeName", maxAllocatableNode.Node().Name, "Instances", maxAllocatableNode.NumInstances(), "Allocatable", maxAllocatableNode.MaxAllocatable(volume.Spec.VolumeClassRef.Name))

	log.V(1).Info("Assuming volume to be on node")
	if err := s.assume(volume, maxAllocatableNode.Node().Name); err != nil {
		return ctrl.Result{}, err
	}

	log.V(1).Info("Running binding asynchronously")
	go func() {
		if err := s.bindingCycle(ctx, log, volume); err != nil {
			if err := s.Cache.ForgetInstance(volume); err != nil {
				log.Error(err, "Error forgetting instance")
			}
		}
	}()
	return ctrl.Result{}, nil
}

func (s *VolumeScheduler) enqueueUnscheduledVolumes(ctx context.Context, queue workqueue.RateLimitingInterface) {
	log := ctrl.LoggerFrom(ctx)
	volumeList := &storagev1alpha1.VolumeList{}
	if err := s.List(ctx, volumeList, client.MatchingFields{storageclient.VolumeSpecVolumePoolRefNameField: ""}); err != nil {
		log.Error(fmt.Errorf("could not list volumes w/o volume pool: %w", err), "Error listing volume pools")
		return
	}

	for _, volume := range volumeList.Items {
		if !volume.DeletionTimestamp.IsZero() {
			continue
		}
		if volume.Spec.VolumePoolRef != nil {
			continue
		}
		queue.Add(ctrl.Request{NamespacedName: client.ObjectKeyFromObject(&volume)})
	}
}

func (s *VolumeScheduler) isVolumeAssigned() predicate.Predicate {
	return predicate.NewPredicateFuncs(func(obj client.Object) bool {
		volume := obj.(*storagev1alpha1.Volume)
		return volume.Spec.VolumePoolRef != nil
	})
}

func (s *VolumeScheduler) isVolumeNotAssigned() predicate.Predicate {
	return predicate.NewPredicateFuncs(func(obj client.Object) bool {
		volume := obj.(*storagev1alpha1.Volume)
		return volume.Spec.VolumePoolRef == nil
	})
}

func (s *VolumeScheduler) handleVolume() handler.EventHandler {
	return handler.Funcs{
		CreateFunc: func(ctx context.Context, evt event.CreateEvent, queue workqueue.RateLimitingInterface) {
			volume := evt.Object.(*storagev1alpha1.Volume)
			log := ctrl.LoggerFrom(ctx)

			if err := s.Cache.AddInstance(volume); err != nil {
				log.Error(err, "Error adding volume to cache")
			}
		},
		UpdateFunc: func(ctx context.Context, evt event.UpdateEvent, queue workqueue.RateLimitingInterface) {
			log := ctrl.LoggerFrom(ctx)

			oldInstance := evt.ObjectOld.(*storagev1alpha1.Volume)
			newInstance := evt.ObjectNew.(*storagev1alpha1.Volume)
			if err := s.Cache.UpdateInstance(oldInstance, newInstance); err != nil {
				log.Error(err, "Error updating volume in cache")
			}
		},
		DeleteFunc: func(ctx context.Context, evt event.DeleteEvent, queue workqueue.RateLimitingInterface) {
			log := ctrl.LoggerFrom(ctx)

			instance := evt.Object.(*storagev1alpha1.Volume)
			if err := s.Cache.RemoveInstance(instance); err != nil {
				log.Error(err, "Error adding volume to cache")
			}
		},
	}
}

func (s *VolumeScheduler) handleVolumePool() handler.EventHandler {
	return handler.Funcs{
		CreateFunc: func(ctx context.Context, evt event.CreateEvent, queue workqueue.RateLimitingInterface) {
			pool := evt.Object.(*storagev1alpha1.VolumePool)
			s.Cache.AddContainer(pool)
			s.enqueueUnscheduledVolumes(ctx, queue)
		},
		UpdateFunc: func(ctx context.Context, evt event.UpdateEvent, queue workqueue.RateLimitingInterface) {
			oldPool := evt.ObjectOld.(*storagev1alpha1.VolumePool)
			newPool := evt.ObjectNew.(*storagev1alpha1.VolumePool)
			s.Cache.UpdateContainer(oldPool, newPool)
			s.enqueueUnscheduledVolumes(ctx, queue)
		},
		DeleteFunc: func(ctx context.Context, evt event.DeleteEvent, queue workqueue.RateLimitingInterface) {
			log := ctrl.LoggerFrom(ctx)

			pool := evt.Object.(*storagev1alpha1.VolumePool)
			if err := s.Cache.RemoveContainer(pool); err != nil {
				log.Error(err, "Error removing volume pool from cache")
			}
		},
	}
}

func (s *VolumeScheduler) SetupWithManager(mgr manager.Manager) error {
	return ctrl.NewControllerManagedBy(mgr).
		Named("volume-scheduler").
		WithOptions(controller.Options{
			// Only a single concurrent reconcile since it is serialized on the scheduling algorithm's node fitting.
			MaxConcurrentReconciles: 1,
		}).
		// Enqueue unscheduled volumes.
		For(&storagev1alpha1.Volume{},
			builder.WithPredicates(
				s.isVolumeNotAssigned(),
			),
		).
		Watches(
			&storagev1alpha1.Volume{},
			s.handleVolume(),
			builder.WithPredicates(
				s.isVolumeAssigned(),
			),
		).
		// Enqueue unscheduled volumes if a volume pool w/ required volume classes becomes available.
		Watches(
			&storagev1alpha1.VolumePool{},
			s.handleVolumePool(),
		).
		Complete(s)
}
