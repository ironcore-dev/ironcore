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
	"strings"

	"github.com/go-logr/logr"
	commonv1alpha1 "github.com/onmetal/onmetal-api/api/common/v1alpha1"
	ipamv1alpha1 "github.com/onmetal/onmetal-api/api/ipam/v1alpha1"
	networkingv1alpha1 "github.com/onmetal/onmetal-api/api/networking/v1alpha1"
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

//+kubebuilder:rbac:groups=networking.api.onmetal.de,resources=networks,verbs=get;list;watch
//+kubebuilder:rbac:groups=networking.api.onmetal.de,resources=networkinterfaces,verbs=get;list;watch;update;patch
//+kubebuilder:rbac:groups=networking.api.onmetal.de,resources=networkinterfaces/status,verbs=get;update;patch
//+kubebuilder:rbac:groups=networking.api.onmetal.de,resources=networkinterfaces/finalizers,verbs=update
//+kubebuilder:rbac:groups=networking.api.onmetal.de,resources=virtualips,verbs=get;create;list;watch
//+kubebuilder:rbac:groups=ipam.api.onmetal.de,resources=prefix,verbs=get;create;list;update;patch;watch

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
	var anyDependencyNotReady bool

	log.V(1).Info("Getting network handle")
	networkHandle, networkNotReadyReason, err := r.getNetworkHandle(ctx, nic)
	if err != nil {
		r.Eventf(nic, corev1.EventTypeWarning, events.ErrorGettingNetworkHandle, "Error getting network handle: %v", err)
		return ctrl.Result{}, nil
	}
	if networkNotReadyReason != "" {
		anyDependencyNotReady = true
		r.Event(nic, corev1.EventTypeNormal, events.NetworkNotReady, networkNotReadyReason)
	}

	log.V(1).Info("Applying IPs")
	ips, ipsNotReadyReason, err := r.applyIPs(ctx, nic)
	if err != nil {
		r.Eventf(nic, corev1.EventTypeWarning, events.ErrorApplyingIPs, "Error applying ips: %v", err)
		return ctrl.Result{}, err
	}
	if ipsNotReadyReason != "" {
		anyDependencyNotReady = true
		r.Event(nic, corev1.EventTypeNormal, events.IPsNotReady, ipsNotReadyReason)
	}

	log.V(1).Info("Applying virtual IPs")
	virtualIP, virtualIPNotReadyReason, err := r.applyVirtualIP(ctx, log, nic)
	if err != nil {
		r.Eventf(nic, corev1.EventTypeWarning, events.ErrorApplyingVirtualIP, "Error applying virtual ip: %v", err)
		return ctrl.Result{}, err
	}
	if virtualIPNotReadyReason != "" {
		anyDependencyNotReady = true
		r.Event(nic, corev1.EventTypeNormal, events.VirtualIPNotReady, virtualIPNotReadyReason)
	}

	var state networkingv1alpha1.NetworkInterfaceState
	if anyDependencyNotReady {
		state = networkingv1alpha1.NetworkInterfaceStatePending
	} else {
		state = networkingv1alpha1.NetworkInterfaceStateAvailable
	}

	log.V(1).Info("Updating network interface status", "State", state)
	if err := r.updateStatus(ctx, nic, state, networkHandle, ips, virtualIP); err != nil {
		return ctrl.Result{}, err
	}
	log.V(1).Info("Successfully updated network status", "State", state)
	return ctrl.Result{}, nil
}

func (r *NetworkInterfaceReconciler) updateStatus(
	ctx context.Context,
	nic *networkingv1alpha1.NetworkInterface,
	state networkingv1alpha1.NetworkInterfaceState,
	networkHandle string,
	ips []commonv1alpha1.IP,
	virtualIP *commonv1alpha1.IP,
) error {
	now := metav1.Now()
	base := nic.DeepCopy()

	if nic.Status.State != state {
		nic.Status.LastStateTransitionTime = &now
	}
	nic.Status.State = state
	nic.Status.NetworkHandle = networkHandle
	nic.Status.IPs = ips
	nic.Status.VirtualIP = virtualIP

	if err := r.Status().Patch(ctx, nic, client.MergeFrom(base)); err != nil {
		return fmt.Errorf("error patching status: %w", err)
	}
	return nil
}

