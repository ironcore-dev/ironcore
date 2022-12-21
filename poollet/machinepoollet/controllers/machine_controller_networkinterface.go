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

package controllers

import (
	"context"
	"fmt"

	"github.com/go-logr/logr"
	"github.com/gogo/protobuf/proto"
	"github.com/onmetal/controller-utils/set"
	commonv1alpha1 "github.com/onmetal/onmetal-api/api/common/v1alpha1"
	computev1alpha1 "github.com/onmetal/onmetal-api/api/compute/v1alpha1"
	networkingv1alpha1 "github.com/onmetal/onmetal-api/api/networking/v1alpha1"
	ori "github.com/onmetal/onmetal-api/ori/apis/machine/v1alpha1"
	orimeta "github.com/onmetal/onmetal-api/ori/apis/meta/v1alpha1"
	machinepoolletv1alpha1 "github.com/onmetal/onmetal-api/poollet/machinepoollet/api/v1alpha1"
	machinepoolletclient "github.com/onmetal/onmetal-api/poollet/machinepoollet/client"
	"github.com/onmetal/onmetal-api/poollet/machinepoollet/controllers/events"
	utilmaps "github.com/onmetal/onmetal-api/utils/maps"
	utilslices "github.com/onmetal/onmetal-api/utils/slices"
	"golang.org/x/exp/slices"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
	corev1 "k8s.io/api/core/v1"
	apierrors "k8s.io/apimachinery/pkg/api/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/types"
	"sigs.k8s.io/controller-runtime/pkg/client"
)

func (r *MachineReconciler) ipsToStrings(ips []commonv1alpha1.IP) []string {
	res := make([]string, len(ips))
	for i, ip := range ips {
		res[i] = ip.Addr.String()
	}
	return res
}

func (r *MachineReconciler) convertProtocol(protocol corev1.Protocol) (ori.Protocol, error) {
	switch protocol {
	case corev1.ProtocolTCP:
		return ori.Protocol_TCP, nil
	case corev1.ProtocolUDP:
		return ori.Protocol_UDP, nil
	case corev1.ProtocolSCTP:
		return ori.Protocol_SCTP, nil
	default:
		return 0, fmt.Errorf("unknown protocol %q", protocol)
	}
}

func (r *MachineReconciler) convertLoadBalancerPorts(ports []networkingv1alpha1.LoadBalancerPort) ([]*ori.LoadBalancerPort, error) {
	res := make([]*ori.LoadBalancerPort, len(ports))
	for i, port := range ports {
		var protocol corev1.Protocol
		if port.Protocol != nil {
			protocol = *port.Protocol
		} else {
			protocol = corev1.ProtocolTCP
		}

		var endPort int32
		if port.EndPort != nil {
			endPort = *port.EndPort
		} else {
			endPort = port.Port
		}

		oriProtocol, err := r.convertProtocol(protocol)
		if err != nil {
			return nil, err
		}

		res[i] = &ori.LoadBalancerPort{
			Protocol: oriProtocol,
			Port:     port.Port,
			EndPort:  endPort,
		}
	}
	return res, nil
}

func (r *MachineReconciler) listORINetworkInterfacesByMachineUID(ctx context.Context, machineUID types.UID) ([]*ori.NetworkInterface, error) {
	res, err := r.MachineRuntime.ListNetworkInterfaces(ctx, &ori.ListNetworkInterfacesRequest{
		Filter: &ori.NetworkInterfaceFilter{
			LabelSelector: map[string]string{
				machinepoolletv1alpha1.MachineUIDLabel: string(machineUID),
			},
		},
	})
	if err != nil {
		return nil, fmt.Errorf("error listing network interfaces: %w", err)
	}
	return res.NetworkInterfaces, nil
}

func (r *MachineReconciler) listORINetworkInterfacesByMachineKey(ctx context.Context, machineKey client.ObjectKey) ([]*ori.NetworkInterface, error) {
	res, err := r.MachineRuntime.ListNetworkInterfaces(ctx, &ori.ListNetworkInterfacesRequest{
		Filter: &ori.NetworkInterfaceFilter{
			LabelSelector: map[string]string{
				machinepoolletv1alpha1.MachineNamespaceLabel: machineKey.Namespace,
			},
		},
	})
	if err != nil {
		return nil, fmt.Errorf("error listing network interfaces: %w", err)
	}
	return res.NetworkInterfaces, nil
}

func (r *MachineReconciler) oriNetworkInterfaceLabels(machine *computev1alpha1.Machine, name string) map[string]string {
	lbls := r.oriMachineLabels(machine)
	lbls[machinepoolletv1alpha1.NetworkInterfaceNameLabel] = name
	return lbls
}

func (r *MachineReconciler) oriNetworkInterfaceFinder(networkInterfaces []*ori.NetworkInterface) func(name, networkHandle string) *ori.NetworkInterface {
	return func(name, networkHandle string) *ori.NetworkInterface {
		networkInterface, _ := utilslices.FindFunc(networkInterfaces, func(networkInterface *ori.NetworkInterface) bool {
			actualName := networkInterface.Metadata.Labels[machinepoolletv1alpha1.NetworkInterfaceNameLabel]
			return networkInterface.Metadata.DeletedAt == 0 &&
				actualName == name &&
				networkInterface.Spec.Network.Handle == networkHandle
		})
		return networkInterface
	}
}

