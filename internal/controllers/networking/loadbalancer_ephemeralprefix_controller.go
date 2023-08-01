// Copyright 2023 OnMetal authors
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

package networking

import (
	"context"
	"errors"
	"fmt"

	"github.com/go-logr/logr"
	ipamv1alpha1 "github.com/onmetal/onmetal-api/api/ipam/v1alpha1"
	networkingv1alpha1 "github.com/onmetal/onmetal-api/api/networking/v1alpha1"
	networkingclient "github.com/onmetal/onmetal-api/internal/client/networking"
	klogutils "github.com/onmetal/onmetal-api/utils/klog"
	apierrors "k8s.io/apimachinery/pkg/api/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/klog/v2"
	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/builder"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/handler"
	"sigs.k8s.io/controller-runtime/pkg/predicate"
)

type LoadBalancerEphemeralPrefixReconciler struct {
	client.Client
}

//+kubebuilder:rbac:groups=networking.api.onmetal.de,resources=loadbalancers,verbs=get;list;watch
//+kubebuilder:rbac:groups=ipam.api.onmetal.de,resources=prefixes,verbs=get;list;watch;create;update;patch;delete

func (r *LoadBalancerEphemeralPrefixReconciler) Reconcile(ctx context.Context, req ctrl.Request) (ctrl.Result, error) {
	log := ctrl.LoggerFrom(ctx)
	loadBalancer := &networkingv1alpha1.LoadBalancer{}
	if err := r.Get(ctx, req.NamespacedName, loadBalancer); err != nil {
		return ctrl.Result{}, client.IgnoreNotFound(err)
	}

	return r.reconcileExists(ctx, log, loadBalancer)
}

func (r *LoadBalancerEphemeralPrefixReconciler) reconcileExists(ctx context.Context, log logr.Logger, loadBalancer *networkingv1alpha1.LoadBalancer) (ctrl.Result, error) {
	if !loadBalancer.DeletionTimestamp.IsZero() {
		return ctrl.Result{}, nil
	}

	return r.reconcile(ctx, log, loadBalancer)
}

func (r *LoadBalancerEphemeralPrefixReconciler) ephemeralLoadBalancerPrefixByName(loadBalancer *networkingv1alpha1.LoadBalancer) map[string]*ipamv1alpha1.Prefix {
	res := make(map[string]*ipamv1alpha1.Prefix)

	for i, loadBalancerPrefix := range loadBalancer.Spec.IPs {
		ephemeral := loadBalancerPrefix.Ephemeral
		if ephemeral == nil {
			continue
		}

		prefixName := networkingv1alpha1.LoadBalancerIPIPAMPrefixName(loadBalancer.Name, i)
		prefix := &ipamv1alpha1.Prefix{
			ObjectMeta: metav1.ObjectMeta{
				Namespace:   loadBalancer.Namespace,
				Name:        prefixName,
				Labels:      ephemeral.PrefixTemplate.Labels,
				Annotations: ephemeral.PrefixTemplate.Annotations,
			},
			Spec: ephemeral.PrefixTemplate.Spec,
		}
		_ = ctrl.SetControllerReference(loadBalancer, prefix, r.Scheme())
		res[prefixName] = prefix
	}

	return res
}

func (r *LoadBalancerEphemeralPrefixReconciler) handleExistingPrefix(ctx context.Context, log logr.Logger, loadBalancer *networkingv1alpha1.LoadBalancer, shouldManage bool, prefix *ipamv1alpha1.Prefix) error {
	if metav1.IsControlledBy(prefix, loadBalancer) {
		if shouldManage {
			log.V(1).Info("Ephemeral prefix is present and controlled by load balancer")
			return nil
		}

		if !prefix.DeletionTimestamp.IsZero() {
			log.V(1).Info("Undesired ephemeral prefix is already deleting")
			return nil
		}

		log.V(1).Info("Deleting undesired ephemeral prefix")
		if err := r.Delete(ctx, prefix); client.IgnoreNotFound(err) != nil {
			return fmt.Errorf("error deleting prefix %s: %w", prefix.Name, err)
		}
		return nil
	}

	if shouldManage {
		return fmt.Errorf("prefix %s was not created for load balancer %s (load balancer is not owner)", prefix.Name, loadBalancer.Name)
	}
	// Prefix is not desired but also not controlled by the load balancer.
	return nil
}

