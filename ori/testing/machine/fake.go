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

package machine

import (
	"context"
	"crypto/rand"
	"encoding/hex"
	"strconv"
	"sync"
	"time"

	ori "github.com/onmetal/onmetal-api/ori/apis/machine/v1alpha1"
	"golang.org/x/exp/slices"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
	"k8s.io/apimachinery/pkg/labels"
)

func filterInLabels(labelSelector, lbls map[string]string) bool {
	return labels.SelectorFromSet(labelSelector).Matches(labels.Set(lbls))
}

const defaultIDLength = 63

func generateID(length int) string {
	data := make([]byte, (length/2)+1)
	for {
		_, _ = rand.Read(data)
		id := hex.EncodeToString(data)

		// Truncated versions of the id should not be numerical.
		if _, err := strconv.ParseInt(id[:12], 10, 64); err != nil {
			continue
		}

		return id[:length]
	}
}

type FakeMachine struct {
	ori.Machine
}

type FakeVolume struct {
	ori.Volume
}

type FakeNetworkInterface struct {
	ori.NetworkInterface
}

type FakeMachineClass struct {
	ori.MachineClass
}

type FakeRuntimeService struct {
	sync.Mutex

	Machines          map[string]*FakeMachine
	Volumes           map[string]*FakeVolume
	NetworkInterfaces map[string]*FakeNetworkInterface
	MachineClasses    map[string]*FakeMachineClass
}

func NewFakeRuntimeService() *FakeRuntimeService {
	return &FakeRuntimeService{
		Machines:          make(map[string]*FakeMachine),
		Volumes:           make(map[string]*FakeVolume),
		NetworkInterfaces: make(map[string]*FakeNetworkInterface),
		MachineClasses:    make(map[string]*FakeMachineClass),
	}
}

func (r *FakeRuntimeService) SetMachines(machines []*FakeMachine) {
	r.Lock()
	defer r.Unlock()

	r.Machines = make(map[string]*FakeMachine)
	for _, machine := range machines {
		r.Machines[machine.Metadata.Id] = machine
	}
}

func (r *FakeRuntimeService) SetVolumes(volumes []*FakeVolume) {
	r.Lock()
	defer r.Unlock()

	r.Volumes = make(map[string]*FakeVolume)
	for _, volume := range volumes {
		r.Volumes[volume.Metadata.Id] = volume
	}
}

func (r *FakeRuntimeService) SetNetworkInterfaces(networkInterfaces []*FakeNetworkInterface) {
	r.Lock()
	defer r.Unlock()

	r.NetworkInterfaces = make(map[string]*FakeNetworkInterface)
	for _, networkInterface := range networkInterfaces {
		r.NetworkInterfaces[networkInterface.Metadata.Id] = networkInterface
	}
}

func (r *FakeRuntimeService) SetMachineClasses(machineClasses []*FakeMachineClass) {
	r.Lock()
	defer r.Unlock()

	r.MachineClasses = make(map[string]*FakeMachineClass)
	for _, machineClass := range machineClasses {
		r.MachineClasses[machineClass.Name] = machineClass
	}
}

func (r *FakeRuntimeService) ListMachines(ctx context.Context, req *ori.ListMachinesRequest) (*ori.ListMachinesResponse, error) {
	r.Lock()
	defer r.Unlock()

	filter := req.Filter

	var res []*ori.Machine
	for _, m := range r.Machines {
		if filter != nil {
			if filter.Id != "" && filter.Id != m.Metadata.Id {
				continue
			}
			if filter.LabelSelector != nil && !filterInLabels(filter.LabelSelector, m.Metadata.Labels) {
				continue
			}
		}

		machine := m.Machine
		res = append(res, &machine)
	}
	return &ori.ListMachinesResponse{Machines: res}, nil
}

func (r *FakeRuntimeService) CreateMachine(ctx context.Context, req *ori.CreateMachineRequest) (*ori.CreateMachineResponse, error) {
	r.Lock()
	defer r.Unlock()

	machine := *req.Machine
	machine.Metadata.Id = generateID(defaultIDLength)
	machine.Metadata.CreatedAt = time.Now().UnixNano()
	machine.Status = &ori.MachineStatus{
		State: ori.MachineState_MACHINE_PENDING,
	}

	r.Machines[machine.Metadata.Id] = &FakeMachine{
		Machine: machine,
	}

	return &ori.CreateMachineResponse{
		Machine: &machine,
	}, nil
}

