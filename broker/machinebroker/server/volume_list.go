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
	"fmt"

	storagev1alpha1 "github.com/onmetal/onmetal-api/api/storage/v1alpha1"
	machinebrokerv1alpha1 "github.com/onmetal/onmetal-api/broker/machinebroker/api/v1alpha1"
	ori "github.com/onmetal/onmetal-api/ori/apis/machine/v1alpha1"
	"github.com/onmetal/onmetal-api/utils/slices"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
	corev1 "k8s.io/api/core/v1"
	apierrors "k8s.io/apimachinery/pkg/api/errors"
	"k8s.io/apimachinery/pkg/labels"
	"sigs.k8s.io/controller-runtime/pkg/client"
)

func (s *Server) listAggregatedOnmetalVolumes(ctx context.Context) ([]AggreagateOnmetalVolume, error) {
	onmetalVolumeList := &storagev1alpha1.VolumeList{}
	if err := s.listManagedAndCreated(ctx, onmetalVolumeList); err != nil {
		return nil, fmt.Errorf("error listing onmetal volumes: %w", err)
	}

	onmetalVolumeAccessSecretList := &corev1.SecretList{}
	if err := s.listWithPurpose(ctx, onmetalVolumeAccessSecretList, machinebrokerv1alpha1.VolumeAccessPurpose); err != nil {
		return nil, fmt.Errorf("error listing onmetal volume access secrets: %w", err)
	}

	onmetalVolumeAccessSecretByName := slices.ToMap(
		onmetalVolumeAccessSecretList.Items,
		func(secret corev1.Secret) string { return secret.Name },
	)
	getAccessSecret := func(name string) (*corev1.Secret, error) {
		secret, ok := onmetalVolumeAccessSecretByName[name]
		if !ok {
			return nil, apierrors.NewNotFound(corev1.Resource("secrets"), name)
		}
		return &secret, nil
	}

	var res []AggreagateOnmetalVolume
	for i := range onmetalVolumeList.Items {
		onmetalVolume := &onmetalVolumeList.Items[i]
		volume, err := s.aggregateOnmetalVolume(onmetalVolume, getAccessSecret)
		if err != nil {
			return nil, fmt.Errorf("error assembling onmetal volume %s: %w", onmetalVolume.Name, err)
		}

		res = append(res, *volume)
	}
	return res, nil
}

func (s *Server) aggregateOnmetalVolume(
	onmetalVolume *storagev1alpha1.Volume,
	getAccessSecret func(name string) (*corev1.Secret, error),
) (*AggreagateOnmetalVolume, error) {
	access := onmetalVolume.Status.Access
	if access == nil {
		return nil, fmt.Errorf("volume does not specify access")
	}

	var onmetalVolumeAccessSecret *corev1.Secret
	if onmetalVolumeSecretRef := access.SecretRef; onmetalVolumeSecretRef != nil {
		secret, err := getAccessSecret(onmetalVolumeSecretRef.Name)
		if err != nil {
			return nil, fmt.Errorf("error access secret %s: %w", onmetalVolumeSecretRef.Name, err)
		}

		onmetalVolumeAccessSecret = secret
	}

	return &AggreagateOnmetalVolume{
		Volume:       onmetalVolume,
		AccessSecret: onmetalVolumeAccessSecret,
	}, nil
}

func (s *Server) getAggregateOnmetalVolume(ctx context.Context, id string) (*AggreagateOnmetalVolume, error) {
	onmetalVolume := &storagev1alpha1.Volume{}
	if err := s.getManagedAndCreated(ctx, id, onmetalVolume); err != nil {
		if !apierrors.IsNotFound(err) {
			return nil, fmt.Errorf("error getting onmetal volume %s: %w", id, err)
		}
		return nil, status.Errorf(codes.NotFound, "volume %s not found", id)
	}

	return s.aggregateOnmetalVolume(onmetalVolume, func(name string) (*corev1.Secret, error) {
		secret := &corev1.Secret{}
		if err := s.cluster.Client().Get(ctx, client.ObjectKey{Namespace: s.cluster.Namespace(), Name: name}, secret); err != nil {
			return nil, err
		}
		return secret, nil
	})
}

func (s *Server) getVolume(ctx context.Context, id string) (*ori.Volume, error) {
	onmetalVolume, err := s.getAggregateOnmetalVolume(ctx, id)
	if err != nil {
		return nil, err
	}

	return s.convertOnmetalVolume(onmetalVolume)
}

func (s *Server) listVolumes(ctx context.Context) ([]*ori.Volume, error) {
	onmetalVolumes, err := s.listAggregatedOnmetalVolumes(ctx)
	if err != nil {
		return nil, err
	}

	res := make([]*ori.Volume, len(onmetalVolumes))
	for i := range onmetalVolumes {
		onmetalVolume := &onmetalVolumes[i]
		volume, err := s.convertOnmetalVolume(onmetalVolume)
		if err != nil {
			return nil, fmt.Errorf("error converting onmetal volume %s: %w", onmetalVolume.Volume.Name, err)
		}

		res[i] = volume
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
	for _, volume := range volumes {
		if !sel.Matches(labels.Set(volume.Metadata.Labels)) {
			continue
		}

		res = append(res, volume)
	}
	return res
}

func (s *Server) ListVolumes(ctx context.Context, req *ori.ListVolumesRequest) (*ori.ListVolumesResponse, error) {
	if filter := req.Filter; filter != nil && filter.Id != "" {
		volume, err := s.getVolume(ctx, filter.Id)
		if err != nil {
			if status.Code(err) != codes.NotFound {
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