func (r *MachineReconciler) oriNetworkInterfaceAttachmentFinder(networkInterfaces []*ori.NetworkInterfaceAttachment) func(name string) *ori.NetworkInterfaceAttachment {
	return func(name string) *ori.NetworkInterfaceAttachment {
		networkInterface, _ := utilslices.FindFunc(networkInterfaces, func(networkInterface *ori.NetworkInterfaceAttachment) bool {
			return networkInterface.Name == name
		})
		return networkInterface
	}
}

func (r *MachineReconciler) machineNetworkInterfaceFinder(networkInterfaces []computev1alpha1.NetworkInterface) func(name string) *computev1alpha1.NetworkInterface {
	return func(name string) *computev1alpha1.NetworkInterface {
		return utilslices.FindRefFunc(networkInterfaces, func(networkInterface computev1alpha1.NetworkInterface) bool {
			return networkInterface.Name == name
		})
	}
}

// isNetworkInterfaceBoundToMachine checks if the referenced network interface is bound to the machine.
func (r *MachineReconciler) isNetworkInterfaceBoundToMachine(machine *computev1alpha1.Machine, machineNetworkInterfaceName string, networkInterface *networkingv1alpha1.NetworkInterface) bool {
	if networkInterfacePhase := networkInterface.Status.Phase; networkInterfacePhase != networkingv1alpha1.NetworkInterfacePhaseBound {
		r.Eventf(machine, corev1.EventTypeNormal, events.NetworkInterfaceNotReady,
			"Network interface %s is in phase %s",
			networkInterface.Name,
			networkInterfacePhase,
		)
		return false
	}

	claimRef := networkInterface.Spec.MachineRef
	if claimRef == nil {
		r.Eventf(machine, corev1.EventTypeNormal, events.NetworkInterfaceNotReady,
			"Network interface %s does not reference any claimer",
			networkInterface.Name,
		)
		return false
	}

	if claimRef.Name != machine.Name || claimRef.UID != machine.UID {
		r.Eventf(machine, corev1.EventTypeNormal, events.NetworkInterfaceNotReady,
			"Network interface %s references a different claimer %s (uid %s)",
			networkInterface.Name,
			claimRef.Name,
			claimRef.UID,
		)
		return false
	}

	for _, networkInterfaceStatus := range machine.Status.NetworkInterfaces {
		if networkInterfaceStatus.Name == machineNetworkInterfaceName {
			if networkInterfaceStatus.Phase == computev1alpha1.NetworkInterfacePhaseBound {
				return true
			}

			r.Eventf(machine, corev1.EventTypeNormal, events.NetworkInterfaceNotReady,
				"Machine network interface status is in phase %s",
				networkInterfaceStatus.Phase,
			)
			return false
		}
	}
	r.Eventf(machine, corev1.EventTypeNormal, events.NetworkInterfaceNotReady,
		"Machine does not yet specify network interface status",
	)
	return false
}

func (r *MachineReconciler) createORINetworkInterface(
	ctx context.Context,
	machine *computev1alpha1.Machine,
	name string,
	spec *ori.NetworkInterfaceSpec,
) (*ori.NetworkInterface, error) {
	res, err := r.MachineRuntime.CreateNetworkInterface(ctx, &ori.CreateNetworkInterfaceRequest{
		NetworkInterface: &ori.NetworkInterface{
			Metadata: &orimeta.ObjectMetadata{
				Labels: r.oriNetworkInterfaceLabels(machine, name),
			},
			Spec: spec,
		},
	})
	if err != nil {
		return nil, err
	}
	return res.NetworkInterface, nil
}

func (r *MachineReconciler) prepareORINetworkInterfaceAttachmentAndSpec(
	ctx context.Context,
	machine *computev1alpha1.Machine,
	networkInterface computev1alpha1.NetworkInterface,
) (*ori.NetworkInterfaceAttachment, *ori.NetworkInterfaceSpec, bool, error) {
	name := networkInterface.Name
	switch {
	case networkInterface.NetworkInterfaceRef != nil || networkInterface.Ephemeral != nil:
		networkInterfaceName := computev1alpha1.MachineNetworkInterfaceName(machine.Name, networkInterface)

		oriNetworkInterfaceSpec, ok, err := r.prepareORINetworkInterfaceSpec(ctx, machine, name, networkInterfaceName)
		if err != nil || !ok {
			return nil, nil, ok, err
		}

		return &ori.NetworkInterfaceAttachment{
			Name: networkInterface.Name,
		}, oriNetworkInterfaceSpec, true, nil
	default:
		return nil, nil, false, fmt.Errorf("unrecognized network interface %#v", networkInterface)
	}
}

