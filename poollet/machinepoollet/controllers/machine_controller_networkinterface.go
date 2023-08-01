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
	"errors"
	"fmt"

	"github.com/go-logr/logr"
	"github.com/gogo/protobuf/proto"
	commonv1alpha1 "github.com/onmetal/onmetal-api/api/common/v1alpha1"
	computev1alpha1 "github.com/onmetal/onmetal-api/api/compute/v1alpha1"
	ipamv1alpha1 "github.com/onmetal/onmetal-api/api/ipam/v1alpha1"
	networkingv1alpha1 "github.com/onmetal/onmetal-api/api/networking/v1alpha1"
	ori "github.com/onmetal/onmetal-api/ori/apis/machine/v1alpha1"
	"github.com/onmetal/onmetal-api/poollet/machinepoollet/controllers/events"
	"github.com/onmetal/onmetal-api/utils/claimmanager"
	utilslices "github.com/onmetal/onmetal-api/utils/slices"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
	corev1 "k8s.io/api/core/v1"
	apierrors "k8s.io/apimachinery/pkg/api/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/util/sets"
	"sigs.k8s.io/controller-runtime/pkg/client"
)

type MachineNetworkInterfaceSelector map[string]string

func (s MachineNetworkInterfaceSelector) Match(nic *networkingv1alpha1.NetworkInterface) bool {
	_, ok := s[nic.Name]
	return ok
}

func (r *MachineReconciler) machineNetworkInterfaceSelector(machine *computev1alpha1.Machine) MachineNetworkInterfaceSelector {
	sel := make(MachineNetworkInterfaceSelector)
	for _, machineNic := range machine.Spec.NetworkInterfaces {
		nicName := computev1alpha1.MachineNetworkInterfaceName(machine.Name, machineNic)
		sel[nicName] = machineNic.Name
	}
	return sel
}

func (r *MachineReconciler) networkInterfaceClaimStrategy() claimmanager.ClaimStrategy[*networkingv1alpha1.NetworkInterface] {
	return claimmanager.LocalUIDReferenceClaimStrategy(r.Client,
		claimmanager.AccessViaLocalUIDReferenceField(func(nic *networkingv1alpha1.NetworkInterface) **commonv1alpha1.LocalUIDReference {
			return &nic.Spec.MachineRef
		}),
	)
}

func (r *MachineReconciler) getNetworkInterfacesForMachine(ctx context.Context, machine *computev1alpha1.Machine, sel MachineNetworkInterfaceSelector) ([]networkingv1alpha1.NetworkInterface, error) {
	nicList := &networkingv1alpha1.NetworkInterfaceList{}
	if err := r.List(ctx, nicList,
		client.InNamespace(machine.Namespace),
	); err != nil {
		return nil, fmt.Errorf("error listing network interfaces: %w", err)
	}

	var (
		claimMgr = claimmanager.New[*networkingv1alpha1.NetworkInterface](machine, sel, r.networkInterfaceClaimStrategy())
		nics     []networkingv1alpha1.NetworkInterface
		errs     []error
	)
	for _, nic := range nicList.Items {
		ok, err := claimMgr.Claim(ctx, &nic)
		if err != nil {
			errs = append(errs, err)
			continue
		}
		if !ok {
			continue
		}

		nics = append(nics, nic)
	}
	return nics, errors.Join(errs...)
}

