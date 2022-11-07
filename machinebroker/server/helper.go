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

package server

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"

	commonv1alpha1 "github.com/onmetal/onmetal-api/apis/common/v1alpha1"
	computev1alpha1 "github.com/onmetal/onmetal-api/apis/compute/v1alpha1"
	networkingv1alpha1 "github.com/onmetal/onmetal-api/apis/networking/v1alpha1"
	storagev1alpha1 "github.com/onmetal/onmetal-api/apis/storage/v1alpha1"
	machinebrokerv1alpha1 "github.com/onmetal/onmetal-api/machinebroker/api/v1alpha1"
	ori "github.com/onmetal/onmetal-api/ori/apis/runtime/v1alpha1"
	corev1 "k8s.io/api/core/v1"
)

var (
	ErrMachineNotFound          = errors.New("machine not found")
	ErrNetworkInterfaceNotFound = errors.New("network interface not found")
)

var onmetalNetworkInterfaceStateToNetworkInterfaceState = map[computev1alpha1.NetworkInterfaceState]ori.NetworkInterfaceState{
	computev1alpha1.NetworkInterfaceStatePending:  ori.NetworkInterfaceState_NETWORK_INTERFACE_PENDING,
	computev1alpha1.NetworkInterfaceStateAttached: ori.NetworkInterfaceState_NETWORK_INTERFACE_ATTACHED,
	computev1alpha1.NetworkInterfaceStateDetached: ori.NetworkInterfaceState_NETWORK_INTERFACE_DETACHED,
	computev1alpha1.NetworkInterfaceStateError:    ori.NetworkInterfaceState_NETWORK_INTERFACE_ERROR,
}

func OnmetalNetworkInterfaceStateToNetworkInterfaceState(state computev1alpha1.NetworkInterfaceState) ori.NetworkInterfaceState {
	return onmetalNetworkInterfaceStateToNetworkInterfaceState[state]
}

func OnmetalNetworkInterfaceToNetworkInterface(
	machineID string,
	machineMetadata *ori.MachineMetadata,
	onmetalNetworkInterface *computev1alpha1.NetworkInterface,
	onmetalNetworkingNetworkInterface *networkingv1alpha1.NetworkInterface,
) (*ori.NetworkInterface, error) {
	ips := make([]string, len(onmetalNetworkingNetworkInterface.Status.IPs))
	for i, ip := range onmetalNetworkingNetworkInterface.Status.IPs {
		ips[i] = ip.String()
	}

	var virtualIPConfig *ori.VirtualIPConfig
	if onmetalVirtualIP := onmetalNetworkingNetworkInterface.Status.VirtualIP; onmetalVirtualIP != nil {
		virtualIPConfig = &ori.VirtualIPConfig{
			Ip: onmetalVirtualIP.String(),
		}
	}

	return &ori.NetworkInterface{
		MachineId:       machineID,
		MachineMetadata: machineMetadata,
		Name:            onmetalNetworkInterface.Name,
		Network:         &ori.NetworkConfig{Handle: onmetalNetworkingNetworkInterface.Status.NetworkHandle},
		Ips:             ips,
		VirtualIp:       virtualIPConfig,
	}, nil
}

var onmetalVolumeStateToVolumeState = map[computev1alpha1.VolumeState]ori.VolumeState{
	computev1alpha1.VolumeStatePending:  ori.VolumeState_VOLUME_PENDING,
	computev1alpha1.VolumeStateAttached: ori.VolumeState_VOLUME_ATTACHED,
	computev1alpha1.VolumeStateDetached: ori.VolumeState_VOLUME_DETACHED,
	computev1alpha1.VolumeStateError:    ori.VolumeState_VOLUME_ERROR,
}

func OnmetalVolumeStateToVolumeState(state computev1alpha1.VolumeState) ori.VolumeState {
	return onmetalVolumeStateToVolumeState[state]
}

func OnmetalVolumeToVolume(
	machineID string,
	machineMetadata *ori.MachineMetadata,
	onmetalVolume *computev1alpha1.Volume,
	onmetalStorageVolume *storagev1alpha1.Volume,
) (*ori.Volume, error) {
	var access *ori.VolumeAccess
	if onmetalVolume.VolumeRef != nil || onmetalVolume.Ephemeral != nil {
		onmetalVolumeAccess := onmetalStorageVolume.Status.Access
		if onmetalVolumeAccess == nil {
			return nil, fmt.Errorf("onmetal volume %s/%s does not specify access", onmetalStorageVolume.Namespace, onmetalStorageVolume.Name)
		}

		access = &ori.VolumeAccess{
			Driver: onmetalVolumeAccess.Driver,
			Handle: onmetalVolumeAccess.Handle,
		}
	}

	var emptyDisk *ori.EmptyDisk
	if onmetalEmptyDisk := onmetalVolume.EmptyDisk; onmetalEmptyDisk != nil {
		var sizeLimitBytes uint64
		if sizeLimit := onmetalEmptyDisk.SizeLimit; sizeLimit != nil {
			sizeLimitBytes = uint64(sizeLimit.Value())
		}

		emptyDisk = &ori.EmptyDisk{
			SizeLimitBytes: sizeLimitBytes,
		}
	}

	return &ori.Volume{
		MachineId:       machineID,
		MachineMetadata: machineMetadata,
		Name:            onmetalVolume.Name,
		Device:          onmetalVolume.Device,
		Access:          access,
		EmptyDisk:       emptyDisk,
	}, nil
}