func (r *FakeRuntimeService) DeleteMachine(ctx context.Context, req *ori.DeleteMachineRequest) (*ori.DeleteMachineResponse, error) {
	r.Lock()
	defer r.Unlock()

	machineID := req.MachineId
	if _, ok := r.Machines[machineID]; !ok {
		return nil, status.Errorf(codes.NotFound, "machine %q not found", machineID)
	}

	delete(r.Machines, machineID)
	return &ori.DeleteMachineResponse{}, nil
}

func (r *FakeRuntimeService) UpdateMachineAnnotations(ctx context.Context, req *ori.UpdateMachineAnnotationsRequest) (*ori.UpdateMachineAnnotationsResponse, error) {
	r.Lock()
	defer r.Unlock()

	machineID := req.MachineId
	machine, ok := r.Machines[machineID]
	if !ok {
		return nil, status.Errorf(codes.NotFound, "machine %q not found", machineID)
	}

	machine.Metadata.Annotations = req.Annotations
	return &ori.UpdateMachineAnnotationsResponse{}, nil
}

func (r *FakeRuntimeService) CreateVolumeAttachment(ctx context.Context, req *ori.CreateVolumeAttachmentRequest) (*ori.CreateVolumeAttachmentResponse, error) {
	r.Lock()
	defer r.Unlock()

	machineID := req.MachineId
	machine, ok := r.Machines[machineID]
	if !ok {
		return nil, status.Errorf(codes.NotFound, "machine %q not found", machineID)
	}

	machine.Spec.Volumes = append(machine.Spec.Volumes, req.Volume)
	return &ori.CreateVolumeAttachmentResponse{}, nil
}

func (r *FakeRuntimeService) DeleteVolumeAttachment(ctx context.Context, req *ori.DeleteVolumeAttachmentRequest) (*ori.DeleteVolumeAttachmentResponse, error) {
	r.Lock()
	defer r.Unlock()

	machineID := req.MachineId
	machine, ok := r.Machines[machineID]
	if !ok {
		return nil, status.Errorf(codes.NotFound, "machine %q not found", machineID)
	}

	var (
		filtered []*ori.VolumeAttachment
		found    bool
	)
	for _, attachment := range machine.Spec.Volumes {
		if attachment.Name == req.Name {
			found = true
			continue
		}

		filtered = append(filtered, attachment)
	}
	if !found {
		return nil, status.Errorf(codes.NotFound, "machine %q volume attachment %q not found", machineID, req.Name)
	}

	machine.Spec.Volumes = filtered
	return &ori.DeleteVolumeAttachmentResponse{}, nil
}

func (r *FakeRuntimeService) CreateNetworkInterfaceAttachment(ctx context.Context, req *ori.CreateNetworkInterfaceAttachmentRequest) (*ori.CreateNetworkInterfaceAttachmentResponse, error) {
	r.Lock()
	defer r.Unlock()

	machineID := req.MachineId
	machine, ok := r.Machines[machineID]
	if !ok {
		return nil, status.Errorf(codes.NotFound, "machine %q not found", machineID)
	}

	machine.Spec.NetworkInterfaces = append(machine.Spec.NetworkInterfaces, req.NetworkInterface)
	return &ori.CreateNetworkInterfaceAttachmentResponse{}, nil
}

func (r *FakeRuntimeService) DeleteNetworkInterfaceAttachment(ctx context.Context, req *ori.DeleteNetworkInterfaceAttachmentRequest) (*ori.DeleteNetworkInterfaceAttachmentResponse, error) {
	r.Lock()
	defer r.Unlock()

	machineID := req.MachineId
	machine, ok := r.Machines[machineID]
	if !ok {
		return nil, status.Errorf(codes.NotFound, "machine %q not found", machineID)
	}

	var (
		filtered []*ori.NetworkInterfaceAttachment
		found    bool
	)
	for _, attachment := range machine.Spec.NetworkInterfaces {
		if attachment.Name == req.Name {
			found = true
			continue
		}

		filtered = append(filtered, attachment)
	}
	if !found {
		return nil, status.Errorf(codes.NotFound, "machine %q network interface attachment %q not found", machineID, req.Name)
	}

	machine.Spec.NetworkInterfaces = filtered
	return &ori.DeleteNetworkInterfaceAttachmentResponse{}, nil

}