func (r *MachineReconciler) prepareORINetworkInterfaceAttachmentAndCreateNetworkInterfaceIfRequired(
	ctx context.Context,
	log logr.Logger,
	machine *computev1alpha1.Machine,
	networkInterface computev1alpha1.NetworkInterface,
	findORINetworkInterface func(name, networkHandle string) *ori.NetworkInterface,
) (networkInterfaceAttachment *ori.NetworkInterfaceAttachment, usedNetworkInterfaceIDs []string, ok bool, err error) {
	name := networkInterface.Name
	oriNetworkInterfaceAttachment, oriNetworkInterfaceSpec, ok, err := r.prepareORINetworkInterfaceAttachmentAndSpec(ctx, machine, networkInterface)
	if err != nil {
		return nil, nil, false, fmt.Errorf("error preparing ori network interface attachment: %w", err)
	}
	if !ok {
		return nil, nil, false, nil
	}

	if oriNetworkInterfaceSpec != nil {
		oriNetworkInterface := findORINetworkInterface(name, oriNetworkInterfaceSpec.Network.Handle)
		if oriNetworkInterface == nil {
			log.V(1).Info("Creating ori network interface")
			n, err := r.createORINetworkInterface(ctx, machine, name, oriNetworkInterfaceSpec)
			if err != nil {
				return nil, nil, false, fmt.Errorf("error creating ori network interface: %w", err)
			}

			oriNetworkInterface = n
		} else {
			if err := r.updateORINetworkInterface(ctx, log, oriNetworkInterface, oriNetworkInterfaceSpec); err != nil {
				usedNetworkInterfaceIDs = append(usedNetworkInterfaceIDs, oriNetworkInterface.Metadata.Id)
				return nil, usedNetworkInterfaceIDs, false, fmt.Errorf("error updating ori network interface: %w", err)
			}
		}

		oriNetworkInterfaceID := oriNetworkInterface.Metadata.Id
		usedNetworkInterfaceIDs = append(usedNetworkInterfaceIDs, oriNetworkInterfaceID)
		oriNetworkInterfaceAttachment.NetworkInterfaceId = oriNetworkInterfaceID
	}

	return oriNetworkInterfaceAttachment, usedNetworkInterfaceIDs, true, nil
}

func (r *MachineReconciler) prepareORINetworkInterfaceAttachments(
	ctx context.Context,
	log logr.Logger,
	machine *computev1alpha1.Machine,
) ([]*ori.NetworkInterfaceAttachment, bool, error) {
	log.V(1).Info("Listing ori network interfaces")
	oriNetworkInterfaces, err := r.listORINetworkInterfacesByMachineUID(ctx, machine.UID)
	if err != nil {
		return nil, false, fmt.Errorf("error listing ori network interfaces: %w", err)
	}

	var (
		findORINetworkInterface        = r.oriNetworkInterfaceFinder(oriNetworkInterfaces)
		oriNetworkInterfaceAttachments []*ori.NetworkInterfaceAttachment
		ok                             = true
		errs                           []error
		usedIDs                        = set.New[string]()
	)

	for _, networkInterface := range machine.Spec.NetworkInterfaces {
		name := networkInterface.Name
		oriNetworkInterfaceAttachment, networkInterfaceUsedORINetworkInterfaceIDs, networkInterfaceOK, err := r.prepareORINetworkInterfaceAttachmentAndCreateNetworkInterfaceIfRequired(ctx, log, machine, networkInterface, findORINetworkInterface)
		usedIDs.Insert(networkInterfaceUsedORINetworkInterfaceIDs...)
		if err != nil {
			errs = append(errs, fmt.Errorf("[network interface %s] %w", name, err))
			continue
		}
		if !networkInterfaceOK {
			ok = false
			continue
		}

		oriNetworkInterfaceAttachments = append(oriNetworkInterfaceAttachments, oriNetworkInterfaceAttachment)
	}

	for _, oriNetworkInterface := range oriNetworkInterfaces {
		oriNetworkInterfaceID := oriNetworkInterface.Metadata.Id
		if usedIDs.Has(oriNetworkInterfaceID) {
			continue
		}

		log := log.WithValues("ORINetworkInterfaceID", oriNetworkInterfaceID)
		log.V(1).Info("Deleting unused ori network interface")
		if _, err := r.MachineRuntime.DeleteNetworkInterface(ctx, &ori.DeleteNetworkInterfaceRequest{
			NetworkInterfaceId: oriNetworkInterfaceID,
		}); err != nil && status.Code(err) != codes.NotFound {
			log.Error(err, "Error deleting unused ori network interface")
		} else {
			log.V(1).Info("Deleted unused ori network interface")
		}
	}

	if len(errs) > 0 {
		return nil, false, fmt.Errorf("error(s) preparing network interface attachments: %v", errs)
	}
	if !ok {
		return nil, false, nil
	}
	return oriNetworkInterfaceAttachments, true, nil
}

