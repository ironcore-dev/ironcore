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

	"github.com/go-logr/logr"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/labels"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apimachinery/pkg/types"
	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/handler"
	"sigs.k8s.io/controller-runtime/pkg/source"

	commonv1alpha1 "github.com/onmetal/onmetal-api/api/common/v1alpha1"
	ipamv1alpha1 "github.com/onmetal/onmetal-api/api/ipam/v1alpha1"
	networkingv1alpha1 "github.com/onmetal/onmetal-api/api/networking/v1alpha1"
	machinebroker "github.com/onmetal/onmetal-api/broker/machinebroker/api/v1alpha1"
	"github.com/onmetal/onmetal-api/internal/client/networking"
	onmetalapiclient "github.com/onmetal/onmetal-api/utils/client"
)

var (
	loadBalancerFieldOwner = client.FieldOwner(networkingv1alpha1.Resource("loadbalancers").String())
)

type LoadBalancerReconciler struct {
	client.Client
	Scheme *runtime.Scheme
}

//+kubebuilder:rbac:groups=networking.api.onmetal.de,resources=loadbalancers,verbs=get;list;watch;create;update;patch;delete
//+kubebuilder:rbac:groups=networking.api.onmetal.de,resources=loadbalancers/status,verbs=get;update;patch
//+kubebuilder:rbac:groups=networking.api.onmetal.de,resources=loadbalancers/finalizers,verbs=update
//+kubebuilder:rbac:groups=networking.api.onmetal.de,resources=networkinterfaces,verbs=get;list;watch
//+kubebuilder:rbac:groups=networking.api.onmetal.de,resources=loadbalancerroutings,verbs=get;list;watch;create;update;patch;delete

func (r *LoadBalancerReconciler) Reconcile(ctx context.Context, req ctrl.Request) (ctrl.Result, error) {
	log := ctrl.LoggerFrom(ctx)
	loadBalancer := &networkingv1alpha1.LoadBalancer{}
	if err := r.Get(ctx, req.NamespacedName, loadBalancer); err != nil {
		return ctrl.Result{}, client.IgnoreNotFound(err)
	}

	return r.reconcileExists(ctx, log, loadBalancer)
}

func (r *LoadBalancerReconciler) reconcileExists(ctx context.Context, log logr.Logger, loadBalancer *networkingv1alpha1.LoadBalancer) (ctrl.Result, error) {
	if !loadBalancer.DeletionTimestamp.IsZero() {
		return r.delete(ctx, log, loadBalancer)
	}
	return r.reconcile(ctx, log, loadBalancer)
}

func (r *LoadBalancerReconciler) delete(ctx context.Context, log logr.Logger, loadBalancer *networkingv1alpha1.LoadBalancer) (ctrl.Result, error) {
	return ctrl.Result{}, nil
}

func (r *LoadBalancerReconciler) reconcile(ctx context.Context, log logr.Logger, loadBalancer *networkingv1alpha1.LoadBalancer) (ctrl.Result, error) {
	log.V(1).Info("Reconcile")

	nicSelector := loadBalancer.Spec.NetworkInterfaceSelector
	if nicSelector != nil {
		log.V(1).Info("Network interface selector is present, managing routing")
		log.V(1).Info("Finding destinations")
		destinations, err := r.findDestinations(ctx, log, loadBalancer)
		if err != nil {
			return ctrl.Result{}, fmt.Errorf("error finding destinations: %w", err)
		}
		log.V(1).Info("Successfully found destinations", "Destinations", destinations)

		log.V(1).Info("Finding network", "Network", loadBalancer.Spec.NetworkRef.Name)
		network, err := r.getNetwork(ctx, log, loadBalancer)
		if err != nil {
			return ctrl.Result{}, fmt.Errorf("error getting network %s: %w", loadBalancer.Spec.NetworkRef.Name, err)
		}
		log.V(1).Info("Successfully found nework", "Network", network.Name)

		log.V(1).Info("Applying routing")
		if err := r.applyRouting(ctx, loadBalancer, destinations, network); err != nil {
			return ctrl.Result{}, fmt.Errorf("error applying routing: %w", err)
		}
		log.V(1).Info("Successfully applied routing")
	}

	var ips []commonv1alpha1.IP
	_, exists := loadBalancer.Labels[machinebroker.ManagerLabel]
	if loadBalancer.Spec.Type == networkingv1alpha1.LoadBalancerTypeInternal && !exists {
		var err error
		ips, err = r.applyInternalIPs(ctx, log, loadBalancer)
		if err != nil {
			return ctrl.Result{}, fmt.Errorf("error getting / applying internal ip: %w", err)
		}

		if err := r.updateStatus(ctx, log, loadBalancer, ips); err != nil {
			return ctrl.Result{}, fmt.Errorf("error patching load balancer status: %w", err)
		}
		log.V(1).Info("Updated load balancer status")
	}

	return ctrl.Result{}, nil
}

