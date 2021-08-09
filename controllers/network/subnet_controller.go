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

package network

import (
	"context"
	"fmt"
	"github.com/onmetal/onmetal-api/pkg/manager"
	"github.com/onmetal/onmetal-api/pkg/utils"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"
	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/log"

	api "github.com/onmetal/onmetal-api/apis/network/v1alpha1"
)

// SubnetReconciler reconciles a Subnet object
type SubnetReconciler struct {
	client.Client
	Scheme  *runtime.Scheme
	manager *manager.Manager
}

//+kubebuilder:rbac:groups=network.onmetal.de,resources=subnets,verbs=get;list;watch;create;update;patch;delete
//+kubebuilder:rbac:groups=network.onmetal.de,resources=subnets/status,verbs=get;update;patch
//+kubebuilder:rbac:groups=network.onmetal.de,resources=subnets/finalizers,verbs=update

// Reconcile
func (r *SubnetReconciler) Reconcile(ctx context.Context, req ctrl.Request) (ctrl.Result, error) {
	log := log.FromContext(ctx)

	log.Info(fmt.Sprintf("reconcile subnet %s", req))

	// wait until ownercache is built
	r.manager.GetOwnerCache().Wait()

	id := utils.NewObjectIdForRequest(req, api.SubnetGK)
	var subnet api.Subnet
	if err := r.Get(ctx, id.ObjectKey, &subnet); err != nil {
		return utils.SucceededIfNotFound(err)
	}
	if len(r.manager.GetOwnerCache().GetSerfsFor(id)) == 0 {
		ipam := api.IPAMRange{
			TypeMeta: metav1.TypeMeta{
				Kind:       "IPAMRange",
				APIVersion: api.GroupVersion.String(),
			},
			ObjectMeta: metav1.ObjectMeta{
				GenerateName: req.Name + "-",
				Namespace:    req.Namespace,
			},
			Spec: api.IPAMRangeSpec{
				Parent: nil,
				Size:   "",
				CIDR:   "10.0.0.0/24",
			},
			Status: api.IPAMRangeStatus{},
		}
		r.manager.GetOwnerCache().CreateSerf(ctx, &subnet, &ipam)
	}

	return ctrl.Result{}, nil
}

// SetupWithManager sets up the controller with the Manager.
func (r *SubnetReconciler) SetupWithManager(mgr *manager.Manager) error {
	r.manager = mgr
	mgr.GetOwnerCache().RegisterGroupKind(context.Background(), api.IPAMRangeGK)
	c, err := ctrl.NewControllerManagedBy(mgr).
		For(&api.Subnet{}).
		Build(r)
	if err == nil {
		mgr.RegisterControllerFor(api.SubnetGK, c)
	}
	return err

}
