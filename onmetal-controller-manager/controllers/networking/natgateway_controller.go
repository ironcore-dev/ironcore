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
	"math"

	"github.com/go-logr/logr"
	"github.com/onmetal/controller-utils/set"
	commonv1alpha1 "github.com/onmetal/onmetal-api/api/common/v1alpha1"
	networkingv1alpha1 "github.com/onmetal/onmetal-api/api/networking/v1alpha1"
	onmetalapiclient "github.com/onmetal/onmetal-api/onmetal-controller-manager/client"
	"github.com/onmetal/onmetal-api/onmetal-controller-manager/controllers/networking/events"
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/labels"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apimachinery/pkg/types"
	"k8s.io/client-go/tools/record"
	"k8s.io/utils/pointer"
	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/handler"
	"sigs.k8s.io/controller-runtime/pkg/source"
)

const (
	MinEphemeralPort int32 = 1024
	MaxEphemeralPort int32 = 65535
)

var (
	natGatewayFieldOwner = client.FieldOwner(networkingv1alpha1.Resource("natgateways").String())
)

type NatGatewayReconciler struct {
	record.EventRecorder
	client.Client
	Scheme *runtime.Scheme
}

//+kubebuilder:rbac:groups=networking.api.onmetal.de,resources=natgateways,verbs=get;list;watch;create;update;patch;delete
//+kubebuilder:rbac:groups=networking.api.onmetal.de,resources=natgateways/status,verbs=get;update;patch
//+kubebuilder:rbac:groups=networking.api.onmetal.de,resources=natgateways/finalizers,verbs=update
//+kubebuilder:rbac:groups=networking.api.onmetal.de,resources=networkinterfaces,verbs=get;list;watch
//+kubebuilder:rbac:groups=networking.api.onmetal.de,resources=natgatewayroutings,verbs=get;list;watch;create;update;patch;delete

func (r *NatGatewayReconciler) Reconcile(ctx context.Context, req ctrl.Request) (ctrl.Result, error) {
	log := ctrl.LoggerFrom(ctx)
	natGateway := &networkingv1alpha1.NATGateway{}
	if err := r.Get(ctx, req.NamespacedName, natGateway); err != nil {
		return ctrl.Result{}, client.IgnoreNotFound(err)
	}

	return r.reconcileExists(ctx, log, natGateway)
}

func (r *NatGatewayReconciler) reconcileExists(ctx context.Context, log logr.Logger, natGateway *networkingv1alpha1.NATGateway) (ctrl.Result, error) {
	if !natGateway.DeletionTimestamp.IsZero() {
		return r.delete(ctx, log, natGateway)
	}
	return r.reconcile(ctx, log, natGateway)
}

func (r *NatGatewayReconciler) delete(ctx context.Context, log logr.Logger, natGateway *networkingv1alpha1.NATGateway) (ctrl.Result, error) {
	return ctrl.Result{}, nil
}

func (r *NatGatewayReconciler) assignPorts(ctx context.Context, log logr.Logger, natGateway *networkingv1alpha1.NATGateway, destinations map[types.UID]*networkingv1alpha1.NATGatewayDestination, slotsPerIP int, portsPerNetworkInterface int32) ([]networkingv1alpha1.NATGatewayDestination, int32, error) {
	natGatewayRouting := &networkingv1alpha1.NATGatewayRouting{}
	natGatewayRoutingKey := client.ObjectKey{Namespace: natGateway.Namespace, Name: natGateway.Name}
	if err := r.Get(ctx, natGatewayRoutingKey, natGatewayRouting); client.IgnoreNotFound(err) != nil {
		return nil, 0, fmt.Errorf("unable to get natgateway routing %s: %w", natGatewayRoutingKey, err)
	}

	used := set.Set[networkingv1alpha1.NATGatewayDestinationIP]{}
	//reuse ip/port combinations if still needed
	for _, destination := range natGatewayRouting.Destinations {
		if v, found := destinations[destination.UID]; found {
			v.IPs = destination.IPs
			for _, ip := range v.IPs {
				used.Insert(ip)
			}
		}
	}

	var (
		currentIp, currentPort int32 = 0, 0
		newDestinations        []networkingv1alpha1.NATGatewayDestination
	)
	for _, destination := range destinations {
		log.V(2).Info("Start processing destination", "destination", destination.Name)
		if len(destination.IPs) != 0 {
			log.V(2).Info("keep previous port assignment of destination", "destination", destination.Name)
			newDestinations = append(newDestinations, *destination)
			continue
		}

		log.V(2).Info("find free port for destination", "destination", destination.Name)
		var candidate *networkingv1alpha1.NATGatewayDestinationIP
		for {
			if int(currentIp) >= len(natGateway.Status.IPs) {
				log.V(1).Info("No ports left: Skipping destination", "name", destination.Name, "UID", destination.UID)
				r.Eventf(natGateway, corev1.EventTypeWarning, events.ErrorNoPortsLeft, "no ports left: skipping destination %s %s", destination.Name, destination.UID, "neededIps", int(math.Ceil(float64(len(destinations))/float64(slotsPerIP))))
				candidate = nil
				break
			}

			port := currentPort*portsPerNetworkInterface + MinEphemeralPort
			candidate = &networkingv1alpha1.NATGatewayDestinationIP{
				IP:      natGateway.Status.IPs[currentIp].IP,
				Port:    port,
				EndPort: port + portsPerNetworkInterface - 1,
			}

			if !used.Has(*candidate) {
				used.Insert(*candidate)
				log.V(2).Info("found port range for destination", "destination", destination.Name, "start", candidate.Port, "end", candidate.EndPort, "ip", candidate.IP.String())
				break
			}

			currentPort++
			if int(currentPort) > slotsPerIP {
				currentPort = 0
				currentIp++
			}
		}

		if candidate != nil {
			destination.IPs = []networkingv1alpha1.NATGatewayDestinationIP{*candidate}
		}
		newDestinations = append(newDestinations, *destination)
	}

	return newDestinations, int32(used.Len()), nil
}

