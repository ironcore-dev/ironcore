// SPDX-FileCopyrightText: 2023 SAP SE or an SAP affiliate company and IronCore contributors
// SPDX-License-Identifier: Apache-2.0

package controllers

import (
	"context"
	"errors"
	"fmt"

	"github.com/go-logr/logr"
	"github.com/gogo/protobuf/proto"
	commonv1alpha1 "github.com/ironcore-dev/ironcore/api/common/v1alpha1"
	computev1alpha1 "github.com/ironcore-dev/ironcore/api/compute/v1alpha1"
	ipamv1alpha1 "github.com/ironcore-dev/ironcore/api/ipam/v1alpha1"
	networkingv1alpha1 "github.com/ironcore-dev/ironcore/api/networking/v1alpha1"
	iri "github.com/ironcore-dev/ironcore/iri/apis/machine/v1alpha1"
	"github.com/ironcore-dev/ironcore/poollet/machinepoollet/api/v1alpha1"
	"github.com/ironcore-dev/ironcore/poollet/machinepoollet/controllers/events"
	"github.com/ironcore-dev/ironcore/utils/claimmanager"
	utilslices "github.com/ironcore-dev/ironcore/utils/slices"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
	corev1 "k8s.io/api/core/v1"
	apierrors "k8s.io/apimachinery/pkg/api/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/util/sets"
	"sigs.k8s.io/controller-runtime/pkg/client"
)

type networkInterfaceClaimStrategy struct {
	client.Client
}

func (s *networkInterfaceClaimStrategy) ClaimState(claimer client.Object, obj client.Object) claimmanager.ClaimState {
	nic := obj.(*networkingv1alpha1.NetworkInterface)
	if machineRef := nic.Spec.MachineRef; machineRef != nil {
		if machineRef.UID == claimer.GetUID() {
			return claimmanager.ClaimStateClaimed
		}
		return claimmanager.ClaimStateTaken
	}
	return claimmanager.ClaimStateFree
}

func (s *networkInterfaceClaimStrategy) Adopt(ctx context.Context, claimer client.Object, obj client.Object) error {
	nic := obj.(*networkingv1alpha1.NetworkInterface)
	base := nic.DeepCopy()
	nic.Spec.MachineRef = commonv1alpha1.NewLocalObjUIDRef(claimer)
	nic.Spec.ProviderID = ""
	return s.Patch(ctx, nic, client.StrategicMergeFrom(base))
}

func (s *networkInterfaceClaimStrategy) Release(ctx context.Context, claimer client.Object, obj client.Object) error {
	nic := obj.(*networkingv1alpha1.NetworkInterface)
	base := nic.DeepCopy()
	nic.Spec.ProviderID = ""
	nic.Spec.MachineRef = nil
	return s.Patch(ctx, nic, client.StrategicMergeFrom(base))
}

func (r *MachineReconciler) networkInterfaceNameToMachineNetworkInterfaceName(machine *computev1alpha1.Machine) map[string]string {
	sel := make(map[string]string)
	for _, machineNic := range machine.Spec.NetworkInterfaces {
		nicName := computev1alpha1.MachineNetworkInterfaceName(machine.Name, machineNic)
		sel[nicName] = machineNic.Name
	}
	return sel
}

func (r *MachineReconciler) machineNetworkInterfaceSelector(machine *computev1alpha1.Machine) claimmanager.Selector {
	names := sets.New(computev1alpha1.MachineNetworkInterfaceNames(machine)...)
	return claimmanager.SelectorFunc(func(obj client.Object) bool {
		nic := obj.(*networkingv1alpha1.NetworkInterface)
		return names.Has(nic.Name)
	})
}

