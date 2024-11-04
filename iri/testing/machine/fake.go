// SPDX-FileCopyrightText: 2023 SAP SE or an SAP affiliate company and IronCore contributors
// SPDX-License-Identifier: Apache-2.0

package machine

import (
	"context"
	"crypto/rand"
	"encoding/hex"
	"strconv"
	"sync"
	"time"

	irievent "github.com/ironcore-dev/ironcore/iri/apis/event/v1alpha1"
	iri "github.com/ironcore-dev/ironcore/iri/apis/machine/v1alpha1"
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
	iri.Machine
}

type FakeReservation struct {
	iri.Reservation
}

type FakeVolume struct {
	iri.Volume
}

type FakeNetworkInterface struct {
	iri.NetworkInterface
}

type FakeMachineClassStatus struct {
	iri.MachineClassStatus
}

type FakeEvent struct {
	irievent.Event
}

type FakeRuntimeService struct {
	sync.Mutex

	Machines           map[string]*FakeMachine
	Reservations       map[string]*FakeReservation
	MachineClassStatus map[string]*FakeMachineClassStatus
	GetExecURL         func(req *iri.ExecRequest) string
	Events             []*FakeEvent
}

// ListEvents implements machine.RuntimeService.
func (r *FakeRuntimeService) ListEvents(ctx context.Context, req *iri.ListEventsRequest) (*iri.ListEventsResponse, error) {
	r.Lock()
	defer r.Unlock()

	var res []*irievent.Event
	for _, e := range r.Events {
		event := e.Event
		res = append(res, &event)
	}

	return &iri.ListEventsResponse{Events: res}, nil
}

