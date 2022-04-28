// Copyright 2022 OnMetal authors
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
//      http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.

package networking

import (
	"context"
	"fmt"

	"github.com/go-logr/logr"
	commonv1alpha1 "github.com/onmetal/onmetal-api/apis/common/v1alpha1"
	computev1alpha1 "github.com/onmetal/onmetal-api/apis/compute/v1alpha1"
	ipamv1alpha1 "github.com/onmetal/onmetal-api/apis/ipam/v1alpha1"
	networkingv1alpha1 "github.com/onmetal/onmetal-api/apis/networking/v1alpha1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"
	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/handler"
	"sigs.k8s.io/controller-runtime/pkg/source"
)

var nicFieldOwner = client.FieldOwner(networkingv1alpha1.Resource("networkinterfaces").String())

type NetworkInterfaceReconciler struct {
	client.Client
	Scheme *runtime.Scheme
}

func (r *NetworkInterfaceReconciler) Reconcile(ctx context.Context, req ctrl.Request) (ctrl.Result, error) {
	log := ctrl.LoggerFrom(ctx)
	nic := &networkingv1alpha1.NetworkInterface{}
	if err := r.Get(ctx, req.NamespacedName, nic); err != nil {
		return ctrl.Result{}, client.IgnoreNotFound(err)
	}

	return r.reconcileExists(ctx, log, nic)
}

func (r *NetworkInterfaceReconciler) reconcileExists(ctx context.Context, log logr.Logger, nic *networkingv1alpha1.NetworkInterface) (ctrl.Result, error) {
	if !nic.DeletionTimestamp.IsZero() {
		return r.delete(ctx, log, nic)
	}
	return r.reconcile(ctx, log, nic)
}

func (r *NetworkInterfaceReconciler) delete(ctx context.Context, log logr.Logger, nic *networkingv1alpha1.NetworkInterface) (ctrl.Result, error) {
	return ctrl.Result{}, nil
}

func (r *NetworkInterfaceReconciler) reconcile(ctx context.Context, log logr.Logger, nic *networkingv1alpha1.NetworkInterface) (ctrl.Result, error) {
	log.V(1).Info("Applying IPs")
	ips, err := r.applyIPs(ctx, nic)
	if err != nil {
		return ctrl.Result{}, err
	}
	log.V(1).Info("Successfully applied IPs", "IPs", ips)

	log.V(1).Info("Applying network interface binding")
	if err := r.applyNetworkInterfaceBinding(ctx, nic, ips); err != nil {
		return ctrl.Result{}, err
	}
	log.V(1).Info("Successfully applied network interface binding")

	log.V(1).Info("Patching status")
	if err := r.patchStatus(ctx, nic, ips); err != nil {
		return ctrl.Result{}, err
	}

	log.V(1).Info("Successfully patched status")
	return ctrl.Result{}, nil
}

func (r *NetworkInterfaceReconciler) applyIPs(ctx context.Context, nic *networkingv1alpha1.NetworkInterface) ([]commonv1alpha1.IP, error) {
	var ips []commonv1alpha1.IP
	for i, ipSource := range nic.Spec.IPs {
		switch {
		case ipSource.Value != nil:
			ips = append(ips, *ipSource.Value)
		case ipSource.EphemeralPrefix != nil:
			template := ipSource.EphemeralPrefix.PrefixTemplateSpec

			prefixMeta := template.ObjectMeta
			prefixMeta.Namespace = nic.Namespace
			prefixMeta.Name = fmt.Sprintf("%s-%d", nic.Name, i)

			prefixSpec := template.Spec

			prefix := &ipamv1alpha1.Prefix{
				TypeMeta: metav1.TypeMeta{
					APIVersion: ipamv1alpha1.SchemeGroupVersion.String(),
					Kind:       ipamv1alpha1.PrefixKind,
				},
				ObjectMeta: prefixMeta,
				Spec:       prefixSpec,
			}

			if err := ctrl.SetControllerReference(nic, prefix, r.Scheme); err != nil {
				return nil, fmt.Errorf("error setting controller reference on %s: %w", prefix.Name, err)
			}

			if err := r.Patch(ctx, prefix, client.Apply, nicFieldOwner, client.ForceOwnership); err != nil {
				return nil, fmt.Errorf("error applying ephemeral prefix %s: %w", prefix.Name, err)
			}

			if ipamv1alpha1.GetPrefixReadiness(prefix) == ipamv1alpha1.ReadinessSucceeded {
				ips = append(ips, prefix.Spec.Prefix.IP())
			}
		}
	}
	return ips, nil
}

func (r *NetworkInterfaceReconciler) applyNetworkInterfaceBinding(ctx context.Context, nic *networkingv1alpha1.NetworkInterface, ips []commonv1alpha1.IP) error {
	nicBinding := &networkingv1alpha1.NetworkInterfaceBinding{
		TypeMeta: metav1.TypeMeta{
			APIVersion: networkingv1alpha1.SchemeGroupVersion.String(),
			Kind:       "NetworkInterfaceBinding",
		},
		ObjectMeta: metav1.ObjectMeta{
			Namespace: nic.Namespace,
			Name:      nic.Name,
		},
		IPs: ips,
	}
	if err := ctrl.SetControllerReference(nic, nicBinding, r.Scheme); err != nil {
		return fmt.Errorf("error setting controller reference on network interface binding: %w", err)
	}

	if err := r.Patch(ctx, nicBinding, client.Apply, nicFieldOwner, client.ForceOwnership); err != nil {
		return fmt.Errorf("error applying network interface binding: %w", err)
	}
	return nil
}

func (r *NetworkInterfaceReconciler) patchStatus(ctx context.Context, nic *networkingv1alpha1.NetworkInterface, ips []commonv1alpha1.IP) error {
	base := nic.DeepCopy()
	nic.Status.IPs = ips
	if err := r.Status().Patch(ctx, nic, client.MergeFrom(base)); err != nil {
		return fmt.Errorf("error patching status: %w", err)
	}
	return nil
}

func (r *NetworkInterfaceReconciler) SetupWithManager(mgr ctrl.Manager) error {
	return ctrl.NewControllerManagedBy(mgr).
		For(&networkingv1alpha1.NetworkInterface{}).
		Owns(&ipamv1alpha1.Prefix{}).
		Owns(&networkingv1alpha1.NetworkInterfaceBinding{}).
		Watches(
			&source.Kind{Type: &computev1alpha1.Machine{}},
			handler.EnqueueRequestsFromMapFunc(func(obj client.Object) []ctrl.Request {
				machine := obj.(*computev1alpha1.Machine)

				var reqs []ctrl.Request
				for _, nic := range machine.Spec.Interfaces {
					if nic.NetworkInterfaceRef != nil {
						reqs = append(reqs, ctrl.Request{NamespacedName: client.ObjectKey{Namespace: machine.Namespace, Name: nic.Name}})
					}
				}

				return reqs
			}),
		).
		Complete(r)
}