func (r *LoadBalancerReconciler) applyInternalIPs(ctx context.Context, log logr.Logger, loadBalancer *networkingv1alpha1.LoadBalancer) ([]commonv1alpha1.IP, error) {
	var ips []commonv1alpha1.IP
	for idx, ipSource := range loadBalancer.Spec.IPs {
		loadBalancerInternalIP, err := r.applyInternalIP(ctx, log, loadBalancer, ipSource, idx)
		if err != nil {
			return nil, fmt.Errorf("[ip %d] %w", idx, err)
		}
		ips = append(ips, loadBalancerInternalIP)
	}
	return ips, nil
}

func (r *LoadBalancerReconciler) applyInternalIP(ctx context.Context, log logr.Logger, loadBalancer *networkingv1alpha1.LoadBalancer, ipSource networkingv1alpha1.IPSource, idx int) (commonv1alpha1.IP, error) {
	switch {
	case ipSource.Value != nil:
		return *ipSource.Value, nil
	case ipSource.Ephemeral != nil:
		template := ipSource.Ephemeral.PrefixTemplate
		prefix := &ipamv1alpha1.Prefix{
			ObjectMeta: metav1.ObjectMeta{
				Namespace: loadBalancer.Namespace,
				Name:      networkingv1alpha1.LoadBalancerIPIPAMPrefixName(loadBalancer.Name, idx),
			},
		}
		log.V(1).Info("Applying prefix for load balancer", "ipFamily", template.Spec.IPFamily)
		if err := onmetalapiclient.ControlledCreateOrGet(ctx, r.Client, loadBalancer, prefix, func() error {
			prefix.Labels = template.Labels
			prefix.Annotations = template.Annotations
			prefix.Spec = template.Spec
			return nil
		}); err != nil {
			if !errors.Is(err, onmetalapiclient.ErrNotControlled) {
				return commonv1alpha1.IP{}, fmt.Errorf("error managing ephemeral prefix %s: %w", prefix.Name, err)
			}
			return commonv1alpha1.IP{}, fmt.Errorf("prefix %s cannot be managed", prefix.Name)
		}

		if prefix.Status.Phase != ipamv1alpha1.PrefixPhaseAllocated {
			return commonv1alpha1.IP{}, fmt.Errorf("prefix %s is not in state %s but %s", prefix.Name, ipamv1alpha1.PrefixPhaseAllocated, prefix.Status.Phase)
		}

		return prefix.Spec.Prefix.IP(), nil
	default:
		return commonv1alpha1.IP{}, fmt.Errorf("unknown ip source %#v", ipSource)
	}
}

func (r *LoadBalancerReconciler) updateStatus(ctx context.Context, log logr.Logger, loadBalancer *networkingv1alpha1.LoadBalancer, ips []commonv1alpha1.IP) error {
	base := loadBalancer.DeepCopy()
	loadBalancer.Status.IPs = ips
	return r.Status().Patch(ctx, loadBalancer, client.MergeFrom(base))
}

func (r *LoadBalancerReconciler) findDestinations(ctx context.Context, log logr.Logger, loadBalancer *networkingv1alpha1.LoadBalancer) ([]commonv1alpha1.LocalUIDReference, error) {
	sel, err := metav1.LabelSelectorAsSelector(loadBalancer.Spec.NetworkInterfaceSelector)
	if err != nil {
		return nil, err
	}

	nicList := &networkingv1alpha1.NetworkInterfaceList{}
	if err := r.List(ctx, nicList,
		client.InNamespace(loadBalancer.Namespace),
		client.MatchingLabelsSelector{Selector: sel},
		client.MatchingFields{networking.NetworkInterfaceSpecNetworkRefNameField: loadBalancer.Spec.NetworkRef.Name},
	); err != nil {
		return nil, fmt.Errorf("error listing network interfaces: %w", err)
	}

	destinations := make([]commonv1alpha1.LocalUIDReference, 0, len(nicList.Items))
	for _, nic := range nicList.Items {
		destinations = append(destinations, commonv1alpha1.LocalUIDReference{Name: nic.Name, UID: nic.UID})
	}
	return destinations, nil
}

