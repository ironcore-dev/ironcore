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

	"github.com/go-logr/logr"
	"github.com/onmetal/controller-utils/conditionutils"
	ipamv1alpha1 "github.com/onmetal/onmetal-api/apis/ipam/v1alpha1"
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

func IsClusterPrefixAllocationInTerminalState(allocation *ipamv1alpha1.ClusterPrefixAllocation) bool {
	readyCond := &ipamv1alpha1.ClusterPrefixAllocationCondition{}
	conditionutils.MustFindSlice(allocation.Status.Conditions, string(ipamv1alpha1.ClusterPrefixAllocationReady), readyCond)
	switch {
	case readyCond.Status == corev1.ConditionFalse && readyCond.Reason == ipamv1alpha1.ClusterPrefixAllocationReadyReasonFailed:
		return true
	case readyCond.Status == corev1.ConditionTrue:
		return true
	default:
		return false
	}
}

type ClusterPrefixAllocationSchedulerReconciler struct {
	client.Client
	Scheme *runtime.Scheme
}

//+kubebuilder:rbac:groups=ipam.api.onmetal.de,resources=clusterprefixes,verbs=get;list;watch
//+kubebuilder:rbac:groups=ipam.api.onmetal.de,resources=clusterprefixallocations,verbs=get;list;watch;update;patch
//+kubebuilder:rbac:groups=ipam.api.onmetal.de,resources=clusterprefixallocations/status,verbs=get;update;patch

// Reconcile is part of the main kubernetes reconciliation loop which aims to
// move the current state of the cluster closer to the desired state.
func (r *ClusterPrefixAllocationSchedulerReconciler) Reconcile(ctx context.Context, req ctrl.Request) (ctrl.Result, error) {
	log := ctrl.LoggerFrom(ctx)
	allocation := &ipamv1alpha1.ClusterPrefixAllocation{}
	if err := r.Get(ctx, req.NamespacedName, allocation); err != nil {
		return ctrl.Result{}, client.IgnoreNotFound(err)
	}

	return r.reconcileExists(ctx, log, allocation)
}

func (r *ClusterPrefixAllocationSchedulerReconciler) reconcileExists(ctx context.Context, log logr.Logger, allocation *ipamv1alpha1.ClusterPrefixAllocation) (ctrl.Result, error) {
	if !allocation.DeletionTimestamp.IsZero() {
		return ctrl.Result{}, nil
	}
	return r.reconcile(ctx, log, allocation)
}

func (r *ClusterPrefixAllocationSchedulerReconciler) patchReadiness(ctx context.Context, allocation *ipamv1alpha1.ClusterPrefixAllocation, readyCond ipamv1alpha1.ClusterPrefixAllocationCondition) error {
	base := allocation.DeepCopy()
	conditionutils.MustUpdateSlice(&allocation.Status.Conditions, string(ipamv1alpha1.ClusterPrefixAllocationReady),
		conditionutils.UpdateStatus(readyCond.Status),
		conditionutils.UpdateReason(readyCond.Reason),
		conditionutils.UpdateMessage(readyCond.Message),
	)
	if err := r.Status().Patch(ctx, allocation, client.MergeFrom(base)); err != nil {
		return fmt.Errorf("error patching ready condition: %w", err)
	}
	return nil
}

func (r *ClusterPrefixAllocationSchedulerReconciler) prefixRefForAllocation(ctx context.Context, allocation *ipamv1alpha1.ClusterPrefixAllocation) (string, error) {
	sel, err := metav1.LabelSelectorAsSelector(allocation.Spec.PrefixSelector)
	if err != nil {
		return "", fmt.Errorf("error building label selector: %w", err)
	}

	list := &ipamv1alpha1.ClusterPrefixList{}
	if err := r.List(ctx, list, client.MatchingLabelsSelector{Selector: sel}); err != nil {
		return "", fmt.Errorf("error listing cluster prefixes: %w", err)
	}

	for _, clusterPrefix := range list.Items {
		if !clusterPrefix.DeletionTimestamp.IsZero() || !IsClusterPrefixReady(&clusterPrefix) {
			continue
		}

		if CanClusterPrefixFitRequest(&clusterPrefix, ipamv1alpha1.PrefixAllocationRequest{
			Prefix:       allocation.Spec.Prefix,
			PrefixLength: allocation.Spec.PrefixLength,
		}) {
			return clusterPrefix.Name, nil
		}
	}
	return "", nil
}

