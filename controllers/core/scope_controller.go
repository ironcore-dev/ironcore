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

	v1 "k8s.io/api/core/v1"
	apierrors "k8s.io/apimachinery/pkg/api/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"
	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/controller/controllerutil"
	"sigs.k8s.io/controller-runtime/pkg/log"

	"github.com/go-logr/logr"
	corev1alpha1 "github.com/onmetal/onmetal-api/apis/core/v1alpha1"
)

const (
	scopeFinilizerName = "scope.core.onmetal.de/finalizer"
)

// ScopeReconciler reconciles a Scope object
type ScopeReconciler struct {
	client.Client
	Log    logr.Logger
	Scheme *runtime.Scheme
}

//+kubebuilder:rbac:groups=core.onmetal.de,resources=scopes,verbs=get;list;watch;create;update;patch;delete
//+kubebuilder:rbac:groups=core.onmetal.de,resources=scopes/status,verbs=get;update;patch
//+kubebuilder:rbac:groups=core.onmetal.de,resources=scopes/finalizers,verbs=update

func (r *ScopeReconciler) Reconcile(ctx context.Context, req ctrl.Request) (ctrl.Result, error) {
	_ = log.FromContext(ctx)

	log := r.Log.WithValues("scope", req.NamespacedName)

	var scope corev1alpha1.Scope
	if err := r.Get(ctx, req.NamespacedName, &scope); err != nil {
		return ctrl.Result{}, client.IgnoreNotFound(err)
	}

	if scope.ObjectMeta.DeletionTimestamp.IsZero() {
		// Add scope finalizer if it does not exist
		if !containsString(scope.GetFinalizers(), scopeFinilizerName) {
			controllerutil.AddFinalizer(&scope, scopeFinilizerName)
			if err := r.Update(ctx, &scope); err != nil {
				return ctrl.Result{}, err
			}
		}
		if scope.Status.Namespace == "" {
			// Create scope namespace
			namespace := v1.Namespace{
				ObjectMeta: metav1.ObjectMeta{
					GenerateName: "scope-",
				},
			}
			if err := r.Create(ctx, &namespace, &client.CreateOptions{}); err != nil {
				log.Error(err, "failed to create namespace", "namespace", namespace)
				return ctrl.Result{}, err
			}
			// Update state with generated namespace name
			scope.Status.Namespace = namespace.Name
			if err := r.Status().Update(ctx, &scope); err != nil {
				return ctrl.Result{}, err
			}
			// TODO: Add state if namespace is ready
			log.V(0).Info("created namespace for scope", "namespace", namespace, "scope", scope)
		}
	} else {
		log.V(0).Info("deleting scope", "scope", scope)
		if containsString(scope.GetFinalizers(), scopeFinilizerName) {
			// Remove external dependencies
			if scope.Status.Namespace != "" {
				namespace := v1.Namespace{
					ObjectMeta: metav1.ObjectMeta{
						Name: scope.Status.Namespace,
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
			controllerutil.RemoveFinalizer(&scope, scopeFinilizerName)
			if err := r.Update(ctx, &scope); err != nil {
				return ctrl.Result{}, err
			}
		}
		return ctrl.Result{}, nil
	}

	return ctrl.Result{}, nil
}

// SetupWithManager sets up the controller with the Manager.
func (r *ScopeReconciler) SetupWithManager(mgr ctrl.Manager) error {
	return ctrl.NewControllerManagedBy(mgr).
		For(&corev1alpha1.Scope{}).
		Complete(r)
}
