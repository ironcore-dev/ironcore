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

package network

import (
	"context"
	"fmt"
	"github.com/go-logr/logr"
	"github.com/onmetal/controller-utils/conditionutils"
	networkv1alpha1 "github.com/onmetal/onmetal-api/apis/network/v1alpha1"
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/labels"
	"k8s.io/apimachinery/pkg/runtime"
	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/handler"
	"sigs.k8s.io/controller-runtime/pkg/manager"
	"sigs.k8s.io/controller-runtime/pkg/source"
)

type PrefixAllocationSchedulerReconciler struct {
	client.Client
	Scheme *runtime.Scheme
}

//+kubebuilder:rbac:groups=network.onmetal.de,resources=prefixes,verbs=get;list;watch
//+kubebuilder:rbac:groups=network.onmetal.de,resources=clusterprefixes,verbs=get;list;watch
//+kubebuilder:rbac:groups=network.onmetal.de,resources=prefixallocations,verbs=get;list;watch;update;patch
//+kubebuilder:rbac:groups=network.onmetal.de,resources=prefixallocations/status,verbs=get;update;patch

// Reconcile is part of the main kubernetes reconciliation loop which aims to
// move the current state of the cluster closer to the desired state.
func (r *PrefixAllocationSchedulerReconciler) Reconcile(ctx context.Context, req ctrl.Request) (ctrl.Result, error) {
	log := ctrl.LoggerFrom(ctx)
	allocation := &networkv1alpha1.PrefixAllocation{}
	if err := r.Get(ctx, req.NamespacedName, allocation); err != nil {
		return ctrl.Result{}, client.IgnoreNotFound(err)
	}

	return r.reconcileExists(ctx, log, allocation)
}

func (r *PrefixAllocationSchedulerReconciler) reconcileExists(ctx context.Context, log logr.Logger, allocation *networkv1alpha1.PrefixAllocation) (ctrl.Result, error) {
	if !allocation.DeletionTimestamp.IsZero() {
		return ctrl.Result{}, nil
	}
	return r.reconcile(ctx, log, allocation)
}

func (r *PrefixAllocationSchedulerReconciler) patchReadiness(ctx context.Context, allocation *networkv1alpha1.PrefixAllocation, readyCond networkv1alpha1.PrefixAllocationCondition) error {
	base := allocation.DeepCopy()
	conditionutils.MustUpdateSlice(&allocation.Status.Conditions, string(networkv1alpha1.PrefixAllocationReady),
		conditionutils.UpdateStatus(readyCond.Status),
		conditionutils.UpdateReason(readyCond.Reason),
		conditionutils.UpdateMessage(readyCond.Message),
	)
	if err := r.Status().Patch(ctx, allocation, client.MergeFrom(base)); err != nil {
		return fmt.Errorf("error patching ready condition: %w", err)
	}
	return nil
}

func (r *PrefixAllocationSchedulerReconciler) prefixRefForAllocation(ctx context.Context, log logr.Logger, allocation *networkv1alpha1.PrefixAllocation) (*networkv1alpha1.PrefixReference, error) {
	sel, err := metav1.LabelSelectorAsSelector(&allocation.Spec.PrefixSelector.LabelSelector)
	if err != nil {
		return nil, fmt.Errorf("error building label selector: %w", err)
	}

	switch kind := allocation.Spec.PrefixSelector.Kind; kind {
	case networkv1alpha1.ClusterPrefixKind:
		name, err := r.clusterPrefixRef(ctx, allocation, sel)
		if err != nil || name == "" {
			return nil, err
		}
		return &networkv1alpha1.PrefixReference{
			Kind: networkv1alpha1.ClusterPrefixKind,
			Name: name,
		}, nil
	case networkv1alpha1.PrefixKind:
		name, err := r.prefixRef(ctx, allocation, sel)
		if err != nil || name == "" {
			return nil, err
		}
		return &networkv1alpha1.PrefixReference{
			Kind: networkv1alpha1.PrefixKind,
			Name: name,
		}, nil
	default:
		return nil, fmt.Errorf("invalid prefix selector kind %q", kind)
	}
}