func NewFakeRuntimeService() *FakeRuntimeService {
	return &FakeRuntimeService{
		Machines:           make(map[string]*FakeMachine),
		MachineClassStatus: make(map[string]*FakeMachineClassStatus),
		Events:             []*FakeEvent{},
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

func (r *FakeRuntimeService) SetGetExecURL(f func(req *iri.ExecRequest) string) {
	r.Lock()
	defer r.Unlock()

	r.GetExecURL = f
}

func (r *FakeRuntimeService) SetEvents(events []*FakeEvent) {
	r.Lock()
	defer r.Unlock()

	r.Events = events
}

func (r *FakeRuntimeService) Version(ctx context.Context, req *iri.VersionRequest) (*iri.VersionResponse, error) {
	return &iri.VersionResponse{
		RuntimeName:    FakeRuntimeName,
		RuntimeVersion: FakeVersion,
	}, nil
}

func (r *FakeRuntimeService) ListMachines(ctx context.Context, req *iri.ListMachinesRequest) (*iri.ListMachinesResponse, error) {
	r.Lock()
	defer r.Unlock()

	filter := req.Filter

	var res []*iri.Machine
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
	return &iri.ListMachinesResponse{Machines: res}, nil
}

func (r *FakeRuntimeService) CreateMachine(ctx context.Context, req *iri.CreateMachineRequest) (*iri.CreateMachineResponse, error) {
	r.Lock()
	defer r.Unlock()

	machine := *req.Machine
	machine.Metadata.Id = generateID(defaultIDLength)
	machine.Metadata.CreatedAt = time.Now().UnixNano()
	machine.Status = &iri.MachineStatus{
		State: iri.MachineState_MACHINE_PENDING,
	}

	r.Machines[machine.Metadata.Id] = &FakeMachine{
		Machine: machine,
	}

	return &iri.CreateMachineResponse{
		Machine: &machine,
	}, nil
}

func (r *FakeRuntimeService) DeleteMachine(ctx context.Context, req *iri.DeleteMachineRequest) (*iri.DeleteMachineResponse, error) {
	r.Lock()
	defer r.Unlock()

	machineID := req.MachineId
	if _, ok := r.Machines[machineID]; !ok {
		return nil, status.Errorf(codes.NotFound, "machine %q not found", machineID)
	}

	delete(r.Machines, machineID)
	return &iri.DeleteMachineResponse{}, nil
}

func (r *FakeRuntimeService) UpdateMachineAnnotations(ctx context.Context, req *iri.UpdateMachineAnnotationsRequest) (*iri.UpdateMachineAnnotationsResponse, error) {
	r.Lock()
	defer r.Unlock()

	machineID := req.MachineId
	machine, ok := r.Machines[machineID]
	if !ok {
		return nil, status.Errorf(codes.NotFound, "machine %q not found", machineID)
	}

	machine.Metadata.Annotations = req.Annotations
	return &iri.UpdateMachineAnnotationsResponse{}, nil
}

func (r *FakeRuntimeService) UpdateMachinePower(ctx context.Context, req *iri.UpdateMachinePowerRequest) (*iri.UpdateMachinePowerResponse, error) {
	r.Lock()
	defer r.Unlock()

	machineID := req.MachineId
	machine, ok := r.Machines[machineID]
	if !ok {
		return nil, status.Errorf(codes.NotFound, "machine %q not found", machineID)
	}

	machine.Spec.Power = req.Power
	return &iri.UpdateMachinePowerResponse{}, nil
}

func (r *FakeRuntimeService) AttachVolume(ctx context.Context, req *iri.AttachVolumeRequest) (*iri.AttachVolumeResponse, error) {
	r.Lock()
	defer r.Unlock()

	machineID := req.MachineId
	machine, ok := r.Machines[machineID]
	if !ok {
		return nil, status.Errorf(codes.NotFound, "machine %q not found", machineID)
	}

	machine.Spec.Volumes = append(machine.Spec.Volumes, req.Volume)
	return &iri.AttachVolumeResponse{}, nil
}

func (r *FakeRuntimeService) DetachVolume(ctx context.Context, req *iri.DetachVolumeRequest) (*iri.DetachVolumeResponse, error) {
	r.Lock()
	defer r.Unlock()

	machineID := req.MachineId
	machine, ok := r.Machines[machineID]
	if !ok {
		return nil, status.Errorf(codes.NotFound, "machine %q not found", machineID)
	}

	var (
		filtered []*iri.Volume
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
	return &iri.DetachVolumeResponse{}, nil
}

func (r *FakeRuntimeService) AttachNetworkInterface(ctx context.Context, req *iri.AttachNetworkInterfaceRequest) (*iri.AttachNetworkInterfaceResponse, error) {
	r.Lock()
	defer r.Unlock()

	machineID := req.MachineId
	machine, ok := r.Machines[machineID]
	if !ok {
		return nil, status.Errorf(codes.NotFound, "machine %q not found", machineID)
	}

	machine.Spec.NetworkInterfaces = append(machine.Spec.NetworkInterfaces, req.NetworkInterface)
	return &iri.AttachNetworkInterfaceResponse{}, nil
}

func (r *FakeRuntimeService) DetachNetworkInterface(ctx context.Context, req *iri.DetachNetworkInterfaceRequest) (*iri.DetachNetworkInterfaceResponse, error) {
	r.Lock()
	defer r.Unlock()

	machineID := req.MachineId
	machine, ok := r.Machines[machineID]
	if !ok {
		return nil, status.Errorf(codes.NotFound, "machine %q not found", machineID)
	}

	var (
		filtered []*iri.NetworkInterface
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
	return &iri.DetachNetworkInterfaceResponse{}, nil
}

func (r *FakeRuntimeService) ListReservations(ctx context.Context, req *iri.ListReservationsRequest) (*iri.ListReservationsResponse, error) {
	r.Lock()
	defer r.Unlock()

	filter := req.Filter

	var res []*iri.Reservation
	for _, m := range r.Reservations {
		if filter != nil {
			if filter.Id != "" && filter.Id != m.Metadata.Id {
				continue
			}
			if filter.LabelSelector != nil && !filterInLabels(filter.LabelSelector, m.Metadata.Labels) {
				continue
			}
		}

		reservation := m.Reservation
		res = append(res, &reservation)
	}
	return &iri.ListReservationsResponse{Reservations: res}, nil
}

func (r *FakeRuntimeService) SetReservations(reservations []*FakeReservation) {
	r.Lock()
	defer r.Unlock()

	r.Reservations = make(map[string]*FakeReservation)
	for _, reservation := range reservations {
		r.Reservations[reservation.Metadata.Id] = reservation
	}
}

func (r *FakeRuntimeService) CreateReservation(ctx context.Context, req *iri.CreateReservationRequest) (*iri.CreateReservationResponse, error) {
	r.Lock()
	defer r.Unlock()

	reservation := *req.Reservation
	reservation.Metadata.Id = generateID(defaultIDLength)
	reservation.Metadata.CreatedAt = time.Now().UnixNano()
	reservation.Status = &iri.ReservationStatus{
		State: 0,
	}

	r.Reservations[reservation.Metadata.Id] = &FakeReservation{
		Reservation: reservation,
	}

	return &iri.CreateReservationResponse{
		Reservation: &reservation,
	}, nil
}

func (r *FakeRuntimeService) DeleteReservation(ctx context.Context, req *iri.DeleteReservationRequest) (*iri.DeleteReservationResponse, error) {
	r.Lock()
	defer r.Unlock()

	reservationID := req.ReservationId
	if _, ok := r.Reservations[reservationID]; !ok {
		return nil, status.Errorf(codes.NotFound, "reservation %q not found", reservationID)
	}

	delete(r.Reservations, reservationID)
	return &iri.DeleteReservationResponse{}, nil
}

func (r *FakeRuntimeService) Status(ctx context.Context, req *iri.StatusRequest) (*iri.StatusResponse, error) {
	r.Lock()
	defer r.Unlock()

	var res []*iri.MachineClassStatus
	for _, m := range r.MachineClassStatus {
		machineClassStatus := m.MachineClassStatus
		res = append(res, &machineClassStatus)
	}
	return &iri.StatusResponse{MachineClassStatus: res}, nil
}

func (r *FakeRuntimeService) Exec(ctx context.Context, req *iri.ExecRequest) (*iri.ExecResponse, error) {
	r.Lock()
	defer r.Unlock()

	var url string
	if r.GetExecURL != nil {
		url = r.GetExecURL(req)
	}
	return &iri.ExecResponse{Url: url}, nil
}
