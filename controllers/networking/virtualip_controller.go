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
	networkingv1alpha1 "github.com/onmetal/onmetal-api/apis/networking/v1alpha1"
	apierrors "k8s.io/apimachinery/pkg/api/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/labels"
	"k8s.io/apimachinery/pkg/runtime"
	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/handler"
	"sigs.k8s.io/controller-runtime/pkg/source"
)

var virtualIPFieldOwner = client.FieldOwner(networkingv1alpha1.Resource("virtualips").String())

type VirtualIPReconciler struct {
	client.Client
	Scheme *runtime.Scheme
}

func (r *VirtualIPReconciler) Reconcile(ctx context.Context, req ctrl.Request) (ctrl.Result, error) {
	log := ctrl.LoggerFrom(ctx)
	virtualIP := &networkingv1alpha1.VirtualIP{}
	if err := r.Get(ctx, req.NamespacedName, virtualIP); err != nil {
		return ctrl.Result{}, client.IgnoreNotFound(err)
	}

	return r.reconcileExists(ctx, log, virtualIP)
}

func (r *VirtualIPReconciler) reconcileExists(ctx context.Context, log logr.Logger, virtualIP *networkingv1alpha1.VirtualIP) (ctrl.Result, error) {
	if !virtualIP.DeletionTimestamp.IsZero() {
		return r.delete(ctx, log, virtualIP)
	}
	return r.reconcile(ctx, log, virtualIP)
}

func (r *VirtualIPReconciler) delete(ctx context.Context, log logr.Logger, virtualIP *networkingv1alpha1.VirtualIP) (ctrl.Result, error) {
	return ctrl.Result{}, nil
}

func (r *VirtualIPReconciler) reconcile(ctx context.Context, log logr.Logger, virtualIP *networkingv1alpha1.VirtualIP) (ctrl.Result, error) {
	if virtualIP.Spec.NetworkInterfaceSelector == nil {
		log.V(1).Info("Network interface selector is nil, assuming the virtual ip is managed by an external process")
		return ctrl.Result{}, nil
	}

	log.V(1).Info("Network interface selector is present, managing the virtual ip")
	sel, err := metav1.LabelSelectorAsSelector(virtualIP.Spec.NetworkInterfaceSelector)
	if err != nil {
		return ctrl.Result{}, fmt.Errorf("invalid network interface selector: %w", err)
	}

	log.V(1).Info("Listing network interfaces for virtual ip")
	nicList := &networkingv1alpha1.NetworkInterfaceList{}
	if err := r.List(ctx, nicList,
		client.InNamespace(virtualIP.Namespace),
		client.MatchingLabelsSelector{Selector: sel},
	); err != nil {
		return ctrl.Result{}, fmt.Errorf("error listing network interfaces: %w", err)
	}

	log.V(1).Info("Constructing virtual ip subsets")
	targetsByNetworkName := make(map[string][]networkingv1alpha1.VirtualIPRoutingSubsetTarget)
	for _, nic := range nicList.Items {
		nicKey := client.ObjectKeyFromObject(&nic)
		nicBinding := &networkingv1alpha1.NetworkInterfaceBinding{}
		if err := r.Get(ctx, nicKey, nicBinding); err != nil {
			if !apierrors.IsNotFound(err) {
				return ctrl.Result{}, fmt.Errorf("non not-found error getting network interface binding: %w", err)
			}

			log.V(1).Info("Network interface binding of network interface not found, omitting it.",
				"NetworkInterfaceKey", nicKey,
			)
			continue
		}

		networkName := nic.Spec.NetworkRef.Name
		for _, ip := range nicBinding.IPs {
			targetsByNetworkName[networkName] = append(targetsByNetworkName[networkName], networkingv1alpha1.VirtualIPRoutingSubsetTarget{
				LocalUIDReference: networkingv1alpha1.LocalUIDReference{
					Name: nic.Name,
					UID:  nic.UID,
				},
				IP: ip,
			})
		}
	}

	subsets := make([]networkingv1alpha1.VirtualIPRoutingSubset, 0, len(targetsByNetworkName))
	for networkName, targets := range targetsByNetworkName {
		network := &networkingv1alpha1.Network{}
		networkKey := client.ObjectKey{Namespace: virtualIP.Namespace, Name: networkName}
		if err := r.Get(ctx, networkKey, network); err != nil {
			if !apierrors.IsNotFound(err) {
				return ctrl.Result{}, fmt.Errorf("non not-found error getting network: %w", err)
			}

			log.V(1).Info("Network not found, omitting all routing targeting it", "NetworkKey", networkKey)
			continue
		}

		subsets = append(subsets, networkingv1alpha1.VirtualIPRoutingSubset{
			NetworkRef: networkingv1alpha1.LocalUIDReference{
				Name: networkName,
				UID:  network.UID,
			},
			Targets: targets,
		})
	}

	log.V(1).Info("Applying virtual ip routing")
	virtualIPRouting := &networkingv1alpha1.VirtualIPRouting{
		TypeMeta: metav1.TypeMeta{
			APIVersion: networkingv1alpha1.SchemeGroupVersion.String(),
			Kind:       "VirtualIPRouting",
		},
		ObjectMeta: metav1.ObjectMeta{
			Namespace: virtualIP.Namespace,
			Name:      virtualIP.Name,
		},
		Subsets: subsets,
	}
	if err := ctrl.SetControllerReference(virtualIP, virtualIPRouting, r.Scheme); err != nil {
		return ctrl.Result{}, fmt.Errorf("error owning virtual ip routing: %w", err)
	}
	if err := r.Patch(ctx, virtualIPRouting, client.Apply, virtualIPFieldOwner, client.ForceOwnership); err != nil {
		return ctrl.Result{}, fmt.Errorf("error applying virtual ip routing: %w", err)
	}

	log.V(1).Info("Successfully applied virtual ip routing")
	return ctrl.Result{}, nil
}

