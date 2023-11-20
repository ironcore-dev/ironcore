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
	ipamv1alpha1 "github.com/ironcore-dev/ironcore/api/ipam/v1alpha1"
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

type NetworkInterfaceEphemeralPrefixReconciler struct {
	client.Client
}

//+kubebuilder:rbac:groups=networking.ironcore.dev,resources=networkinterfaces,verbs=get;list;watch
//+kubebuilder:rbac:groups=ipam.ironcore.dev,resources=prefixes,verbs=get;list;watch;create;update;patch;delete

func (r *NetworkInterfaceEphemeralPrefixReconciler) Reconcile(ctx context.Context, req ctrl.Request) (ctrl.Result, error) {
	log := ctrl.LoggerFrom(ctx)
	nic := &networkingv1alpha1.NetworkInterface{}
	if err := r.Get(ctx, req.NamespacedName, nic); err != nil {
		return ctrl.Result{}, client.IgnoreNotFound(err)
	}

	return r.reconcileExists(ctx, log, nic)
}

func (r *NetworkInterfaceEphemeralPrefixReconciler) reconcileExists(ctx context.Context, log logr.Logger, nic *networkingv1alpha1.NetworkInterface) (ctrl.Result, error) {
	if !nic.DeletionTimestamp.IsZero() {
		return ctrl.Result{}, nil
	}

	return r.reconcile(ctx, log, nic)
}

func (r *NetworkInterfaceEphemeralPrefixReconciler) ephemeralNetworkInterfacePrefixByName(nic *networkingv1alpha1.NetworkInterface) map[string]*ipamv1alpha1.Prefix {
	res := make(map[string]*ipamv1alpha1.Prefix)

	for i, nicIP := range nic.Spec.IPs {
		ephemeral := nicIP.Ephemeral
		if ephemeral == nil {
			continue
		}

		prefixName := networkingv1alpha1.NetworkInterfaceIPIPAMPrefixName(nic.Name, i)
		prefix := &ipamv1alpha1.Prefix{
			ObjectMeta: metav1.ObjectMeta{
				Namespace:   nic.Namespace,
				Name:        prefixName,
				Labels:      ephemeral.PrefixTemplate.Labels,
				Annotations: maps.Clone(ephemeral.PrefixTemplate.Annotations),
			},
			Spec: ephemeral.PrefixTemplate.Spec,
		}
		annotations.SetDefaultEphemeralManagedBy(prefix)
		_ = ctrl.SetControllerReference(nic, prefix, r.Scheme())
		res[prefixName] = prefix
	}

	for i, nicPrefix := range nic.Spec.Prefixes {
		ephemeral := nicPrefix.Ephemeral
		if ephemeral == nil {
			continue
		}

		prefixName := networkingv1alpha1.NetworkInterfacePrefixIPAMPrefixName(nic.Name, i)
		prefix := &ipamv1alpha1.Prefix{
			ObjectMeta: metav1.ObjectMeta{
				Namespace:   nic.Namespace,
				Name:        prefixName,
				Labels:      ephemeral.PrefixTemplate.Labels,
				Annotations: ephemeral.PrefixTemplate.Annotations,
			},
			Spec: ephemeral.PrefixTemplate.Spec,
		}
		_ = ctrl.SetControllerReference(nic, prefix, r.Scheme())
		res[prefixName] = prefix
	}

	return res
}

func (r *NetworkInterfaceEphemeralPrefixReconciler) handleExistingPrefix(ctx context.Context, log logr.Logger, nic *networkingv1alpha1.NetworkInterface, shouldManage bool, prefix *ipamv1alpha1.Prefix) error {
	if annotations.IsDefaultEphemeralControlledBy(prefix, nic) {
		if shouldManage {
			log.V(1).Info("Ephemeral prefix is present and controlled by network interface")
			return nil
		}

		if !prefix.DeletionTimestamp.IsZero() {
			log.V(1).Info("Undesired ephemeral prefix is already deleting")
			return nil
		}

		log.V(1).Info("Deleting undesired ephemeral prefix")
		if err := r.Delete(ctx, prefix); client.IgnoreNotFound(err) != nil {
			return fmt.Errorf("error deleting prefix %s: %w", prefix.Name, err)
		}
		return nil
	}

	if shouldManage {
		log.V(1).Info("Won't adopt unmanaged prefix")
	}
	return nil
}

