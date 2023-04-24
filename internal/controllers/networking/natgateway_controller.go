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

	"github.com/bits-and-blooms/bitset"
	"github.com/go-logr/logr"
	commonv1alpha1 "github.com/onmetal/onmetal-api/api/common/v1alpha1"
	networkingv1alpha1 "github.com/onmetal/onmetal-api/api/networking/v1alpha1"
	"github.com/onmetal/onmetal-api/internal/client/networking"
	"github.com/onmetal/onmetal-api/utils/generic"
	"golang.org/x/exp/maps"
	"golang.org/x/exp/slices"
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/labels"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apimachinery/pkg/types"
	"k8s.io/apimachinery/pkg/util/sets"
	"k8s.io/client-go/tools/record"
	"k8s.io/utils/pointer"
	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/builder"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/handler"
	"sigs.k8s.io/controller-runtime/pkg/predicate"
	"sigs.k8s.io/controller-runtime/pkg/source"
)

const (
	MinEphemeralPort   int32 = 1024
	MaxEphemeralPort   int32 = 65535
	NoOfEphemeralPorts       = MaxEphemeralPort + 1 - MinEphemeralPort
)

var (
	natGatewayFieldOwner = client.FieldOwner(networkingv1alpha1.Resource("natgateways").String())
)

type NATGatewayReconciler struct {
	record.EventRecorder
	client.Client
	Scheme *runtime.Scheme
}

//+kubebuilder:rbac:groups="",resources=events,verbs=create;patch
//+kubebuilder:rbac:groups=networking.api.onmetal.de,resources=natgateways,verbs=get;list;watch;create;update;patch;delete
//+kubebuilder:rbac:groups=networking.api.onmetal.de,resources=natgateways/status,verbs=get;update;patch
//+kubebuilder:rbac:groups=networking.api.onmetal.de,resources=natgateways/finalizers,verbs=update
//+kubebuilder:rbac:groups=networking.api.onmetal.de,resources=networkinterfaces,verbs=get;list;watch
//+kubebuilder:rbac:groups=networking.api.onmetal.de,resources=natgatewayroutings,verbs=get;list;watch;create;update;patch;delete

func (r *NATGatewayReconciler) Reconcile(ctx context.Context, req ctrl.Request) (ctrl.Result, error) {
	log := ctrl.LoggerFrom(ctx)
	natGateway := &networkingv1alpha1.NATGateway{}
	if err := r.Get(ctx, req.NamespacedName, natGateway); err != nil {
		return ctrl.Result{}, client.IgnoreNotFound(err)
	}

	return r.reconcileExists(ctx, log, natGateway)
}

func (r *NATGatewayReconciler) reconcileExists(ctx context.Context, log logr.Logger, natGateway *networkingv1alpha1.NATGateway) (ctrl.Result, error) {
	if !natGateway.DeletionTimestamp.IsZero() {
		return r.delete(ctx, log, natGateway)
	}
	return r.reconcile(ctx, log, natGateway)
}

func (r *NATGatewayReconciler) delete(ctx context.Context, log logr.Logger, natGateway *networkingv1alpha1.NATGateway) (ctrl.Result, error) {
	return ctrl.Result{}, nil
}

type slots struct {
	numOfIPs             uint
	slotsPerIP           uint
	used                 uint
	slotsByIPFamilyAndIP map[corev1.IPFamily]map[commonv1alpha1.IP]*bitset.BitSet
}

func newSlots(slotsPerIP uint, ips []commonv1alpha1.IP) *slots {
	var l uint
	slotsByIPFamilyAndIP := make(map[corev1.IPFamily]map[commonv1alpha1.IP]*bitset.BitSet)
	for _, ip := range ips {
		slotsByIP := slotsByIPFamilyAndIP[ip.Family()]
		if slotsByIP == nil {
			slotsByIP = make(map[commonv1alpha1.IP]*bitset.BitSet)
			slotsByIPFamilyAndIP[ip.Family()] = slotsByIP
		}

		if _, ok := slotsByIP[ip]; ok {
			// don't re-initialize on duplicate ips
			continue
		}
		slotsByIP[ip] = bitset.New(slotsPerIP)
		l++
	}

	return &slots{
		numOfIPs:             l,
		slotsPerIP:           slotsPerIP,
		slotsByIPFamilyAndIP: slotsByIPFamilyAndIP,
	}
}

