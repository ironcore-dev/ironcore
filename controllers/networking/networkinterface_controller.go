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
	onmetalapiclientutils "github.com/onmetal/onmetal-api/clientutils"
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
	log.V(1).Info("Getting network")
	network := &networkingv1alpha1.Network{}
	if err := r.Get(ctx, client.ObjectKey{Namespace: nic.Namespace, Name: nic.Spec.NetworkRef.Name}, network); err != nil {
		return ctrl.Result{}, fmt.Errorf("error getting network: %w", err)
	}
	log.V(1).Info("Successfully got network")

	log.V(1).Info("Applying IPs")
	ips, err := r.applyIPs(ctx, nic)
	if err != nil {
		return ctrl.Result{}, err
	}
	log.V(1).Info("Successfully applied IPs", "IPs", ips)

	log.V(1).Info("Applying virtual IPs")
	virtualIP, err := r.applyVirtualIP(ctx, log, nic)
	if err != nil {
		return ctrl.Result{}, err
	}
	log.V(1).Info("Successfully applied virtual IP")

	log.V(1).Info("Applying network interface binding")
	if err := r.applyNetworkInterfaceBinding(ctx, nic, network, ips, virtualIP); err != nil {
		return ctrl.Result{}, err
	}
	log.V(1).Info("Successfully applied network interface binding")

	log.V(1).Info("Patching status")
	var vip *commonv1alpha1.IP
	if virtualIP != nil {
		vip = virtualIP.Status.IP
	}
	if err := r.patchStatus(ctx, nic, ips, vip); err != nil {
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
			template := ipSource.EphemeralPrefix.PrefixTemplate
			prefix := &ipamv1alpha1.Prefix{
				ObjectMeta: metav1.ObjectMeta{
					Namespace: nic.Namespace,
					Name:      fmt.Sprintf("%s-%d", nic.Name, i),
				},
			}
			if err := onmetalapiclientutils.ControlledCreateOrGet(ctx, r.Client, nic, prefix, func() error {
				prefix.Labels = template.Labels
				prefix.Annotations = template.Annotations
				prefix.Spec = template.Spec
				return nil
			}); err != nil {
				return nil, fmt.Errorf("error managing ephemeral prefix %s: %w", prefix.Name, err)
			}

			if ipamv1alpha1.GetPrefixReadiness(prefix) == ipamv1alpha1.ReadinessSucceeded {
				ips = append(ips, prefix.Spec.Prefix.IP())
			}
		}
	}
	return ips, nil
}

func (r *NetworkInterfaceReconciler) applyVirtualIP(ctx context.Context, log logr.Logger, nic *networkingv1alpha1.NetworkInterface) (*networkingv1alpha1.VirtualIP, error) {
	if nic.Spec.VirtualIP == nil {
		log.V(1).Info("Network interface does not specify any virtual ip")
		return nil, nil
	}

	vipClaim, err := r.getOrManageVirtualIPClaim(ctx, log, nic)
	if err != nil {
		return nil, err
	}

	if phase := vipClaim.Status.Phase; phase != networkingv1alpha1.VirtualIPClaimPhaseBound {
		log.V(1).Info("Virtual ip claim phase is not yet bound", "Phase", phase)
		return nil, nil
	}

	vip := &networkingv1alpha1.VirtualIP{}
	vipKey := client.ObjectKey{Namespace: nic.Namespace, Name: vipClaim.Spec.VirtualIPRef.Name}
	log.V(1).Info("Getting virtual ip bound by claim", "VirtualIPKey", vipKey)
	if err := r.Get(ctx, vipKey, vip); err != nil {
		return nil, fmt.Errorf("error getting virtual ip %s: %w", vipKey, err)
	}

	log.V(1).Info("Successfully retrieved virtual ip", "VirtualIPIP", vip.Status.IP)
	return vip, nil
}

func (r *NetworkInterfaceReconciler) getOrManageVirtualIPClaim(ctx context.Context, log logr.Logger, nic *networkingv1alpha1.NetworkInterface) (*networkingv1alpha1.VirtualIPClaim, error) {
	if vipClaimRef := nic.Spec.VirtualIP.VirtualIPClaimRef; vipClaimRef != nil {
		vipClaim := &networkingv1alpha1.VirtualIPClaim{}
		vipClaimKey := client.ObjectKey{Namespace: nic.Namespace, Name: vipClaimRef.Name}
		log.V(1).Info("Getting referenced virtual ip claim", "VirtualIPClaimKey", vipClaimKey)
		if err := r.Get(ctx, vipClaimKey, vipClaim); err != nil {
			return nil, fmt.Errorf("error getting referenced virtual ip claim %s: %w", vipClaimKey, err)
		}

		return vipClaim, nil
	}

	log.V(1).Info("Managing ephemeral virtual ip claim")
	ephemeralVip := &networkingv1alpha1.VirtualIPClaim{
		ObjectMeta: metav1.ObjectMeta{
			Namespace: nic.Namespace,
			Name:      nic.Name,
		},
	}
	if err := onmetalapiclientutils.ControlledCreateOrGet(ctx, r.Client, nic, ephemeralVip, func() error {
		ephemeral := nic.Spec.VirtualIP.Ephemeral
		ephemeralVip.Labels = ephemeral.VirtualIPClaimTemplate.Labels
		ephemeralVip.Annotations = ephemeral.VirtualIPClaimTemplate.Annotations
		ephemeralVip.Spec = ephemeral.VirtualIPClaimTemplate.Spec
		return nil
	}); err != nil {
		return nil, fmt.Errorf("error managing ephemeral virtual ip claim: %w", err)
	}
	return ephemeralVip, nil
}