func (r *PrefixAllocationSchedulerReconciler) clusterPrefixRef(ctx context.Context, allocation *networkv1alpha1.PrefixAllocation, sel labels.Selector) (string, error) {
	list := &networkv1alpha1.ClusterPrefixList{}
	if err := r.List(ctx, list, client.MatchingLabelsSelector{Selector: sel}); err != nil {
		return "", fmt.Errorf("error listing cluster prefixes: %w", err)
	}

	for _, clusterPrefix := range list.Items {
		if !clusterPrefix.DeletionTimestamp.IsZero() || !IsClusterPrefixReady(&clusterPrefix) {
			continue
		}

		if CanClusterPrefixFitRequest(&clusterPrefix, allocation.Spec.PrefixAllocationRequest) {
			return clusterPrefix.Name, nil
		}
	}
	return "", nil
}

func (r *PrefixAllocationSchedulerReconciler) prefixRef(ctx context.Context, allocation *networkv1alpha1.PrefixAllocation, sel labels.Selector) (string, error) {
	list := &networkv1alpha1.PrefixList{}
	if err := r.List(ctx, list, client.InNamespace(allocation.Namespace), client.MatchingLabelsSelector{Selector: sel}); err != nil {
		return "", fmt.Errorf("error listing prefixes: %w", err)
	}

	for _, prefix := range list.Items {
		if !prefix.DeletionTimestamp.IsZero() || !IsPrefixReady(&prefix) {
			continue
		}

		if CanPrefixFitRequest(&prefix, allocation.Spec.PrefixAllocationRequest) {
			return prefix.Name, nil
		}
	}
	return "", nil
}

func IsPrefixAllocationInTerminalState(allocation *networkv1alpha1.PrefixAllocation) bool {
	readyCond := &networkv1alpha1.PrefixAllocationCondition{}
	conditionutils.MustFindSlice(allocation.Status.Conditions, string(networkv1alpha1.PrefixAllocationReady), readyCond)
	switch {
	case readyCond.Status == corev1.ConditionFalse && readyCond.Reason == networkv1alpha1.PrefixAllocationReadyReasonFailed:
		return true
	case readyCond.Status == corev1.ConditionTrue:
		return true
	default:
		return false
	}
}

func (r *PrefixAllocationSchedulerReconciler) reconcile(ctx context.Context, log logr.Logger, allocation *networkv1alpha1.PrefixAllocation) (ctrl.Result, error) {
	if allocation.Spec.PrefixRef != nil {
		log.V(1).Info("Allocation has already been scheduled")
		return ctrl.Result{}, nil
	}
	if allocation.Spec.PrefixSelector == nil {
		log.V(1).Info("Allocation has no selector")
		return ctrl.Result{}, nil
	}
	if IsPrefixAllocationInTerminalState(allocation) {
		log.V(1).Info("Allocation is in terminal state")
		return ctrl.Result{}, nil
	}

	log.V(1).Info("Determining suitable prefix ref for allocation")
	ref, err := r.prefixRefForAllocation(ctx, log, allocation)
	if err != nil {
		return ctrl.Result{}, fmt.Errorf("error computing prefix ref for allocation: %w", err)
	}
	if ref == nil {
		log.V(1).Info("No suitable prefix ref found")
		if err := r.patchReadiness(ctx, allocation, networkv1alpha1.PrefixAllocationCondition{
			Status:  corev1.ConditionFalse,
			Reason:  networkv1alpha1.PrefixAllocationReadyReasonPending,
			Message: "There is currently no prefix that satisfies the requirements.",
		}); err != nil {
			return ctrl.Result{}, fmt.Errorf("error patching readiness: %w", err)
		}
		return ctrl.Result{}, nil
	}

	log.V(1).Info("Suitable prefix ref found, assigning allocation", "PrefixRef", *ref)
	base := allocation.DeepCopy()
	allocation.Spec.PrefixRef = ref
	if err := r.Patch(ctx, allocation, client.MergeFrom(base)); err != nil {
		return ctrl.Result{}, fmt.Errorf("error patching ref: %w", err)
	}
	return ctrl.Result{}, nil
}

const (
	prefixAllocationSpecIfUnscheduledOnClusterPrefixField = ".spec[?(@.prefixRef.name == '' && @.prefixSelector.kind == 'ClusterPrefix')]"
	prefixAllocationSpecIfUnscheduledOnPrefixField        = ".spec[?(@.prefixRef.name == '' && @.prefixSelector.kind == 'Prefix')]"
)