func (s *slots) Total() uint {
	return s.numOfIPs * s.slotsPerIP
}

func (s *slots) Used() uint {
	return s.used
}

func (s *slots) Use(ip commonv1alpha1.IP, slot uint) bool {
	slotsByIP := s.slotsByIPFamilyAndIP[ip.Family()]
	slots := slotsByIP[ip]
	if slots == nil || slot >= slots.Len() || slots.Test(slot) {
		return false
	}

	slots.Set(slot)
	s.used++
	if slots.All() {
		delete(slotsByIP, ip)
		if len(slotsByIP) == 0 {
			delete(s.slotsByIPFamilyAndIP, ip.Family())
		}
	}
	return true
}

func (s *slots) TryUseNextFree(ipFamily corev1.IPFamily) (ip commonv1alpha1.IP, slot uint, ok bool) {
	slotsByIP := s.slotsByIPFamilyAndIP[ipFamily]
	for ip, slots := range slotsByIP {
		ip := ip
		slot, ok = slots.NextClear(0)
		if ok {
			s.Use(ip, slot)
			return ip, slot, true
		}
	}
	return commonv1alpha1.IP{}, 0, false
}

type slotPortMapper struct {
	portsPerNetworkInterface int32
}

func newSlotPortMapper(portsPerNetworkInterface int32) *slotPortMapper {
	return &slotPortMapper{portsPerNetworkInterface}
}

func (m *slotPortMapper) SlotsPerIP() uint {
	return uint(NoOfEphemeralPorts / m.portsPerNetworkInterface)
}

func (m *slotPortMapper) EndPort(port int32) int32 {
	return port + m.portsPerNetworkInterface - 1
}

func (m *slotPortMapper) SlotForPorts(port, endPort int32) (uint, bool) {
	if port < MinEphemeralPort || port >= endPort || endPort > MaxEphemeralPort {
		return 0, false
	}
	if m.EndPort(port) != endPort {
		return 0, false
	}
	return uint((port - MinEphemeralPort) / m.portsPerNetworkInterface), true
}

func (m *slotPortMapper) PortsForSlot(slot uint) (port, endPort int32) {
	port = int32(slot)*m.portsPerNetworkInterface + MinEphemeralPort
	endPort = m.EndPort(port)
	return port, endPort
}