func (r *MachineReconciler) updateORINetworkInterface(
	ctx context.Context,
	log logr.Logger,
	oriNetworkInterface *ori.NetworkInterface,
	oriNetworkInterfaceSpec *ori.NetworkInterfaceSpec,
) error {
	id := oriNetworkInterface.Metadata.Id
	log.V(1).Info("Found existing ori network interface", "ID", id)

	actualIPs := oriNetworkInterface.Spec.Ips
	desiredIPs := oriNetworkInterfaceSpec.Ips
	if !slices.Equal(actualIPs, desiredIPs) {
		log.V(1).Info("Updating ori network interface ips")
		if _, err := r.MachineRuntime.UpdateNetworkInterfaceIPs(ctx, &ori.UpdateNetworkInterfaceIPsRequest{
			NetworkInterfaceId: id,
			Ips:                desiredIPs,
		}); err != nil {
			return fmt.Errorf("error updating ori network interface ips: %w", err)
		}
	}

	actualVirtualIP := oriNetworkInterface.Spec.VirtualIp
	desiredVirtualIP := oriNetworkInterfaceSpec.VirtualIp
	switch {
	case actualVirtualIP == nil && desiredVirtualIP != nil:
		log.V(1).Info("Creating ori network interface virtual ip")
		if _, err := r.MachineRuntime.CreateNetworkInterfaceVirtualIP(ctx, &ori.CreateNetworkInterfaceVirtualIPRequest{
			NetworkInterfaceId: id,
			VirtualIp:          desiredVirtualIP,
		}); err != nil {
			return fmt.Errorf("error creating ori network interface virtual ip: %w", err)
		}
	case actualVirtualIP != nil && desiredVirtualIP != nil && !proto.Equal(actualVirtualIP, desiredVirtualIP):
		log.V(1).Info("Updating ori network interface virtual ip")
		if _, err := r.MachineRuntime.UpdateNetworkInterfaceVirtualIP(ctx, &ori.UpdateNetworkInterfaceVirtualIPRequest{
			NetworkInterfaceId: id,
			VirtualIp:          desiredVirtualIP,
		}); err != nil {
			return fmt.Errorf("error updating ori network interface virtual ip: %w", err)
		}
	case actualVirtualIP != nil && desiredVirtualIP == nil:
		log.V(1).Info("Deleting ori network interface virtual ip")
		if _, err := r.MachineRuntime.DeleteNetworkInterfaceVirtualIP(ctx, &ori.DeleteNetworkInterfaceVirtualIPRequest{
			NetworkInterfaceId: id,
		}); err != nil {
			return fmt.Errorf("error deleting ori network interface virtual ip: %w", err)
		}
	}

	if err := r.updateORINetworkInterfacePrefixes(ctx, log, oriNetworkInterface, oriNetworkInterfaceSpec); err != nil {
		return fmt.Errorf("error updating ori network interface prefixes: %w", err)
	}

	if err := r.updateORINetworkInterfaceLoadBalancerTargets(ctx, log, oriNetworkInterface, oriNetworkInterfaceSpec); err != nil {
		return fmt.Errorf("error updating ori network interface load balancer targets: %w", err)
	}

	return nil
}

func (r *MachineReconciler) updateORINetworkInterfaceLoadBalancerTargets(
	ctx context.Context,
	log logr.Logger,
	oriNetworkInterface *ori.NetworkInterface,
	oriNetworkInterfaceSpec *ori.NetworkInterfaceSpec,
) error {
	id := oriNetworkInterface.Metadata.Id
	actualLoadBalancerTargets := r.buildLoadBalancerTargetMap(oriNetworkInterface.Spec.LoadBalancerTargets)
	desiredLoadBalancerTargets := r.buildLoadBalancerTargetMap(oriNetworkInterfaceSpec.LoadBalancerTargets)

	delLoadBalancerTargets := utilmaps.KeysDifference(actualLoadBalancerTargets, desiredLoadBalancerTargets)
	newLoadBalancerTargets := utilmaps.KeysDifference(desiredLoadBalancerTargets, actualLoadBalancerTargets)

	var errs []error
	for key := range delLoadBalancerTargets {
		delTgt := actualLoadBalancerTargets[key]
		log.V(1).Info("Deleting outdated ori network interface load balancer target", "LoadBalancerTarget", key)
		if _, err := r.MachineRuntime.DeleteNetworkInterfaceLoadBalancerTarget(ctx, &ori.DeleteNetworkInterfaceLoadBalancerTargetRequest{
			NetworkInterfaceId: id,
			LoadBalancerTarget: delTgt,
		}); err != nil {
			errs = append(errs, fmt.Errorf("error deleting ori network interface load balancer target %s: %w", key, err))
		}
	}
	for key := range newLoadBalancerTargets {
		newTgt := desiredLoadBalancerTargets[key]
		log.V(1).Info("Creating new ori network interface load balancer target", "LoadBalancerTarget", key)
		if _, err := r.MachineRuntime.CreateNetworkInterfaceLoadBalancerTarget(ctx, &ori.CreateNetworkInterfaceLoadBalancerTargetRequest{
			NetworkInterfaceId: id,
			LoadBalancerTarget: newTgt,
		}); err != nil {
			errs = append(errs, fmt.Errorf("error creating ori network interface load balancer target %s: %w", key, err))
		}
	}

	if len(errs) > 0 {
		return fmt.Errorf("error(s) updating ori network interface load balancer targets: %v", errs)
	}
	return nil
}

func (r *MachineReconciler) updateORINetworkInterfacePrefixes(
	ctx context.Context,
	log logr.Logger,
	oriNetworkInterface *ori.NetworkInterface,
	oriNetworkInterfaceSpec *ori.NetworkInterfaceSpec,
) error {
	id := oriNetworkInterface.Metadata.Id
	actualPrefixes := set.New(oriNetworkInterface.Spec.Prefixes...)
	desiredPrefixes := set.New(oriNetworkInterfaceSpec.Prefixes...)

	delPrefixes := actualPrefixes.Difference(desiredPrefixes)
	newPrefixes := desiredPrefixes.Difference(actualPrefixes)

	var errs []error
	for delPrefix := range delPrefixes {
		log.V(1).Info("Deleting outdated ori network interface prefix", "Prefix", delPrefix)
		if _, err := r.MachineRuntime.DeleteNetworkInterfacePrefix(ctx, &ori.DeleteNetworkInterfacePrefixRequest{
			NetworkInterfaceId: id,
			Prefix:             delPrefix,
		}); err != nil {
			errs = append(errs, fmt.Errorf("error deleting ori network interface prefix %s: %w", delPrefix, err))
		}
	}
	for newPrefix := range newPrefixes {
		log.V(1).Info("Creating new ori network interface prefix", "Prefix", newPrefix)
		if _, err := r.MachineRuntime.CreateNetworkInterfacePrefix(ctx, &ori.CreateNetworkInterfacePrefixRequest{
			NetworkInterfaceId: id,
			Prefix:             newPrefix,
		}); err != nil {
			errs = append(errs, fmt.Errorf("error creating ori network interface prefix %s: %w", newPrefix, err))
		}
	}

	if len(errs) > 0 {
		return fmt.Errorf("error(s) updating ori network interface prefixes: %v", errs)
	}
	return nil
}

