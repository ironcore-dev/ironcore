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
	computev1alpha1 "github.com/onmetal/onmetal-api/api/compute/v1alpha1"
	networkingv1alpha1 "github.com/onmetal/onmetal-api/api/networking/v1alpha1"
	ori "github.com/onmetal/onmetal-api/ori/apis/machine/v1alpha1"
	"github.com/onmetal/onmetal-api/poollet/machinepoollet/controllers/events"
	utilslices "github.com/onmetal/onmetal-api/utils/slices"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
	corev1 "k8s.io/api/core/v1"
	apierrors "k8s.io/apimachinery/pkg/api/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"sigs.k8s.io/controller-runtime/pkg/client"
)

func stringersToStrings[E fmt.Stringer](stringers []E) []string {
	res := make([]string, len(stringers))
	for i, s := range stringers {
		res[i] = s.String()
	}
	return res
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

func (r *MachineReconciler) prepareORINetworkInterfaces(
	ctx context.Context,
	log logr.Logger,
	machine *computev1alpha1.Machine,
) ([]*ori.NetworkInterface, []string, error) {
	var (
		oriNics     []*ori.NetworkInterface
		unreadyNics []string
		errs        []error
	)
	for _, machineNic := range machine.Spec.NetworkInterfaces {
		log := log.WithValues("MachineNetworkInterface", machineNic.Name)
		oriNic, ok, err := r.prepareORINetworkInterface(ctx, log, machine, &machineNic)
		if err != nil {
			errs = append(errs, fmt.Errorf("[network interface %s] error preparing: %w", machineNic.Name, err))
			continue
		}
		if !ok {
			unreadyNics = append(unreadyNics, machineNic.Name)
			continue
		}

		oriNics = append(oriNics, oriNic)
	}

	if len(errs) > 0 {
		return nil, nil, errors.Join(errs...)
	}
	return oriNics, unreadyNics, nil
}

func (r *MachineReconciler) prepareORINetworkInterface(
	ctx context.Context,
	log logr.Logger,
	machine *computev1alpha1.Machine,
	machineNic *computev1alpha1.NetworkInterface,
) (*ori.NetworkInterface, bool, error) {
	switch {
	case machineNic.NetworkInterfaceRef != nil || machineNic.Ephemeral != nil:
		networkInterface := &networkingv1alpha1.NetworkInterface{}
		networkInterfaceKey := client.ObjectKey{Namespace: machine.Namespace, Name: computev1alpha1.MachineNetworkInterfaceName(machine.Name, *machineNic)}
		if err := r.Get(ctx, networkInterfaceKey, networkInterface); err != nil {
			if !apierrors.IsNotFound(err) {
				return nil, false, fmt.Errorf("error getting network interface: %w", err)
			}
			r.Eventf(machine, corev1.EventTypeNormal, events.NetworkInterfaceNotReady, "Network interface %s not found", networkInterfaceKey.Name)
			return nil, false, nil
		}

		if state := networkInterface.Status.State; state != networkingv1alpha1.NetworkInterfaceStateAvailable {
			r.Eventf(machine, corev1.EventTypeNormal, events.NetworkInterfaceNotReady, "Network interface %s is in state %s", networkInterfaceKey.Name, state)
			return nil, false, nil
		}

		if !r.isNetworkInterfaceBoundToMachine(machine, machineNic.Name, networkInterface) {
			return nil, false, nil
		}

		return &ori.NetworkInterface{
			Name:      machineNic.Name,
			NetworkId: networkInterface.Status.NetworkHandle,
			Ips:       stringersToStrings(networkInterface.Status.IPs),
		}, true, nil
	default:
		return nil, false, fmt.Errorf("unrecognized machine volume %#v", machineNic)
	}
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
	desiredORINics, _, err := r.prepareORINetworkInterfaces(ctx, log, machine)
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
		Handle: status.NetworkInterfaceHandle,
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

func (r *MachineReconciler) updateNetworkInterfaceProviderIDIfExists(
	ctx context.Context,
	machine *computev1alpha1.Machine,
	machineNic *computev1alpha1.NetworkInterface,
	providerID string,
) error {
	nic := &networkingv1alpha1.NetworkInterface{}
	nicKey := client.ObjectKey{Namespace: machine.Namespace, Name: computev1alpha1.MachineNetworkInterfaceName(machine.Name, *machineNic)}
	if err := r.Get(ctx, nicKey, nic); err != nil {
		if !apierrors.IsNotFound(err) {
			return fmt.Errorf("error getting network interface %s: %w", nicKey.Name, err)
		}
		// Be graceful if the actual network interface does not exist anymore.
		return nil
	}

	if nic.Status.ProviderID == providerID {
		return nil
	}

	baseNic := nic.DeepCopy()
	nic.Status.ProviderID = providerID
	if err := r.Status().Patch(ctx, nic, client.MergeFrom(baseNic)); err != nil {
		// Be graceful if the actual network interface does not exist anymore.
		if apierrors.IsNotFound(err) {
			return nil
		}
		return fmt.Errorf("error patching network interface %s status: %w", nicKey.Name, err)
	}
	return nil
}

func (r *MachineReconciler) getNetworkInterfaceStatusesForMachine(
	ctx context.Context,
	log logr.Logger,
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

		if err := r.updateNetworkInterfaceProviderIDIfExists(ctx, machine, &machineNic, nicStatusValues.Handle); err != nil {
			errs = append(errs, fmt.Errorf("[network interface %s] %w", machineNic.Name, err))
			continue
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