func (r *LoadBalancerReconciler) getNetwork(ctx context.Context, log logr.Logger, loadBalancer *networkingv1alpha1.LoadBalancer) (*networkingv1alpha1.Network, error) {
	network := &networkingv1alpha1.Network{}
	if err := r.Get(ctx, types.NamespacedName{Name: loadBalancer.Spec.NetworkRef.Name, Namespace: loadBalancer.Namespace}, network); err != nil {
		return nil, fmt.Errorf("error getting network %s: %w", loadBalancer.Spec.NetworkRef.Name, err)
	}
	return network, nil
}

func (r *LoadBalancerReconciler) applyRouting(ctx context.Context, loadBalancer *networkingv1alpha1.LoadBalancer, destinations []commonv1alpha1.LocalUIDReference, network *networkingv1alpha1.Network) error {
	loadBalancerRouting := &networkingv1alpha1.LoadBalancerRouting{
		TypeMeta: metav1.TypeMeta{
			Kind:       "LoadBalancerRouting",
			APIVersion: networkingv1alpha1.SchemeGroupVersion.String(),
		},
		ObjectMeta: metav1.ObjectMeta{
			Namespace: loadBalancer.Namespace,
			Name:      loadBalancer.Name,
		},
		Destinations: destinations,
		NetworkRef: commonv1alpha1.LocalUIDReference{
			Name: network.Name,
			UID:  network.UID,
		},
	}
	if err := ctrl.SetControllerReference(loadBalancer, loadBalancerRouting, r.Scheme); err != nil {
		return fmt.Errorf("error setting controller reference: %w", err)
	}
	if err := r.Patch(ctx, loadBalancerRouting, client.Apply, loadBalancerFieldOwner, client.ForceOwnership); err != nil {
		return fmt.Errorf("error applying loadbalancer routing: %w", err)
	}
	return nil
}

func (r *LoadBalancerReconciler) SetupWithManager(mgr ctrl.Manager) error {
	ctx := context.Background()
	log := ctrl.Log.WithName("loadbalancer").WithName("setup")

	return ctrl.NewControllerManagedBy(mgr).
		For(&networkingv1alpha1.LoadBalancer{}).
		Owns(&networkingv1alpha1.LoadBalancerRouting{}).
		Watches(
			&source.Kind{Type: &networkingv1alpha1.NetworkInterface{}},
			r.enqueueByLoadBalancerMatchingNetworkInterface(ctx, log),
		).
		Complete(r)
}

func (r *LoadBalancerReconciler) enqueueByLoadBalancerMatchingNetworkInterface(ctx context.Context, log logr.Logger) handler.EventHandler {
	return handler.EnqueueRequestsFromMapFunc(func(obj client.Object) []ctrl.Request {
		nic := obj.(*networkingv1alpha1.NetworkInterface)
		log = log.WithValues("NetworkInterfaceKey", client.ObjectKeyFromObject(nic))

		loadBalancerList := &networkingv1alpha1.LoadBalancerList{}
		if err := r.List(ctx, loadBalancerList,
			client.InNamespace(nic.Namespace),
			client.MatchingFields{networking.LoadBalancerNetworkNameField: nic.Spec.NetworkRef.Name},
		); err != nil {
			log.Error(err, "Error listing loadbalancers for network")
			return nil
		}

		var res []ctrl.Request
		for _, loadBalancer := range loadBalancerList.Items {
			loadBalancerKey := client.ObjectKeyFromObject(&loadBalancer)
			log := log.WithValues("LoadBalancerKey", loadBalancerKey)
			nicSelector := loadBalancer.Spec.NetworkInterfaceSelector
			if nicSelector == nil {
				return nil
			}

			sel, err := metav1.LabelSelectorAsSelector(nicSelector)
			if err != nil {
				log.Error(err, "Invalid network interface selector")
				continue
			}

			if sel.Matches(labels.Set(nic.Labels)) {
				res = append(res, ctrl.Request{NamespacedName: loadBalancerKey})
			}
		}
		return res
	})
}