func (r *FakeRuntimeService) ListVolumes(ctx context.Context, req *ori.ListVolumesRequest) (*ori.ListVolumesResponse, error) {
	r.Lock()
	defer r.Unlock()

	filter := req.Filter

	var res []*ori.Volume
	for _, v := range r.Volumes {
		if filter != nil {
			if filter.Id != "" && filter.Id != v.Metadata.Id {
				continue
			}
			if filter.LabelSelector != nil && !filterInLabels(filter.LabelSelector, v.Metadata.Labels) {
				continue
			}
		}

		volume := v.Volume
		res = append(res, &volume)
	}
	return &ori.ListVolumesResponse{Volumes: res}, nil
}

func (r *FakeRuntimeService) CreateVolume(ctx context.Context, req *ori.CreateVolumeRequest) (*ori.CreateVolumeResponse, error) {
	r.Lock()
	defer r.Unlock()

	volume := *req.Volume
	volume.Metadata.Id = generateID(defaultIDLength)
	volume.Metadata.CreatedAt = time.Now().UnixNano()

	r.Volumes[volume.Metadata.Id] = &FakeVolume{
		Volume: volume,
	}

	return &ori.CreateVolumeResponse{
		Volume: &volume,
	}, nil
}

func (r *FakeRuntimeService) DeleteVolume(ctx context.Context, req *ori.DeleteVolumeRequest) (*ori.DeleteVolumeResponse, error) {
	r.Lock()
	defer r.Unlock()

	volumeID := req.VolumeId
	if _, ok := r.Volumes[volumeID]; !ok {
		return nil, status.Errorf(codes.NotFound, "volume %q not found", volumeID)
	}

	delete(r.Volumes, volumeID)
	return &ori.DeleteVolumeResponse{}, nil
}

func (r *FakeRuntimeService) ListNetworkInterfaces(ctx context.Context, req *ori.ListNetworkInterfacesRequest) (*ori.ListNetworkInterfacesResponse, error) {
	r.Lock()
	defer r.Unlock()

	filter := req.Filter

	var res []*ori.NetworkInterface
	for _, v := range r.NetworkInterfaces {
		if filter != nil {
			if filter.Id != "" && filter.Id != v.Metadata.Id {
				continue
			}
			if filter.LabelSelector != nil && !filterInLabels(filter.LabelSelector, v.Metadata.Labels) {
				continue
			}
		}

		networkInterface := v.NetworkInterface
		res = append(res, &networkInterface)
	}
	return &ori.ListNetworkInterfacesResponse{NetworkInterfaces: res}, nil
}

func (r *FakeRuntimeService) CreateNetworkInterface(ctx context.Context, req *ori.CreateNetworkInterfaceRequest) (*ori.CreateNetworkInterfaceResponse, error) {
	r.Lock()
	defer r.Unlock()

	networkInterface := *req.NetworkInterface
	networkInterface.Metadata.Id = generateID(defaultIDLength)
	networkInterface.Metadata.CreatedAt = time.Now().UnixNano()

	r.NetworkInterfaces[networkInterface.Metadata.Id] = &FakeNetworkInterface{
		NetworkInterface: networkInterface,
	}

	return &ori.CreateNetworkInterfaceResponse{
		NetworkInterface: &networkInterface,
	}, nil
}

func (r *FakeRuntimeService) DeleteNetworkInterface(ctx context.Context, req *ori.DeleteNetworkInterfaceRequest) (*ori.DeleteNetworkInterfaceResponse, error) {
	r.Lock()
	defer r.Unlock()

	networkInterfaceID := req.NetworkInterfaceId
	if _, ok := r.NetworkInterfaces[networkInterfaceID]; !ok {
		return nil, status.Errorf(codes.NotFound, "network interface %q not found", networkInterfaceID)
	}

	delete(r.NetworkInterfaces, networkInterfaceID)
	return &ori.DeleteNetworkInterfaceResponse{}, nil
}