func (r *NetworkInterfaceReconciler) getNetworkHandle(ctx context.Context, nic *networkingv1alpha1.NetworkInterface) (networkHandle string, notReadyReason string, err error) {
	network := &networkingv1alpha1.Network{}
	networkKey := client.ObjectKey{Namespace: nic.Namespace, Name: nic.Spec.NetworkRef.Name}
	if err := r.Get(ctx, networkKey, network); err != nil {
		if !apierrors.IsNotFound(err) {
			return "", "", fmt.Errorf("error getting network %s: %w", networkKey, err)
		}
		return "", fmt.Sprintf("Network %s not found", networkKey.Name), nil
	}

	switch state := network.Status.State; state {
	case networkingv1alpha1.NetworkStateAvailable:
		return network.Spec.Handle, "", nil
	default:
		return "", fmt.Sprintf("Network %s is not in state %s but %s", networkKey.Name, networkingv1alpha1.NetworkStateAvailable, state), nil
	}
}

func (r *NetworkInterfaceReconciler) applyIPs(ctx context.Context, nic *networkingv1alpha1.NetworkInterface) ([]commonv1alpha1.IP, string, error) {
	var (
		ips             []commonv1alpha1.IP
		notReadyReasons []string
	)
	for idx, ipSource := range nic.Spec.IPs {
		ip, notReadyReason, err := r.applyIP(ctx, nic, ipSource, idx)
		if err != nil {
			return nil, "", fmt.Errorf("[ip %d] %w", idx, err)
		}
		if notReadyReason != "" {
			notReadyReasons = append(notReadyReasons, fmt.Sprintf("[ip %d] %s", idx, notReadyReason))
			continue
		}

		ips = append(ips, ip)
	}

	return ips, strings.Join(notReadyReasons, ", "), nil
}

func (r *NetworkInterfaceReconciler) applyIP(ctx context.Context, nic *networkingv1alpha1.NetworkInterface, ipSource networkingv1alpha1.IPSource, idx int) (commonv1alpha1.IP, string, error) {
	switch {
	case ipSource.Value != nil:
		return *ipSource.Value, "", nil
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
				return commonv1alpha1.IP{}, "", fmt.Errorf("error managing ephemeral prefix %s: %w", prefix.Name, err)
			}
			return commonv1alpha1.IP{}, fmt.Sprintf("Prefix %s cannot be managed", prefix.Name), nil
		}

		if prefix.Status.Phase != ipamv1alpha1.PrefixPhaseAllocated {
			return commonv1alpha1.IP{}, fmt.Sprintf("Prefix %s is not in state %s but %s", prefix.Name, ipamv1alpha1.PrefixPhaseAllocated, prefix.Status.Phase), nil
		}
		return prefix.Spec.Prefix.IP(), "", nil
	default:
		return commonv1alpha1.IP{}, "", fmt.Errorf("unknown ip source %#v", ipSource)
	}
}

func (r *NetworkInterfaceReconciler) applyVirtualIP(ctx context.Context, log logr.Logger, nic *networkingv1alpha1.NetworkInterface) (*commonv1alpha1.IP, string, error) {
	virtualIP := nic.Spec.VirtualIP
	switch {
	case virtualIP == nil:
		log.V(1).Info("Network interface does not specify any virtual ip")
		return nil, "", nil
	case virtualIP.VirtualIPRef != nil:
		vip := &networkingv1alpha1.VirtualIP{}
		vipKey := client.ObjectKey{Namespace: nic.Namespace, Name: virtualIP.VirtualIPRef.Name}
		log.V(1).Info("Getting referenced virtual ip", "VirtualIPKey", vipKey)
		if err := r.Get(ctx, vipKey, vip); err != nil {
			if !apierrors.IsNotFound(err) {
				return nil, "", fmt.Errorf("error getting referenced virtual ip claim %s: %w", vipKey, err)
			}

			return nil, fmt.Sprintf("Virtual IP %s not found", vipKey.Name), nil
		}

		res, notReadyReason := r.getVirtualIPIP(nic, vip)
		return res, notReadyReason, nil
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
				return nil, "", fmt.Errorf("error managing ephemeral virtual ip: %w", err)
			}

			return nil, fmt.Sprintf("Virtual IP %s cannot be managed", vip.Name), nil
		}
		res, notReadyReason := r.getVirtualIPIP(nic, vip)
		return res, notReadyReason, nil
	default:
		return nil, "", fmt.Errorf("unknown virtual ip %#v", virtualIP)
	}
}