func (r *MachineReconciler) getNetworkInterfacesForMachine(ctx context.Context, machine *computev1alpha1.Machine) ([]networkingv1alpha1.NetworkInterface, error) {
	nicList := &networkingv1alpha1.NetworkInterfaceList{}
	if err := r.List(ctx, nicList,
		client.InNamespace(machine.Namespace),
	); err != nil {
		return nil, fmt.Errorf("error listing network interfaces: %w", err)
	}

	var (
		sel      = r.machineNetworkInterfaceSelector(machine)
		claimMgr = claimmanager.New(machine, sel, &networkInterfaceClaimStrategy{r.Client})
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

func (r *MachineReconciler) prepareIRINetworkInterfacesForMachine(
	ctx context.Context,
	machine *computev1alpha1.Machine,
	nics []networkingv1alpha1.NetworkInterface,
) ([]*iri.NetworkInterface, map[string]v1alpha1.ObjectUIDRef, bool, error) {
	iriNics, mapping, err := r.getIRINetworkInterfacesForMachine(ctx, machine, nics)
	if err != nil {
		return nil, nil, false, err
	}

	if len(iriNics) != len(machine.Spec.NetworkInterfaces) {
		expectedNicNames := utilslices.ToSetFunc(machine.Spec.NetworkInterfaces, func(v computev1alpha1.NetworkInterface) string { return v.Name })
		actualNicNames := utilslices.ToSetFunc(iriNics, (*iri.NetworkInterface).GetName)
		missingNicNames := sets.List(expectedNicNames.Difference(actualNicNames))
		r.Eventf(machine, corev1.EventTypeNormal, events.NetworkInterfaceNotReady, "Machine network interfaces are not ready: %v", missingNicNames)
		return nil, nil, false, nil
	}

	return iriNics, mapping, true, err
}

func (r *MachineReconciler) getIRINetworkInterfacesForMachine(
	ctx context.Context,
	machine *computev1alpha1.Machine,
	nics []networkingv1alpha1.NetworkInterface,
) ([]*iri.NetworkInterface, map[string]v1alpha1.ObjectUIDRef, error) {
	var (
		nicNameToMachineNicName = r.networkInterfaceNameToMachineNetworkInterfaceName(machine)

		iriNics                []*iri.NetworkInterface
		machineNicNameToUIDRef = make(map[string]v1alpha1.ObjectUIDRef)
		errs                   []error
	)
	for _, nic := range nics {
		machineNicName := nicNameToMachineNicName[nic.Name]
		iriNic, ok, err := r.prepareIRINetworkInterface(ctx, machine, &nic, machineNicName)
		if err != nil {
			errs = append(errs, fmt.Errorf("[network interface %s] error preparing: %w", machineNicName, err))
			continue
		}
		if !ok {
			continue
		}

		iriNics = append(iriNics, iriNic)
		machineNicNameToUIDRef[machineNicName] = v1alpha1.ObjUID(&nic)
	}
	if err := errors.Join(errs...); err != nil {
		return nil, nil, err
	}
	return iriNics, machineNicNameToUIDRef, nil
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

func (r *MachineReconciler) prepareIRINetworkInterface(
	ctx context.Context,
	machine *computev1alpha1.Machine,
	nic *networkingv1alpha1.NetworkInterface,
	machineNicName string,
) (*iri.NetworkInterface, bool, error) {
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

	return &iri.NetworkInterface{
		Name:       machineNicName,
		NetworkId:  network.Spec.ProviderID,
		Ips:        utilslices.Map(ips, commonv1alpha1.IP.String),
		Attributes: nic.Spec.Attributes,
	}, true, nil
}

func (r *MachineReconciler) getExistingIRINetworkInterfacesForMachine(
	ctx context.Context,
	log logr.Logger,
	iriMachine *iri.Machine,
	desiredIRINics []*iri.NetworkInterface,
) ([]*iri.NetworkInterface, error) {
	var (
		iriNics              []*iri.NetworkInterface
		desiredIRINicsByName = utilslices.ToMapByKey(desiredIRINics, (*iri.NetworkInterface).GetName)
		errs                 []error
	)

	for _, iriNic := range iriMachine.Spec.NetworkInterfaces {
		log := log.WithValues("NetworkInterface", iriNic.Name)

		desiredIRINic, desiredNicPresent := desiredIRINicsByName[iriNic.Name]
		if desiredNicPresent && proto.Equal(desiredIRINic, iriNic) {
			log.V(1).Info("Existing IRI network interface is up-to-date")
			iriNics = append(iriNics, iriNic)
			continue
		}

		log.V(1).Info("Detaching outdated IRI network interface")
		_, err := r.MachineRuntime.DetachNetworkInterface(ctx, &iri.DetachNetworkInterfaceRequest{
			MachineId: iriMachine.Metadata.Id,
			Name:      iriNic.Name,
		})

		if err != nil {
			if status.Code(err) != codes.NotFound {
				errs = append(errs, fmt.Errorf("[network interface %s] %w", iriNic.Name, err))
				continue
			}
		}
	}
	if len(errs) > 0 {
		return nil, errors.Join(errs...)
	}
	return iriNics, nil
}

func (r *MachineReconciler) getNewAttachIRINetworkInterfaces(
	ctx context.Context,
	log logr.Logger,
	iriMachine *iri.Machine,
	desiredIRINics, existingIRINics []*iri.NetworkInterface,
) ([]*iri.NetworkInterface, error) {
	var (
		desiredNewIRINics = FindNewIRINetworkInterfaces(desiredIRINics, existingIRINics)
		iriNics           []*iri.NetworkInterface
		errs              []error
	)
	for _, newIRINic := range desiredNewIRINics {
		log := log.WithValues("NetworkInterface", newIRINic.Name)
		log.V(1).Info("Attaching new network interface")
		if _, err := r.MachineRuntime.AttachNetworkInterface(ctx, &iri.AttachNetworkInterfaceRequest{
			MachineId:        iriMachine.Metadata.Id,
			NetworkInterface: newIRINic,
		}); err != nil {
			errs = append(errs, fmt.Errorf("[network interface %s] %w", newIRINic.Name, err))
			continue
		}

		iriNics = append(iriNics, newIRINic)
	}
	if len(errs) > 0 {
		return nil, errors.Join(errs...)
	}
	return iriNics, nil
}

func (r *MachineReconciler) updateIRINetworkInterfaces(
	ctx context.Context,
	log logr.Logger,
	machine *computev1alpha1.Machine,
	iriMachine *iri.Machine,
	nics []networkingv1alpha1.NetworkInterface,
) ([]*iri.NetworkInterface, error) {
	desiredIRINics, _, err := r.getIRINetworkInterfacesForMachine(ctx, machine, nics)
	if err != nil {
		return nil, fmt.Errorf("error preparing iri network interfaces: %w", err)
	}

	existingIRINics, err := r.getExistingIRINetworkInterfacesForMachine(ctx, log, iriMachine, desiredIRINics)
	if err != nil {
		return nil, fmt.Errorf("error getting existing iri network interfaces for machine: %w", err)
	}

	_, err = r.getNewAttachIRINetworkInterfaces(ctx, log, iriMachine, desiredIRINics, existingIRINics)
	if err != nil {
		return nil, fmt.Errorf("error getting new iri network interfaces for machine: %w", err)
	}

	return desiredIRINics, nil
}

func (r *MachineReconciler) computeNetworkInterfaceMapping(
	machine *computev1alpha1.Machine,
	nics []networkingv1alpha1.NetworkInterface,
	iriNics []*iri.NetworkInterface,
) map[string]v1alpha1.ObjectUIDRef {
	nicByName := utilslices.ToMapByKey(nics,
		func(nic networkingv1alpha1.NetworkInterface) string { return nic.Name },
	)

	machineNicNameToNicName := make(map[string]string, len(machine.Spec.NetworkInterfaces))
	for _, machineNic := range machine.Spec.NetworkInterfaces {
		nicName := computev1alpha1.MachineNetworkInterfaceName(machine.Name, machineNic)
		machineNicNameToNicName[machineNic.Name] = nicName
	}

	nicMapping := make(map[string]v1alpha1.ObjectUIDRef, len(iriNics))
	for _, iriNic := range iriNics {
		nicName := machineNicNameToNicName[iriNic.Name]
		nic := nicByName[nicName]

		nicMapping[iriNic.Name] = v1alpha1.ObjUID(&nic)
	}
	return nicMapping
}

var iriNetworkInterfaceStateToMachineNetworkInterfaceState = map[iri.NetworkInterfaceState]computev1alpha1.NetworkInterfaceState{
	iri.NetworkInterfaceState_NETWORK_INTERFACE_PENDING:  computev1alpha1.NetworkInterfaceStatePending,
	iri.NetworkInterfaceState_NETWORK_INTERFACE_ATTACHED: computev1alpha1.NetworkInterfaceStateAttached,
}

func (r *MachineReconciler) convertIRINetworkInterfaceState(state iri.NetworkInterfaceState) (computev1alpha1.NetworkInterfaceState, error) {
	if res, ok := iriNetworkInterfaceStateToMachineNetworkInterfaceState[state]; ok {
		return res, nil
	}
	return "", fmt.Errorf("unknown network interface attachment state %v", state)
}

func (r *MachineReconciler) convertIRINetworkInterfaceStatus(status *iri.NetworkInterfaceStatus, machine *computev1alpha1.Machine, machineNic computev1alpha1.NetworkInterface) (computev1alpha1.NetworkInterfaceStatus, error) {
	state, err := r.convertIRINetworkInterfaceState(status.State)
	if err != nil {
		return computev1alpha1.NetworkInterfaceStatus{}, err
	}

	return computev1alpha1.NetworkInterfaceStatus{
		Name:                status.Name,
		Handle:              status.Handle,
		State:               state,
		NetworkInterfaceRef: corev1.LocalObjectReference{Name: computev1alpha1.MachineNetworkInterfaceName(machine.Name, machineNic)},
	}, nil
}

func (r *MachineReconciler) addNetworkInterfaceStatusValues(now metav1.Time, existing, newValues *computev1alpha1.NetworkInterfaceStatus) {
	if existing.State != newValues.State {
		existing.LastStateTransitionTime = &now
	}
	existing.Name = newValues.Name
	existing.NetworkInterfaceRef = newValues.NetworkInterfaceRef
	existing.State = newValues.State
	existing.Handle = newValues.Handle
}

func (r *MachineReconciler) getNetworkInterfaceStatusesForMachine(
	machine *computev1alpha1.Machine,
	iriMachine *iri.Machine,
	now metav1.Time,
) ([]computev1alpha1.NetworkInterfaceStatus, error) {
	var (
		iriNicStatusByName        = utilslices.ToMapByKey(iriMachine.Status.NetworkInterfaces, (*iri.NetworkInterfaceStatus).GetName)
		existingNicStatusesByName = utilslices.ToMapByKey(machine.Status.NetworkInterfaces, func(status computev1alpha1.NetworkInterfaceStatus) string { return status.Name })
		nicStatuses               []computev1alpha1.NetworkInterfaceStatus
		errs                      []error
	)

	for _, machineNic := range machine.Spec.NetworkInterfaces {
		var (
			iriNicStatus, ok = iriNicStatusByName[machineNic.Name]
			nicStatusValues  computev1alpha1.NetworkInterfaceStatus
		)
		if ok {
			var err error
			nicStatusValues, err = r.convertIRINetworkInterfaceStatus(iriNicStatus, machine, machineNic)
			if err != nil {
				return nil, fmt.Errorf("[network interface %s] %w", machineNic.Name, err)
			}
		} else {
			nicStatusValues = computev1alpha1.NetworkInterfaceStatus{
				Name:                machineNic.Name,
				State:               computev1alpha1.NetworkInterfaceStatePending,
				NetworkInterfaceRef: corev1.LocalObjectReference{Name: computev1alpha1.MachineNetworkInterfaceName(machine.Name, machineNic)},
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