func (r *MachineReconciler) prepareORINetworkInterfaceSpec(
	ctx context.Context,
	machine *computev1alpha1.Machine,
	name, networkInterfaceName string,
) (*ori.NetworkInterfaceSpec, bool, error) {
	networkInterface := &networkingv1alpha1.NetworkInterface{}
	networkInterfaceKey := client.ObjectKey{Namespace: machine.Namespace, Name: networkInterfaceName}
	if err := r.Get(ctx, networkInterfaceKey, networkInterface); err != nil {
		if !apierrors.IsNotFound(err) {
			return nil, false, fmt.Errorf("error getting network interface: %w", err)
		}
		r.Eventf(machine, corev1.EventTypeNormal, events.NetworkInterfaceNotReady, "Network interface %s not found", networkInterfaceName)
		return nil, false, nil
	}

	if state := networkInterface.Status.State; state != networkingv1alpha1.NetworkInterfaceStateAvailable {
		r.Eventf(machine, corev1.EventTypeNormal, events.NetworkInterfaceNotReady, "Network interface %s is in state %s", networkInterfaceName, state)
		return nil, false, nil
	}

	if !r.isNetworkInterfaceBoundToMachine(machine, name, networkInterface) {
		return nil, false, nil
	}

	var virtualIPSpec *ori.VirtualIPSpec
	if virtualIP := networkInterface.Status.VirtualIP; virtualIP != nil {
		virtualIPSpec = &ori.VirtualIPSpec{
			Ip: virtualIP.String(),
		}
	}

	prefixes, err := r.aliasPrefixesForNetworkInterface(ctx, networkInterface)
	if err != nil {
		return nil, false, fmt.Errorf("error getting alias prefixes for network interface: %w", err)
	}

	loadBalancerTargets, err := r.loadBalancerTargetsForNetworkInterface(ctx, networkInterface)
	if err != nil {
		return nil, false, fmt.Errorf("error getting load balancer targets for network interface: %w", err)
	}

	return &ori.NetworkInterfaceSpec{
		Network: &ori.NetworkSpec{
			Handle: networkInterface.Status.NetworkHandle,
		},
		Ips:                 r.ipsToStrings(networkInterface.Status.IPs),
		VirtualIp:           virtualIPSpec,
		Prefixes:            prefixes,
		LoadBalancerTargets: loadBalancerTargets,
	}, true, nil
}

func (r *MachineReconciler) aliasPrefixesForNetworkInterface(
	ctx context.Context,
	networkInterface *networkingv1alpha1.NetworkInterface,
) ([]string, error) {
	aliasPrefixList := &networkingv1alpha1.AliasPrefixList{}
	if err := r.List(ctx, aliasPrefixList,
		client.InNamespace(networkInterface.Namespace),
	); err != nil {
		return nil, fmt.Errorf("error listing alias prefixes: %w", err)
	}

	aliasPrefixRoutingList := &networkingv1alpha1.AliasPrefixRoutingList{}
	if err := r.List(ctx, aliasPrefixRoutingList,
		client.InNamespace(networkInterface.Namespace),
		client.MatchingFields{
			machinepoolletclient.AliasPrefixRoutingNetworkRefNameField: networkInterface.Spec.NetworkRef.Name,
		},
	); err != nil {
		return nil, fmt.Errorf("error listing alias prefix routings: %w", err)
	}

	aliasPrefixRoutingsByName := utilslices.ToMap(
		aliasPrefixRoutingList.Items,
		func(apr networkingv1alpha1.AliasPrefixRouting) string { return apr.Name },
	)

	prefixes := set.New[string]()

	for _, aliasPrefix := range aliasPrefixList.Items {
		prefix := aliasPrefix.Status.Prefix
		if prefix == nil {
			continue
		}

		aliasPrefixRouting, ok := aliasPrefixRoutingsByName[aliasPrefix.Name]
		if !ok {
			continue
		}

		if slices.ContainsFunc(
			aliasPrefixRouting.Destinations,
			func(ref commonv1alpha1.LocalUIDReference) bool { return ref.UID == networkInterface.UID },
		) {
			prefixes.Insert(prefix.String())
		}
	}
	return set.SortedSlice(prefixes), nil
}