func (r *NatGatewayReconciler) reconcile(ctx context.Context, log logr.Logger, natGateway *networkingv1alpha1.NATGateway) (ctrl.Result, error) {
	log.V(1).Info("Reconcile")

	nicSelector := natGateway.Spec.NetworkInterfaceSelector
	if nicSelector == nil {
		log.V(1).Info("Network interface selector is empty")
		return ctrl.Result{}, nil
	}
	log.V(1).Info("Network interface selector is present, managing routing")

	if natGateway.Spec.PortsPerNetworkInterface == nil {
		log.V(1).Info("PortsPerNetworkInterface is empty")
		return ctrl.Result{}, nil
	}
	log.V(1).Info("PortsPerNetworkInterface is present")
	portsPerNetworkInterface := *natGateway.Spec.PortsPerNetworkInterface

	log.V(1).Info("Finding destinations")
	destinations, err := r.findDestinations(ctx, log, natGateway)
	if err != nil {
		return ctrl.Result{}, fmt.Errorf("error finding destinations: %w", err)
	}
	log.V(1).Info("Successfully found destinations", "Destinations", destinations)

	slotsPerIP := int((MaxEphemeralPort - MinEphemeralPort) / portsPerNetworkInterface)

	log.V(1).Info("Assigning ports")
	updatedDestinations, slotsInUse, err := r.assignPorts(ctx, log, natGateway, destinations, slotsPerIP, portsPerNetworkInterface)
	if err != nil {
		return ctrl.Result{}, fmt.Errorf("error assigning ports: %w", err)
	}

	log.V(1).Info("Applying routing")
	if err := r.applyRouting(ctx, natGateway, updatedDestinations); err != nil {
		return ctrl.Result{}, fmt.Errorf("error applying routing: %w", err)
	}
	log.V(1).Info("Successfully applied routing")

	if err := r.patchStatus(ctx, natGateway, slotsInUse*portsPerNetworkInterface); err != nil {
		return ctrl.Result{}, fmt.Errorf("unable to patch status: %w", err)
	}

	return ctrl.Result{}, nil
}

func (r *NatGatewayReconciler) patchStatus(ctx context.Context, natGateway *networkingv1alpha1.NATGateway, portsInUse int32) error {
	natGatewayBase := natGateway.DeepCopy()
	natGateway.Status.PortsUsed = pointer.Int32(portsInUse)
	return r.Patch(ctx, natGateway, client.MergeFrom(natGatewayBase))
}

