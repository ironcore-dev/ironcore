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

package ipam

import (
	"context"
	"fmt"
	"math/rand"

	"github.com/go-logr/logr"
	ipamv1alpha1 "github.com/onmetal/onmetal-api/apis/ipam/v1alpha1"
	"inet.af/netaddr"
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/labels"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/client-go/tools/record"
	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/handler"
	"sigs.k8s.io/controller-runtime/pkg/manager"
	"sigs.k8s.io/controller-runtime/pkg/source"
)

type PrefixAllocationScheduler struct {
	client.Client
	Scheme *runtime.Scheme
	Events record.EventRecorder
}

//+kubebuilder:rbac:groups=ipam.api.onmetal.de,resources=prefixes,verbs=get;list;watch
//+kubebuilder:rbac:groups=ipam.api.onmetal.de,resources=prefixallocations,verbs=get;list;watch;update;patch
//+kubebuilder:rbac:groups=ipam.api.onmetal.de,resources=prefixallocations/status,verbs=get;update;patch

// Reconcile is part of the main kubernetes reconciliation loop which aims to
// move the current state of the cluster closer to the desired state.
func (s *PrefixAllocationScheduler) Reconcile(ctx context.Context, req ctrl.Request) (ctrl.Result, error) {
	log := ctrl.LoggerFrom(ctx)
	allocation := &ipamv1alpha1.PrefixAllocation{}
	if err := s.Get(ctx, req.NamespacedName, allocation); err != nil {
		return ctrl.Result{}, client.IgnoreNotFound(err)
	}

	return s.reconcileExists(ctx, log, allocation)
}

func (s *PrefixAllocationScheduler) reconcileExists(ctx context.Context, log logr.Logger, allocation *ipamv1alpha1.PrefixAllocation) (ctrl.Result, error) {
	if !allocation.DeletionTimestamp.IsZero() {
		return ctrl.Result{}, nil
	}
	return s.reconcile(ctx, log, allocation)
}

func isPrefixAllocatedAndNotDeleting(prefix *ipamv1alpha1.Prefix) bool {
	return prefix.DeletionTimestamp.IsZero() &&
		prefix.Status.Phase == ipamv1alpha1.PrefixPhaseAllocated
}

func (s *PrefixAllocationScheduler) prefixForAllocation(ctx context.Context, log logr.Logger, allocation *ipamv1alpha1.PrefixAllocation) (string, error) {
	sel, err := metav1.LabelSelectorAsSelector(allocation.Spec.PrefixSelector)
	if err != nil {
		return "", fmt.Errorf("error building label selector: %w", err)
	}

	list := &ipamv1alpha1.PrefixList{}
	if err := s.List(ctx, list,
		client.InNamespace(allocation.Namespace),
		client.MatchingLabelsSelector{Selector: sel},
		client.MatchingFields{PrefixSpecIPFamilyField: string(allocation.Spec.IPFamily)},
	); err != nil {
		return "", fmt.Errorf("error listing prefixes: %w", err)
	}

	var suitable []ipamv1alpha1.Prefix
	for _, prefix := range list.Items {
		if !isPrefixAllocatedAndNotDeleting(&prefix) {
			continue
		}

		if prefixFitsAllocation(&prefix, allocation) {
			suitable = append(suitable, prefix)
		}
	}

	if len(suitable) == 0 {
		return "", nil
	}
	// TODO: Currently, random is good enough. In the future, we should consider the load-factor of a prefix as well.
	return suitable[rand.Intn(len(suitable))].Name, nil
}

func prefixCompatibleWithAllocation(prefix *ipamv1alpha1.Prefix, allocation *ipamv1alpha1.PrefixAllocation) bool {
	if prefix.Spec.IPFamily != allocation.Spec.IPFamily {
		return false
	}
	if allocation.Spec.Prefix.IsValid() && *prefix.Spec.Prefix == *allocation.Spec.Prefix {
		return false
	}
	if allocation.Spec.PrefixLength > 0 && int32(prefix.Spec.Prefix.Bits()) >= allocation.Spec.PrefixLength {
		return false
	}
	return true
}