func (r *MachineReconciler) loadBalancerTargetsForNetworkInterface(
	ctx context.Context,
	networkInterface *networkingv1alpha1.NetworkInterface,
) ([]*ori.LoadBalancerTargetSpec, error) {
	loadBalancerList := &networkingv1alpha1.LoadBalancerList{}
	if err := r.List(ctx, loadBalancerList,
		client.InNamespace(networkInterface.Namespace),
	); err != nil {
		return nil, fmt.Errorf("error listing load balancers: %w", err)
	}

	loadBalancerRoutingList := &networkingv1alpha1.LoadBalancerRoutingList{}
	if err := r.List(ctx, loadBalancerRoutingList,
		client.InNamespace(networkInterface.Namespace),
		client.MatchingFields{
			machinepoolletclient.LoadBalancerRoutingNetworkRefNameField: networkInterface.Spec.NetworkRef.Name,
		},
	); err != nil {
		return nil, fmt.Errorf("error listing load balancer routings: %w", err)
	}

	loadBalancerRoutingsByName := utilslices.ToMap(
		loadBalancerRoutingList.Items,
		func(apr networkingv1alpha1.LoadBalancerRouting) string { return apr.Name },
	)

	// Construct a map by the key / hash code of the load balancer target to eliminate duplicates.
	loadBalancerTargetsByKey := make(map[string]*ori.LoadBalancerTargetSpec)
	for _, loadBalancer := range loadBalancerList.Items {
		loadBalancerRouting, ok := loadBalancerRoutingsByName[loadBalancer.Name]
		if !ok {
			continue
		}

		if len(loadBalancer.Status.IPs) == 0 {
			continue
		}

		if !slices.ContainsFunc(
			loadBalancerRouting.Destinations,
			func(r commonv1alpha1.LocalUIDReference) bool { return r.UID == networkInterface.UID },
		) {
			continue
		}

		for _, ip := range loadBalancer.Status.IPs {
			ports, err := r.convertLoadBalancerPorts(loadBalancer.Spec.Ports)
			if err != nil {
				return nil, err
			}

			loadBalancerTarget := &ori.LoadBalancerTargetSpec{
				Ip:    ip.String(),
				Ports: ports,
			}

			loadBalancerTargetsByKey[loadBalancerTarget.Key()] = loadBalancerTarget
		}
	}

	res := make([]*ori.LoadBalancerTargetSpec, 0, len(loadBalancerTargetsByKey))
	for _, loadBalancerTarget := range loadBalancerTargetsByKey {
		res = append(res, loadBalancerTarget)
	}
	return res, nil
}

func (r *MachineReconciler) buildLoadBalancerTargetMap(tgts []*ori.LoadBalancerTargetSpec) map[string]*ori.LoadBalancerTargetSpec {
	res := make(map[string]*ori.LoadBalancerTargetSpec)
	for _, tgt := range tgts {
		res[tgt.Key()] = tgt
	}
	return res
}

func (r *MachineReconciler) updateNetworkInterface(
	ctx context.Context,
	log logr.Logger,
	machine *computev1alpha1.Machine,
	networkInterface computev1alpha1.NetworkInterface,
	findORINetworkInterface func(name, networkHandle string) *ori.NetworkInterface,
	oriMachine *ori.Machine,
	existingORINetworkInterfaceAttachment *ori.NetworkInterfaceAttachment,
) (usedNetworkInterfaceIDs []string, err error) {
	name := networkInterface.Name
	machineID := oriMachine.Metadata.Id

	addExistingNetworkInterfaceIDIfPresent := func() {
		if existingORINetworkInterfaceAttachment != nil {
			if networkInterfaceID := existingORINetworkInterfaceAttachment.NetworkInterfaceId; networkInterfaceID != "" {
				usedNetworkInterfaceIDs = append(usedNetworkInterfaceIDs, networkInterfaceID)
			}
		}
	}

	oriNetworkInterfaceAttachment, oriNetworkInterfaceSpec, ok, err := r.prepareORINetworkInterfaceAttachmentAndSpec(ctx, machine, networkInterface)
	if err != nil {
		addExistingNetworkInterfaceIDIfPresent()
		return usedNetworkInterfaceIDs, fmt.Errorf("error preparing ori network interface attachment: %w", err)
	}
	if !ok {
		if existingORINetworkInterfaceAttachment != nil {
			log.V(1).Info("Deleting outdated ori network interface attachment")
			if _, err := r.MachineRuntime.DeleteNetworkInterfaceAttachment(ctx, &ori.DeleteNetworkInterfaceAttachmentRequest{
				MachineId: machineID,
				Name:      name,
			}); err != nil && status.Code(err) != codes.NotFound {
				addExistingNetworkInterfaceIDIfPresent()
				return usedNetworkInterfaceIDs, fmt.Errorf("error deleting outdated ori network interface attachment: %w", err)
			}
		}
		return nil, nil
	}

	if oriNetworkInterfaceSpec != nil {
		oriNetworkInterface := findORINetworkInterface(name, oriNetworkInterfaceSpec.Network.Handle)
		if oriNetworkInterface == nil {
			n, err := r.createORINetworkInterface(ctx, machine, name, oriNetworkInterfaceSpec)
			if err != nil {
				addExistingNetworkInterfaceIDIfPresent()
				return usedNetworkInterfaceIDs, fmt.Errorf("error creating ori network interface: %w", err)
			}

			oriNetworkInterface = n
		} else {
			if err := r.updateORINetworkInterface(ctx, log, oriNetworkInterface, oriNetworkInterfaceSpec); err != nil {
				addExistingNetworkInterfaceIDIfPresent()
				usedNetworkInterfaceIDs = append(usedNetworkInterfaceIDs, oriNetworkInterface.Metadata.Id)
				return usedNetworkInterfaceIDs, fmt.Errorf("error updating ori network interface: %w", err)
			}
		}

		oriNetworkInterfaceID := oriNetworkInterface.Metadata.Id
		usedNetworkInterfaceIDs = append(usedNetworkInterfaceIDs, oriNetworkInterfaceID)
		oriNetworkInterfaceAttachment.NetworkInterfaceId = oriNetworkInterfaceID
	}

	if existingORINetworkInterfaceAttachment != nil {
		if proto.Equal(existingORINetworkInterfaceAttachment, oriNetworkInterfaceAttachment) {
			log.V(1).Info("Existing ori network interface attachment is up-to-date")
			return usedNetworkInterfaceIDs, nil
		}

		log.V(1).Info("Existing ori network interface attachment is outdated, deleting")
		if _, err := r.MachineRuntime.DeleteNetworkInterfaceAttachment(ctx, &ori.DeleteNetworkInterfaceAttachmentRequest{
			MachineId: machineID,
			Name:      name,
		}); err != nil && status.Code(err) != codes.NotFound {
			addExistingNetworkInterfaceIDIfPresent()
			return usedNetworkInterfaceIDs, fmt.Errorf("error deleting outdated ori network interface attachment: %w", err)
		}
	}

	log.V(1).Info("Creating network interface attachment")
	if _, err := r.MachineRuntime.CreateNetworkInterfaceAttachment(ctx, &ori.CreateNetworkInterfaceAttachmentRequest{
		MachineId:        oriMachine.Metadata.Id,
		NetworkInterface: oriNetworkInterfaceAttachment,
	}); err != nil {
		return usedNetworkInterfaceIDs, fmt.Errorf("error creating network interface attachmetn: %w", err)
	}

	return usedNetworkInterfaceIDs, nil
}

