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
	commonv1alpha1 "github.com/onmetal/onmetal-api/apis/common/v1alpha1"
	networkv1alpha1 "github.com/onmetal/onmetal-api/apis/network/v1alpha1"
	"github.com/onmetal/onmetal-api/equality"
	"github.com/onmetal/onmetal-api/predicates"
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"
	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/builder"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/manager"
)

const subnetFieldOwner = client.FieldOwner("networking.onmetal.de/subnet")

// SubnetReconciler reconciles a Subnet object
type SubnetReconciler struct {
	client.Client
	Scheme *runtime.Scheme
}

//+kubebuilder:rbac:groups=network.onmetal.de,resources=subnets,verbs=get;list;watch;create;update;patch;delete
//+kubebuilder:rbac:groups=network.onmetal.de,resources=subnets/status,verbs=get;update;patch
//+kubebuilder:rbac:groups=network.onmetal.de,resources=subnets/finalizers,verbs=update
//+kubebuilder:rbac:groups=network.onmetal.de,resources=ipamranges,verbs=get;list;watch;create;update;patch

// Reconcile reconciles the spec with the real world.
func (r *SubnetReconciler) Reconcile(ctx context.Context, req ctrl.Request) (ctrl.Result, error) {
	log := ctrl.LoggerFrom(ctx)
	subnet := &networkv1alpha1.Subnet{}
	if err := r.Get(ctx, req.NamespacedName, subnet); err != nil {
		return ctrl.Result{}, client.IgnoreNotFound(err)
	}

	return r.reconcileExists(ctx, log, subnet)
}

func (r *SubnetReconciler) reconcileExists(ctx context.Context, log logr.Logger, subnet *networkv1alpha1.Subnet) (ctrl.Result, error) {
	if !subnet.DeletionTimestamp.IsZero() {
		return ctrl.Result{}, nil
	}

	return r.reconcile(ctx, log, subnet)
}

func (r *SubnetReconciler) reconcile(ctx context.Context, log logr.Logger, subnet *networkv1alpha1.Subnet) (ctrl.Result, error) {
	var (
		ipamRangeParent *corev1.LocalObjectReference
		requests        []networkv1alpha1.IPAMRangeRequest
		rootCIDRs       []commonv1alpha1.CIDR
	)
	if parent := subnet.Spec.Parent; parent != nil {
		ipamRangeParent = &corev1.LocalObjectReference{Name: networkv1alpha1.SubnetIPAMName(parent.Name)}
		for _, rng := range subnet.Spec.Ranges {
			cidr := rng.CIDR
			requests = append(requests, networkv1alpha1.IPAMRangeRequest{
				Size: rng.Size,
				CIDR: cidr,
			})
		}
	} else {
		for _, rng := range subnet.Spec.Ranges {
			if rng.CIDR != nil {
				rootCIDRs = append(rootCIDRs, *rng.CIDR)
			}
		}
	}

	ipamRange := &networkv1alpha1.IPAMRange{
		TypeMeta: metav1.TypeMeta{
			APIVersion: networkv1alpha1.GroupVersion.String(),
			Kind:       networkv1alpha1.IPAMRangeGK.Kind,
		},
		ObjectMeta: metav1.ObjectMeta{
			Namespace: subnet.Namespace,
			Name:      networkv1alpha1.SubnetIPAMName(subnet.Name),
		},
		Spec: networkv1alpha1.IPAMRangeSpec{
			Parent:   ipamRangeParent,
			CIDRs:    rootCIDRs,
			Requests: requests,
		},
	}
	if err := ctrl.SetControllerReference(subnet, ipamRange, r.Scheme); err != nil {
		return ctrl.Result{}, fmt.Errorf("could not own ipam range: %w", err)
	}

	if err := r.Patch(ctx, ipamRange, client.Apply, subnetFieldOwner); err != nil {
		return ctrl.Result{}, fmt.Errorf("could not apply ipam range: %w", err)
	}

	var cidrs []commonv1alpha1.CIDR
	for _, allocation := range ipamRange.Status.Allocations {
		if allocation.State == networkv1alpha1.IPAMRangeAllocationFailed || allocation.CIDR == nil {
			continue
		}

		for _, request := range requests {
			if equality.Semantic.DeepEqual(allocation.Request, request) {
				cidrs = append(cidrs, *request.CIDR)
			}
		}
	}

	updated := subnet.DeepCopy()
	updated.Status.CIDRs = cidrs
	updated.Status.State = networkv1alpha1.SubnetStateUp
	if err := r.Status().Patch(ctx, updated, client.MergeFrom(subnet)); err != nil {
		return ctrl.Result{}, fmt.Errorf("could not update subnet status")
	}

	return ctrl.Result{}, nil
}

// SetupWithManager sets up the controller with the Manager.
func (r *SubnetReconciler) SetupWithManager(mgr manager.Manager) error {
	return ctrl.NewControllerManagedBy(mgr).
		For(&networkv1alpha1.Subnet{}).
		Owns(&networkv1alpha1.IPAMRange{}, builder.WithPredicates(
			predicates.IPAMRangeAllocationsChangedPredicate{},
		)).
		Complete(r)
}