func (r *FakeRuntimeService) UpdateNetworkInterfaceIPs(ctx context.Context, req *ori.UpdateNetworkInterfaceIPsRequest) (*ori.UpdateNetworkInterfaceIPsResponse, error) {
	r.Lock()
	defer r.Unlock()

	networkInterfaceID := req.NetworkInterfaceId
	networkInterface, ok := r.NetworkInterfaces[networkInterfaceID]
	if !ok {
		return nil, status.Errorf(codes.NotFound, "network interface %q not found", networkInterfaceID)
	}

	networkInterface.Spec.Ips = req.Ips
	return &ori.UpdateNetworkInterfaceIPsResponse{}, nil
}

func (r *FakeRuntimeService) CreateNetworkInterfaceVirtualIP(ctx context.Context, req *ori.CreateNetworkInterfaceVirtualIPRequest) (*ori.CreateNetworkInterfaceVirtualIPResponse, error) {
	r.Lock()
	defer r.Unlock()

	networkInterfaceID := req.NetworkInterfaceId
	networkInterface, ok := r.NetworkInterfaces[networkInterfaceID]
	if !ok {
		return nil, status.Errorf(codes.NotFound, "network interface %q not found", networkInterfaceID)
	}

	if networkInterface.Spec.VirtualIp != nil {
		return nil, status.Errorf(codes.AlreadyExists, "network interface %q virtual ip already exists", networkInterfaceID)
	}

	networkInterface.Spec.VirtualIp = req.VirtualIp
	return &ori.CreateNetworkInterfaceVirtualIPResponse{}, nil
}

func (r *FakeRuntimeService) UpdateNetworkInterfaceVirtualIP(ctx context.Context, req *ori.UpdateNetworkInterfaceVirtualIPRequest) (*ori.UpdateNetworkInterfaceVirtualIPResponse, error) {
	r.Lock()
	defer r.Unlock()

	networkInterfaceID := req.NetworkInterfaceId
	networkInterface, ok := r.NetworkInterfaces[networkInterfaceID]
	if !ok {
		return nil, status.Errorf(codes.NotFound, "network interface %q not found", networkInterfaceID)
	}

	if networkInterface.Spec.VirtualIp == nil {
		return nil, status.Errorf(codes.NotFound, "network interface %q virtual ip not found", networkInterfaceID)
	}

	networkInterface.Spec.VirtualIp = req.VirtualIp
	return &ori.UpdateNetworkInterfaceVirtualIPResponse{}, nil
}

func (r *FakeRuntimeService) DeleteNetworkInterfaceVirtualIP(ctx context.Context, req *ori.DeleteNetworkInterfaceVirtualIPRequest) (*ori.DeleteNetworkInterfaceVirtualIPResponse, error) {
	r.Lock()
	defer r.Unlock()

	networkInterfaceID := req.NetworkInterfaceId
	networkInterface, ok := r.NetworkInterfaces[networkInterfaceID]
	if !ok {
		return nil, status.Errorf(codes.NotFound, "network interface %q not found", networkInterfaceID)
	}

	if networkInterface.Spec.VirtualIp == nil {
		return nil, status.Errorf(codes.NotFound, "network interface %q virtual ip not found", networkInterfaceID)
	}

	networkInterface.Spec.VirtualIp = nil
	return &ori.DeleteNetworkInterfaceVirtualIPResponse{}, nil
}

func (r *FakeRuntimeService) CreateNetworkInterfacePrefix(ctx context.Context, req *ori.CreateNetworkInterfacePrefixRequest) (*ori.CreateNetworkInterfacePrefixResponse, error) {
	r.Lock()
	defer r.Unlock()

	networkInterfaceID := req.NetworkInterfaceId
	networkInterface, ok := r.NetworkInterfaces[networkInterfaceID]
	if !ok {
		return nil, status.Errorf(codes.NotFound, "network interface %q not found", networkInterfaceID)
	}

	prefix := req.Prefix
	if slices.Contains(networkInterface.Spec.Prefixes, prefix) {
		return nil, status.Errorf(codes.AlreadyExists, "network interface %q prefix %q already exists", networkInterfaceID, prefix)
	}

	networkInterface.Spec.Prefixes = append(networkInterface.Spec.Prefixes, prefix)
	return &ori.CreateNetworkInterfacePrefixResponse{}, nil
}

