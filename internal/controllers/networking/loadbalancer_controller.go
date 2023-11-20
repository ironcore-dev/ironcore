// Copyright 2022 IronCore authors
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
	commonv1alpha1 "github.com/ironcore-dev/ironcore/api/common/v1alpha1"
	networkingv1alpha1 "github.com/ironcore-dev/ironcore/api/networking/v1alpha1"
	"github.com/ironcore-dev/ironcore/internal/client/networking"
	clientutils "github.com/ironcore-dev/ironcore/utils/client"
	apierrors "k8s.io/apimachinery/pkg/api/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/labels"
	"k8s.io/client-go/util/workqueue"
	"k8s.io/klog/v2"
	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/builder"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/event"
	"sigs.k8s.io/controller-runtime/pkg/handler"
	"sigs.k8s.io/controller-runtime/pkg/predicate"
)

var (
	loadBalancerFieldOwner = client.FieldOwner(networkingv1alpha1.Resource("loadbalancers").String())
)

type LoadBalancerReconciler struct {
	client.Client
}

//+kubebuilder:rbac:groups=networking.ironcore.dev,resources=loadbalancers,verbs=get;list;watch;create;update;patch;delete
//+kubebuilder:rbac:groups=networking.ironcore.dev,resources=loadbalancers/status,verbs=get;update;patch
//+kubebuilder:rbac:groups=networking.ironcore.dev,resources=loadbalancers/finalizers,verbs=update
//+kubebuilder:rbac:groups=networking.ironcore.dev,resources=networkinterfaces,verbs=get;list;watch
//+kubebuilder:rbac:groups=networking.ironcore.dev,resources=networks,verbs=get;list;watch
//+kubebuilder:rbac:groups=networking.ironcore.dev,resources=loadbalancerroutings,verbs=get;list;watch;create;update;patch;delete

