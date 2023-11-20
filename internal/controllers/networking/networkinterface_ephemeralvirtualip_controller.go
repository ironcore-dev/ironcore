// SPDX-FileCopyrightText: 2023 SAP SE or an SAP affiliate company and IronCore contributors
// SPDX-License-Identifier: Apache-2.0

package networking

import (
	"context"
	"errors"
	"fmt"
	"maps"

	"github.com/ironcore-dev/ironcore/utils/annotations"

	"github.com/go-logr/logr"
	commonv1alpha1 "github.com/ironcore-dev/ironcore/api/common/v1alpha1"
	networkingv1alpha1 "github.com/ironcore-dev/ironcore/api/networking/v1alpha1"
	networkingclient "github.com/ironcore-dev/ironcore/internal/client/networking"
	klogutils "github.com/ironcore-dev/ironcore/utils/klog"
	apierrors "k8s.io/apimachinery/pkg/api/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/klog/v2"
	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/builder"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/handler"
	"sigs.k8s.io/controller-runtime/pkg/predicate"
)

type NetworkInterfaceEphemeralVirtualIPReconciler struct {
	client.Client
}

//+kubebuilder:rbac:groups=networking.ironcore.dev,resources=networkinterfaces,verbs=get;list;watch
//+kubebuilder:rbac:groups=networking.ironcore.dev,resources=virtualips,verbs=get;list;watch;create;update;patch;delete

func (r *NetworkInterfaceEphemeralVirtualIPReconciler) Reconcile(ctx context.Context, req ctrl.Request) (ctrl.Result, error) {
	log := ctrl.LoggerFrom(ctx)
	nic := &networkingv1alpha1.NetworkInterface{}
	if err := r.Get(ctx, req.NamespacedName, nic); err != nil {
		return ctrl.Result{}, client.IgnoreNotFound(err)
	}

	return r.reconcileExists(ctx, log, nic)
}

func (r *NetworkInterfaceEphemeralVirtualIPReconciler) reconcileExists(ctx context.Context, log logr.Logger, nic *networkingv1alpha1.NetworkInterface) (ctrl.Result, error) {
	if !nic.DeletionTimestamp.IsZero() {
		return ctrl.Result{}, nil
	}

	return r.reconcile(ctx, log, nic)
}

func (r *NetworkInterfaceEphemeralVirtualIPReconciler) ephemeralNetworkInterfaceVirtualIPByName(nic *networkingv1alpha1.NetworkInterface) map[string]*networkingv1alpha1.VirtualIP {
	res := make(map[string]*networkingv1alpha1.VirtualIP)

	vipSrc := nic.Spec.VirtualIP
	if vipSrc == nil {
		return res
	}
	ephemeral := vipSrc.Ephemeral
	if ephemeral == nil {
		return res
	}

	virtualIPName := networkingv1alpha1.NetworkInterfaceVirtualIPName(nic.Name, *vipSrc)
	virtualIP := &networkingv1alpha1.VirtualIP{
		ObjectMeta: metav1.ObjectMeta{
			Namespace:   nic.Namespace,
			Name:        virtualIPName,
			Labels:      ephemeral.VirtualIPTemplate.Labels,
			Annotations: maps.Clone(ephemeral.VirtualIPTemplate.Annotations),
		},
		Spec: ephemeral.VirtualIPTemplate.Spec,
	}
	annotations.SetDefaultEphemeralManagedBy(virtualIP)
	_ = ctrl.SetControllerReference(nic, virtualIP, r.Scheme())
	virtualIP.Spec.TargetRef = &commonv1alpha1.LocalUIDReference{
		Name: nic.Name,
		UID:  nic.UID,
	}
	res[virtualIPName] = virtualIP

	return res
}

func (r *NetworkInterfaceEphemeralVirtualIPReconciler) handleExistingVirtualIP(ctx context.Context, log logr.Logger, nic *networkingv1alpha1.NetworkInterface, shouldManage bool, virtualIP *networkingv1alpha1.VirtualIP) error {
	if annotations.IsDefaultEphemeralControlledBy(virtualIP, nic) {
		if shouldManage {
			log.V(1).Info("Ephemeral virtual IP is present and controlled by network interface")
			return nil
		}

		if !virtualIP.DeletionTimestamp.IsZero() {
			log.V(1).Info("Undesired ephemeral virtual IP is already deleting")
			return nil
		}

		log.V(1).Info("Deleting undesired ephemeral virtual IP")
		if err := r.Delete(ctx, virtualIP); client.IgnoreNotFound(err) != nil {
			return fmt.Errorf("error deleting virtual IP %s: %w", virtualIP.Name, err)
		}
		return nil
	}

	if shouldManage {
		log.V(1).Info("Won't adopt unmanaged virtual IP")
	}
	return nil
}

