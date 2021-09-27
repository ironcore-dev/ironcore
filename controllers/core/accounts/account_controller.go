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

package accounts

import (
	"context"
	"github.com/onmetal/onmetal-api/apis/core"
	"github.com/onmetal/onmetal-api/pkg/utils"
	"k8s.io/apimachinery/pkg/labels"
	"k8s.io/apimachinery/pkg/selection"
	"sigs.k8s.io/controller-runtime/pkg/manager"

	"github.com/go-logr/logr"
	corev1alpha1 "github.com/onmetal/onmetal-api/apis/core/v1alpha1"
	v1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"
	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/client"
)

const (
	finilizerName = core.LabelDomain + "/account"
)

// Reconciler reconciles a Account object
type Reconciler struct {
	client.Client
	Log    logr.Logger
	Scheme *runtime.Scheme
}

//+kubebuilder:rbac:groups=core.onmetal.de,resources=accounts,verbs=get;list;watch;create;update;patch;delete
//+kubebuilder:rbac:groups=core.onmetal.de,resources=accounts/status,verbs=get;update;patch
//+kubebuilder:rbac:groups=core.onmetal.de,resources=accounts/finalizers,verbs=update

func (r *Reconciler) Reconcile(ctx context.Context, req ctrl.Request) (ctrl.Result, error) {
	log := logr.FromContextOrDiscard(ctx)

	var account corev1alpha1.Account
	if err := r.Get(ctx, req.NamespacedName, &account); err != nil {
		return utils.SucceededIfNotFound(err)
	}

	if account.ObjectMeta.DeletionTimestamp.IsZero() {
		// Add account finalizer if it does not exist
		if err := utils.AssureFinalizer(ctx, r.Client, finilizerName, &account); err != nil {
			return utils.Requeue(err)
		}
		namespaces, err := r.listNamespacesForAccount(ctx, req)
		if err != nil {
			return utils.Requeue(err)
		}
		var namespace *v1.Namespace
		for _, n := range namespaces {
			if namespace == nil {
				namespace = &n
			} else {
				if namespace.CreationTimestamp.After(n.CreationTimestamp.Time) {
					if err := utils.AssureFinalizerRemoved(ctx, r.Client, finilizerName, namespace); client.IgnoreNotFound(err) != nil {
						return utils.Requeue(err)
					}
					if err := utils.AssureDeleting(ctx, r.Client, namespace); err != nil {
						return utils.Requeue(err)
					}
					namespace = &n
				}
			}
		}
		if namespace == nil {
			// Create account namespace
			log.Info("creating account namespace", "account", account.Name)
			namespace = &v1.Namespace{
				ObjectMeta: metav1.ObjectMeta{
					Finalizers:   []string{finilizerName},
					GenerateName: "account-",
					Labels: map[string]string{
						core.AccountLabel: account.Name,
						core.ScopeLabel:   account.Name,
					},
				},
			}
			if err := r.Create(ctx, namespace, &client.CreateOptions{}); err != nil {
				log.Error(err, "failed to create account namespace", "account", account.Name)
				return utils.Requeue(err)
			}
			log.Info("created account namespace", "account", account.Name, "namespace", namespace.Name)
		} else {
			if err := utils.AssureFinalizer(ctx, r.Client, finilizerName, namespace); err != nil {
				return utils.Requeue(err)
			}
		}
		// TODO: update on change only
		// Update state with generated namespace name
		account.Status.Namespace = namespace.Name
		account.Status.State = corev1alpha1.AccountStateReady
		if err := r.Status().Update(ctx, &account); err != nil {
			return utils.Requeue(err)
		}
	} else {
		log.Info("deleting account", "account", account.Name)
		// Remove external dependencies
		namespaces, err := r.listNamespacesForAccount(ctx, req)
		if err != nil {
			return utils.Requeue(err)
		}
		var namespace *v1.Namespace
		for _, n := range namespaces {
			if n.Name != account.Status.Namespace {
				if err := utils.AssureFinalizerRemoved(ctx, r.Client, finilizerName, &n); client.IgnoreNotFound(err) != nil {
					return utils.Requeue(err)
				}
				if err := utils.AssureDeleting(ctx, r.Client, &n); err != nil {
					return utils.Requeue(err)
				}
			} else {
				namespace = &n
			}
		}

		if namespace != nil {
			account.Status.State = corev1alpha1.AccountStateTerminating
			if err := r.Status().Update(ctx, &account); err != nil {
				return utils.Requeue(err)
			}
			if err := utils.AssureFinalizerRemoved(ctx, r.Client, finilizerName, namespace); client.IgnoreNotFound(err) != nil {
				return utils.Requeue(err)
			}
			log.Info("deleting account namespace", "account", account.Name, "namespace", namespace.Name)
			if err := utils.AssureDeleting(ctx, r.Client, namespace); err != nil {
				return utils.Requeue(err)
			}
			if err := r.Get(ctx, client.ObjectKey{Name: namespace.Name}, namespace); err == nil || client.IgnoreNotFound(err) != nil {
				return utils.Requeue(err)
			}
		}
		// Remove finalizer
		if err := utils.AssureFinalizerRemoved(ctx, r.Client, finilizerName, &account); err != nil {
			return utils.Requeue(err)
		}
	}
	return utils.Succeeded()
}

func (r *Reconciler) listNamespacesForAccount(ctx context.Context, req ctrl.Request) ([]v1.Namespace, error) {
	var namespaces v1.NamespaceList
	requirementScope, _ := labels.NewRequirement(core.ParentNamespace, selection.DoesNotExist, nil)
	requirementAccount, _ := labels.NewRequirement(core.AccountLabel, selection.DoubleEquals, []string{req.Name})
	err := r.List(ctx, &namespaces, &client.ListOptions{
		LabelSelector: labels.NewSelector().Add(*requirementScope).Add(*requirementAccount),
	})
	if err != nil {
		return nil, err
	}
	return namespaces.Items, nil
}

// SetupWithManager sets up the controller with the Manager.
func (r *Reconciler) SetupWithManager(mgr manager.Manager) error {
	return ctrl.NewControllerManagedBy(mgr).
		For(&corev1alpha1.Account{}).
		Complete(r)
}
