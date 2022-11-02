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
	"errors"
	"fmt"
	"reflect"

	"github.com/go-logr/logr"
	commonv1alpha1 "github.com/onmetal/onmetal-api/apis/common/v1alpha1"
	ipamv1alpha1 "github.com/onmetal/onmetal-api/apis/ipam/v1alpha1"
	networkingv1alpha1 "github.com/onmetal/onmetal-api/apis/networking/v1alpha1"
	onmetalapiclient "github.com/onmetal/onmetal-api/onmetal-controller-manager/client"
	onmetalapiclientutils "github.com/onmetal/onmetal-api/onmetal-controller-manager/clientutils"
	"github.com/onmetal/onmetal-api/onmetal-controller-manager/controllers/networking/events"
	corev1 "k8s.io/api/core/v1"
	apierrors "k8s.io/apimachinery/pkg/api/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/client-go/tools/record"
	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/handler"
	"sigs.k8s.io/controller-runtime/pkg/source"
)

type NetworkInterfaceReconciler struct {
	record.EventRecorder
	client.Client
	Scheme *runtime.Scheme
}

//+kubebuilder:rbac:groups=networking.api.onmetal.de,resources=networkinterfaces,verbs=get;list;watch;create;update;patch;delete
//+kubebuilder:rbac:groups=networking.api.onmetal.de,resources=networkinterfaces/status,verbs=get;update;patch
//+kubebuilder:rbac:groups=networking.api.onmetal.de,resources=networkinterfaces/finalizers,verbs=update
//+kubebuilder:rbac:groups=networking.api.onmetal.de,resources=virtualips,verbs=get;create;list;watch
//+kubebuilder:rbac:groups=networking.api.onmetal.de,resources=networkinterfacebindings,verbs=get;list;watch;create;update;patch
//+kubebuilder:rbac:groups=ipam.api.onmetal.de,resources=prefix,verbs=get;create;list;update;patch;watch
//+kubebuilder:rbac:groups=networking.api.onmetal.de,resources=networks,verbs=get;list;watch

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

	log.V(1).Info("Applying virtual IPs")
	virtualIP, err := r.applyVirtualIP(ctx, log, nic)
	if err != nil {
		return ctrl.Result{}, err
	}
	log.V(1).Info("Successfully applied virtual IP")

	log.V(1).Info("Patching status")
	if err := r.patchStatus(ctx, nic, ips, virtualIP); err != nil {
		return ctrl.Result{}, err
	}
	log.V(1).Info("Successfully patched status")
	return ctrl.Result{}, nil
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

func (r *NetworkInterfaceReconciler) applyIPs(ctx context.Context, nic *networkingv1alpha1.NetworkInterface) ([]commonv1alpha1.IP, error) {
	var (
		ips           []commonv1alpha1.IP
		applyWarnings []string
	)
	for idx, ipSource := range nic.Spec.IPs {
		ip, ok, applyWarning, err := r.applyIP(ctx, nic, ipSource, idx)
		if err != nil {
			return nil, fmt.Errorf("[ip %d] %w", idx, err)
		}
		if ok {
			ips = append(ips, ip)
		}
		if applyWarning != "" {
			applyWarnings = append(applyWarnings, applyWarning)
		}
	}

	if len(applyWarnings) > 0 {
		r.Eventf(nic, corev1.EventTypeWarning, events.FailedApplyingIPs, "Failed applying ip(s): %v", applyWarnings)
	}

	return ips, nil
}

func (r *NetworkInterfaceReconciler) applyIP(ctx context.Context, nic *networkingv1alpha1.NetworkInterface, ipSource networkingv1alpha1.IPSource, idx int) (commonv1alpha1.IP, bool, string, error) {
	switch {
	case ipSource.Value != nil:
		return *ipSource.Value, true, "", nil
	case ipSource.Ephemeral != nil:
		template := ipSource.Ephemeral.PrefixTemplate
		prefix := &ipamv1alpha1.Prefix{
			ObjectMeta: metav1.ObjectMeta{
				Namespace: nic.Namespace,
				Name:      networkingv1alpha1.NetworkInterfaceIPSourceEphemeralPrefixName(nic.Name, idx),
			},
		}
		if err := onmetalapiclientutils.ControlledCreateOrGet(ctx, r.Client, nic, prefix, func() error {
			prefix.Labels = template.Labels
			prefix.Annotations = template.Annotations
			prefix.Spec = template.Spec
			return nil
		}); err != nil {
			if !errors.Is(err, onmetalapiclientutils.ErrNotControlled) {
				return commonv1alpha1.IP{}, false, "", fmt.Errorf("error managing ephemeral prefix %s: %w", prefix.Name, err)
			}
			return commonv1alpha1.IP{}, false, fmt.Sprintf("prefix %s cannot be managed", prefix.Name), nil
		}

		if prefix.Status.Phase == ipamv1alpha1.PrefixPhaseAllocated {
			return prefix.Spec.Prefix.IP(), true, "", nil
		}
		return commonv1alpha1.IP{}, false, "", nil
	default:
		return commonv1alpha1.IP{}, false, "", fmt.Errorf("unknown ip source %#v", ipSource)
	}
}