func (r *MachineReconciler) prepareORINetworkInterfaces(ctx context.Context, machine *computev1alpha1.Machine) ([]*ori.NetworkInterface, bool, error) {
	sel := r.machineNetworkInterfaceSelector(machine)
	nics, err := r.getNetworkInterfacesForMachine(ctx, machine, sel)
	if err != nil {
		return nil, false, err
	}

	var (
		oriNics []*ori.NetworkInterface
		errs    []error
	)
	for _, nic := range nics {
		machineNicName := sel[nic.Name]
		oriNic, ok, err := r.prepareORINetworkInterface(ctx, machine, &nic, machineNicName)
		if err != nil {
			errs = append(errs, fmt.Errorf("[network interface %s] error preparing: %w", machineNicName, err))
			continue
		}
		if !ok {
			continue
		}

		oriNics = append(oriNics, oriNic)
	}
	if err := errors.Join(errs...); err != nil {
		return nil, false, err
	}

	if len(oriNics) != len(machine.Spec.Volumes) {
		expectedNicNames := utilslices.ToSetFunc(machine.Spec.NetworkInterfaces, func(v computev1alpha1.NetworkInterface) string { return v.Name })
		actualNicNames := utilslices.ToSetFunc(oriNics, (*ori.NetworkInterface).GetName)
		missingNicNames := sets.List(expectedNicNames.Difference(actualNicNames))
		r.Eventf(machine, corev1.EventTypeNormal, events.NetworkInterfaceNotReady, "Machine network interfaces are not ready: %v", missingNicNames)
		return nil, false, nil
	}
	return oriNics, true, nil
}

func (r *MachineReconciler) getNetworkInterfaceIP(
	ctx context.Context,
	machine *computev1alpha1.Machine,
	nic *networkingv1alpha1.NetworkInterface,
	idx int,
	nicIP networkingv1alpha1.IPSource,
) (commonv1alpha1.IP, bool, error) {
	switch {
	case nicIP.Value != nil:
		return *nicIP.Value, true, nil
	case nicIP.Ephemeral != nil:
		prefix := &ipamv1alpha1.Prefix{}
		prefixName := networkingv1alpha1.NetworkInterfaceIPIPAMPrefixName(nic.Name, idx)
		prefixKey := client.ObjectKey{Namespace: nic.Namespace, Name: prefixName}
		if err := r.Get(ctx, prefixKey, prefix); err != nil {
			if !apierrors.IsNotFound(err) {
				return commonv1alpha1.IP{}, false, fmt.Errorf("error getting prefix %s: %w", prefixName, err)
			}

			r.Eventf(machine, corev1.EventTypeNormal, events.NetworkInterfaceNotReady, "Network interface prefix %s not found", prefixName)
			return commonv1alpha1.IP{}, false, nil
		}

		if !metav1.IsControlledBy(prefix, nic) {
			r.Eventf(machine, corev1.EventTypeNormal, events.NetworkInterfaceNotReady, "Network interface prefix %s not controlled by network interface", prefixName, nic.Name)
			return commonv1alpha1.IP{}, false, nil
		}

		if prefix.Status.Phase != ipamv1alpha1.PrefixPhaseAllocated {
			r.Eventf(machine, corev1.EventTypeNormal, events.NetworkInterfaceNotReady, "Network interface prefix %s is not yet allocated", prefixName)
			return commonv1alpha1.IP{}, false, nil
		}

		return prefix.Spec.Prefix.IP(), true, nil
	default:
		return commonv1alpha1.IP{}, false, fmt.Errorf("unrecognized network interface ip %#v", nicIP)
	}
}

func (r *MachineReconciler) getNetworkInterfaceIPs(
	ctx context.Context,
	machine *computev1alpha1.Machine,
	nic *networkingv1alpha1.NetworkInterface,
) ([]commonv1alpha1.IP, bool, error) {
	var ips []commonv1alpha1.IP
	for i, nicIP := range nic.Spec.IPs {
		ip, ok, err := r.getNetworkInterfaceIP(ctx, machine, nic, i, nicIP)
		if err != nil || !ok {
			return nil, false, err
		}

		ips = append(ips, ip)
	}
	return ips, true, nil
}

