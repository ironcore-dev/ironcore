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

	"github.com/go-logr/logr"
	"inet.af/netaddr"
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"
	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/client"

	commonv1alpha1 "github.com/onmetal/onmetal-api/apis/common/v1alpha1"
	networkv1alpha1 "github.com/onmetal/onmetal-api/apis/network/v1alpha1"
)

const reservedIPFieldOwner = client.FieldOwner("networking.onmetal.de/reservedip")

// ReservedIPReconciler reconciles a ReservedIP object
type ReservedIPReconciler struct {
	client.Client
	Scheme *runtime.Scheme
}

//+kubebuilder:rbac:groups=network.onmetal.de,resources=reservedips,verbs=get;list;watch;create;update;patch;delete
//+kubebuilder:rbac:groups=network.onmetal.de,resources=reservedips/status,verbs=get;update;patch
//+kubebuilder:rbac:groups=network.onmetal.de,resources=reservedips/finalizers,verbs=update

// Reconcile is part of the main kubernetes reconciliation loop which aims to
// move the current state of the cluster closer to the desired state.
// TODO(user): Modify the Reconcile function to compare the state specified by
// the ReservedIP object against the actual cluster state, and then
// perform operations to make the cluster state reflect the state specified by
// the user.
//
// For more details, check Reconcile and its Result here:
// - https://pkg.go.dev/sigs.k8s.io/controller-runtime@v0.8.3/pkg/reconcile
func (r *ReservedIPReconciler) Reconcile(ctx context.Context, req ctrl.Request) (ctrl.Result, error) {
	log := ctrl.LoggerFrom(ctx)
	reservedIP := &networkv1alpha1.ReservedIP{}
	if err := r.Get(ctx, req.NamespacedName, reservedIP); err != nil {
		return ctrl.Result{}, client.IgnoreNotFound(err)
	}
	return r.reconcileExists(ctx, log, reservedIP)
}

func (r *ReservedIPReconciler) reconcileExists(ctx context.Context, log logr.Logger, reservedip *networkv1alpha1.ReservedIP) (ctrl.Result, error) {
	if !reservedip.DeletionTimestamp.IsZero() {
		return r.delete(ctx, log, reservedip)
	}
	return r.reconcile(ctx, log, reservedip)
}

func (r *ReservedIPReconciler) reconcile(ctx context.Context, log logr.Logger, reservedip *networkv1alpha1.ReservedIP) (ctrl.Result, error) {

	reservedipStatus := networkv1alpha1.ReservedIPStatus{
		State: networkv1alpha1.ReservedIPStateInitial,
	}
	var request networkv1alpha1.IPAMRangeRequest
	if reservedip.Spec.IP.String() != "" {
		request.IPs = commonv1alpha1.NewIPRangePtr(netaddr.IPRangeFrom(reservedip.Spec.IP.IP, reservedip.Spec.IP.IP))
	} else {
		request.IPCount = 1
	}

	subIPAMRange := &networkv1alpha1.IPAMRange{
		TypeMeta: metav1.TypeMeta{
			APIVersion: networkv1alpha1.GroupVersion.String(),
			Kind:       networkv1alpha1.IPAMRangeGK.Kind,
		},
		ObjectMeta: metav1.ObjectMeta{
			Namespace: reservedip.Namespace,
			Name:      fmt.Sprintf("reservedip-subnet-%s-%s", reservedip.Name, reservedip.Spec.Subnet.Name),
		},
		Spec: networkv1alpha1.IPAMRangeSpec{
			Parent: &corev1.LocalObjectReference{
				Name: networkv1alpha1.SubnetIPAMName(reservedip.Spec.Subnet.Name),
			},
			Requests: []networkv1alpha1.IPAMRangeRequest{request},
		},
	}
	if err := ctrl.SetControllerReference(reservedip, subIPAMRange, r.Scheme); err != nil {
		reservedipStatus.State = networkv1alpha1.ReservedIPStateError
		return ctrl.Result{}, fmt.Errorf("could not own subnet %s ipam range: %w", reservedip.Spec.Subnet.Name, err)
	}
	if err := r.Patch(ctx, subIPAMRange, client.Apply, reservedIPFieldOwner); err != nil {
		reservedipStatus.State = networkv1alpha1.ReservedIPStateError
		return ctrl.Result{}, fmt.Errorf("could not create subnet %s ipam range: %w", reservedip.Spec.Subnet.Name, err)
	}
	for _, allocation := range subIPAMRange.Status.Allocations {
		if allocation.State != networkv1alpha1.IPAMRangeAllocationFree || allocation.IPs == nil {
			continue
		}
		ip := allocation.IPs.From
		reservedipStatus = networkv1alpha1.ReservedIPStatus{
			State: networkv1alpha1.ReservedIPStateReady,
			IP:    ip,
		}
		break
	}
	oldReservedIP := reservedip.DeepCopy()
	reservedip.Status = reservedipStatus
	if err := r.Status().Patch(ctx, reservedip, client.MergeFrom(oldReservedIP)); err != nil {
		return ctrl.Result{}, fmt.Errorf("could not update status: %w", err)
	}
	return ctrl.Result{}, nil
}

func (r *ReservedIPReconciler) delete(ctx context.Context, log logr.Logger, reservedip *networkv1alpha1.ReservedIP) (ctrl.Result, error) {
	return ctrl.Result{}, nil
}

// SetupWithManager sets up the controller with the Manager.
func (r *ReservedIPReconciler) SetupWithManager(mgr ctrl.Manager) error {
	return ctrl.NewControllerManagedBy(mgr).
		For(&networkv1alpha1.ReservedIP{}).
		Complete(r)
}