func (r *NetworkInterfaceReconciler) applyVirtualIP(ctx context.Context, log logr.Logger, nic *networkingv1alpha1.NetworkInterface) (*commonv1alpha1.IP, error) {
	virtualIP := nic.Spec.VirtualIP
	switch {
	case virtualIP == nil:
		log.V(1).Info("Network interface does not specify any virtual ip")
		return nil, nil
	case virtualIP.VirtualIPRef != nil:
		vip := &networkingv1alpha1.VirtualIP{}
		vipKey := client.ObjectKey{Namespace: nic.Namespace, Name: virtualIP.VirtualIPRef.Name}
		log.V(1).Info("Getting referenced virtual ip", "VirtualIPKey", vipKey)
		if err := r.Get(ctx, vipKey, vip); err != nil {
			if !apierrors.IsNotFound(err) {
				return nil, fmt.Errorf("error getting referenced virtual ip claim %s: %w", vipKey, err)
			}

			r.Eventf(nic, corev1.EventTypeWarning, events.FailedApplyingVirtualIP, "Virtual IP %s not found", vipKey.Name)
			return nil, nil
		}

		return r.getVirtualIPIP(log, nic, vip), nil
	case virtualIP.Ephemeral != nil:
		log.V(1).Info("Managing ephemeral virtual ip claim")
		vip := &networkingv1alpha1.VirtualIP{
			ObjectMeta: metav1.ObjectMeta{
				Namespace: nic.Namespace,
				Name:      nic.Name,
			},
		}
		if err := onmetalapiclientutils.ControlledCreateOrGet(ctx, r.Client, nic, vip, func() error {
			ephemeral := nic.Spec.VirtualIP.Ephemeral
			vip.Labels = ephemeral.VirtualIPTemplate.Labels
			vip.Annotations = ephemeral.VirtualIPTemplate.Annotations
			vip.Spec = ephemeral.VirtualIPTemplate.Spec
			vip.Spec.TargetRef = &commonv1alpha1.LocalUIDReference{Name: nic.Name, UID: nic.UID}
			return nil
		}); err != nil {
			if !errors.Is(onmetalapiclientutils.ErrNotControlled, err) {
				return nil, fmt.Errorf("error managing ephemeral virtual ip: %w", err)
			}

			r.Eventf(nic, corev1.EventTypeWarning, events.FailedApplyingVirtualIP, "Cannot manage virtual ip %s", vip.Name)
			return nil, nil
		}
		return r.getVirtualIPIP(log, nic, vip), nil
	default:
		return nil, fmt.Errorf("unknown virtual ip %#v", virtualIP)
	}
}

func (r *NetworkInterfaceReconciler) getVirtualIPIP(log logr.Logger, nic *networkingv1alpha1.NetworkInterface, vip *networkingv1alpha1.VirtualIP) *commonv1alpha1.IP {
	if !reflect.DeepEqual(vip.Spec.TargetRef, &commonv1alpha1.LocalUIDReference{Name: nic.Name, UID: nic.UID}) {
		log.V(1).Info("Virtual ip does not target network interface", "TargetRef", vip.Spec.TargetRef)
		return nil
	}

	if phase := vip.Status.Phase; phase != networkingv1alpha1.VirtualIPPhaseBound {
		log.V(1).Info("Virtual ip is not bound", "Phase", phase)
		return nil
	}

	log.V(1).Info("Virtual ip is up and bound", "IP", vip.Status.IP)
	return vip.Status.IP
}

func (r *NetworkInterfaceReconciler) SetupWithManager(mgr ctrl.Manager) error {
	ctx := context.Background()
	log := ctrl.Log.WithName("networkinterface").WithName("setup")

	return ctrl.NewControllerManagedBy(mgr).
		For(&networkingv1alpha1.NetworkInterface{}).
		Owns(&ipamv1alpha1.Prefix{}).
		Owns(&networkingv1alpha1.VirtualIP{}).
		Watches(
			&source.Kind{Type: &networkingv1alpha1.VirtualIP{}},
			r.enqueueByNetworkInterfaceVirtualIPReferences(log, ctx),
		).
		Complete(r)
}

func (r *NetworkInterfaceReconciler) enqueueByNetworkInterfaceVirtualIPReferences(log logr.Logger, ctx context.Context) handler.EventHandler {
	return handler.EnqueueRequestsFromMapFunc(func(obj client.Object) []ctrl.Request {
		vip := obj.(*networkingv1alpha1.VirtualIP)
		log = log.WithValues("VirtualIPKey", client.ObjectKeyFromObject(vip))

		nicList := &networkingv1alpha1.NetworkInterfaceList{}
		if err := r.List(ctx, nicList,
			client.InNamespace(vip.Namespace),
			client.MatchingFields{onmetalapiclient.NetworkInterfaceVirtualIPNames: vip.Name},
		); err != nil {
			log.Error(err, "Error listing network interfaces using virtual ip")
			return nil
		}

		reqs := make([]ctrl.Request, 0, len(nicList.Items))
		for _, nic := range nicList.Items {
			reqs = append(reqs, ctrl.Request{NamespacedName: client.ObjectKeyFromObject(&nic)})
		}
		return reqs
	})
}
