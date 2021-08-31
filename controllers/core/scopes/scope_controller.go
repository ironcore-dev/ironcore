/*
 * Copyright (c) 2021 by the OnMetal authors.
 *
 * Licensed under the Apache License, Version 2.0 (the "License");
 * you may not use this file except in compliance with the License.
 * You may obtain a copy of the License at
 *
 *     http://www.apache.org/licenses/LICENSE-2.0
 *
 * Unless required by applicable law or agreed to in writing, software
 * distributed under the License is distributed on an "AS IS" BASIS,
 * WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
 * See the License for the specific language governing permissions and
 * limitations under the License.
 */

package scopes

import (
	"context"
	"fmt"
	"github.com/onmetal/onmetal-api/apis/core"
	"github.com/onmetal/onmetal-api/pkg/logging"
	"github.com/onmetal/onmetal-api/pkg/utils"
	"k8s.io/apimachinery/pkg/labels"
	"k8s.io/apimachinery/pkg/selection"

	v1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"
	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/log"

	"github.com/go-logr/logr"
	corev1alpha1 "github.com/onmetal/onmetal-api/apis/core/v1alpha1"
)

const (
	scopeFinilizerName = core.LabelDomain + "/scope"
)

// Reconciler reconciles a Scope object
type Reconciler struct {
	client.Client
	Log    logr.Logger
	Scheme *runtime.Scheme
}

//+kubebuilder:rbac:groups=core.onmetal.de,resources=scopes,verbs=get;list;watch;create;update;patch;delete
//+kubebuilder:rbac:groups=core.onmetal.de,resources=scopes/status,verbs=get;update;patch
//+kubebuilder:rbac:groups=core.onmetal.de,resources=scopes/finalizers,verbs=update

func (r *Reconciler) Reconcile(ctx context.Context, req ctrl.Request) (ctrl.Result, error) {
	_ = log.FromContext(ctx)

	log := logging.NewLogger(r.Log, "scope", req.NamespacedName)

	var scope corev1alpha1.Scope
	if err := r.Get(ctx, req.NamespacedName, &scope); err != nil {
		return utils.SucceededIfNotFound(err)
	}

	if scope.ObjectMeta.DeletionTimestamp.IsZero() {
		var parentNamespace v1.Namespace
		if err := r.Client.Get(ctx, client.ObjectKey{Name: req.Namespace}, &parentNamespace); err != nil {
			return utils.Requeue(err)
		}
		accountName := utils.GetLabel(&parentNamespace, core.AccountLabel)
		scopeName := utils.GetLabel(&parentNamespace, core.ScopeLabel, accountName)

		// Add scope finalizer if it does not exist
		if err := utils.AssureFinalizer(ctx, log, r.Client, scopeFinilizerName, &scope); err != nil {
			return utils.Requeue(err)
		}
		namespaces, err := r.listNamespacesForScopes(ctx, req)
		if err != nil {
			return utils.Requeue(err)
		}
		var namespace *v1.Namespace
		for _, n := range namespaces {
			if namespace == nil {
				namespace = &n
			} else {
				if namespace.CreationTimestamp.After(n.CreationTimestamp.Time) {
					if err := utils.AssureFinalizerRemoved(ctx, log, r.Client, scopeFinilizerName, namespace); client.IgnoreNotFound(err) != nil {
						return utils.Requeue(err)
					}
					if err := utils.AssureDeleting(ctx, log, r.Client, namespace); err != nil {
						return utils.Requeue(err)
					}
					namespace = &n
				}
			}
		}
		if namespace == nil {
			// Create scope namespace
			log.Infof("creating namespace for scope %s", scope.Name)
			namespace = &v1.Namespace{
				ObjectMeta: metav1.ObjectMeta{
					Finalizers:   []string{scopeFinilizerName},
					GenerateName: fmt.Sprintf("scope-%s-", req.Name),
					Labels: map[string]string{
						core.ScopeLabel:      scope.Name,
						core.AccountLabel:    accountName,
						core.ParentNamespace: req.Namespace,
						core.ParentScope:     scopeName,
					},
				},
			}
			if err := r.Create(ctx, namespace, &client.CreateOptions{}); err != nil {
				log.Errorf(err, "failed to create namespace for scope %s", scope.Name)
				return utils.Requeue(err)
			}
			log.Infof("using namespace %s for scope %s", namespace.Name, scope.Name)
		} else {
			if err := utils.AssureFinalizer(ctx, log, r.Client, scopeFinilizerName, namespace); err != nil {
				return utils.Requeue(err)
			}
		}
		// TODO: update on change only
		// Update state with generated namespace name
		scope.Status.Namespace = namespace.Name
		scope.Status.State = corev1alpha1.ScopeStateReady
		scope.Status.Account = accountName
		scope.Status.ParentScope = scopeName
		scope.Status.ParentNamespace = req.Namespace
		if err := r.Status().Update(ctx, &scope); err != nil {
			return utils.Requeue(err)
		}
	} else {
		log.Infof("deleting scope %s", scope.Name)
		// Remove external dependencies
		namespaces, err := r.listNamespacesForScopes(ctx, req)
		if err != nil {
			return utils.Requeue(err)
		}
		var namespace *v1.Namespace
		for _, n := range namespaces {
			if n.Name != scope.Status.Namespace {
				if err := utils.AssureFinalizerRemoved(ctx, log, r.Client, scopeFinilizerName, &n); client.IgnoreNotFound(err) != nil {
					return utils.Requeue(err)
				}
				if err := utils.AssureDeleting(ctx, log, r.Client, &n); err != nil {
					return utils.Requeue(err)
				}
			} else {
				namespace = &n
			}
		}

		if namespace != nil {
			scope.Status.State = corev1alpha1.ScopeStateTerminating
			if err := r.Status().Update(ctx, &scope); err != nil {
				return utils.Requeue(err)
			}
			if err := utils.AssureFinalizerRemoved(ctx, log, r.Client, scopeFinilizerName, namespace); client.IgnoreNotFound(err) != nil {
				return utils.Requeue(err)
			}
			if err := utils.AssureDeleting(ctx, log, r.Client, namespace); err != nil {
				return utils.Requeue(err)
			}
			if err := r.Get(ctx, client.ObjectKey{Name: namespace.Name}, namespace); err == nil || client.IgnoreNotFound(err) != nil {
				return utils.Requeue(err)
			}
		}
		// Remove finalizer
		if err := utils.AssureFinalizerRemoved(ctx, log, r.Client, scopeFinilizerName, &scope); err != nil {
			return utils.Requeue(err)
		}
	}
	return utils.Succeeded()
}

func (r *Reconciler) listNamespacesForScopes(ctx context.Context, req ctrl.Request) ([]v1.Namespace, error) {
	var namespaces v1.NamespaceList
	requirementScope, _ := labels.NewRequirement(core.ScopeLabel, selection.DoubleEquals, []string{req.Name})
	requirementParentNamespace, _ := labels.NewRequirement(core.ParentNamespace, selection.DoubleEquals, []string{req.Namespace})
	err := r.List(ctx, &namespaces, &client.ListOptions{
		LabelSelector: labels.NewSelector().Add(*requirementScope).Add(*requirementParentNamespace),
	})
	if err != nil {
		return nil, err
	}
	return namespaces.Items, nil
}

// SetupWithManager sets up the controller with the Manager.
func (r *Reconciler) SetupWithManager(mgr ctrl.Manager) error {
	return ctrl.NewControllerManagedBy(mgr).
		For(&corev1alpha1.Scope{}).
		Complete(r)
}
