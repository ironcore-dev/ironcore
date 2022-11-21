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
	"errors"
	"fmt"

	ori "github.com/onmetal/onmetal-api/ori/apis/storage/v1alpha1"
	"github.com/onmetal/onmetal-api/volumebroker/apiutils"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

func (s *Server) VolumeStatus(ctx context.Context, req *ori.VolumeStatusRequest) (*ori.VolumeStatusResponse, error) {
	log := s.loggerFrom(ctx)
	id := req.VolumeId
	log = log.WithValues("VolumeID", id)

	log.V(1).Info("Getting volume")
	onmetalVolume, err := s.getOnmetalVolume(ctx, id)
	if err != nil {
		var volumeNotFound *volumeNotFoundError
		if !errors.As(err, &volumeNotFound) {
			return nil, fmt.Errorf("error getting onmetal volume: %w", err)
		}
		return nil, status.Error(codes.NotFound, volumeNotFound.Error())
	}

	metadata, err := apiutils.GetMetadataAnnotation(onmetalVolume)
	if err != nil {
		return nil, fmt.Errorf("error getting metadata: %w", err)
	}

	annotations, err := apiutils.GetAnnotationsAnnotation(onmetalVolume)
	if err != nil {
		return nil, fmt.Errorf("error getting annotations: %w", err)
	}

	labels, err := apiutils.GetLabelsAnnotation(onmetalVolume)
	if err != nil {
		return nil, fmt.Errorf("error getting labels: %w", err)
	}

	state := s.convertOnmetalVolumeState(onmetalVolume.Status.State)

	access, err := s.convertOnmetalVolumeAccess(ctx, onmetalVolume.Status.Access)
	if err != nil {
		return nil, fmt.Errorf("error converting access: %w", err)
	}

	return &ori.VolumeStatusResponse{
		Status: &ori.VolumeStatus{
			Id:          id,
			Metadata:    metadata,
			Image:       onmetalVolume.Spec.Image,
			ImageRef:    "", // TODO: Fill
			State:       state,
			Access:      access,
			Annotations: annotations,
			Labels:      labels,
		},
	}, nil
}