func (r *LoadBalancerReconciler) Reconcile(ctx context.Context, req ctrl.Request) (ctrl.Result, error) {
	log := ctrl.LoggerFrom(ctx)
	loadBalancer := &networkingv1alpha1.LoadBalancer{}
	if err := r.Get(ctx, req.NamespacedName, loadBalancer); err != nil {
		if !apierrors.IsNotFound(err) {
			return ctrl.Result{}, err
		}
		return ctrl.Result{}, nil
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
	if nicSelector == nil {
		log.V(1).Info("No network interface selector present, assuming external process is managing routing")
		return ctrl.Result{}, nil
	}

	log.V(1).Info("Managing load balancer routing")

	networkName := loadBalancer.Spec.NetworkRef.Name
	log.V(1).Info("Getting network", "Network", networkName)
	network, err := r.getNetwork(ctx, loadBalancer)
	if err != nil {
		return ctrl.Result{}, err
	}
	if network == nil {
		log.V(1).Info("Network not ready", "Network", networkName)
		return ctrl.Result{}, nil
	}

	log.V(1).Info("Finding destinations")
	destinations, err := r.findDestinations(ctx, loadBalancer)
	if err != nil {
		return ctrl.Result{}, fmt.Errorf("error finding destinations: %w", err)
	}

	log.V(1).Info("Applying routing", "Destinations", destinations, "Network", klog.KObj(network))
	if err := r.applyRouting(ctx, loadBalancer, destinations, network); err != nil {
		return ctrl.Result{}, fmt.Errorf("error applying routing: %w", err)
	}

	log.V(1).Info("Reconciled")
	return ctrl.Result{}, nil
}

func (r *LoadBalancerReconciler) findDestinations(ctx context.Context, loadBalancer *networkingv1alpha1.LoadBalancer) ([]networkingv1alpha1.LoadBalancerDestination, error) {
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

	// Make slice non-nil so omitempty does not file.
	destinations := make([]networkingv1alpha1.LoadBalancerDestination, 0)
	for _, nic := range nicList.Items {
		if nic.Status.State != networkingv1alpha1.NetworkInterfaceStateAvailable {
			continue
		}

		for _, ip := range nic.Status.IPs {
			destinations = append(destinations, networkingv1alpha1.LoadBalancerDestination{
				IP: ip,
				TargetRef: &networkingv1alpha1.LoadBalancerTargetRef{
					UID:        nic.UID,
					Name:       nic.Name,
					ProviderID: nic.Spec.ProviderID,
				},
			})
		}
	}
	return destinations, nil
}

func (r *LoadBalancerReconciler) getNetwork(ctx context.Context, loadBalancer *networkingv1alpha1.LoadBalancer) (*networkingv1alpha1.Network, error) {
	network := &networkingv1alpha1.Network{}
	networkKey := client.ObjectKey{Namespace: loadBalancer.Namespace, Name: loadBalancer.Spec.NetworkRef.Name}
	if err := r.Get(ctx, networkKey, network); err != nil {
		if !apierrors.IsNotFound(err) {
			return nil, fmt.Errorf("error getting network %s: %w", networkKey.Name, err)
		}
		return nil, nil
	}
	return network, nil
}

func (r *LoadBalancerReconciler) applyRouting(
	ctx context.Context,
	loadBalancer *networkingv1alpha1.LoadBalancer,
	destinations []networkingv1alpha1.LoadBalancerDestination,
	network *networkingv1alpha1.Network,
) error {
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
	_ = ctrl.SetControllerReference(loadBalancer, loadBalancerRouting, r.Scheme())
	if err := r.Patch(ctx, loadBalancerRouting, client.Apply, loadBalancerFieldOwner, client.ForceOwnership); err != nil {
		return fmt.Errorf("error applying loadbalancer routing: %w", err)
	}
	return nil
}

func (r *LoadBalancerReconciler) enqueueByNetworkInterface() handler.EventHandler {
	getEnqueueFunc := func(ctx context.Context, nic *networkingv1alpha1.NetworkInterface) func(nics []*networkingv1alpha1.NetworkInterface, queue workqueue.RateLimitingInterface) {
		log := ctrl.LoggerFrom(ctx)
		loadBalancerList := &networkingv1alpha1.LoadBalancerList{}
		if err := r.List(ctx, loadBalancerList,
			client.InNamespace(nic.Namespace),
			client.MatchingFields{networking.LoadBalancerNetworkNameField: nic.Spec.NetworkRef.Name},
		); err != nil {
			log.Error(err, "Error listing network interfaces")
			return nil
		}

		return func(nics []*networkingv1alpha1.NetworkInterface, queue workqueue.RateLimitingInterface) {
			for _, loadBalancer := range loadBalancerList.Items {
				loadBalancerKey := client.ObjectKeyFromObject(&loadBalancer)
				log := log.WithValues("LoadBalancerKey", loadBalancerKey)
				nicSelector := loadBalancer.Spec.NetworkInterfaceSelector
				if nicSelector == nil {
					return
				}

				sel, err := metav1.LabelSelectorAsSelector(nicSelector)
				if err != nil {
					log.Error(err, "Invalid network interface selector")
					continue
				}

				for _, nic := range nics {
					if sel.Matches(labels.Set(nic.Labels)) {
						queue.Add(ctrl.Request{NamespacedName: loadBalancerKey})
						break
					}
				}
			}
		}
	}

	return handler.Funcs{
		CreateFunc: func(ctx context.Context, evt event.CreateEvent, queue workqueue.RateLimitingInterface) {
			nic := evt.Object.(*networkingv1alpha1.NetworkInterface)
			enqueueFunc := getEnqueueFunc(ctx, nic)
			if enqueueFunc != nil {
				enqueueFunc([]*networkingv1alpha1.NetworkInterface{nic}, queue)
			}
		},
		UpdateFunc: func(ctx context.Context, evt event.UpdateEvent, queue workqueue.RateLimitingInterface) {
			newNic := evt.ObjectNew.(*networkingv1alpha1.NetworkInterface)
			oldNic := evt.ObjectOld.(*networkingv1alpha1.NetworkInterface)
			enqueueFunc := getEnqueueFunc(ctx, newNic)
			if enqueueFunc != nil {
				enqueueFunc([]*networkingv1alpha1.NetworkInterface{newNic, oldNic}, queue)
			}
		},
		DeleteFunc: func(ctx context.Context, evt event.DeleteEvent, queue workqueue.RateLimitingInterface) {
			nic := evt.Object.(*networkingv1alpha1.NetworkInterface)
			enqueueFunc := getEnqueueFunc(ctx, nic)
			if enqueueFunc != nil {
				enqueueFunc([]*networkingv1alpha1.NetworkInterface{nic}, queue)
			}
		},
		GenericFunc: func(ctx context.Context, evt event.GenericEvent, queue workqueue.RateLimitingInterface) {
			nic := evt.Object.(*networkingv1alpha1.NetworkInterface)
			enqueueFunc := getEnqueueFunc(ctx, nic)
			if enqueueFunc != nil {
				enqueueFunc([]*networkingv1alpha1.NetworkInterface{nic}, queue)
			}
		},
	}
}

func (r *LoadBalancerReconciler) enqueueByNetwork() handler.EventHandler {
	return handler.EnqueueRequestsFromMapFunc(func(ctx context.Context, obj client.Object) []ctrl.Request {
		log := ctrl.LoggerFrom(ctx)
		network := obj.(*networkingv1alpha1.Network)

		loadBalancerList := &networkingv1alpha1.LoadBalancerList{}
		if err := r.List(ctx, loadBalancerList,
			client.InNamespace(network.Namespace),
			client.MatchingFields{networking.LoadBalancerNetworkNameField: network.Name},
		); err != nil {
			log.Error(err, "Error listing load balancers for network")
			return nil
		}

		return clientutils.ReconcileRequestsFromObjectStructSlice[*networkingv1alpha1.LoadBalancer](loadBalancerList.Items)
	})
}

func (r *LoadBalancerReconciler) networkStateChangedPredicate() predicate.Predicate {
	return predicate.Funcs{
		UpdateFunc: func(evt event.UpdateEvent) bool {
			oldNetwork := evt.ObjectOld.(*networkingv1alpha1.Network)
			newNetwork := evt.ObjectNew.(*networkingv1alpha1.Network)
			return oldNetwork.Status.State != newNetwork.Status.State
		},
	}
}

func (r *LoadBalancerReconciler) networkInterfaceAvailablePredicate() predicate.Predicate {
	isNetworkInterfaceAvailable := func(nic *networkingv1alpha1.NetworkInterface) bool {
		return nic.Status.State == networkingv1alpha1.NetworkInterfaceStateAvailable
	}
	return predicate.Funcs{
		CreateFunc: func(evt event.CreateEvent) bool {
			nic := evt.Object.(*networkingv1alpha1.NetworkInterface)
			return isNetworkInterfaceAvailable(nic)
		},
		UpdateFunc: func(evt event.UpdateEvent) bool {
			oldNic := evt.ObjectOld.(*networkingv1alpha1.NetworkInterface)
			newNic := evt.ObjectNew.(*networkingv1alpha1.NetworkInterface)
			return isNetworkInterfaceAvailable(oldNic) || isNetworkInterfaceAvailable(newNic)
		},
		DeleteFunc: func(evt event.DeleteEvent) bool {
			nic := evt.Object.(*networkingv1alpha1.NetworkInterface)
			return isNetworkInterfaceAvailable(nic)
		},
		GenericFunc: func(evt event.GenericEvent) bool {
			nic := evt.Object.(*networkingv1alpha1.NetworkInterface)
			return isNetworkInterfaceAvailable(nic)
		},
	}
}

func (r *LoadBalancerReconciler) SetupWithManager(mgr ctrl.Manager) error {
	return ctrl.NewControllerManagedBy(mgr).
		For(&networkingv1alpha1.LoadBalancer{}).
		Owns(&networkingv1alpha1.LoadBalancerRouting{}).
		Watches(
			&networkingv1alpha1.Network{},
			r.enqueueByNetwork(),
			builder.WithPredicates(r.networkStateChangedPredicate()),
		).
		Watches(
			&networkingv1alpha1.NetworkInterface{},
			r.enqueueByNetworkInterface(),
			builder.WithPredicates(r.networkInterfaceAvailablePredicate()),
		).
		Complete(r)
}