func (r *NetworkInterfaceEphemeralPrefixReconciler) handleCreatePrefix(
	ctx context.Context,
	log logr.Logger,
	nic *networkingv1alpha1.NetworkInterface,
	prefix *ipamv1alpha1.Prefix,
) error {
	log.V(1).Info("Creating prefix")
	prefixKey := client.ObjectKeyFromObject(prefix)
	err := r.Create(ctx, prefix)
	if err == nil {
		return nil
	}
	if !apierrors.IsAlreadyExists(err) {
		return err
	}

	// Due to a fast resync, we might get an already exists error.
	// In this case, try to fetch the prefix again and, when successful, treat it as managing
	// an existing prefix.
	if err := r.Get(ctx, prefixKey, prefix); err != nil {
		return fmt.Errorf("error getting prefix %s after already exists: %w", prefixKey.Name, err)
	}

	// Treat a retrieved prefix as an existing we should manage.
	log.V(1).Info("Retrieved prefix after already exists conflict")
	return r.handleExistingPrefix(ctx, log, nic, true, prefix)
}

func (r *NetworkInterfaceEphemeralPrefixReconciler) reconcile(ctx context.Context, log logr.Logger, nic *networkingv1alpha1.NetworkInterface) (ctrl.Result, error) {
	log.V(1).Info("Reconcile")

	log.V(1).Info("Listing prefixes")
	prefixList := &ipamv1alpha1.PrefixList{}
	if err := r.List(ctx, prefixList,
		client.InNamespace(nic.Namespace),
	); err != nil {
		return ctrl.Result{}, fmt.Errorf("error listing prefixes: %w", err)
	}
	log.V(5).Info("Listed prefixes", "Prefixes", klogutils.KObjStructSlice(prefixList.Items))

	var (
		ephemNicByName = r.ephemeralNetworkInterfacePrefixByName(nic)
		errs           []error
	)
	for _, prefix := range prefixList.Items {
		prefixName := prefix.Name
		_, shouldManage := ephemNicByName[prefixName]
		delete(ephemNicByName, prefixName)
		log := log.WithValues("Prefix", klog.KObj(&prefix), "ShouldManage", shouldManage)
		if err := r.handleExistingPrefix(ctx, log, nic, shouldManage, &prefix); err != nil {
			errs = append(errs, err)
		}
	}

	for _, prefix := range ephemNicByName {
		log := log.WithValues("Prefix", klog.KObj(prefix))
		if err := r.handleCreatePrefix(ctx, log, nic, prefix); err != nil {
			errs = append(errs, err)
		}
	}

	if err := errors.Join(errs...); err != nil {
		return ctrl.Result{}, fmt.Errorf("error managing ephemeral prefixes: %w", err)
	}

	log.V(1).Info("Reconciled")
	return ctrl.Result{}, nil
}

func (r *NetworkInterfaceEphemeralPrefixReconciler) networkInterfaceNotDeletingPredicate() predicate.Predicate {
	return predicate.NewPredicateFuncs(func(obj client.Object) bool {
		networkInterface := obj.(*networkingv1alpha1.NetworkInterface)
		return networkInterface.DeletionTimestamp.IsZero()
	})
}

func (r *NetworkInterfaceEphemeralPrefixReconciler) enqueueByPrefix() handler.EventHandler {
	return handler.EnqueueRequestsFromMapFunc(func(ctx context.Context, obj client.Object) []ctrl.Request {
		prefix := obj.(*ipamv1alpha1.Prefix)
		log := ctrl.LoggerFrom(ctx)

		nicList := &networkingv1alpha1.NetworkInterfaceList{}
		if err := r.List(ctx, nicList,
			client.InNamespace(prefix.Namespace),
			client.MatchingFields{
				networkingclient.NetworkInterfacePrefixNamesField: prefix.Name,
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

func (r *NetworkInterfaceEphemeralPrefixReconciler) SetupWithManager(mgr ctrl.Manager) error {
	return ctrl.NewControllerManagedBy(mgr).
		Named("networkinterfaceephemeralprefix").
		For(
			&networkingv1alpha1.NetworkInterface{},
			builder.WithPredicates(
				r.networkInterfaceNotDeletingPredicate(),
			),
		).
		Owns(&ipamv1alpha1.Prefix{}).
		Watches(
			&ipamv1alpha1.Prefix{},
			r.enqueueByPrefix(),
		).
		Complete(r)
}
