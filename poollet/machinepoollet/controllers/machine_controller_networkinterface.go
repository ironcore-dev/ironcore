// SPDX-FileCopyrightText: SAP SE or an SAP affiliate company and IronCore contributors
// SPDX-License-Identifier: Apache-2.0

package controllers

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"maps"
	"strings"

	"github.com/go-logr/logr"
	commonv1alpha1 "github.com/ironcore-dev/ironcore/api/common/v1alpha1"
	computev1alpha1 "github.com/ironcore-dev/ironcore/api/compute/v1alpha1"
	ipamv1alpha1 "github.com/ironcore-dev/ironcore/api/ipam/v1alpha1"
	networkingv1alpha1 "github.com/ironcore-dev/ironcore/api/networking/v1alpha1"
	iri "github.com/ironcore-dev/ironcore/iri/apis/machine/v1alpha1"
	irimeta "github.com/ironcore-dev/ironcore/iri/apis/meta/v1alpha1"
	poolletutils "github.com/ironcore-dev/ironcore/poollet/common/utils"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
	"google.golang.org/protobuf/proto"

	"github.com/ironcore-dev/ironcore/poollet/machinepoollet/api/v1alpha1"
	"github.com/ironcore-dev/ironcore/poollet/machinepoollet/controllers/events"
	"github.com/ironcore-dev/ironcore/utils/claimmanager"
	"github.com/ironcore-dev/ironcore/utils/equality"
	utilsmaps "github.com/ironcore-dev/ironcore/utils/maps"
	utilslices "github.com/ironcore-dev/ironcore/utils/slices"
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
	if err := s.Patch(ctx, nic, client.StrategicMergeFrom(base)); err != nil {
		return err
	}

	// The poollet maintains the network interface status: once the claim is released,
	// no attachment is maintained anymore and the status has to reflect that.
	return patchNetworkInterfaceStatus(ctx, s.Client, nic, resetNetworkInterfaceStatusValues())
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
		r.Eventf(machine, nil, corev1.EventTypeNormal, events.NetworkInterfaceNotReady, events.AttachingNetworkInterface, "Machine network interfaces are not ready: %s", strings.Join(missingNicNames, ", "))
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

			r.Eventf(machine, nil, corev1.EventTypeNormal, events.NetworkInterfaceNotReady, events.AttachingNetworkInterface, "Network interface prefix %s not found", prefixName)
			return commonv1alpha1.IP{}, false, nil
		}

		if !metav1.IsControlledBy(prefix, nic) {
			r.Eventf(machine, nil, corev1.EventTypeNormal, events.NetworkInterfaceNotReady, events.AttachingNetworkInterface, "Network interface prefix %s not controlled by network interface %s", prefixName, nic.Name)
			return commonv1alpha1.IP{}, false, nil
		}

		if prefix.Status.Phase != ipamv1alpha1.PrefixPhaseAllocated {
			r.Eventf(machine, nil, corev1.EventTypeNormal, events.NetworkInterfaceNotReady, events.AttachingNetworkInterface, "Network interface prefix %s is not yet allocated", prefixName)
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

func (r *MachineReconciler) iriNetworkInterfaceLabels(networkinterface *networkingv1alpha1.NetworkInterface) (map[string]string, error) {
	labels := map[string]string{
		v1alpha1.NetworkInterfaceUIDLabel:       string(networkinterface.UID),
		v1alpha1.NetworkInterfaceNamespaceLabel: networkinterface.Namespace,
		v1alpha1.NetworkInterfaceNameLabel:      networkinterface.Name,
	}
	apiLabels, err := poolletutils.PrepareDownwardAPILabels(networkinterface, r.NicDownwardAPILabels, v1alpha1.MachineDownwardAPIPrefix)
	if err != nil {
		return nil, err
	}
	labels = utilsmaps.AppendMap(labels, apiLabels)
	return labels, nil
}

func (r *MachineReconciler) iriNetworkLabels(network *networkingv1alpha1.Network) (map[string]string, error) {
	labels := map[string]string{
		v1alpha1.NetworkUIDLabel:       string(network.UID),
		v1alpha1.NetworkNamespaceLabel: network.Namespace,
		v1alpha1.NetworkNameLabel:      network.Name,
	}
	apiLabels, err := poolletutils.PrepareDownwardAPILabels(network, r.NetworkDownwardAPILabels, v1alpha1.MachineDownwardAPIPrefix)
	if err != nil {
		return nil, err
	}
	labels = utilsmaps.AppendMap(labels, apiLabels)
	return labels, nil
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
		r.Eventf(machine, nil, corev1.EventTypeNormal, events.NetworkInterfaceNotReady, events.AttachingNetworkInterface, "Network interface %s network %s not found", nic.Name, networkKey.Name)
		return nil, false, nil
	}
	nicLabels, err := r.iriNetworkInterfaceLabels(nic)
	if err != nil {
		return nil, false, fmt.Errorf("error preparing iri networkinterface labels: %w", err)
	}

	networkLabels, err := r.iriNetworkLabels(network)
	if err != nil {
		return nil, false, fmt.Errorf("error preparing iri network labels: %w", err)
	}

	attributes, err := r.prepareNetworkInterfaceAttributes(nic, nicLabels, networkLabels)
	if err != nil {
		return nil, false, err
	}

	ips, ok, err := r.getNetworkInterfaceIPs(ctx, machine, nic)
	if err != nil || !ok {
		return nil, false, err
	}
	return &iri.NetworkInterface{
		Name:      machineNicName,
		NetworkId: network.Spec.ProviderID,
		Ips:       utilslices.Map(ips, commonv1alpha1.IP.String),
		Metadata: &irimeta.ObjectMetadata{
			Id:        string(nic.UID),
			Name:      nic.Name,
			Namespace: nic.Namespace,
		},
		Attributes: attributes,
	}, true, nil
}

