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

package ipam

import (
	"context"
	"fmt"

	"github.com/go-logr/logr"
	"github.com/onmetal/controller-utils/conditionutils"
	commonv1alpha1 "github.com/onmetal/onmetal-api/apis/common/v1alpha1"
	ipamv1alpha1 "github.com/onmetal/onmetal-api/apis/ipam/v1alpha1"
	corev1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/runtime"
	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/client"
)

func IsIPReady(ip *ipamv1alpha1.IP) bool {
	return conditionutils.MustFindSliceStatus(ip.Status.Conditions, string(ipamv1alpha1.IPReady)) == corev1.ConditionTrue
}

// IPReconciler reconciles a IP object
type IPReconciler struct {
	client.Client
	Scheme *runtime.Scheme
}

//+kubebuilder:rbac:groups=ipam.api.onmetal.de,resources=ips,verbs=get;list;watch;create;update;patch;delete
//+kubebuilder:rbac:groups=ipam.api.onmetal.de,resources=ips/status,verbs=get;update;patch
//+kubebuilder:rbac:groups=ipam.api.onmetal.de,resources=ips/finalizers,verbs=update
//+kubebuilder:rbac:groups=ipam.api.onmetal.de,resources=prefixallocations,verbs=get;list;watch;create;update;patch;delete
//+kubebuilder:rbac:groups=ipam.api.onmetal.de,resources=prefixallocations/status,verbs=get;update;patch

// Reconcile is part of the main kubernetes reconciliation loop which aims to
// move the current state of the cluster closer to the desired state.
func (r *IPReconciler) Reconcile(ctx context.Context, req ctrl.Request) (ctrl.Result, error) {
	log := ctrl.LoggerFrom(ctx)
	ip := &ipamv1alpha1.IP{}
	if err := r.Get(ctx, req.NamespacedName, ip); err != nil {
		return ctrl.Result{}, client.IgnoreNotFound(err)
	}

	return r.reconcileExists(ctx, log, ip)
}

func (r *IPReconciler) reconcileExists(ctx context.Context, log logr.Logger, ip *ipamv1alpha1.IP) (ctrl.Result, error) {
	if !ip.DeletionTimestamp.IsZero() {
		return r.delete(ctx, log, ip)
	}
	return r.reconcile(ctx, log, ip)
}

func (r *IPReconciler) delete(ctx context.Context, log logr.Logger, ip *ipamv1alpha1.IP) (ctrl.Result, error) {
	return ctrl.Result{}, nil
}

func (r *IPReconciler) patchReadiness(ctx context.Context, ip *ipamv1alpha1.IP, readyCond ipamv1alpha1.IPCondition) error {
	base := ip.DeepCopy()
	conditionutils.MustUpdateSlice(&ip.Status.Conditions, string(ipamv1alpha1.IPReady),
		conditionutils.UpdateStatus(readyCond.Status),
		conditionutils.UpdateReason(readyCond.Reason),
		conditionutils.UpdateMessage(readyCond.Message),
	)
	if err := r.Status().Patch(ctx, ip, client.MergeFrom(base)); err != nil {
		return fmt.Errorf("error patching ready condition: %w", err)
	}
	return nil
}

func (r *IPReconciler) patchAssignment(ctx context.Context, ip *ipamv1alpha1.IP, res PrefixAllocationer) error {
	base := ip.DeepCopy()
	ip.Spec.IP = res.Result().Range.To
	prefixRef := res.PrefixRef()
	ip.Spec.PrefixRef = prefixRef
	if err := r.Patch(ctx, ip, client.MergeFrom(base)); err != nil {
		return fmt.Errorf("error patching ip assignment: %w", err)
	}
	return nil
}

func (r *IPReconciler) reconcile(ctx context.Context, log logr.Logger, ip *ipamv1alpha1.IP) (ctrl.Result, error) {
	switch {
	case IsIPReady(ip):
		log.V(1).Info("IP is ready")
		return ctrl.Result{}, nil
	default:
		log.V(1).Info("IP readiness needs to be computed")
		m := NewPrefixAllocator(r.Client, r.Scheme)
		res, err := m.Apply(ctx, (*IP)(ip))
		switch {
		case err != nil:
			if err := r.patchReadiness(ctx, ip, ipamv1alpha1.IPCondition{
				Status:  corev1.ConditionUnknown,
				Reason:  "ErrorAllocating",
				Message: fmt.Sprintf("Allocating resulted in an error: %v.", err),
			}); err != nil {
				log.Error(err, "Error patching readiness")
			}
			return ctrl.Result{}, fmt.Errorf("error applying allocation: %w", err)
		case res.ReadyState() == PrefixAllocationSucceeded && !ip.Spec.IP.IsValid():
			log.V(1).Info("Patching IP assignment")
			if err := r.patchAssignment(ctx, ip, res); err != nil {
				return ctrl.Result{}, fmt.Errorf("error patching ip assignment: %w", err)
			}
			return ctrl.Result{}, nil
		case res.ReadyState() == PrefixAllocationSucceeded:
			log.V(1).Info("Marking ip as ready")
			if err := r.patchReadiness(ctx, ip, ipamv1alpha1.IPCondition{
				Status:  corev1.ConditionTrue,
				Reason:  ipamv1alpha1.IPReadyReasonAllocated,
				Message: "IP has been successfully allocated.",
			}); err != nil {
				return ctrl.Result{}, fmt.Errorf("error marking ip as allocated: %w", err)
			}
			return ctrl.Result{}, nil
		default:
			log.V(1).Info("IP is pending")
			if err := r.patchReadiness(ctx, ip, ipamv1alpha1.IPCondition{
				Status:  corev1.ConditionFalse,
				Reason:  ipamv1alpha1.IPReadyReasonPending,
				Message: "IP is not yet allocated.",
			}); err != nil {
				return ctrl.Result{}, fmt.Errorf("error marking ip as pending: %w", err)
			}
			return ctrl.Result{}, nil
		}
	}
}

// SetupWithManager sets up the controller with the Manager.
func (r *IPReconciler) SetupWithManager(mgr ctrl.Manager) error {
	return ctrl.NewControllerManagedBy(mgr).
		For(&ipamv1alpha1.IP{}).
		Owns(&ipamv1alpha1.PrefixAllocation{}).
		Complete(r)
}

const (
	PrefixAllocationIPLabel = "ipam.api.onmetal.de/ip"
)

type IP ipamv1alpha1.IP

func (i *IP) Object() client.Object {
	return (*ipamv1alpha1.IP)(i)
}

func (i *IP) Label() string {
	return PrefixAllocationIPLabel
}

func (i *IP) Request() ipamv1alpha1.PrefixAllocationRequest {
	if i.Spec.IP.IsValid() {
		return ipamv1alpha1.PrefixAllocationRequest{
			Range: commonv1alpha1.PtrToIPRange(commonv1alpha1.IPRange{From: i.Spec.IP, To: i.Spec.IP}),
		}
	}
	return ipamv1alpha1.PrefixAllocationRequest{
		RangeLength: 1,
	}
}

func (i *IP) PrefixRef() *ipamv1alpha1.PrefixReference {
	return i.Spec.PrefixRef
}

func (i *IP) PrefixSelector() *ipamv1alpha1.PrefixSelector {
	return i.Spec.PrefixSelector
}