func (r *VirtualIPReconciler) SetupWithManager(mgr ctrl.Manager) error {
	log := ctrl.Log.WithName("virtualip").WithName("setup")
	ctx := ctrl.LoggerInto(context.Background(), log)

	return ctrl.NewControllerManagedBy(mgr).
		For(&networkingv1alpha1.VirtualIP{}).
		Owns(&networkingv1alpha1.VirtualIPRouting{}).
		Watches(
			&source.Kind{Type: &networkingv1alpha1.NetworkInterface{}},
			handler.EnqueueRequestsFromMapFunc(func(obj client.Object) []ctrl.Request {
				nic := obj.(*networkingv1alpha1.NetworkInterface)
				return r.requestsByNetworkInterface(ctx, log, nic)
			}),
		).
		Watches(
			&source.Kind{Type: &networkingv1alpha1.NetworkInterfaceBinding{}},
			handler.EnqueueRequestsFromMapFunc(func(obj client.Object) []ctrl.Request {
				nicBinding := obj.(*networkingv1alpha1.NetworkInterfaceBinding)
				return r.requestsByNetworkInterfaceBinding(ctx, log, nicBinding)
			}),
		).
		Complete(r)
}

func (r *VirtualIPReconciler) requestsByNetworkInterface(ctx context.Context, log logr.Logger, nic *networkingv1alpha1.NetworkInterface) []ctrl.Request {
	virtualIPList := &networkingv1alpha1.VirtualIPList{}
	if err := r.List(ctx, virtualIPList, client.InNamespace(nic.Namespace)); err != nil {
		log.Error(err, "Error listing virtual ips", "Namespace", nic.Namespace)
		return nil
	}

	var reqs []ctrl.Request
	for _, virtualIP := range virtualIPList.Items {
		nicSelector := virtualIP.Spec.NetworkInterfaceSelector
		if nicSelector == nil {
			// This means the virtual ip routing is managed externally.
			continue
		}

		sel, err := metav1.LabelSelectorAsSelector(nicSelector)
		if err != nil {
			log.V(1).Error(err, "Network interface has invalid selector",
				"NetworkInterfaceKey", client.ObjectKeyFromObject(nic),
			)
			continue
		}

		if sel.Matches(labels.Set(nic.Labels)) {
			reqs = append(reqs, ctrl.Request{NamespacedName: client.ObjectKeyFromObject(&virtualIP)})
		}
	}
	return reqs
}

func (r *VirtualIPReconciler) requestsByNetworkInterfaceBinding(ctx context.Context, log logr.Logger, nicBinding *networkingv1alpha1.NetworkInterfaceBinding) []ctrl.Request {
	nic := &networkingv1alpha1.NetworkInterface{}
	nicKey := client.ObjectKeyFromObject(nicBinding)
	if err := r.Get(ctx, nicKey, nic); err != nil {
		log.Error(err, "Error getting network interface for binding", "NetworkInterfaceKey", nicKey)
		return nil
	}

	return r.requestsByNetworkInterface(ctx, log, nic)
}
