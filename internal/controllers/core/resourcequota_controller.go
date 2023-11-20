// SPDX-FileCopyrightText: 2023 SAP SE or an SAP affiliate company and IronCore contributors
// SPDX-License-Identifier: Apache-2.0

package core

import (
	"context"
	"fmt"

	"github.com/go-logr/logr"
	"github.com/ironcore-dev/controller-utils/metautils"
	corev1alpha1 "github.com/ironcore-dev/ironcore/api/core/v1alpha1"
	ironcoreclient "github.com/ironcore-dev/ironcore/utils/client"
	"github.com/ironcore-dev/ironcore/utils/quota"
	corev1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/api/meta"
	"k8s.io/apimachinery/pkg/api/resource"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apimachinery/pkg/util/sets"
	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/builder"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/handler"
	"sigs.k8s.io/controller-runtime/pkg/predicate"
)

type ResourceQuotaReconciler struct {
	client.Client
	APIReader client.Reader
	Scheme    *runtime.Scheme
	Registry  quota.Registry
}

//+kubebuilder:rbac:groups=core.ironcore.dev,resources=resourcequotas,verbs=get;list;watch;create;update;patch;delete
//+kubebuilder:rbac:groups=core.ironcore.dev,resources=resourcequotas/status,verbs=get;update;patch
//+kubebuilder:rbac:groups=core.ironcore.dev,resources=resourcequotas/finalizers,verbs=update

//+kubebuilder:rbac:groups="",resources=namespaces,verbs=get;list;watch;update;patch

//+kubebuilder:rbac:groups=compute.ironcore.dev,resources=machines,verbs=get;list;watch
//+kubebuilder:rbac:groups=compute.ironcore.dev,resources=machineclasses,verbs=get;list;watch
//+kubebuilder:rbac:groups=storage.ironcore.dev,resources=volumes,verbs=get;list;watch
//+kubebuilder:rbac:groups=storage.ironcore.dev,resources=volumeclasses,verbs=get;list;watch

func (r *ResourceQuotaReconciler) Reconcile(ctx context.Context, req ctrl.Request) (ctrl.Result, error) {
	log := ctrl.LoggerFrom(ctx)

	resourceQuota := &corev1alpha1.ResourceQuota{}
	if err := r.Get(ctx, req.NamespacedName, resourceQuota); err != nil {
		return ctrl.Result{}, client.IgnoreNotFound(err)
	}

	return r.reconcileExists(ctx, log, resourceQuota)
}

func (r *ResourceQuotaReconciler) reconcileExists(ctx context.Context, log logr.Logger, resourceQuota *corev1alpha1.ResourceQuota) (ctrl.Result, error) {
	if !resourceQuota.DeletionTimestamp.IsZero() {
		return r.delete(ctx, log, resourceQuota)
	}
	return r.reconcile(ctx, log, resourceQuota)
}

func (r *ResourceQuotaReconciler) delete(ctx context.Context, log logr.Logger, resourceQuota *corev1alpha1.ResourceQuota) (ctrl.Result, error) {
	return ctrl.Result{}, nil
}

func (r *ResourceQuotaReconciler) reconcile(ctx context.Context, log logr.Logger, resourceQuota *corev1alpha1.ResourceQuota) (ctrl.Result, error) {
	log.V(1).Info("Reconcile")

	log.V(1).Info("Gathering matching evaluators")
	matchingEvaluators, coveredResourceNames := r.getMatchingEvaluators(resourceQuota)

	log.V(1).Info("Calculating resource usage")
	used, err := r.calculateUsage(ctx, log, resourceQuota, matchingEvaluators, coveredResourceNames)
	if err != nil {
		return ctrl.Result{}, fmt.Errorf("error calculating usage: %w", err)
	}

	log.V(1).Info("Updating resource quota status", "Used", used)
	if err := r.updateStatus(ctx, resourceQuota, resourceQuota.Spec.Hard, used); err != nil {
		return ctrl.Result{}, fmt.Errorf("error updating resource quota status: %w", err)
	}

	log.V(1).Info("Reconciled")
	return ctrl.Result{}, nil
}

func (r *ResourceQuotaReconciler) getMatchingEvaluators(resourceQuota *corev1alpha1.ResourceQuota) ([]quota.Evaluator, sets.Set[corev1alpha1.ResourceName]) {
	var (
		evaluators           []quota.Evaluator
		hardResourceNames    = quota.ResourceNames(resourceQuota.Spec.Hard)
		coveredResourceNames = sets.New[corev1alpha1.ResourceName]()
	)
	for _, evaluator := range r.Registry.List() {
		var matches bool
		for resourceName := range hardResourceNames {
			if evaluator.MatchesResourceName(resourceName) {
				matches = true
				coveredResourceNames.Insert(resourceName)
			}
		}

		if matches {
			evaluators = append(evaluators, evaluator)
		}
	}
	return evaluators, coveredResourceNames
}