func OnmetalIPsToIPs(ips []commonv1alpha1.IP) []string {
	res := make([]string, len(ips))
	for i, ip := range ips {
		res[i] = ip.String()
	}
	return res
}

var onmetalMachineStateToORIState = map[computev1alpha1.MachineState]ori.MachineState{
	computev1alpha1.MachineStatePending:  ori.MachineState_MACHINE_PENDING,
	computev1alpha1.MachineStateRunning:  ori.MachineState_MACHINE_RUNNING,
	computev1alpha1.MachineStateShutdown: ori.MachineState_MACHINE_SHUTDOWN,
	computev1alpha1.MachineStateError:    ori.MachineState_MACHINE_ERROR,
	computev1alpha1.MachineStateUnknown:  ori.MachineState_MACHINE_UNKNOWN,
}

func OnmetalMachineStateToORIState(state computev1alpha1.MachineState) ori.MachineState {
	if oriState, ok := onmetalMachineStateToORIState[state]; ok {
		return oriState
	}
	return ori.MachineState_MACHINE_UNKNOWN
}

func IPsIPFamilies(ips []commonv1alpha1.IP) []corev1.IPFamily {
	res := make([]corev1.IPFamily, len(ips))
	for i, ip := range ips {
		res[i] = ip.Family()
	}
	return res
}

func IPsIPSource(ips []commonv1alpha1.IP) []networkingv1alpha1.IPSource {
	res := make([]networkingv1alpha1.IPSource, len(ips))
	for i := range ips {
		res[i] = networkingv1alpha1.IPSource{
			Value: &ips[i],
		}
	}
	return res
}

func ParseIPs(ipStrings []string) ([]commonv1alpha1.IP, error) {
	var ips []commonv1alpha1.IP
	for _, ipString := range ipStrings {
		ip, err := commonv1alpha1.ParseIP(ipString)
		if err != nil {
			return nil, fmt.Errorf("error parsing ip %q: %w", ipString, err)
		}

		ips = append(ips, ip)
	}
	return ips, nil
}

func ParseIPsAndFamilies(ipStrings []string) ([]commonv1alpha1.IP, []corev1.IPFamily, error) {
	var (
		ips        []commonv1alpha1.IP
		ipFamilies []corev1.IPFamily
	)
	for _, ipString := range ipStrings {
		ip, err := commonv1alpha1.ParseIP(ipString)
		if err != nil {
			return nil, nil, fmt.Errorf("error parsing ip %q: %w", ipString, err)
		}

		ips = append(ips, ip)
		ipFamilies = append(ipFamilies, ip.Family())
	}
	return ips, ipFamilies, nil
}

func MachineToORIMachine(machine *computev1alpha1.Machine) (*ori.Machine, error) {
	metadata := &ori.MachineMetadata{}
	if err := json.Unmarshal([]byte(machine.Annotations[machinebrokerv1alpha1.MetadataAnnotation]), metadata); err != nil {
		return nil, fmt.Errorf("error unmarshalling metadata: %w", err)
	}

	labels := make(map[string]string)
	if err := json.Unmarshal([]byte(machine.Annotations[machinebrokerv1alpha1.LabelsAnnotation]), &labels); err != nil {
		return nil, fmt.Errorf("error unmarshalling labels: %w", err)
	}

	annotations := make(map[string]string)
	if err := json.Unmarshal([]byte(machine.Annotations[machinebrokerv1alpha1.AnnotationsAnnotation]), &annotations); err != nil {
		return nil, fmt.Errorf("error unmarshalling annotations: %w", err)
	}

	return &ori.Machine{
		Id:          string(machine.UID),
		Metadata:    metadata,
		Annotations: annotations,
		Labels:      labels,
	}, nil
}

type Cleaner struct {
	funcs []func(ctx context.Context) error
}

func NewCleaner() *Cleaner {
	return &Cleaner{}
}

func (c *Cleaner) Add(f func(ctx context.Context) error) {
	// funcs need to be added in reverse order (cleanup stack)
	c.funcs = append([]func(ctx context.Context) error{f}, c.funcs...)
}

func (c *Cleaner) Cleanup(ctx context.Context) error {
	for _, f := range c.funcs {
		if err := f(ctx); err != nil {
			return err
		}
	}
	return nil
}
