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

package machine

import (
	"context"
	"crypto/rand"
	"encoding/hex"
	"strconv"
	"sync"
	"time"

	ori "github.com/ironcore-dev/ironcore/ori/apis/machine/v1alpha1"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
	"k8s.io/apimachinery/pkg/labels"
)

var (
	// FakeVersion is the version of the fake runtime.
	FakeVersion = "0.1.0"

	// FakeRuntimeName is the name of the fake runtime.
	FakeRuntimeName = "fakeRuntime"
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

type FakeMachineClassStatus struct {
	ori.MachineClassStatus
}

type FakeRuntimeService struct {
	sync.Mutex

	Machines           map[string]*FakeMachine
	MachineClassStatus map[string]*FakeMachineClassStatus
	GetExecURL         func(req *ori.ExecRequest) string
}

func NewFakeRuntimeService() *FakeRuntimeService {
	return &FakeRuntimeService{
		Machines:           make(map[string]*FakeMachine),
		MachineClassStatus: make(map[string]*FakeMachineClassStatus),
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

func (r *FakeRuntimeService) SetMachineClasses(machineClassStatus []*FakeMachineClassStatus) {
	r.Lock()
	defer r.Unlock()

	r.MachineClassStatus = make(map[string]*FakeMachineClassStatus)
	for _, status := range machineClassStatus {
		r.MachineClassStatus[status.MachineClass.Name] = status
	}
}

func (r *FakeRuntimeService) SetGetExecURL(f func(req *ori.ExecRequest) string) {
	r.Lock()
	defer r.Unlock()

	r.GetExecURL = f
}

func (r *FakeRuntimeService) Version(ctx context.Context, req *ori.VersionRequest) (*ori.VersionResponse, error) {
	return &ori.VersionResponse{
		RuntimeName:    FakeRuntimeName,
		RuntimeVersion: FakeVersion,
	}, nil
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

func (r *FakeRuntimeService) UpdateMachinePower(ctx context.Context, req *ori.UpdateMachinePowerRequest) (*ori.UpdateMachinePowerResponse, error) {
	r.Lock()
	defer r.Unlock()

	machineID := req.MachineId
	machine, ok := r.Machines[machineID]
	if !ok {
		return nil, status.Errorf(codes.NotFound, "machine %q not found", machineID)
	}

	machine.Spec.Power = req.Power
	return &ori.UpdateMachinePowerResponse{}, nil
}

func (r *FakeRuntimeService) AttachVolume(ctx context.Context, req *ori.AttachVolumeRequest) (*ori.AttachVolumeResponse, error) {
	r.Lock()
	defer r.Unlock()

	machineID := req.MachineId
	machine, ok := r.Machines[machineID]
	if !ok {
		return nil, status.Errorf(codes.NotFound, "machine %q not found", machineID)
	}

	machine.Spec.Volumes = append(machine.Spec.Volumes, req.Volume)
	return &ori.AttachVolumeResponse{}, nil
}

func (r *FakeRuntimeService) DetachVolume(ctx context.Context, req *ori.DetachVolumeRequest) (*ori.DetachVolumeResponse, error) {
	r.Lock()
	defer r.Unlock()

	machineID := req.MachineId
	machine, ok := r.Machines[machineID]
	if !ok {
		return nil, status.Errorf(codes.NotFound, "machine %q not found", machineID)
	}

	var (
		filtered []*ori.Volume
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
	return &ori.DetachVolumeResponse{}, nil
}

func (r *FakeRuntimeService) AttachNetworkInterface(ctx context.Context, req *ori.AttachNetworkInterfaceRequest) (*ori.AttachNetworkInterfaceResponse, error) {
	r.Lock()
	defer r.Unlock()

	machineID := req.MachineId
	machine, ok := r.Machines[machineID]
	if !ok {
		return nil, status.Errorf(codes.NotFound, "machine %q not found", machineID)
	}

	machine.Spec.NetworkInterfaces = append(machine.Spec.NetworkInterfaces, req.NetworkInterface)
	return &ori.AttachNetworkInterfaceResponse{}, nil
}

func (r *FakeRuntimeService) DetachNetworkInterface(ctx context.Context, req *ori.DetachNetworkInterfaceRequest) (*ori.DetachNetworkInterfaceResponse, error) {
	r.Lock()
	defer r.Unlock()

	machineID := req.MachineId
	machine, ok := r.Machines[machineID]
	if !ok {
		return nil, status.Errorf(codes.NotFound, "machine %q not found", machineID)
	}

	var (
		filtered []*ori.NetworkInterface
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
	return &ori.DetachNetworkInterfaceResponse{}, nil
}

func (r *FakeRuntimeService) Status(ctx context.Context, req *ori.StatusRequest) (*ori.StatusResponse, error) {
	r.Lock()
	defer r.Unlock()

	var res []*ori.MachineClassStatus
	for _, m := range r.MachineClassStatus {
		machineClassStatus := m.MachineClassStatus
		res = append(res, &machineClassStatus)
	}
	return &ori.StatusResponse{MachineClassStatus: res}, nil
}

func (r *FakeRuntimeService) Exec(ctx context.Context, req *ori.ExecRequest) (*ori.ExecResponse, error) {
	r.Lock()
	defer r.Unlock()

	var url string
	if r.GetExecURL != nil {
		url = r.GetExecURL(req)
	}
	return &ori.ExecResponse{Url: url}, nil
}