func (r *ClusterPrefixAllocationSchedulerReconciler) reconcile(ctx context.Context, log logr.Logger, allocation *ipamv1alpha1.ClusterPrefixAllocation) (ctrl.Result, error) {
	if allocation.Spec.PrefixRef != nil {
		log.V(1).Info("Allocation has already been scheduled")
		return ctrl.Result{}, nil
	}
	if allocation.Spec.PrefixSelector == nil {
		log.V(1).Info("Allocation has no selector")
		return ctrl.Result{}, nil
	}
	if IsClusterPrefixAllocationInTerminalState(allocation) {
		log.V(1).Info("Allocation is in terminal state")
		return ctrl.Result{}, nil
	}

	log.V(1).Info("Determining suitable prefix ref for allocation")
	ref, err := r.prefixRefForAllocation(ctx, allocation)
	if err != nil {
		return ctrl.Result{}, fmt.Errorf("error computing prefix ref for allocation: %w", err)
	}
	if ref == "" {
		log.V(1).Info("No suitable prefix ref found")
		if err := r.patchReadiness(ctx, allocation, ipamv1alpha1.ClusterPrefixAllocationCondition{
			Status:  corev1.ConditionFalse,
			Reason:  ipamv1alpha1.ClusterPrefixAllocationReadyReasonPending,
			Message: "There is currently no prefix that satisfies the requirements.",
		}); err != nil {
			return ctrl.Result{}, fmt.Errorf("error patching readiness: %w", err)
		}
		return ctrl.Result{}, nil
	}

	log.V(1).Info("Suitable prefix ref found, assigning allocation", "PrefixRef", ref)
	base := allocation.DeepCopy()
	allocation.Spec.PrefixRef = &corev1.LocalObjectReference{Name: ref}
	if err := r.Patch(ctx, allocation, client.MergeFrom(base)); err != nil {
		return ctrl.Result{}, fmt.Errorf("error patching ref: %w", err)
	}
	return ctrl.Result{}, nil
}

const (
	clusterPrefixAllocationSpecIfUnscheduledPrefixField = ".spec[?(@.prefixRef.name == '' && @.prefixSelector)]"
)

func (r *ClusterPrefixAllocationSchedulerReconciler) SetupWithManager(mgr manager.Manager) error {
	ctx := context.Background()
	log := ctrl.Log.WithName("clusterprefixallocationscheduler").WithName("setup")

	if err := mgr.GetFieldIndexer().IndexField(ctx, &ipamv1alpha1.ClusterPrefixAllocation{}, clusterPrefixAllocationSpecIfUnscheduledPrefixField, func(obj client.Object) []string {
		allocation := obj.(*ipamv1alpha1.ClusterPrefixAllocation)
		if allocation.Spec.PrefixRef != nil || allocation.Spec.PrefixSelector == nil {
			return nil
		}
		return []string{""}
	}); err != nil {
		return fmt.Errorf("error indexing field %q: %w", clusterPrefixAllocationSpecIfUnscheduledPrefixField, err)
	}

	return ctrl.NewControllerManagedBy(mgr).
		Named("clusterprefixallocationscheduler").
		For(&ipamv1alpha1.ClusterPrefixAllocation{}).
		Watches(
			&source.Kind{Type: &ipamv1alpha1.ClusterPrefix{}},
			r.enqueueByMatchingClusterPrefix(ctx, log),
		).
		Complete(r)
}

func (r *ClusterPrefixAllocationSchedulerReconciler) enqueueByMatchingClusterPrefix(ctx context.Context, log logr.Logger) handler.EventHandler {
	return handler.EnqueueRequestsFromMapFunc(func(obj client.Object) []ctrl.Request {
		clusterPrefix := obj.(*ipamv1alpha1.ClusterPrefix)
		if !IsClusterPrefixReady(clusterPrefix) {
			return nil
		}

		list := &ipamv1alpha1.ClusterPrefixAllocationList{}
		if err := r.List(ctx, list, client.MatchingFields{clusterPrefixAllocationSpecIfUnscheduledPrefixField: ""}); err != nil {
			log.Error(err, "Error listing cluster prefix allocations")
			return nil
		}

		var res []ctrl.Request
		for _, clusterPrefixAllocation := range list.Items {
			sel, err := metav1.LabelSelectorAsSelector(clusterPrefixAllocation.Spec.PrefixSelector)
			if err != nil {
				log.Error(err, "Invalid label selector")
				continue
			}

			if sel.Matches(labels.Set(clusterPrefix.Labels)) {
				res = append(res, ctrl.Request{NamespacedName: client.ObjectKeyFromObject(&clusterPrefixAllocation)})
			}
		}
		return res
	})
}