func (r *NATGatewayReconciler) assignPorts(
	ctx context.Context,
	log logr.Logger,
	natGateway *networkingv1alpha1.NATGateway,
	targetNameByUID map[types.UID]string,
) ([]networkingv1alpha1.NATGatewayDestination, int32, error) {
	natGatewayRouting := &networkingv1alpha1.NATGatewayRouting{}
	natGatewayRoutingKey := client.ObjectKey{Namespace: natGateway.Namespace, Name: natGateway.Name}
	if err := r.Get(ctx, natGatewayRoutingKey, natGatewayRouting); client.IgnoreNotFound(err) != nil {
		return nil, 0, fmt.Errorf("unable to get natgateway routing %s: %w", natGatewayRoutingKey, err)
	}

	var (
		portsPerNetworkInterface = generic.Deref(natGateway.Spec.PortsPerNetworkInterface, networkingv1alpha1.DefaultPortsPerNetworkInterface)
		mapper                   = newSlotPortMapper(portsPerNetworkInterface)
	)

	ips := make([]commonv1alpha1.IP, 0, len(natGateway.Status.IPs))
	for _, ip := range natGateway.Status.IPs {
		ips = append(ips, ip.IP)
	}

	slots := newSlots(mapper.SlotsPerIP(), ips)

	var (
		destinationsByUID = make(map[types.UID]networkingv1alpha1.NATGatewayDestination)
	)

	log.V(2).Info("Determining destinations to re-use")
	for _, destination := range natGatewayRouting.Destinations {
		targetUID := destination.UID
		if targetName, found := targetNameByUID[targetUID]; found {
			log := log.WithValues("Destination", destination)
			log.V(2).Info("Found existing destination")
			var (
				dstIPs         []networkingv1alpha1.NATGatewayDestinationIP
				seenIPFamilies = sets.New[corev1.IPFamily]()
			)
			for _, dstIP := range destination.IPs {
				log := log.WithValues("DestinationIP", dstIP)
				ip := dstIP.IP

				if seenIPFamilies.Has(ip.Family()) {
					log.V(2).Info("Dropping duplicate ip-family ip", "IP", ip)
					continue
				}

				slot, ok := mapper.SlotForPorts(dstIP.Port, dstIP.EndPort)
				if !ok {
					log.V(2).Info("Dropping invalid destination ip")
					continue
				}

				seenIPFamilies.Insert(ip.Family())
				if !slots.Use(ip, slot) {
					log.V(2).Info("Dropping non-available ip", "IP", ip)
					continue
				}

				log.V(2).Info("Re-using ip")
				dstIPs = append(dstIPs, dstIP)
			}

			if len(dstIPs) > 0 {
				destinationsByUID[targetUID] = networkingv1alpha1.NATGatewayDestination{
					Name: targetName,
					UID:  targetUID,
					IPs:  dstIPs,
				}

				if len(dstIPs) == len(natGateway.Spec.IPFamilies) {
					log.V(2).Info("Re-using all previous destination ips", "IPs", dstIPs)
				} else {
					log.V(2).Info("Re-using some previous destination ips", "IPs", dstIPs)
				}
			} else {
				log.V(2).Info("Not re-using any previous destination ip")
			}
		}
	}

	log.V(2).Info("Calculating new required destinations")
	for targetUID, targetName := range targetNameByUID {
		log := log.WithValues("TargetUID", targetUID, "TargetName", targetName)
		destination := destinationsByUID[targetUID]
		destination.Name = targetName
		destination.UID = targetUID

		for _, ipFamily := range natGateway.Spec.IPFamilies {
			log := log.WithValues("IPFamily", ipFamily)
			if slices.ContainsFunc(destination.IPs, func(dstIP networkingv1alpha1.NATGatewayDestinationIP) bool {
				return dstIP.IP.Family() == ipFamily
			}) {
				log.V(2).Info("IP family already satisfied")
				continue
			}

			ip, slot, ok := slots.TryUseNextFree(ipFamily)
			if !ok {
				log.V(2).Info("No slots available for ip family")
				continue
			}

			port, endPort := mapper.PortsForSlot(slot)
			destination.IPs = append(destination.IPs, networkingv1alpha1.NATGatewayDestinationIP{
				IP:      ip,
				Port:    port,
				EndPort: endPort,
			})
		}

		if len(destination.IPs) > 0 {
			destinationsByUID[targetUID] = destination
			slices.SortFunc(destination.IPs, func(a, b networkingv1alpha1.NATGatewayDestinationIP) bool {
				return a.IP.Family() < b.IP.Family()
			})
		} else {
			log.V(2).Info("Could not create destination for target")
		}
	}

	res := maps.Values(destinationsByUID)
	slices.SortFunc(res, func(a, b networkingv1alpha1.NATGatewayDestination) bool {
		return a.UID < b.UID
	})

	portsUsed := int32(slots.Used()) * portsPerNetworkInterface
	return res, portsUsed, nil
}