func (r *MachineReconciler) prepareORINetworkInterface(
	ctx context.Context,
	machine *computev1alpha1.Machine,
	nic *networkingv1alpha1.NetworkInterface,
	machineNicName string,
) (*ori.NetworkInterface, bool, error) {
	network := &networkingv1alpha1.Network{}
	networkKey := client.ObjectKey{Namespace: nic.Namespace, Name: nic.Spec.NetworkRef.Name}
	if err := r.Get(ctx, networkKey, network); err != nil {
		if !apierrors.IsNotFound(err) {
			return nil, false, fmt.Errorf("error getting network %s: %w", networkKey.Name, err)
		}
		r.Eventf(machine, corev1.EventTypeNormal, events.NetworkInterfaceNotReady, "Network interface %s network %s not found", nic.Name, networkKey.Name)
		return nil, false, nil
	}

	ips, ok, err := r.getNetworkInterfaceIPs(ctx, machine, nic)
	if err != nil || !ok {
		return nil, false, err
	}

	return &ori.NetworkInterface{
		Name:      machineNicName,
		NetworkId: network.Spec.ProviderID,
		Ips:       utilslices.Map(ips, commonv1alpha1.IP.String),
	}, true, nil
}

func (r *MachineReconciler) getExistingORINetworkInterfacesForMachine(
	ctx context.Context,
	log logr.Logger,
	oriMachine *ori.Machine,
	desiredORINics []*ori.NetworkInterface,
) ([]*ori.NetworkInterface, error) {
	var (
		oriNics              []*ori.NetworkInterface
		desiredORINicsByName = utilslices.ToMapByKey(desiredORINics, (*ori.NetworkInterface).GetName)
		errs                 []error
	)

	for _, oriNic := range oriMachine.Spec.NetworkInterfaces {
		log := log.WithValues("NetworkInterface", oriNic.Name)

		desiredORINic, desiredNicPresent := desiredORINicsByName[oriNic.Name]
		if desiredNicPresent && proto.Equal(desiredORINic, oriNic) {
			log.V(1).Info("Existing ORI network interface is up-to-date")
			oriNics = append(oriNics, oriNic)
			continue
		}

		log.V(1).Info("Detaching outdated ORI network interface")
		_, err := r.MachineRuntime.DetachNetworkInterface(ctx, &ori.DetachNetworkInterfaceRequest{
			MachineId: oriMachine.Metadata.Id,
			Name:      oriNic.Name,
		})
		if err != nil {
			if status.Code(err) != codes.NotFound {
				errs = append(errs, fmt.Errorf("[network interface %s] %w", oriNic.Name, err))
				continue
			}
		}
	}
	if len(errs) > 0 {
		return nil, errors.Join(errs...)
	}
	return oriNics, nil
}

func (r *MachineReconciler) getNewORINetworkInterfacesForMachine(
	ctx context.Context,
	log logr.Logger,
	oriMachine *ori.Machine,
	desiredORINics, existingORINics []*ori.NetworkInterface,
) ([]*ori.NetworkInterface, error) {
	var (
		desiredNewORINics = FindNewORINetworkInterfaces(desiredORINics, existingORINics)
		oriNics           []*ori.NetworkInterface
		errs              []error
	)
	for _, newORINic := range desiredNewORINics {
		log := log.WithValues("NetworkInterface", newORINic.Name)
		log.V(1).Info("Attaching new network interface")
		if _, err := r.MachineRuntime.AttachNetworkInterface(ctx, &ori.AttachNetworkInterfaceRequest{
			MachineId:        oriMachine.Metadata.Id,
			NetworkInterface: newORINic,
		}); err != nil {
			errs = append(errs, fmt.Errorf("[network interface %s] %w", newORINic.Name, err))
			continue
		}

		oriNics = append(oriNics, newORINic)
	}
	if len(errs) > 0 {
		return nil, errors.Join(errs...)
	}
	return oriNics, nil
}