func (r *NetworkInterfaceReconciler) getVirtualIPIP(nic *networkingv1alpha1.NetworkInterface, vip *networkingv1alpha1.VirtualIP) (*commonv1alpha1.IP, string) {
	if !reflect.DeepEqual(vip.Spec.TargetRef, &commonv1alpha1.LocalUIDReference{Name: nic.Name, UID: nic.UID}) {
		return nil, fmt.Sprintf("Virtual IP %s does not target network interface", vip.Name)
	}

	if phase := vip.Status.Phase; phase != networkingv1alpha1.VirtualIPPhaseBound {
		return nil, fmt.Sprintf("Virtual IP %s is not bound", vip.Name)
	}

	return vip.Status.IP, ""
}

func (r *NetworkInterfaceReconciler) SetupWithManager(mgr ctrl.Manager) error {
	ctx := context.Background()
	log := ctrl.Log.WithName("networkinterface").WithName("setup")

	return ctrl.NewControllerManagedBy(mgr).
		For(&networkingv1alpha1.NetworkInterface{}).
		Owns(&ipamv1alpha1.Prefix{}).
		Owns(&networkingv1alpha1.VirtualIP{}).
		Watches(
			&source.Kind{Type: &networkingv1alpha1.Network{}},
			r.enqueueByNetworkInterfaceNetworkReferences(ctx, log),
		).
		Watches(
			&source.Kind{Type: &networkingv1alpha1.VirtualIP{}},
			r.enqueueByNetworkInterfaceVirtualIPReferences(ctx, log),
		).
		Complete(r)
}

func (r *NetworkInterfaceReconciler) enqueueByNetworkInterfaceVirtualIPReferences(ctx context.Context, log logr.Logger) handler.EventHandler {
	return handler.EnqueueRequestsFromMapFunc(func(obj client.Object) []ctrl.Request {
		vip := obj.(*networkingv1alpha1.VirtualIP)
		log = log.WithValues("VirtualIPKey", client.ObjectKeyFromObject(vip))

		nicList := &networkingv1alpha1.NetworkInterfaceList{}
		if err := r.List(ctx, nicList,
			client.InNamespace(vip.Namespace),
			client.MatchingFields{onmetalapiclient.NetworkInterfaceVirtualIPNamesField: vip.Name},
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

func (r *NetworkInterfaceReconciler) enqueueByNetworkInterfaceNetworkReferences(ctx context.Context, log logr.Logger) handler.EventHandler {
	return handler.EnqueueRequestsFromMapFunc(func(obj client.Object) []ctrl.Request {
		network := obj.(*networkingv1alpha1.Network)
		log = log.WithValues("NetworkKey", client.ObjectKeyFromObject(network))

		nicList := &networkingv1alpha1.NetworkInterfaceList{}
		if err := r.List(ctx, nicList,
			client.InNamespace(network.Namespace),
			client.MatchingFields{onmetalapiclient.NetworkInterfaceNetworkNameField: network.Name},
		); err != nil {
			log.Error(err, "Error listing network interface using network")
			return nil
		}

		reqs := make([]ctrl.Request, 0, len(nicList.Items))
		for _, nic := range nicList.Items {
			reqs = append(reqs, ctrl.Request{NamespacedName: client.ObjectKeyFromObject(&nic)})
		}
		return reqs
	})
}