func (r *NatGatewayReconciler) findDestinations(ctx context.Context, log logr.Logger, natGateway *networkingv1alpha1.NATGateway) (map[types.UID]*networkingv1alpha1.NATGatewayDestination, error) {
	sel, err := metav1.LabelSelectorAsSelector(natGateway.Spec.NetworkInterfaceSelector)
	if err != nil {
		return nil, err
	}

	nicList := &networkingv1alpha1.NetworkInterfaceList{}
	if err := r.List(ctx, nicList,
		client.InNamespace(natGateway.Namespace),
		client.MatchingLabelsSelector{Selector: sel},
		client.MatchingFields{onmetalapiclient.NetworkInterfaceNetworkNameField: natGateway.Spec.NetworkRef.Name},
	); err != nil {
		return nil, fmt.Errorf("error listing network interfaces: %w", err)
	}

	destinations := map[types.UID]*networkingv1alpha1.NATGatewayDestination{}
	for _, nic := range nicList.Items {
		if nic.Spec.VirtualIP != nil {
			log.V(1).Info("Ignored network interface because it is exposed through a VirtualIP", "NetworkInterfaceKey", client.ObjectKeyFromObject(&nic))
			continue
		}

		destinations[nic.UID] = &networkingv1alpha1.NATGatewayDestination{
			Name: nic.Name,
			UID:  nic.UID,
		}
	}
	return destinations, nil
}

func (r *NatGatewayReconciler) applyRouting(ctx context.Context, natGateway *networkingv1alpha1.NATGateway, destinations []networkingv1alpha1.NATGatewayDestination) error {
	network := &networkingv1alpha1.Network{}
	if err := r.Get(ctx, types.NamespacedName{Name: natGateway.Spec.NetworkRef.Name, Namespace: natGateway.Namespace}, network); err != nil {
		return fmt.Errorf("error getting network %s: %w", natGateway.Spec.NetworkRef.Name, err)
	}

	natGatewayRouting := &networkingv1alpha1.NATGatewayRouting{
		TypeMeta: metav1.TypeMeta{
			Kind:       "NATGatewayRouting",
			APIVersion: networkingv1alpha1.SchemeGroupVersion.String(),
		},
		ObjectMeta: metav1.ObjectMeta{
			Namespace: natGateway.Namespace,
			Name:      natGateway.Name,
		},
		NetworkRef: commonv1alpha1.LocalUIDReference{
			Name: network.Name,
			UID:  network.UID,
		},
		Destinations: destinations,
	}
	if err := ctrl.SetControllerReference(natGateway, natGatewayRouting, r.Scheme); err != nil {
		return fmt.Errorf("error setting controller reference: %w", err)
	}
	if err := r.Patch(ctx, natGatewayRouting, client.Apply, natGatewayFieldOwner, client.ForceOwnership); err != nil {
		return fmt.Errorf("error applying natgateway routing: %w", err)
	}
	return nil
}

func (r *NatGatewayReconciler) SetupWithManager(mgr ctrl.Manager) error {
	ctx := context.Background()
	log := ctrl.Log.WithName("natgateway").WithName("setup")

	return ctrl.NewControllerManagedBy(mgr).
		For(&networkingv1alpha1.NATGateway{}).
		Owns(&networkingv1alpha1.NATGatewayRouting{}).
		Watches(
			&source.Kind{Type: &networkingv1alpha1.NetworkInterface{}},
			r.enqueueByNatGatewayMatchingNetworkInterface(ctx, log),
		).
		Complete(r)
}

func (r *NatGatewayReconciler) enqueueByNatGatewayMatchingNetworkInterface(ctx context.Context, log logr.Logger) handler.EventHandler {
	return handler.EnqueueRequestsFromMapFunc(func(obj client.Object) []ctrl.Request {
		nic := obj.(*networkingv1alpha1.NetworkInterface)
		log = log.WithValues("NetworkInterfaceKey", client.ObjectKeyFromObject(nic))

		natGatewayList := &networkingv1alpha1.NATGatewayList{}
		if err := r.List(ctx, natGatewayList,
			client.InNamespace(nic.Namespace),
			client.MatchingFields{onmetalapiclient.NATGatewayNetworkNameField: nic.Spec.NetworkRef.Name},
		); err != nil {
			log.Error(err, "Error listing natgateways for network")
			return nil
		}

		var res []ctrl.Request
		for _, natGateway := range natGatewayList.Items {
			natGatewayKey := client.ObjectKeyFromObject(&natGateway)
			log := log.WithValues("NATGatewayKey", natGatewayKey)
			nicSelector := natGateway.Spec.NetworkInterfaceSelector
			if nicSelector == nil {
				return nil
			}

			sel, err := metav1.LabelSelectorAsSelector(nicSelector)
			if err != nil {
				log.Error(err, "Invalid network interface selector")
				continue
			}

			if sel.Matches(labels.Set(nic.Labels)) {
				res = append(res, ctrl.Request{NamespacedName: natGatewayKey})
			}
		}
		return res
	})
}
