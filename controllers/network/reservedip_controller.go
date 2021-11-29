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

	"github.com/onmetal/onmetal-api/predicates"
	"sigs.k8s.io/controller-runtime/pkg/builder"

	"github.com/go-logr/logr"
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"
	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/client"

	commonv1alpha1 "github.com/onmetal/onmetal-api/apis/common/v1alpha1"
	networkv1alpha1 "github.com/onmetal/onmetal-api/apis/network/v1alpha1"
)

const reservedIPFieldOwner = client.FieldOwner("networking.onmetal.de/reserved-ip")

// ReservedIPReconciler reconciles a ReservedIP object
type ReservedIPReconciler struct {
	client.Client
	Scheme *runtime.Scheme
}

//+kubebuilder:rbac:groups=network.onmetal.de,resources=reservedips,verbs=get;list;watch;create;update;patch;delete
//+kubebuilder:rbac:groups=network.onmetal.de,resources=reservedips/status,verbs=get;update;patch
//+kubebuilder:rbac:groups=network.onmetal.de,resources=reservedips/finalizers,verbs=update
//+kubebuilder:rbac:groups=network.onmetal.de,resources=ipamranges,verbs=get;list;watch;create;update;patch

// Reconcile is part of the main kubernetes reconciliation loop which aims to
// move the current state of the cluster closer to the desired state.
func (r *ReservedIPReconciler) Reconcile(ctx context.Context, req ctrl.Request) (ctrl.Result, error) {
	log := ctrl.LoggerFrom(ctx)
	reservedIP := &networkv1alpha1.ReservedIP{}
	if err := r.Get(ctx, req.NamespacedName, reservedIP); err != nil {
		return ctrl.Result{}, client.IgnoreNotFound(err)
	}
	return r.reconcileExists(ctx, log, reservedIP)
}

func (r *ReservedIPReconciler) reconcileExists(ctx context.Context, log logr.Logger, reservedIP *networkv1alpha1.ReservedIP) (ctrl.Result, error) {
	if !reservedIP.DeletionTimestamp.IsZero() {
		return r.delete(ctx, log, reservedIP)
	}
	return r.reconcile(ctx, log, reservedIP)
}

func (r *ReservedIPReconciler) reconcile(ctx context.Context, log logr.Logger, reservedIP *networkv1alpha1.ReservedIP) (ctrl.Result, error) {
	var request networkv1alpha1.IPAMRangeItem
	if ip := reservedIP.Spec.IP; ip != nil {
		request.IPs = commonv1alpha1.PtrToIPRange(commonv1alpha1.IPRangeFrom(*ip, *ip))
	} else {
		request.IPCount = 1
	}

	ipamRange := &networkv1alpha1.IPAMRange{
		TypeMeta: metav1.TypeMeta{
			APIVersion: networkv1alpha1.GroupVersion.String(),
			Kind:       networkv1alpha1.IPAMRangeGK.Kind,
		},
		ObjectMeta: metav1.ObjectMeta{
			Namespace: reservedIP.Namespace,
			Name:      networkv1alpha1.ReservedIPIPAMName(reservedIP.Name),
		},
		Spec: networkv1alpha1.IPAMRangeSpec{
			Parent: &corev1.LocalObjectReference{
				Name: networkv1alpha1.SubnetIPAMName(reservedIP.Spec.Subnet.Name),
			},
			Items: []networkv1alpha1.IPAMRangeItem{request},
		},
	}
	if err := ctrl.SetControllerReference(reservedIP, ipamRange, r.Scheme); err != nil {
		base := reservedIP.DeepCopy()
		reservedIP.Status.State = networkv1alpha1.ReservedIPStateError
		if err := r.Status().Patch(ctx, reservedIP, client.MergeFrom(base)); err != nil {
			log.Error(err, "Could not update reserved ip status")
		}
		return ctrl.Result{}, fmt.Errorf("could not own ipam range: %w", err)
	}

	if err := r.Patch(ctx, ipamRange, client.Apply, reservedIPFieldOwner); err != nil {
		base := reservedIP.DeepCopy()
		reservedIP.Status.State = networkv1alpha1.ReservedIPStateError
		if err := r.Status().Patch(ctx, reservedIP, client.MergeFrom(base)); err != nil {
			log.Error(err, "Could not update reserved ip status")
		}
		return ctrl.Result{}, fmt.Errorf("could not apply ipam range: %w", err)
	}

	var ip *commonv1alpha1.IPAddr
	for _, allocation := range ipamRange.Status.Allocations {
		if allocation.State != networkv1alpha1.IPAMRangeAllocationFree || allocation.IPs == nil {
			continue
		}

		allocation := allocation
		ip = &allocation.IPs.From
	}

	base := reservedIP.DeepCopy()
	reservedIP.Status.IP = ip
	if ip != nil {
		reservedIP.Status.State = networkv1alpha1.ReservedIPStateReady
	} else {
		reservedIP.Status.State = networkv1alpha1.ReservedIPStatePending
	}
	if err := r.Status().Patch(ctx, reservedIP, client.MergeFrom(base)); err != nil {
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
		Owns(&networkv1alpha1.IPAMRange{}, builder.WithPredicates(predicates.IPAMRangeAllocationsChangedPredicate{})).
		Complete(r)
}