func (r *NetworkInterfaceReconciler) applyNetworkInterfaceBinding(
	ctx context.Context,
	nic *networkingv1alpha1.NetworkInterface,
	network *networkingv1alpha1.Network,
	ips []commonv1alpha1.IP,
	virtualIP *networkingv1alpha1.VirtualIP,
) error {
	var vipRef *commonv1alpha1.LocalUIDReference
	if virtualIP != nil {
		vipRef = &commonv1alpha1.LocalUIDReference{
			Name: virtualIP.Name,
			UID:  virtualIP.UID,
		}
	}
	nicBinding := &networkingv1alpha1.NetworkInterfaceBinding{
		TypeMeta: metav1.TypeMeta{
			APIVersion: networkingv1alpha1.SchemeGroupVersion.String(),
			Kind:       "NetworkInterfaceBinding",
		},
		ObjectMeta: metav1.ObjectMeta{
			Namespace: nic.Namespace,
			Name:      nic.Name,
		},
		NetworkRef: commonv1alpha1.LocalUIDReference{
			Name: network.Name,
			UID:  network.UID,
		},
		IPs:          ips,
		VirtualIPRef: vipRef,
	}
	if err := ctrl.SetControllerReference(nic, nicBinding, r.Scheme); err != nil {
		return fmt.Errorf("error setting controller reference on network interface binding: %w", err)
	}
	if err := r.Patch(ctx, nicBinding, client.Apply, nicFieldOwner, client.ForceOwnership); err != nil {
		return fmt.Errorf("error applying network interface binding: %w", err)
	}
	return nil
}

func (r *NetworkInterfaceReconciler) patchStatus(ctx context.Context, nic *networkingv1alpha1.NetworkInterface, ips []commonv1alpha1.IP, virtualIP *commonv1alpha1.IP) error {
	base := nic.DeepCopy()
	nic.Status.IPs = ips
	nic.Status.VirtualIP = virtualIP
	if err := r.Status().Patch(ctx, nic, client.MergeFrom(base)); err != nil {
		return fmt.Errorf("error patching status: %w", err)
	}
	return nil
}

const networkInterfaceSpecVirtualIPVirtualIPClaimRefNameField = ".spec.virtualIP.virtualIPClaimRef.name"

func (r *NetworkInterfaceReconciler) SetupWithManager(mgr ctrl.Manager) error {
	ctx := context.Background()
	log := ctrl.Log.WithName("networkinterface").WithName("setup")

	if err := mgr.GetFieldIndexer().IndexField(ctx, &networkingv1alpha1.NetworkInterface{}, networkInterfaceSpecVirtualIPVirtualIPClaimRefNameField, func(obj client.Object) []string {
		nic := obj.(*networkingv1alpha1.NetworkInterface)
		virtualIP := nic.Spec.VirtualIP
		if virtualIP == nil {
			return nil
		}

		claimRef := virtualIP.VirtualIPClaimRef
		if claimRef == nil {
			return nil
		}

		return []string{claimRef.Name}
	}); err != nil {
		return err
	}

	return ctrl.NewControllerManagedBy(mgr).
		For(&networkingv1alpha1.NetworkInterface{}).
		Owns(&ipamv1alpha1.Prefix{}).
		Owns(&networkingv1alpha1.NetworkInterfaceBinding{}).
		Owns(&networkingv1alpha1.VirtualIPClaim{}).
		Watches(
			&source.Kind{Type: &networkingv1alpha1.VirtualIPClaim{}},
			r.enqueueByReverseVirtualIPClaimReference(log, ctx)).
		Watches(
			&source.Kind{Type: &computev1alpha1.Machine{}},
			r.enqueueByMachineNetworkInterfaceReference(),
		).
		Complete(r)
}

func (r *NetworkInterfaceReconciler) enqueueByMachineNetworkInterfaceReference() handler.EventHandler {
	return handler.EnqueueRequestsFromMapFunc(func(obj client.Object) []ctrl.Request {
		machine := obj.(*computev1alpha1.Machine)

		var reqs []ctrl.Request
		for _, nic := range machine.Spec.Interfaces {
			if nic.NetworkInterfaceRef != nil {
				reqs = append(reqs, ctrl.Request{NamespacedName: client.ObjectKey{Namespace: machine.Namespace, Name: nic.Name}})
			}
		}

		return reqs
	})
}

func (r *NetworkInterfaceReconciler) enqueueByReverseVirtualIPClaimReference(log logr.Logger, ctx context.Context) handler.EventHandler {
	return handler.EnqueueRequestsFromMapFunc(func(obj client.Object) []ctrl.Request {
		vipClaim := obj.(*networkingv1alpha1.VirtualIPClaim)
		log = log.WithValues("VirtualIPClaimKey", client.ObjectKeyFromObject(vipClaim))

		nicList := &networkingv1alpha1.NetworkInterfaceList{}
		if err := r.List(ctx, nicList,
			client.InNamespace(vipClaim.Namespace),
			client.MatchingFields{networkInterfaceSpecVirtualIPVirtualIPClaimRefNameField: vipClaim.Name},
		); err != nil {
			log.Error(err, "Error listing network interfaces using virtual ip claim")
			return nil
		}

		reqs := make([]ctrl.Request, 0, len(nicList.Items))
		for _, nic := range nicList.Items {
			reqs = append(reqs, ctrl.Request{NamespacedName: client.ObjectKeyFromObject(&nic)})
		}
		return reqs
	})
}
