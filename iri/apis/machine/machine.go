// SPDX-FileCopyrightText: 2023 SAP SE or an SAP affiliate company and IronCore contributors
// SPDX-License-Identifier: Apache-2.0

package machine

import (
	"context"

	api "github.com/ironcore-dev/ironcore/iri/apis/machine/v1alpha1"
)

type RuntimeService interface {
	Version(context.Context, *api.VersionRequest) (*api.VersionResponse, error)
	ListEvents(context.Context, *api.ListEventsRequest) (*api.ListEventsResponse, error)
	ListMachines(context.Context, *api.ListMachinesRequest) (*api.ListMachinesResponse, error)
	CreateMachine(context.Context, *api.CreateMachineRequest) (*api.CreateMachineResponse, error)
	DeleteMachine(context.Context, *api.DeleteMachineRequest) (*api.DeleteMachineResponse, error)
	UpdateMachineAnnotations(context.Context, *api.UpdateMachineAnnotationsRequest) (*api.UpdateMachineAnnotationsResponse, error)
	UpdateMachinePower(context.Context, *api.UpdateMachinePowerRequest) (*api.UpdateMachinePowerResponse, error)
	AttachVolume(context.Context, *api.AttachVolumeRequest) (*api.AttachVolumeResponse, error)
	DetachVolume(context.Context, *api.DetachVolumeRequest) (*api.DetachVolumeResponse, error)
	AttachNetworkInterface(context.Context, *api.AttachNetworkInterfaceRequest) (*api.AttachNetworkInterfaceResponse, error)
	DetachNetworkInterface(context.Context, *api.DetachNetworkInterfaceRequest) (*api.DetachNetworkInterfaceResponse, error)
	ListReservations(context.Context, *api.ListReservationsRequest) (*api.ListReservationsResponse, error)
	CreateReservation(context.Context, *api.CreateReservationRequest) (*api.CreateReservationResponse, error)
	DeleteReservation(context.Context, *api.DeleteReservationRequest) (*api.DeleteReservationResponse, error)
	Status(context.Context, *api.StatusRequest) (*api.StatusResponse, error)
	Exec(context.Context, *api.ExecRequest) (*api.ExecResponse, error)
}