func (r *ResourceQuotaReconciler) calculateUsage(
	ctx context.Context,
	log logr.Logger,
	resourceQuota *corev1alpha1.ResourceQuota,
	evaluators []quota.Evaluator,
	coveredResourceNames sets.Set[corev1alpha1.ResourceName],
) (corev1alpha1.ResourceList, error) {
	usage := make(corev1alpha1.ResourceList, len(coveredResourceNames))
	zero := resource.MustParse("0")
	for resourceName := range coveredResourceNames {
		usage[resourceName] = zero
	}

	for _, evaluator := range evaluators {
		gvk, list, err := metautils.NewListForObject(r.Scheme, evaluator.Type())
		if err != nil {
			return nil, fmt.Errorf("error creating list for type %T: %w", evaluator.Type, err)
		}

		// We use the APIReader to list here since using the cached reader might cause objects created slightly before
		// a new resource quota to be missed.
		// TODO: Re-evaluate how we can avoid APIReader here.
		log := log.WithValues("GVK", gvk)
		log.V(1).Info("Listing resources")
		if err := r.APIReader.List(ctx, list, client.InNamespace(resourceQuota.Namespace)); err != nil {
			return nil, fmt.Errorf("[%s] error listing objects: %w", gvk, err)
		}

		log.V(1).Info("Listed resources", "NoOfResources", meta.LenList(list))

		log.V(1).Info("Calculating usage for type")
		typeUsage := corev1alpha1.ResourceList{}
		if err := metautils.EachListItem(list, func(obj client.Object) error {
			matches, err := quota.EvaluatorMatchesResourceScopeSelector(evaluator, obj, resourceQuota.Spec.ScopeSelector)
			if err != nil {
				return fmt.Errorf("error matching resource scope selector: %w", err)
			}
			if !matches {
				return nil
			}

			itemUsage, err := evaluator.Usage(ctx, obj)
			if err != nil {
				return fmt.Errorf("error computing usage: %w", err)
			}

			typeUsage = quota.Add(typeUsage, itemUsage)
			return nil
		}); err != nil {
			return nil, fmt.Errorf("[%s] error iterating list: %w", gvk, err)
		}

		log.V(1).Info("Calculated type usage", "TypeUsage", typeUsage)
		usage = quota.Add(usage, typeUsage)
	}

	usage = quota.Mask(usage, coveredResourceNames)
	return usage, nil
}

func (r *ResourceQuotaReconciler) updateStatus(
	ctx context.Context,
	resourceQuota *corev1alpha1.ResourceQuota,
	hard, used corev1alpha1.ResourceList,
) error {
	base := resourceQuota.DeepCopy()
	resourceQuota.Status.Hard = hard
	resourceQuota.Status.Used = used
	if err := r.Status().Patch(ctx, resourceQuota, client.MergeFrom(base)); err != nil {
		return fmt.Errorf("error patching resource quota status: %w", err)
	}
	return nil
}

func (r *ResourceQuotaReconciler) enqueueResourceQuotasByNamespace() handler.EventHandler {
	return handler.EnqueueRequestsFromMapFunc(func(ctx context.Context, obj client.Object) []ctrl.Request {
		namespace := obj.(*corev1.Namespace)
		log := ctrl.LoggerFrom(ctx, "Namespace", namespace.Name)

		resourceQuotaList := &corev1alpha1.ResourceQuotaList{}
		if err := r.List(ctx, resourceQuotaList,
			client.InNamespace(namespace.Name),
		); err != nil {
			log.Error(err, "Error listing resource quotas in namespace")
			return nil
		}

		// Clean up replenishment annotation to mark it as 'received'
		go func() {
			log.V(1).Info("Removing replenish resource quota annotation from namespace")
			if _, err := PatchRemoveReplenishResourceQuotaAnnotation(ctx, r.Client, namespace); err != nil {
				log.Error(err, "Error removing replenish resource quota annotation from namespace")
			}
		}()

		return ironcoreclient.ReconcileRequestsFromObjectStructSlice(resourceQuotaList.Items)
	})
}

var resourceQuotaDirtyPredicate = predicate.NewPredicateFuncs(func(obj client.Object) bool {
	resourceQuota := obj.(*corev1alpha1.ResourceQuota)

	// If we did not calculate any usage yet (i.e. hard being unset), we need to recalc.
	if resourceQuota.Status.Hard == nil {
		return true
	}

	// If the spec enforced quota does not match with the status enforced quota, we also need to recalc.
	if !quota.Equals(resourceQuota.Spec.Hard, resourceQuota.Status.Hard) {
		return true
	}

	// If we're deleting, we have to enqueue.
	if !resourceQuota.DeletionTimestamp.IsZero() {
		return true
	}

	// Ignore any other case. We're triggered by watching namespaces when resource changes occur.
	return false
})

func (r *ResourceQuotaReconciler) SetupWithManager(mgr ctrl.Manager) error {
	return ctrl.NewControllerManagedBy(mgr).
		For(
			&corev1alpha1.ResourceQuota{},
			builder.WithPredicates(resourceQuotaDirtyPredicate),
		).
		Watches(
			&corev1.Namespace{},
			r.enqueueResourceQuotasByNamespace(),
			builder.WithPredicates(HasReplenishResourceQuotaPredicate),
		).
		Complete(r)
}