func (r *NATGatewayReconciler) reconcile(ctx context.Context, log logr.Logger, natGateway *networkingv1alpha1.NATGateway) (ctrl.Result, error) {
	log.V(1).Info("Reconcile")

	nicSelector := natGateway.Spec.NetworkInterfaceSelector
	if nicSelector == nil {
		log.V(1).Info("Network interface selector is empty")
		return ctrl.Result{}, nil
	}
	log.V(1).Info("Network interface selector is present, managing routing")

	log.V(1).Info("Selecting targets")
	targetNameByUID, err := r.selectTargets(ctx, log, natGateway)
	if err != nil {
		return ctrl.Result{}, fmt.Errorf("error selecting targets: %w", err)
	}
	log.V(1).Info("Selected targets", "TargetNameByUID", targetNameByUID)

	log.V(1).Info("Assigning ports")
	updatedDestinations, portsInUse, err := r.assignPorts(ctx, log, natGateway, targetNameByUID)
	if err != nil {
		return ctrl.Result{}, fmt.Errorf("error assigning ports: %w", err)
	}

	log.V(1).Info("Applying routing")
	if err := r.applyRouting(ctx, natGateway, updatedDestinations); err != nil {
		return ctrl.Result{}, fmt.Errorf("error applying routing: %w", err)
	}
	log.V(1).Info("Successfully applied routing")

	if err := r.patchStatus(ctx, natGateway, portsInUse); err != nil {
		return ctrl.Result{}, fmt.Errorf("unable to patch status: %w", err)
	}

	return ctrl.Result{}, nil
}

func (r *NATGatewayReconciler) patchStatus(ctx context.Context, natGateway *networkingv1alpha1.NATGateway, portsInUse int32) error {
	natGatewayBase := natGateway.DeepCopy()
	natGateway.Status.PortsUsed = pointer.Int32(portsInUse)
	return r.Patch(ctx, natGateway, client.MergeFrom(natGatewayBase))
}

func (r *NATGatewayReconciler) selectTargets(ctx context.Context, log logr.Logger, natGateway *networkingv1alpha1.NATGateway) (map[types.UID]string, error) {
	sel, err := metav1.LabelSelectorAsSelector(natGateway.Spec.NetworkInterfaceSelector)
	if err != nil {
		return nil, err
	}

	nicList := &networkingv1alpha1.NetworkInterfaceList{}
	if err := r.List(ctx, nicList,
		client.InNamespace(natGateway.Namespace),
		client.MatchingLabelsSelector{Selector: sel},
		client.MatchingFields{networking.NetworkInterfaceSpecNetworkRefNameField: natGateway.Spec.NetworkRef.Name},
	); err != nil {
		return nil, fmt.Errorf("error listing network interfaces: %w", err)
	}

	networkInterfaceNameByUID := make(map[types.UID]string)
	for _, nic := range nicList.Items {
		if nic.Spec.VirtualIP != nil {
			log.V(1).Info("Ignored network interface because it is exposed through a VirtualIP", "NetworkInterfaceKey", client.ObjectKeyFromObject(&nic))
			continue
		}

		networkInterfaceNameByUID[nic.UID] = nic.Name
	}
	return networkInterfaceNameByUID, nil
}

func (r *NATGatewayReconciler) applyRouting(ctx context.Context, natGateway *networkingv1alpha1.NATGateway, destinations []networkingv1alpha1.NATGatewayDestination) error {
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

func (r *NATGatewayReconciler) SetupWithManager(mgr ctrl.Manager) error {
	ctx := context.Background()
	log := ctrl.Log.WithName("natgateway").WithName("setup")

	return ctrl.NewControllerManagedBy(mgr).
		For(&networkingv1alpha1.NATGateway{}).
		Owns(
			&networkingv1alpha1.NATGatewayRouting{},
			builder.WithPredicates(&predicate.ResourceVersionChangedPredicate{}),
		).
		Watches(
			&source.Kind{Type: &networkingv1alpha1.NetworkInterface{}},
			r.enqueueByNetworkInterface(ctx, log),
		).
		Complete(r)
}

func (r *NATGatewayReconciler) enqueueByNetworkInterface(ctx context.Context, log logr.Logger) handler.EventHandler {
	return handler.EnqueueRequestsFromMapFunc(func(obj client.Object) []ctrl.Request {
		nic := obj.(*networkingv1alpha1.NetworkInterface)
		log = log.WithValues("NetworkInterfaceKey", client.ObjectKeyFromObject(nic))

		natGatewayList := &networkingv1alpha1.NATGatewayList{}
		if err := r.List(ctx, natGatewayList,
			client.InNamespace(nic.Namespace),
			client.MatchingFields{networking.NATGatewayNetworkNameField: nic.Spec.NetworkRef.Name},
		); err != nil {
			log.Error(err, "Error listing NAT gateways for network")
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
