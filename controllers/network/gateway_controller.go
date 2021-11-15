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

	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"
	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/log"

	networkv1alpha1 "github.com/onmetal/onmetal-api/apis/network/v1alpha1"
)

// GatewayReconciler reconciles a Gateway object
type GatewayReconciler struct {
	client.Client
	Scheme *runtime.Scheme
}

//+kubebuilder:rbac:groups=network.onmetal.de,resources=gateways,verbs=get;list;watch;create;update;patch;delete
//+kubebuilder:rbac:groups=network.onmetal.de,resources=gateways/status,verbs=get;update;patch
//+kubebuilder:rbac:groups=network.onmetal.de,resources=gateways/finalizers,verbs=update
//+kubebuilder:rbac:groups=network.onmetal.de,resources=ipamranges,verbs=get;list;watch;create;update;patch

// Reconcile moves the current state of the cluster closer to the desired state.
func (r *GatewayReconciler) Reconcile(ctx context.Context, req ctrl.Request) (ctrl.Result, error) {
	_ = log.FromContext(ctx)

	gw := &networkv1alpha1.Gateway{}
	if err := r.Get(ctx, req.NamespacedName, gw); err != nil {
		return ctrl.Result{}, client.IgnoreNotFound(err)
	}

	if gw.IsBeingDeleted() {
		return ctrl.Result{}, nil
	}

	ipamRange := newIPAMRangeFromGateway(gw)
	if err := ctrl.SetControllerReference(gw, ipamRange, r.Scheme); err != nil {
		return ctrl.Result{}, fmt.Errorf("setting the controller reference of the ipam range: %w", err)
	}

	if err := r.Patch(ctx, ipamRange, client.Apply, gatewayFieldOwner); err != nil {
		return ctrl.Result{}, fmt.Errorf("server-side applying the ipam range: %w", err)
	}

	oldGW := gw.DeepCopy()
	gw.Status.IPs = append(gw.Status.IPs, ipamRange.Status.Allocations[0].IPs.From)
	if err := r.Status().Patch(ctx, gw, client.MergeFrom(oldGW)); err != nil {
		return ctrl.Result{}, fmt.Errorf("updating status: %w", err)
	}
	return ctrl.Result{}, nil
}

// SetupWithManager sets up the controller with the Manager.
func (r *GatewayReconciler) SetupWithManager(mgr ctrl.Manager) error {
	return ctrl.NewControllerManagedBy(mgr).
		For(&networkv1alpha1.Gateway{}).
		Complete(r)
}

const gatewayFieldOwner = client.FieldOwner("compute.onmetal.de/gateway")

func newIPAMRangeFromGateway(gw *networkv1alpha1.Gateway) *networkv1alpha1.IPAMRange {
	ipamRange := &networkv1alpha1.IPAMRange{
		TypeMeta: metav1.TypeMeta{
			APIVersion: networkv1alpha1.GroupVersion.String(),
			Kind:       networkv1alpha1.IPAMRangeGK.Kind,
		},
		ObjectMeta: metav1.ObjectMeta{
			Namespace: gw.Namespace,
			Name:      gw.IPAMRangeName(),
		},
		Spec: networkv1alpha1.IPAMRangeSpec{
			Parent: &corev1.LocalObjectReference{
				Name: networkv1alpha1.SubnetIPAMName(gw.Spec.SourceIPAMRange.Name),
			},
			Requests: []networkv1alpha1.IPAMRangeRequest{{IPCount: 1}},
		},
	}
	return ipamRange
}