func (r *PrefixAllocationSchedulerReconciler) SetupWithManager(mgr manager.Manager) error {
	ctx := context.Background()
	log := ctrl.Log.WithName("prefixallocationscheduler").WithName("setup")

	if err := mgr.GetFieldIndexer().IndexField(ctx, &networkv1alpha1.PrefixAllocation{}, prefixAllocationSpecIfUnscheduledOnClusterPrefixField, func(obj client.Object) []string {
		allocation := obj.(*networkv1alpha1.PrefixAllocation)
		if allocation.Spec.PrefixRef != nil {
			return nil
		}
		if prefixSelector := allocation.Spec.PrefixSelector; prefixSelector == nil || prefixSelector.Kind != networkv1alpha1.ClusterPrefixKind {
			return nil
		}
		return []string{""}
	}); err != nil {
		return fmt.Errorf("error indexing field %q: %w", prefixAllocationSpecIfUnscheduledOnClusterPrefixField, err)
	}

	if err := mgr.GetFieldIndexer().IndexField(ctx, &networkv1alpha1.PrefixAllocation{}, prefixAllocationSpecIfUnscheduledOnPrefixField, func(obj client.Object) []string {
		allocation := obj.(*networkv1alpha1.PrefixAllocation)
		if allocation.Spec.PrefixRef != nil {
			return nil
		}
		if prefixSelector := allocation.Spec.PrefixSelector; prefixSelector == nil || prefixSelector.Kind != networkv1alpha1.PrefixKind {
			return nil
		}
		return []string{""}
	}); err != nil {
		return fmt.Errorf("error indexing field %q: %w", prefixAllocationSpecIfUnscheduledOnPrefixField, err)
	}

	return ctrl.NewControllerManagedBy(mgr).
		Named("prefixallocationscheduler").
		For(&networkv1alpha1.PrefixAllocation{}).
		Watches(
			&source.Kind{Type: &networkv1alpha1.Prefix{}},
			r.enqueueByMatchingPrefix(ctx, log),
		).
		Watches(
			&source.Kind{Type: &networkv1alpha1.ClusterPrefix{}},
			r.enqueueByMatchingClusterPrefix(ctx, log),
		).
		Complete(r)
}

func (r *PrefixAllocationSchedulerReconciler) enqueueByMatchingClusterPrefix(ctx context.Context, log logr.Logger) handler.EventHandler {
	return handler.EnqueueRequestsFromMapFunc(func(obj client.Object) []ctrl.Request {
		clusterPrefix := obj.(*networkv1alpha1.ClusterPrefix)
		if !IsClusterPrefixReady(clusterPrefix) {
			return nil
		}

		list := &networkv1alpha1.PrefixAllocationList{}
		if err := r.List(ctx, list, client.MatchingFields{prefixAllocationSpecIfUnscheduledOnClusterPrefixField: ""}); err != nil {
			log.Error(err, "Error listing cluster prefix allocations")
			return nil
		}

		var res []ctrl.Request
		for _, prefixAllocation := range list.Items {
			sel, err := metav1.LabelSelectorAsSelector(&prefixAllocation.Spec.PrefixSelector.LabelSelector)
			if err != nil {
				log.Error(err, "Invalid label selector")
				continue
			}

			if sel.Matches(labels.Set(clusterPrefix.Labels)) {
				res = append(res, ctrl.Request{NamespacedName: client.ObjectKeyFromObject(&prefixAllocation)})
			}
		}
		return res
	})
}

func (r *PrefixAllocationSchedulerReconciler) enqueueByMatchingPrefix(ctx context.Context, log logr.Logger) handler.EventHandler {
	return handler.EnqueueRequestsFromMapFunc(func(obj client.Object) []ctrl.Request {
		prefix := obj.(*networkv1alpha1.Prefix)
		if !IsPrefixReady(prefix) {
			return nil
		}

		list := &networkv1alpha1.PrefixAllocationList{}
		if err := r.List(ctx, list, client.MatchingFields{prefixAllocationSpecIfUnscheduledOnPrefixField: ""}); err != nil {
			log.Error(err, "Error listing prefix allocations")
			return nil
		}

		var res []ctrl.Request
		for _, prefixAllocation := range list.Items {
			sel, err := metav1.LabelSelectorAsSelector(&prefixAllocation.Spec.PrefixSelector.LabelSelector)
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