func (r *MachineReconciler) updateORINetworkInterfaces(
	ctx context.Context,
	log logr.Logger,
	machine *computev1alpha1.Machine,
	oriMachine *ori.Machine,
) error {
	desiredORINics, _, err := r.prepareORINetworkInterfaces(ctx, machine)
	if err != nil {
		return fmt.Errorf("error preparing ori network interfaces: %w", err)
	}

	existingORINics, err := r.getExistingORINetworkInterfacesForMachine(ctx, log, oriMachine, desiredORINics)
	if err != nil {
		return fmt.Errorf("error getting existing ori network interfaces for machine: %w", err)
	}

	_, err = r.getNewORINetworkInterfacesForMachine(ctx, log, oriMachine, desiredORINics, existingORINics)
	if err != nil {
		return fmt.Errorf("error getting new ori network interfaces for machine: %w", err)
	}

	return nil
}

var oriNetworkInterfaceStateToNetworkInterfaceState = map[ori.NetworkInterfaceState]computev1alpha1.NetworkInterfaceState{
	ori.NetworkInterfaceState_NETWORK_INTERFACE_PENDING:  computev1alpha1.NetworkInterfaceStatePending,
	ori.NetworkInterfaceState_NETWORK_INTERFACE_ATTACHED: computev1alpha1.NetworkInterfaceStateAttached,
}

func (r *MachineReconciler) convertORINetworkInterfaceState(state ori.NetworkInterfaceState) (computev1alpha1.NetworkInterfaceState, error) {
	if res, ok := oriNetworkInterfaceStateToNetworkInterfaceState[state]; ok {
		return res, nil
	}
	return "", fmt.Errorf("unknown network interface attachment state %v", state)
}

func (r *MachineReconciler) convertORINetworkInterfaceStatus(status *ori.NetworkInterfaceStatus) (computev1alpha1.NetworkInterfaceStatus, error) {
	state, err := r.convertORINetworkInterfaceState(status.State)
	if err != nil {
		return computev1alpha1.NetworkInterfaceStatus{}, err
	}

	return computev1alpha1.NetworkInterfaceStatus{
		Name:   status.Name,
		Handle: status.Handle,
		State:  state,
	}, nil
}

func (r *MachineReconciler) addNetworkInterfaceStatusValues(now metav1.Time, existing, newValues *computev1alpha1.NetworkInterfaceStatus) {
	if existing.State != newValues.State {
		existing.LastStateTransitionTime = &now
	}
	existing.Name = newValues.Name
	existing.State = newValues.State
	existing.Handle = newValues.Handle
}

func (r *MachineReconciler) getNetworkInterfaceStatusesForMachine(
	machine *computev1alpha1.Machine,
	oriMachine *ori.Machine,
	now metav1.Time,
) ([]computev1alpha1.NetworkInterfaceStatus, error) {
	var (
		oriNicStatusByName        = utilslices.ToMapByKey(oriMachine.Status.NetworkInterfaces, (*ori.NetworkInterfaceStatus).GetName)
		existingNicStatusesByName = utilslices.ToMapByKey(machine.Status.NetworkInterfaces, func(status computev1alpha1.NetworkInterfaceStatus) string { return status.Name })
		nicStatuses               []computev1alpha1.NetworkInterfaceStatus
		errs                      []error
	)

	for _, machineNic := range machine.Spec.NetworkInterfaces {
		var (
			oriNicStatus, ok = oriNicStatusByName[machineNic.Name]
			nicStatusValues  computev1alpha1.NetworkInterfaceStatus
		)
		if ok {
			var err error
			nicStatusValues, err = r.convertORINetworkInterfaceStatus(oriNicStatus)
			if err != nil {
				return nil, fmt.Errorf("[network interface %s] %w", machineNic.Name, err)
			}
		} else {
			nicStatusValues = computev1alpha1.NetworkInterfaceStatus{
				Name:  machineNic.Name,
				State: computev1alpha1.NetworkInterfaceStatePending,
			}
		}

		nicStatus := existingNicStatusesByName[machineNic.Name]
		r.addNetworkInterfaceStatusValues(now, &nicStatus, &nicStatusValues)
		nicStatuses = append(nicStatuses, nicStatus)
	}
	if len(errs) > 0 {
		return nil, errors.Join(errs...)
	}
	return nicStatuses, nil
}