func prefixFitsAllocation(prefix *ipamv1alpha1.Prefix, allocation *ipamv1alpha1.PrefixAllocation) bool {
	if !prefixCompatibleWithAllocation(prefix, allocation) {
		return false
	}

	var bldr netaddr.IPSetBuilder
	bldr.AddPrefix(prefix.Spec.Prefix.IPPrefix)
	for _, used := range prefix.Status.Used {
		bldr.RemovePrefix(used.IPPrefix)
	}
	set, _ := bldr.IPSet()

	switch {
	case allocation.Spec.Prefix.IsValid():
		return set.ContainsPrefix(allocation.Spec.Prefix.IPPrefix)
	case allocation.Spec.PrefixLength > 0:
		_, _, ok := set.RemoveFreePrefix(uint8(allocation.Spec.PrefixLength))
		return ok
	default:
		panic(fmt.Sprintf("unhandled allocation %#v", allocation))
	}
}

func (s *PrefixAllocationScheduler) reconcile(ctx context.Context, log logr.Logger, allocation *ipamv1alpha1.PrefixAllocation) (ctrl.Result, error) {
	if allocation.Spec.PrefixRef != nil {
		log.V(1).Info("Allocation has already been scheduled")
		return ctrl.Result{}, nil
	}
	if allocation.Spec.PrefixSelector == nil {
		log.V(1).Info("Allocation has no selector")
		return ctrl.Result{}, nil
	}
	if allocation.Status.Phase.IsTerminal() {
		log.V(1).Info("Allocation is in terminal state", "Phase", allocation.Status.Phase)
		return ctrl.Result{}, nil
	}

	log.V(1).Info("Determining suitable prefix for allocation")
	ref, err := s.prefixForAllocation(ctx, log, allocation)
	if err != nil {
		return ctrl.Result{}, fmt.Errorf("error finding prefix for allocation: %w", err)
	}
	if ref == "" {
		log.V(1).Info("No suitable prefix found")
		s.Events.Event(allocation, corev1.EventTypeNormal, "NoSuitablePrefix", "No suitable prefix for scheduling allocation found.")
		return ctrl.Result{}, nil
	}

	log.V(1).Info("Suitable prefix found, assigning allocation", "Prefix", ref)
	base := allocation.DeepCopy()
	allocation.Spec.PrefixRef = &corev1.LocalObjectReference{Name: ref}
	if err := s.Patch(ctx, allocation, client.MergeFrom(base)); err != nil {
		return ctrl.Result{}, fmt.Errorf("error patching ref: %w", err)
	}
	return ctrl.Result{}, nil
}

const (
	prefixAllocationSpecIPFamilyField = "spec.ipFamily"
)

func (s *PrefixAllocationScheduler) SetupWithManager(mgr manager.Manager) error {
	ctx := context.Background()
	log := ctrl.Log.WithName("prefixallocationscheduler").WithName("setup")

	if err := mgr.GetFieldIndexer().IndexField(ctx, &ipamv1alpha1.PrefixAllocation{}, prefixAllocationSpecIPFamilyField, func(obj client.Object) []string {
		allocation := obj.(*ipamv1alpha1.PrefixAllocation)
		return []string{string(allocation.Spec.IPFamily)}
	}); err != nil {
		return err
	}

	return ctrl.NewControllerManagedBy(mgr).
		Named("prefixallocationscheduler").
		For(&ipamv1alpha1.PrefixAllocation{}).
		Watches(
			&source.Kind{Type: &ipamv1alpha1.Prefix{}},
			s.enqueueByMatchingPrefix(ctx, log),
		).
		Complete(s)
}

func (s *PrefixAllocationScheduler) enqueueByMatchingPrefix(ctx context.Context, log logr.Logger) handler.EventHandler {
	return handler.EnqueueRequestsFromMapFunc(func(obj client.Object) []ctrl.Request {
		prefix := obj.(*ipamv1alpha1.Prefix)
		if !isPrefixAllocatedAndNotDeleting(prefix) {
			return nil
		}

		list := &ipamv1alpha1.PrefixAllocationList{}
		if err := s.List(ctx, list, client.MatchingFields{PrefixSpecIPFamilyField: string(prefix.Spec.IPFamily)}); err != nil {
			log.Error(err, "Error listing prefix allocations")
			return nil
		}

		var res []ctrl.Request
		for _, prefixAllocation := range list.Items {
			if prefixAllocation.Spec.PrefixRef != nil {
				continue
			}

			sel, err := metav1.LabelSelectorAsSelector(prefixAllocation.Spec.PrefixSelector)
			if err != nil {
				log.Error(err, "Invalid label selector")
				continue
			}

			if sel.Matches(labels.Set(prefix.Labels)) {
				res = append(res, ctrl.Request{NamespacedName: client.ObjectKeyFromObject(&prefixAllocation)})
			}
		}
		return res
	})
}
