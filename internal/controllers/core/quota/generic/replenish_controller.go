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

package generic

import (
	"context"
	"fmt"
	"strings"

	"github.com/go-logr/logr"
	corev1alpha1 "github.com/onmetal/onmetal-api/api/core/v1alpha1"
	"github.com/onmetal/onmetal-api/internal/controllers/core"
	"github.com/onmetal/onmetal-api/utils/quota"
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/builder"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/client/apiutil"
	"sigs.k8s.io/controller-runtime/pkg/event"
	"sigs.k8s.io/controller-runtime/pkg/predicate"
)

type ReplenishReconciler struct {
	client    client.Client
	typ       client.Object
	evaluator quota.Evaluator
}

type ReplenishReconcilerOptions struct {
	Client    client.Client
	Type      client.Object
	Evaluator quota.Evaluator
}

func NewReplenishReconciler(opts ReplenishReconcilerOptions) (*ReplenishReconciler, error) {
	if opts.Client == nil {
		return nil, fmt.Errorf("must specify Client")
	}
	if opts.Type == nil {
		return nil, fmt.Errorf("must specify Type")
	}
	if opts.Evaluator == nil {
		return nil, fmt.Errorf("must specify Evaluator")
	}

	return &ReplenishReconciler{
		client:    opts.Client,
		typ:       opts.Type,
		evaluator: opts.Evaluator,
	}, nil
}

func (r *ReplenishReconciler) anyResourceQuotaMatches(resourceQuotas []corev1alpha1.ResourceQuota) bool {
	for _, resourceQuota := range resourceQuotas {
		for resourceName := range resourceQuota.Spec.Hard {
			if r.evaluator.MatchesResourceName(resourceName) {
				return true
			}
		}
	}
	return false
}

func (r *ReplenishReconciler) Reconcile(ctx context.Context, req ctrl.Request) (ctrl.Result, error) {
	log := ctrl.LoggerFrom(ctx)

	log.V(1).Info("Listing quotas")
	resourceQuotaList := &corev1alpha1.ResourceQuotaList{}
	if err := r.client.List(ctx, resourceQuotaList,
		client.InNamespace(req.Namespace),
	); err != nil {
		return ctrl.Result{}, fmt.Errorf("error listing resource quotas: %w", err)
	}

	if len(resourceQuotaList.Items) == 0 {
		log.V(1).Info("No resource quota present in namespace")
		return ctrl.Result{}, nil
	}

	if !r.anyResourceQuotaMatches(resourceQuotaList.Items) {
		log.V(1).Info("No resource quota matched")
		return ctrl.Result{}, nil
	}

	log.V(1).Info("Adding replenish resource quota annotation")
	namespace := &corev1.Namespace{
		ObjectMeta: metav1.ObjectMeta{
			Name: req.Namespace,
		},
	}
	if _, err := core.PatchAddReplenishResourceQuotaAnnotation(ctx, r.client, namespace); err != nil {
		return ctrl.Result{}, fmt.Errorf("error adding replenish resource quota annotation")
	}

	log.V(1).Info("Added replenish resource quota annotation")
	return ctrl.Result{}, nil
}

func (r *ReplenishReconciler) SetupWithManager(mgr ctrl.Manager) error {
	gvk, err := apiutil.GVKForObject(r.typ, mgr.GetScheme())
	if err != nil {
		return fmt.Errorf("error getting gvk for %T: %w", r.typ, err)
	}

	name := fmt.Sprintf("resourcequota-replenish-%s", strings.ToLower(gvk.GroupKind().String()))
	log := ctrl.Log.WithName(name)
	ctx := ctrl.LoggerInto(context.TODO(), log)

	return ctrl.NewControllerManagedBy(mgr).
		Named(name).
		For(
			r.typ,
			builder.WithPredicates(ResourceDeletedOrUsageChangedPredicate(ctx, log, r.evaluator)),
		).
		Complete(r)
}

func ResourceDeletedOrUsageChangedPredicate(ctx context.Context, log logr.Logger, evaluator quota.Evaluator) predicate.Funcs {
	return predicate.Funcs{
		CreateFunc: func(event event.CreateEvent) bool {
			return false
		},
		DeleteFunc: func(event event.DeleteEvent) bool {
			return true
		},
		UpdateFunc: func(event event.UpdateEvent) bool {
			oldUsage, err := evaluator.Usage(ctx, event.ObjectOld)
			if err != nil {
				log.Error(err, "Error determining old usage")
				return false
			}

			newUsage, err := evaluator.Usage(ctx, event.ObjectNew)
			if err != nil {
				log.Error(err, "Error determining new usage")
				return false
			}

			return !quota.Equals(oldUsage, newUsage)
		},
		GenericFunc: func(event event.GenericEvent) bool {
			return false
		},
	}
}