func (r *FakeRuntimeService) DeleteNetworkInterfacePrefix(ctx context.Context, req *ori.DeleteNetworkInterfacePrefixRequest) (*ori.DeleteNetworkInterfacePrefixResponse, error) {
	r.Lock()
	defer r.Unlock()

	networkInterfaceID := req.NetworkInterfaceId
	networkInterface, ok := r.NetworkInterfaces[networkInterfaceID]
	if !ok {
		return nil, status.Errorf(codes.NotFound, "network interface %q not found", networkInterfaceID)
	}

	prefix := req.Prefix
	idx := slices.Index(networkInterface.Spec.Prefixes, prefix)
	if idx < 0 {
		return nil, status.Errorf(codes.NotFound, "network interface %q prefix %q not found", networkInterfaceID, prefix)
	}

	networkInterface.Spec.Prefixes = slices.Delete(networkInterface.Spec.Prefixes, idx, idx+1)
	return &ori.DeleteNetworkInterfacePrefixResponse{}, nil
}

func (r *FakeRuntimeService) CreateNetworkInterfaceLoadBalancerTarget(ctx context.Context, req *ori.CreateNetworkInterfaceLoadBalancerTargetRequest) (*ori.CreateNetworkInterfaceLoadBalancerTargetResponse, error) {
	r.Lock()
	defer r.Unlock()

	networkInterfaceID := req.NetworkInterfaceId
	networkInterface, ok := r.NetworkInterfaces[networkInterfaceID]
	if !ok {
		return nil, status.Errorf(codes.NotFound, "network interface %q not found", networkInterfaceID)
	}

	loadBalancerTarget := req.LoadBalancerTarget
	if slices.ContainsFunc(
		networkInterface.Spec.LoadBalancerTargets,
		func(tgt *ori.LoadBalancerTargetSpec) bool {
			return tgt.Key() == loadBalancerTarget.Key()
		},
	) {
		return nil, status.Errorf(codes.AlreadyExists, "network interface %q load balancer target %q already exists", networkInterfaceID, loadBalancerTarget)
	}

	networkInterface.Spec.LoadBalancerTargets = append(networkInterface.Spec.LoadBalancerTargets, loadBalancerTarget)
	return &ori.CreateNetworkInterfaceLoadBalancerTargetResponse{}, nil
}

func (r *FakeRuntimeService) DeleteNetworkInterfaceLoadBalancerTarget(ctx context.Context, req *ori.DeleteNetworkInterfaceLoadBalancerTargetRequest) (*ori.DeleteNetworkInterfaceLoadBalancerTargetResponse, error) {
	r.Lock()
	defer r.Unlock()

	networkInterfaceID := req.NetworkInterfaceId
	networkInterface, ok := r.NetworkInterfaces[networkInterfaceID]
	if !ok {
		return nil, status.Errorf(codes.NotFound, "network interface %q not found", networkInterfaceID)
	}

	loadBalancerTarget := req.LoadBalancerTarget
	idx := slices.IndexFunc(
		networkInterface.Spec.LoadBalancerTargets,
		func(tgt *ori.LoadBalancerTargetSpec) bool {
			return tgt.Key() == loadBalancerTarget.Key()
		},
	)
	if idx < 0 {
		return nil, status.Errorf(codes.NotFound, "network interface %q load balancer target %q not found", networkInterfaceID, loadBalancerTarget)
	}

	networkInterface.Spec.LoadBalancerTargets = slices.Delete(networkInterface.Spec.LoadBalancerTargets, idx, idx+1)
	return &ori.DeleteNetworkInterfaceLoadBalancerTargetResponse{}, nil
}

func (r *FakeRuntimeService) ListMachineClasses(ctx context.Context, req *ori.ListMachineClassesRequest) (*ori.ListMachineClassesResponse, error) {
	r.Lock()
	defer r.Unlock()

	var res []*ori.MachineClass
	for _, m := range r.MachineClasses {
		machineClass := m.MachineClass
		res = append(res, &machineClass)
	}
	return &ori.ListMachineClassesResponse{MachineClasses: res}, nil
}
