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

	storagev1alpha1 "github.com/onmetal/onmetal-api/api/storage/v1alpha1"
	volumebrokerv1alpha1 "github.com/onmetal/onmetal-api/broker/volumebroker/api/v1alpha1"
	ori "github.com/onmetal/onmetal-api/ori/apis/volume/v1alpha1"
	apierrors "k8s.io/apimachinery/pkg/api/errors"
	"k8s.io/apimachinery/pkg/labels"
	"sigs.k8s.io/controller-runtime/pkg/client"
)

func (s *Server) listOnmetalVolumes(ctx context.Context) ([]storagev1alpha1.Volume, error) {
	onmetalVolumeList := &storagev1alpha1.VolumeList{}

	if err := s.client.List(ctx, onmetalVolumeList,
		client.InNamespace(s.namespace),
		client.MatchingLabels{volumebrokerv1alpha1.VolumeManagerLabel: volumebrokerv1alpha1.VolumeBrokerManager},
	); err != nil {
		return nil, fmt.Errorf("error listing volumes: %w", err)
	}

	return onmetalVolumeList.Items, nil
}

func (s *Server) listVolumes(ctx context.Context) ([]*ori.Volume, error) {
	onmetalVolumes, err := s.listOnmetalVolumes(ctx)
	if err != nil {
		return nil, fmt.Errorf("error listing volumes: %w", err)
	}

	var res []*ori.Volume
	for _, onmetalVolume := range onmetalVolumes {
		volume, err := s.convertOnmetalVolume(&onmetalVolume)
		if err != nil {
			return nil, err
		}

		res = append(res, volume)
	}
	return res, nil
}

func (s *Server) filterVolumes(volumes []*ori.Volume, filter *ori.VolumeFilter) []*ori.Volume {
	if filter == nil {
		return volumes
	}

	var (
		res []*ori.Volume
		sel = labels.SelectorFromSet(filter.LabelSelector)
	)
	for _, oriVolume := range volumes {
		if !sel.Matches(labels.Set(oriVolume.Labels)) {
			continue
		}

		res = append(res, oriVolume)
	}
	return res
}

func (s *Server) getOnmetalVolume(ctx context.Context, volumeID string) (*storagev1alpha1.Volume, error) {
	onmetalVolume := &storagev1alpha1.Volume{}
	onmetalVolumeKey := client.ObjectKey{Namespace: s.namespace, Name: volumeID}
	if err := s.client.Get(ctx, onmetalVolumeKey, onmetalVolume); err != nil {
		if !apierrors.IsNotFound(err) {
			return nil, fmt.Errorf("error getting volume %s: %w", onmetalVolumeKey, err)
		}
		return nil, newVolumeNotFoundError(volumeID)
	}
	return onmetalVolume, nil
}

func (s *Server) getVolume(ctx context.Context, id string) (*ori.Volume, error) {
	onmetalVolume, err := s.getOnmetalVolume(ctx, id)
	if err != nil {
		return nil, err
	}

	return s.convertOnmetalVolume(onmetalVolume)
}

func (s *Server) ListVolumes(ctx context.Context, req *ori.ListVolumesRequest) (*ori.ListVolumesResponse, error) {
	if filter := req.Filter; filter != nil && filter.Id != "" {
		volume, err := s.getVolume(ctx, filter.Id)
		if err != nil {
			if !errors.As(err, new(*volumeNotFoundError)) {
				return nil, err
			}
			return &ori.ListVolumesResponse{
				Volumes: []*ori.Volume{},
			}, nil
		}

		return &ori.ListVolumesResponse{
			Volumes: []*ori.Volume{volume},
		}, nil
	}

	volumes, err := s.listVolumes(ctx)
	if err != nil {
		return nil, err
	}

	volumes = s.filterVolumes(volumes, req.Filter)

	return &ori.ListVolumesResponse{
		Volumes: volumes,
	}, nil
}