func (r *MachineReconciler) updateORINetworkInterfaces(ctx context.Context, log logr.Logger, machine *computev1alpha1.Machine, oriMachine *ori.Machine) error {
	machineID := oriMachine.Metadata.Id

	log.V(1).Info("Listing ori network interfaces")
	oriNetworkInterfaces, err := r.listORINetworkInterfacesByMachineUID(ctx, machine.UID)
	if err != nil {
		return fmt.Errorf("error listing ori network interfaces: %w", err)
	}

	var (
		errs                              []error
		findNetworkInterface              = r.machineNetworkInterfaceFinder(machine.Spec.NetworkInterfaces)
		findORINetworkInterface           = r.oriNetworkInterfaceFinder(oriNetworkInterfaces)
		findORINetworkInterfaceAttachment = r.oriNetworkInterfaceAttachmentFinder(oriMachine.Spec.NetworkInterfaces)
		usedORINetworkInterfaceIDs        = set.New[string]()
	)

	for _, oriNetworkInterface := range oriMachine.Spec.NetworkInterfaces {
		if networkInterface := findNetworkInterface(oriNetworkInterface.Name); networkInterface != nil {
			continue
		}

		log.V(1).Info("Deleting outdated ori network interface attachment")
		if _, err := r.MachineRuntime.DeleteNetworkInterfaceAttachment(ctx, &ori.DeleteNetworkInterfaceAttachmentRequest{
			MachineId: machineID,
			Name:      oriNetworkInterface.Name,
		}); err != nil && status.Code(err) != codes.NotFound {
			if oriNetworkInterfaceID := oriNetworkInterface.NetworkInterfaceId; oriNetworkInterfaceID != "" {
				usedORINetworkInterfaceIDs.Insert(oriNetworkInterfaceID)
			}
			errs = append(errs, fmt.Errorf("[network interface %s] error deleting outdated ori network interface attachment: %w", oriNetworkInterface.Name, err))
		}
	}

	for _, networkInterface := range machine.Spec.NetworkInterfaces {
		name := networkInterface.Name
		existingORINetworkInterfaceAttachment := findORINetworkInterfaceAttachment(name)
		networkInterfaceUsedNetworkInterfaceIDs, err := r.updateNetworkInterface(ctx, log, machine, networkInterface, findORINetworkInterface, oriMachine, existingORINetworkInterfaceAttachment)
		usedORINetworkInterfaceIDs.Insert(networkInterfaceUsedNetworkInterfaceIDs...)
		if err != nil {
			errs = append(errs, fmt.Errorf("[network interface %s] %w", name, err))
			continue
		}
	}

	for _, oriNetworkInterface := range oriNetworkInterfaces {
		oriNetworkInterfaceID := oriNetworkInterface.Metadata.Id
		if usedORINetworkInterfaceIDs.Has(oriNetworkInterfaceID) {
			continue
		}

		log := log.WithValues("ORINetworkInterfaceID", oriNetworkInterfaceID)
		log.V(1).Info("Deleting unused ori network interface")
		if _, err := r.MachineRuntime.DeleteNetworkInterface(ctx, &ori.DeleteNetworkInterfaceRequest{
			NetworkInterfaceId: oriNetworkInterfaceID,
		}); err != nil && status.Code(err) != codes.NotFound {
			log.Error(err, "Error deleting unused ori network interface")
		} else {
			log.V(1).Info("Deleted unused ori network interface")
		}
	}

	if len(errs) > 0 {
		return fmt.Errorf("error(s) updating ori network interface(s): %v", errs)
	}
	return nil
}

var oriNetworkInterfaceAttachmentStateToNetworkInterfaceState = map[ori.NetworkInterfaceAttachmentState]computev1alpha1.NetworkInterfaceState{
	ori.NetworkInterfaceAttachmentState_NETWORK_INTERFACE_ATTACHMENT_PENDING:  computev1alpha1.NetworkInterfaceStatePending,
	ori.NetworkInterfaceAttachmentState_NETWORK_INTERFACE_ATTACHMENT_ATTACHED: computev1alpha1.NetworkInterfaceStateAttached,
	ori.NetworkInterfaceAttachmentState_NETWORK_INTERFACE_ATTACHMENT_DETACHED: computev1alpha1.NetworkInterfaceStateDetached,
}

