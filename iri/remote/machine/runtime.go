// SPDX-FileCopyrightText: 2023 SAP SE or an SAP affiliate company and IronCore contributors
// SPDX-License-Identifier: Apache-2.0

package machine

import (
	"context"
	"fmt"

	"github.com/ironcore-dev/ironcore/iri/apis/machine"
	iri "github.com/ironcore-dev/ironcore/iri/apis/machine/v1alpha1"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"
)

type remoteRuntime struct {
	client iri.MachineRuntimeClient
}

func NewRemoteRuntime(endpoint string) (machine.RuntimeService, error) {
	conn, err := grpc.NewClient(endpoint,
		grpc.WithTransportCredentials(insecure.NewCredentials()),
	)
	if err != nil {
		return nil, fmt.Errorf("error dialing: %w", err)
	}

	return &remoteRuntime{
		client: iri.NewMachineRuntimeClient(conn),
	}, nil
}

func (r *remoteRuntime) Version(ctx context.Context, req *iri.VersionRequest) (*iri.VersionResponse, error) {
	return r.client.Version(ctx, req)
}

func (r *remoteRuntime) ListEvents(ctx context.Context, req *iri.ListEventsRequest) (*iri.ListEventsResponse, error) {
	return r.client.ListEvents(ctx, req)
}

func (r *remoteRuntime) ListMachines(ctx context.Context, req *iri.ListMachinesRequest) (*iri.ListMachinesResponse, error) {
	return r.client.ListMachines(ctx, req)
}

func (r *remoteRuntime) CreateMachine(ctx context.Context, req *iri.CreateMachineRequest) (*iri.CreateMachineResponse, error) {
	return r.client.CreateMachine(ctx, req)
}

func (r *remoteRuntime) DeleteMachine(ctx context.Context, req *iri.DeleteMachineRequest) (*iri.DeleteMachineResponse, error) {
	return r.client.DeleteMachine(ctx, req)
}

func (r *remoteRuntime) UpdateMachineAnnotations(ctx context.Context, req *iri.UpdateMachineAnnotationsRequest) (*iri.UpdateMachineAnnotationsResponse, error) {
	return r.client.UpdateMachineAnnotations(ctx, req)
}

func (r *remoteRuntime) UpdateMachinePower(ctx context.Context, req *iri.UpdateMachinePowerRequest) (*iri.UpdateMachinePowerResponse, error) {
	return r.client.UpdateMachinePower(ctx, req)
}

func (r *remoteRuntime) AttachVolume(ctx context.Context, req *iri.AttachVolumeRequest) (*iri.AttachVolumeResponse, error) {
	return r.client.AttachVolume(ctx, req)
}

func (r *remoteRuntime) DetachVolume(ctx context.Context, req *iri.DetachVolumeRequest) (*iri.DetachVolumeResponse, error) {
	return r.client.DetachVolume(ctx, req)
}

func (r *remoteRuntime) AttachNetworkInterface(ctx context.Context, req *iri.AttachNetworkInterfaceRequest) (*iri.AttachNetworkInterfaceResponse, error) {
	return r.client.AttachNetworkInterface(ctx, req)
}

func (r *remoteRuntime) DetachNetworkInterface(ctx context.Context, req *iri.DetachNetworkInterfaceRequest) (*iri.DetachNetworkInterfaceResponse, error) {
	return r.client.DetachNetworkInterface(ctx, req)
}

func (r *remoteRuntime) ListReservations(ctx context.Context, req *iri.ListReservationsRequest) (*iri.ListReservationsResponse, error) {
	return r.client.ListReservations(ctx, req)
}

func (r *remoteRuntime) CreateReservation(ctx context.Context, req *iri.CreateReservationRequest) (*iri.CreateReservationResponse, error) {
	return r.client.CreateReservation(ctx, req)
}

func (r *remoteRuntime) DeleteReservation(ctx context.Context, req *iri.DeleteReservationRequest) (*iri.DeleteReservationResponse, error) {
	return r.client.DeleteReservation(ctx, req)
}

func (r *remoteRuntime) Status(ctx context.Context, req *iri.StatusRequest) (*iri.StatusResponse, error) {
	return r.client.Status(ctx, req)
}

func (r *remoteRuntime) Exec(ctx context.Context, req *iri.ExecRequest) (*iri.ExecResponse, error) {
	return r.client.Exec(ctx, req)
}