func (r *NetworkInterfaceEphemeralVirtualIPReconciler) handleCreateVirtualIP(
	ctx context.Context,
	log logr.Logger,
	nic *networkingv1alpha1.NetworkInterface,
	virtualIP *networkingv1alpha1.VirtualIP,
) error {
	log.V(1).Info("Creating virtual IP")
	virtualIPKey := client.ObjectKeyFromObject(virtualIP)
	err := r.Create(ctx, virtualIP)
	if err == nil {
		return nil
	}
	if !apierrors.IsAlreadyExists(err) {
		return err
	}

	// Due to a fast resync, we might get an already exists error.
	// In this case, try to fetch the virtual IP again and, when successful, treat it as managing
	// an existing virtual IP.
	if err := r.Get(ctx, virtualIPKey, virtualIP); err != nil {
		return fmt.Errorf("error getting virtual IP %s after already exists: %w", virtualIPKey.Name, err)
	}

	// Treat a retrieved virtual IP as an existing we should manage.
	log.V(1).Info("Retrieved virtual IP after already exists conflict")
	return r.handleExistingVirtualIP(ctx, log, nic, true, virtualIP)
}

func (r *NetworkInterfaceEphemeralVirtualIPReconciler) reconcile(ctx context.Context, log logr.Logger, nic *networkingv1alpha1.NetworkInterface) (ctrl.Result, error) {
	log.V(1).Info("Reconcile")

	log.V(1).Info("Listing virtual IPs")
	virtualIPList := &networkingv1alpha1.VirtualIPList{}
	if err := r.List(ctx, virtualIPList,
		client.InNamespace(nic.Namespace),
	); err != nil {
		return ctrl.Result{}, fmt.Errorf("error listing virtual IPs: %w", err)
	}
	log.V(5).Info("Listed virtual IPs", "VirtualIPs", klogutils.KObjStructSlice(virtualIPList.Items))

	var (
		ephemNicByName = r.ephemeralNetworkInterfaceVirtualIPByName(nic)
		errs           []error
	)
	for _, virtualIP := range virtualIPList.Items {
		virtualIPName := virtualIP.Name
		_, shouldManage := ephemNicByName[virtualIPName]
		delete(ephemNicByName, virtualIPName)
		log := log.WithValues("VirtualIP", klog.KObj(&virtualIP), "ShouldManage", shouldManage)
		if err := r.handleExistingVirtualIP(ctx, log, nic, shouldManage, &virtualIP); err != nil {
			errs = append(errs, err)
		}
	}

	for _, virtualIP := range ephemNicByName {
		log := log.WithValues("VirtualIP", klog.KObj(virtualIP))
		if err := r.handleCreateVirtualIP(ctx, log, nic, virtualIP); err != nil {
			errs = append(errs, err)
		}
	}

	if err := errors.Join(errs...); err != nil {
		return ctrl.Result{}, fmt.Errorf("error managing ephemeral virtual IPs: %w", err)
	}

	log.V(1).Info("Reconciled")
	return ctrl.Result{}, nil
}

func (r *NetworkInterfaceEphemeralVirtualIPReconciler) networkInterfaceNotDeletingPredicate() predicate.Predicate {
	return predicate.NewPredicateFuncs(func(obj client.Object) bool {
		networkInterface := obj.(*networkingv1alpha1.NetworkInterface)
		return networkInterface.DeletionTimestamp.IsZero()
	})
}

func (r *NetworkInterfaceEphemeralVirtualIPReconciler) enqueueByVirtualIP() handler.EventHandler {
	return handler.EnqueueRequestsFromMapFunc(func(ctx context.Context, obj client.Object) []ctrl.Request {
		virtualIP := obj.(*networkingv1alpha1.VirtualIP)
		log := ctrl.LoggerFrom(ctx)

		nicList := &networkingv1alpha1.NetworkInterfaceList{}
		if err := r.List(ctx, nicList,
			client.InNamespace(virtualIP.Namespace),
			client.MatchingFields{
				networkingclient.NetworkInterfaceVirtualIPNamesField: virtualIP.Name,
			},
		); err != nil {
			log.Error(err, "Error listing network interfaces")
			return nil
		}

		var reqs []ctrl.Request
		for _, nic := range nicList.Items {
			if !nic.DeletionTimestamp.IsZero() {
				continue
			}

			reqs = append(reqs, ctrl.Request{NamespacedName: client.ObjectKeyFromObject(&nic)})
		}
		return reqs
	})
}

func (r *NetworkInterfaceEphemeralVirtualIPReconciler) SetupWithManager(mgr ctrl.Manager) error {
	return ctrl.NewControllerManagedBy(mgr).
		Named("networkinterfaceephemeralvirtualip").
		For(
			&networkingv1alpha1.NetworkInterface{},
			builder.WithPredicates(
				r.networkInterfaceNotDeletingPredicate(),
			),
		).
		Owns(&networkingv1alpha1.VirtualIP{}).
		Watches(
			&networkingv1alpha1.VirtualIP{},
			r.enqueueByVirtualIP(),
		).
		Complete(r)
}