func (r *MachineReconciler) prepareNetworkInterfaceAttributes(
	nic *networkingv1alpha1.NetworkInterface,
	nicLabels map[string]string,
	networkLabels map[string]string,
) (map[string]string, error) {
	var attributes map[string]string
	if nic.Spec.Attributes != nil {
		attributes = maps.Clone(nic.Spec.Attributes)
	} else {
		attributes = make(map[string]string)
	}
	nicLabelsJSON, err := json.Marshal(nicLabels)
	if err != nil {
		return nil, fmt.Errorf("error marshaling NIC labels: %w", err)
	}
	attributes[v1alpha1.NICLabelsAttributeKey] = string(nicLabelsJSON)

	networkLabelsJSON, err := json.Marshal(networkLabels)
	if err != nil {
		return nil, fmt.Errorf("error marshaling network labels: %w", err)
	}
	attributes[v1alpha1.NetworkLabelsAttributeKey] = string(networkLabelsJSON)

	return attributes, nil
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
		if desiredNicPresent && iriNetworkInterfaceUpToDate(desiredIRINic, iriNic) {
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

// iriNetworkInterfaceUpToDate reports whether the existing IRI network interface matches
// the desired one. The metadata is only compared if the provider actually reports it:
// providers that predate the metadata field never echo it, and comparing it would cause
// an endless detach/attach loop. Providers that persist and report the metadata do see
// metadata changes (through one detach/attach cycle), all others simply keep working.
func iriNetworkInterfaceUpToDate(desired, existing *iri.NetworkInterface) bool {
	if existing.Metadata == nil && desired.Metadata != nil {
		desired = proto.Clone(desired).(*iri.NetworkInterface)
		desired.Metadata = nil
	}
	return proto.Equal(desired, existing)
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
	// The machine status only distinguishes between pending and attached network interfaces:
	// NETWORK_INTERFACE_READY (networking realized but not yet attached) and
	// NETWORK_INTERFACE_ERROR both map to pending. The authoritative networking view of
	// a network interface is the status of the NetworkInterface object itself.
	iri.NetworkInterfaceState_NETWORK_INTERFACE_READY: computev1alpha1.NetworkInterfaceStatePending,
	iri.NetworkInterfaceState_NETWORK_INTERFACE_ERROR: computev1alpha1.NetworkInterfaceStatePending,
}

var iriNetworkInterfaceStateToNetworkInterfaceState = map[iri.NetworkInterfaceState]networkingv1alpha1.NetworkInterfaceState{
	iri.NetworkInterfaceState_NETWORK_INTERFACE_PENDING:  networkingv1alpha1.NetworkInterfaceStatePending,
	iri.NetworkInterfaceState_NETWORK_INTERFACE_READY:    networkingv1alpha1.NetworkInterfaceStateAvailable,
	iri.NetworkInterfaceState_NETWORK_INTERFACE_ATTACHED: networkingv1alpha1.NetworkInterfaceStateAvailable,
	iri.NetworkInterfaceState_NETWORK_INTERFACE_ERROR:    networkingv1alpha1.NetworkInterfaceStateError,
}

// resetNetworkInterfaceStatusValues returns the status values of a network interface that is
// not (or no longer) reported by the network interface's provider.
func resetNetworkInterfaceStatusValues() networkingv1alpha1.NetworkInterfaceStatus {
	return networkingv1alpha1.NetworkInterfaceStatus{
		State: networkingv1alpha1.NetworkInterfaceStatePending,
	}
}

// patchNetworkInterfaceStatus applies the given status values to the network interface,
// keeping the LastStateTransitionTime if the state did not change. If the resulting status
// is equal to the current status, no patch is issued.
//
// The patch is optimistic-locking: if the network interface was modified since it was read
// (e.g. concurrently released by the NetworkInterfaceReleaseReconciler), the patch fails
// with a conflict error. The conflict propagates up to the machine reconciliation, which
// retries and re-reads the network interface.
func patchNetworkInterfaceStatus(
	ctx context.Context,
	c client.Client,
	nic *networkingv1alpha1.NetworkInterface,
	values networkingv1alpha1.NetworkInterfaceStatus,
) error {
	now := metav1.Now()
	base := nic.DeepCopy()

	if nic.Status.State != values.State {
		nic.Status.LastStateTransitionTime = &now
	}
	nic.Status.State = values.State
	nic.Status.IPs = values.IPs
	nic.Status.Prefixes = values.Prefixes
	nic.Status.VirtualIP = values.VirtualIP

	// Note: apiequality.Semantic.DeepEqual must not be used here, as it panics on the
	// unexported fields of the netip types embedded in the commonv1alpha1 IP types (see
	// k8s.io/apimachinery/third_party/forked/golang/reflect). The local equality.Semantic
	// registers custom equality funcs for the IP types and hence is safe to use.
	if equality.Semantic.DeepEqual(base.Status, nic.Status) {
		return nil
	}

	return c.Status().Patch(ctx, nic, client.MergeFromWithOptions(base, client.MergeFromWithOptimisticLock{}))
}

// networkInterfaceStatusValues computes the NetworkInterface status values from the status
// reported via IRI. The provider is expected to report all effective values (ips, prefixes,
// virtual ip); anything not reported results in the corresponding status field being unset.
func networkInterfaceStatusValues(
	iriNicStatus *iri.NetworkInterfaceStatus,
) (networkingv1alpha1.NetworkInterfaceStatus, error) {
	state, ok := iriNetworkInterfaceStateToNetworkInterfaceState[iriNicStatus.State]
	if !ok {
		return networkingv1alpha1.NetworkInterfaceStatus{}, fmt.Errorf("unknown network interface state %v", iriNicStatus.State)
	}

	var ips []commonv1alpha1.IP
	if len(iriNicStatus.Ips) > 0 {
		parsedIPs, err := commonv1alpha1.ParseIPs(iriNicStatus.Ips...)
		if err != nil {
			return networkingv1alpha1.NetworkInterfaceStatus{}, fmt.Errorf("error parsing reported ips %v: %w", iriNicStatus.Ips, err)
		}
		ips = parsedIPs
	}

	var prefixes []commonv1alpha1.IPPrefix
	for _, reportedPrefix := range iriNicStatus.Prefixes {
		prefix, err := commonv1alpha1.ParseIPPrefix(reportedPrefix)
		if err != nil {
			return networkingv1alpha1.NetworkInterfaceStatus{}, fmt.Errorf("error parsing reported prefix %q: %w", reportedPrefix, err)
		}
		prefixes = append(prefixes, prefix)
	}

	var virtualIP *commonv1alpha1.IP
	if iriNicStatus.VirtualIp != "" {
		parsedVirtualIP, err := commonv1alpha1.ParseIP(iriNicStatus.VirtualIp)
		if err != nil {
			return networkingv1alpha1.NetworkInterfaceStatus{}, fmt.Errorf("error parsing reported virtual ip %q: %w", iriNicStatus.VirtualIp, err)
		}
		virtualIP = &parsedVirtualIP
	}

	return networkingv1alpha1.NetworkInterfaceStatus{
		State:     state,
		IPs:       ips,
		Prefixes:  prefixes,
		VirtualIP: virtualIP,
	}, nil
}

func (r *MachineReconciler) convertIRINetworkInterfaceState(state iri.NetworkInterfaceState) (computev1alpha1.NetworkInterfaceState, error) {
	if res, ok := iriNetworkInterfaceStateToMachineNetworkInterfaceState[state]; ok {
		return res, nil
	}
	return "", fmt.Errorf("unknown network interface attachment state %v", state)
}

func (r *MachineReconciler) convertIRINetworkInterfaceStatus(status *iri.NetworkInterfaceStatus, nicName string) (computev1alpha1.NetworkInterfaceStatus, error) {
	state, err := r.convertIRINetworkInterfaceState(status.State)
	if err != nil {
		return computev1alpha1.NetworkInterfaceStatus{}, err
	}

	return computev1alpha1.NetworkInterfaceStatus{
		Name:                status.Name,
		Handle:              status.Handle,
		State:               state,
		NetworkInterfaceRef: corev1.LocalObjectReference{Name: nicName},
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
		nicName := computev1alpha1.MachineNetworkInterfaceName(machine.Name, machineNic)
		if ok {
			var err error
			nicStatusValues, err = r.convertIRINetworkInterfaceStatus(iriNicStatus, nicName)
			if err != nil {
				return nil, fmt.Errorf("[network interface %s] %w", machineNic.Name, err)
			}
		} else {
			nicStatusValues = computev1alpha1.NetworkInterfaceStatus{
				Name:                machineNic.Name,
				State:               computev1alpha1.NetworkInterfaceStatePending,
				NetworkInterfaceRef: corev1.LocalObjectReference{Name: nicName},
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
