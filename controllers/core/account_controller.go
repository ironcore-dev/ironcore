/*
Copyright 2021.

Licensed under the Apache License, Version 2.0 (the "License");
you may not use this file except in compliance with the License.
You may obtain a copy of the License at

    http://www.apache.org/licenses/LICENSE-2.0

Unless required by applicable law or agreed to in writing, software
distributed under the License is distributed on an "AS IS" BASIS,
WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
See the License for the specific language governing permissions and
limitations under the License.
*/

package core

import (
	"context"
	"github.com/go-logr/logr"
	v1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"
	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/controller/controllerutil"
	"sigs.k8s.io/controller-runtime/pkg/log"

	corev1alpha1 "github.com/onmetal/onmetal-api/apis/core/v1alpha1"
	apierrors "k8s.io/apimachinery/pkg/api/errors"
)

const (
	accountFinilizerName = "account.core.onmetal.de/finalizer"
)

// AccountReconciler reconciles a Account object
type AccountReconciler struct {
	client.Client
	Log    logr.Logger
	Scheme *runtime.Scheme
}

//+kubebuilder:rbac:groups=core.onmetal.de,resources=accounts,verbs=get;list;watch;create;update;patch;delete
//+kubebuilder:rbac:groups=core.onmetal.de,resources=accounts/status,verbs=get;update;patch
//+kubebuilder:rbac:groups=core.onmetal.de,resources=accounts/finalizers,verbs=update

func (r *AccountReconciler) Reconcile(ctx context.Context, req ctrl.Request) (ctrl.Result, error) {
	_ = log.FromContext(ctx)
	log := r.Log.WithValues("account", req.NamespacedName)

	var account corev1alpha1.Account
	if err := r.Get(ctx, req.NamespacedName, &account); err != nil {
		return ctrl.Result{}, client.IgnoreNotFound(err)
	}

	if account.ObjectMeta.DeletionTimestamp.IsZero() {
		// Add account finalizer if it does not exist
		if !containsString(account.GetFinalizers(), accountFinilizerName) {
			controllerutil.AddFinalizer(&account, accountFinilizerName)
			if err := r.Update(ctx, &account); err != nil {
				return ctrl.Result{}, err
			}
		}
		if account.Status.Namespace == "" {
			// Create account namespace
			namespace := v1.Namespace{
				ObjectMeta: metav1.ObjectMeta{
					GenerateName: "account-",
				},
			}
			if err := r.Create(ctx, &namespace, &client.CreateOptions{}); err != nil {
				log.Error(err, "failed to create namespace", "namespace", namespace)
				return ctrl.Result{}, err
			}
			// Update state with generated namespace name
			account.Status.Namespace = namespace.Name
			if err := r.Status().Update(ctx, &account); err != nil {
				return ctrl.Result{}, err
			}
			log.V(0).Info("created namespace for account", "namespace", namespace, "account", account)
		}
	} else {
		log.V(0).Info("deleting acccount", "account", account)
		if containsString(account.GetFinalizers(), accountFinilizerName) {
			// Remove external dependencies
			if account.Status.Namespace != "" {
				namespace := v1.Namespace{
					ObjectMeta: metav1.ObjectMeta{
						Name: account.Status.Namespace,
					},
				}
				if err := r.Delete(ctx, &namespace, client.PropagationPolicy(metav1.DeletePropagationBackground)); client.IgnoreNotFound(err) != nil {
					log.Error(err, "unable to delete namespace", "namespace", namespace)
					if !apierrors.IsNotFound(err) {
						return ctrl.Result{}, err
					}
				} else {
					log.V(0).Info("deleted namespace", "namespace", namespace)
				}
			}
			// Remove finalizer
			controllerutil.RemoveFinalizer(&account, accountFinilizerName)
			if err := r.Update(ctx, &account); err != nil {
				return ctrl.Result{}, err
			}
		}
		return ctrl.Result{}, nil
	}
	return ctrl.Result{}, nil
}

// Helper functions to check and remove string from a slice of strings.
func containsString(slice []string, s string) bool {
	for _, item := range slice {
		if item == s {
			return true
		}
	}
	return false
}

// SetupWithManager sets up the controller with the Manager.
func (r *AccountReconciler) SetupWithManager(mgr ctrl.Manager) error {
	return ctrl.NewControllerManagedBy(mgr).
		For(&corev1alpha1.Account{}).
		Complete(r)
}
