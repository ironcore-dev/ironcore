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

package compute

import (
	"context"
	"fmt"

	"github.com/go-logr/logr"
	"k8s.io/apimachinery/pkg/runtime"
	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/client"

	computev1alpha1 "github.com/onmetal/onmetal-api/apis/compute/v1alpha1"
)

// ComputePoolReconciler reconciles a ComputePool object
type ComputePoolReconciler struct {
	client.Client
	Log    logr.Logger
	Scheme *runtime.Scheme
}

//+kubebuilder:rbac:groups=compute.onmetal.de,resources=computepools,verbs=get;list;watch;create;update;patch;delete
//+kubebuilder:rbac:groups=compute.onmetal.de,resources=computepools/status,verbs=get;update;patch
//+kubebuilder:rbac:groups=compute.onmetal.de,resources=computepools/finalizers,verbs=update

// Reconcile is part of the main kubernetes reconciliation loop which aims to
// move the current state of the cluster closer to the desired state.
// TODO(user): Modify the Reconcile function to compare the state specified by
// the ComputePool object against the actual cluster state, and then
// perform operations to make the cluster state reflect the state specified by
// the user.
//
// For more details, check Reconcile and its Result here:
// - https://pkg.go.dev/sigs.k8s.io/controller-runtime@v0.7.2/pkg/reconcile
func (r *ComputePoolReconciler) Reconcile(ctx context.Context, req ctrl.Request) (ctrl.Result, error) {
	_ = r.Log.WithValues("computepool", req.NamespacedName)

	r.Log.Info(fmt.Sprintf("reconcile computepool %s/%s", req.Namespace, req.Name))

	return ctrl.Result{}, nil
}

// SetupWithManager sets up the controller with the Manager.
func (r *ComputePoolReconciler) SetupWithManager(mgr ctrl.Manager) error {
	return ctrl.NewControllerManagedBy(mgr).
		For(&computev1alpha1.ComputePool{}).
		Complete(r)
}