func (r *LoadBalancerEphemeralPrefixReconciler) handleCreatePrefix(
	ctx context.Context,
	log logr.Logger,
	loadBalancer *networkingv1alpha1.LoadBalancer,
	prefix *ipamv1alpha1.Prefix,
) error {
	log.V(1).Info("Creating prefix")
	prefixKey := client.ObjectKeyFromObject(prefix)
	err := r.Create(ctx, prefix)
	if err == nil {
		return nil
	}
	if !apierrors.IsAlreadyExists(err) {
		return err
	}

	// Due to a fast resync, we might get an already exists error.
	// In this case, try to fetch the prefix again and, when successful, treat it as managing
	// an existing prefix.
	if err := r.Get(ctx, prefixKey, prefix); err != nil {
		return fmt.Errorf("error getting prefix %s after already exists: %w", prefixKey.Name, err)
	}

	// Treat a retrieved prefix as an existing we should manage.
	log.V(1).Info("Retrieved prefix after already exists conflict")
	return r.handleExistingPrefix(ctx, log, loadBalancer, true, prefix)
}

func (r *LoadBalancerEphemeralPrefixReconciler) reconcile(ctx context.Context, log logr.Logger, loadBalancer *networkingv1alpha1.LoadBalancer) (ctrl.Result, error) {
	log.V(1).Info("Reconcile")

	log.V(1).Info("Listing prefixes")
	prefixList := &ipamv1alpha1.PrefixList{}
	if err := r.List(ctx, prefixList,
		client.InNamespace(loadBalancer.Namespace),
	); err != nil {
		return ctrl.Result{}, fmt.Errorf("error listing prefixes: %w", err)
	}
	log.V(5).Info("Listed prefixes", "Prefixes", klogutils.KObjStructSlice(prefixList.Items))

	var (
		ephemLoadBalancerByName = r.ephemeralLoadBalancerPrefixByName(loadBalancer)
		errs                    []error
	)
	for _, prefix := range prefixList.Items {
		prefixName := prefix.Name
		_, shouldManage := ephemLoadBalancerByName[prefixName]
		delete(ephemLoadBalancerByName, prefixName)
		log := log.WithValues("Prefix", klog.KObj(&prefix), "ShouldManage", shouldManage)
		if err := r.handleExistingPrefix(ctx, log, loadBalancer, shouldManage, &prefix); err != nil {
			errs = append(errs, err)
		}
	}

	for _, prefix := range ephemLoadBalancerByName {
		log := log.WithValues("Prefix", klog.KObj(prefix))
		if err := r.handleCreatePrefix(ctx, log, loadBalancer, prefix); err != nil {
			errs = append(errs, err)
		}
	}

	if err := errors.Join(errs...); err != nil {
		return ctrl.Result{}, fmt.Errorf("error managing ephemeral prefixes: %w", err)
	}

	log.V(1).Info("Reconciled")
	return ctrl.Result{}, nil
}

func (r *LoadBalancerEphemeralPrefixReconciler) loadBalancerNotDeletingPredicate() predicate.Predicate {
	return predicate.NewPredicateFuncs(func(obj client.Object) bool {
		loadBalancer := obj.(*networkingv1alpha1.LoadBalancer)
		return loadBalancer.DeletionTimestamp.IsZero()
	})
}

func (r *LoadBalancerEphemeralPrefixReconciler) enqueueByPrefix() handler.EventHandler {
	return handler.EnqueueRequestsFromMapFunc(func(ctx context.Context, obj client.Object) []ctrl.Request {
		prefix := obj.(*ipamv1alpha1.Prefix)
		log := ctrl.LoggerFrom(ctx)

		loadBalancerList := &networkingv1alpha1.LoadBalancerList{}
		if err := r.List(ctx, loadBalancerList,
			client.InNamespace(prefix.Namespace),
			client.MatchingFields{
				networkingclient.LoadBalancerPrefixNamesField: prefix.Name,
			},
		); err != nil {
			log.Error(err, "Error listing load balancers")
			return nil
		}

		var reqs []ctrl.Request
		for _, loadBalancer := range loadBalancerList.Items {
			if !loadBalancer.DeletionTimestamp.IsZero() {
				continue
			}

			reqs = append(reqs, ctrl.Request{NamespacedName: client.ObjectKeyFromObject(&loadBalancer)})
		}
		return reqs
	})
}

func (r *LoadBalancerEphemeralPrefixReconciler) SetupWithManager(mgr ctrl.Manager) error {
	return ctrl.NewControllerManagedBy(mgr).
		Named("loadbalancerephemeralprefix").
		For(
			&networkingv1alpha1.LoadBalancer{},
			builder.WithPredicates(
				r.loadBalancerNotDeletingPredicate(),
			),
		).
		Owns(&ipamv1alpha1.Prefix{}).
		Watches(
			&ipamv1alpha1.Prefix{},
			r.enqueueByPrefix(),
		).
		Complete(r)
}