func (r *MachineReconciler) convertORINetworkInterfaceAttachmentState(state ori.NetworkInterfaceAttachmentState) (computev1alpha1.NetworkInterfaceState, error) {
	if res, ok := oriNetworkInterfaceAttachmentStateToNetworkInterfaceState[state]; ok {
		return res, nil
	}
	return "", fmt.Errorf("unknown network interface attachment state %v", state)
}

func (r *MachineReconciler) convertORINetworkInterfaceAttachmentStatus(status *ori.NetworkInterfaceAttachmentStatus) (computev1alpha1.NetworkInterfaceStatus, error) {
	state, err := r.convertORINetworkInterfaceAttachmentState(status.State)
	if err != nil {
		return computev1alpha1.NetworkInterfaceStatus{}, err
	}

	return computev1alpha1.NetworkInterfaceStatus{
		Name:   status.Name,
		Handle: status.NetworkInterfaceHandle,
		State:  state,
	}, nil
}

func (r *MachineReconciler) updateNetworkInterfaceStates(machine *computev1alpha1.Machine, oriMachine *ori.Machine, now metav1.Time) error {
	seenNames := set.New[string]()
	for _, oriNetworkInterfaceAttachmentStatus := range oriMachine.Status.NetworkInterfaces {
		name := oriNetworkInterfaceAttachmentStatus.Name
		seenNames.Insert(name)
		newNetworkInterfaceStatus, err := r.convertORINetworkInterfaceAttachmentStatus(oriNetworkInterfaceAttachmentStatus)
		if err != nil {
			return fmt.Errorf("error converting ori network interface status %s: %w", name, err)
		}

		idx := slices.IndexFunc(
			machine.Status.NetworkInterfaces,
			func(status computev1alpha1.NetworkInterfaceStatus) bool { return status.Name == name },
		)
		if idx < 0 {
			newNetworkInterfaceStatus.LastStateTransitionTime = &now
			machine.Status.NetworkInterfaces = append(machine.Status.NetworkInterfaces, newNetworkInterfaceStatus)
		} else {
			networkInterfaceStatus := &machine.Status.NetworkInterfaces[idx]
			networkInterfaceStatus.Handle = newNetworkInterfaceStatus.Handle
			lastStateTransitionTime := networkInterfaceStatus.LastStateTransitionTime
			if networkInterfaceStatus.State != newNetworkInterfaceStatus.State {
				lastStateTransitionTime = &now
			}
			networkInterfaceStatus.LastStateTransitionTime = lastStateTransitionTime
			networkInterfaceStatus.State = newNetworkInterfaceStatus.State
		}
	}

	for i := range machine.Status.NetworkInterfaces {
		networkInterfaceStatus := &machine.Status.NetworkInterfaces[i]
		if seenNames.Has(networkInterfaceStatus.Name) {
			continue
		}

		newState := computev1alpha1.NetworkInterfaceStateDetached
		if networkInterfaceStatus.State != newState {
			networkInterfaceStatus.LastStateTransitionTime = &now
		}
		networkInterfaceStatus.State = newState
	}
	return nil
}

func (r *MachineReconciler) deleteNetworkInterfaces(ctx context.Context, log logr.Logger, networkInterfaces []*ori.NetworkInterface) (bool, error) {
	var (
		errs                        []error
		deletingNetworkInterfaceIDs []string
	)
	for _, networkInterface := range networkInterfaces {
		networkInterfaceID := networkInterface.Metadata.Id
		log := log.WithValues("NetworkInterfaceID", networkInterfaceID)

		log.V(1).Info("Deleting network interface")
		if _, err := r.MachineRuntime.DeleteNetworkInterface(ctx, &ori.DeleteNetworkInterfaceRequest{
			NetworkInterfaceId: networkInterfaceID,
		}); err != nil {
			if status.Code(err) != codes.NotFound {
				errs = append(errs, fmt.Errorf("error deleting network interface %s: %w", networkInterfaceID, err))
				continue
			}

			deletingNetworkInterfaceIDs = append(deletingNetworkInterfaceIDs, networkInterfaceID)
		}
	}

	switch {
	case len(errs) > 0:
		return false, fmt.Errorf("error(s) deleting network interface(s): %v", errs)
	case len(deletingNetworkInterfaceIDs) > 0:
		log.V(1).Info("NetworkInterfaces are deleting", "DeletingNetworkInterfaceIDs", deletingNetworkInterfaceIDs)
		return false, nil
	default:
		log.V(1).Info("All network interfaces are gone")
		return true, nil
	}
}

func (r *MachineReconciler) deleteNetworkInterfacesByMachineUID(ctx context.Context, log logr.Logger, machineUID types.UID) (bool, error) {
	log.V(1).Info("Listing network interfaces by machine uid")
	networkInterfaces, err := r.listORINetworkInterfacesByMachineUID(ctx, machineUID)
	if err != nil {
		return false, fmt.Errorf("error listing network interfaces by machine uid: %w", err)
	}

	return r.deleteNetworkInterfaces(ctx, log, networkInterfaces)
}

func (r *MachineReconciler) deleteNetworkInterfacesByMachineKey(ctx context.Context, log logr.Logger, machineKey client.ObjectKey) (bool, error) {
	log.V(1).Info("Listing network interfaces by machine key")
	networkInterfaces, err := r.listORINetworkInterfacesByMachineKey(ctx, machineKey)
	if err != nil {
		return false, fmt.Errorf("error listing network interfaces by machine key: %w", err)
	}

	return r.deleteNetworkInterfaces(ctx, log, networkInterfaces)
}
