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
	"github.com/onmetal/onmetal-api/broker/common"
	volumebrokerv1alpha1 "github.com/onmetal/onmetal-api/broker/volumebroker/api/v1alpha1"
	ori "github.com/onmetal/onmetal-api/ori/apis/volume/v1alpha1"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
	corev1 "k8s.io/api/core/v1"
	apierrors "k8s.io/apimachinery/pkg/api/errors"
	"k8s.io/apimachinery/pkg/labels"
	"sigs.k8s.io/controller-runtime/pkg/client"
)

func (s *Server) listManagedAndCreated(ctx context.Context, list client.ObjectList) error {
	return s.client.List(ctx, list,
		client.InNamespace(s.namespace),
		client.MatchingLabels{
			volumebrokerv1alpha1.ManagerLabel: volumebrokerv1alpha1.VolumeBrokerManager,
			volumebrokerv1alpha1.CreatedLabel: "true",
		},
	)
}

func (s *Server) listAggregateOnmetalVolumes(ctx context.Context) ([]AggregateOnmetalVolume, error) {
	onmetalVolumeList := &storagev1alpha1.VolumeList{}
	if err := s.listManagedAndCreated(ctx, onmetalVolumeList); err != nil {
		return nil, fmt.Errorf("error listing onmetal volumes: %w", err)
	}

	secretList := &corev1.SecretList{}
	if err := s.client.List(ctx, secretList,
		client.InNamespace(s.namespace),
	); err != nil {
		return nil, fmt.Errorf("error listing secrets: %w", err)
	}

	secretByNameGetter, err := common.NewObjectGetter[string, *corev1.Secret](
		corev1.Resource("secrets"),
		common.ByObjectName[*corev1.Secret](),
		common.ObjectSlice[string](secretList.Items),
	)
	if err != nil {
		return nil, fmt.Errorf("error constructing secret getter: %w", err)
	}

	var res []AggregateOnmetalVolume
	for i := range onmetalVolumeList.Items {
		onmetalVolume := &onmetalVolumeList.Items[i]

		var encryptionSecret *corev1.Secret
		if encryption := onmetalVolume.Spec.Encryption; encryption != nil {
			encryptionSecret, err = secretByNameGetter.Get(encryption.SecretRef.Name)
			if err != nil {
				return nil, fmt.Errorf("error getting onmetal encryption secret %s: %w", encryption.SecretRef.Name, err)
			}
		}

		aggregateOnmetalVolume, err := s.aggregateOnmetalVolume(onmetalVolume, encryptionSecret, secretByNameGetter.Get)
		if err != nil {
			return nil, fmt.Errorf("error aggregating onmetal volume %s: %w", onmetalVolume.Name, err)
		}

		res = append(res, *aggregateOnmetalVolume)
	}

	return res, nil
}

func (s *Server) clientGetSecretFunc(ctx context.Context) func(string) (*corev1.Secret, error) {
	return func(name string) (*corev1.Secret, error) {
		secret := &corev1.Secret{}
		if err := s.client.Get(ctx, client.ObjectKey{Namespace: s.namespace, Name: name}, secret); err != nil {
			return nil, err
		}
		return secret, nil
	}
}

func (s *Server) getOnmetalVolumeAccessSecretIfRequired(
	onmetalVolume *storagev1alpha1.Volume,
	getSecret func(string) (*corev1.Secret, error),
) (*corev1.Secret, error) {
	if onmetalVolume.Status.State != storagev1alpha1.VolumeStateAvailable {
		return nil, nil
	}

	access := onmetalVolume.Status.Access
	if access == nil {
		return nil, nil
	}

	secretRef := access.SecretRef
	if secretRef == nil {
		return nil, nil
	}

	secretName := secretRef.Name
	return getSecret(secretName)
}

func (s *Server) aggregateOnmetalVolume(
	onmetalVolume *storagev1alpha1.Volume,
	onmetalEncryptionSecret *corev1.Secret,
	getSecret func(string) (*corev1.Secret, error),
) (*AggregateOnmetalVolume, error) {
	accessSecret, err := s.getOnmetalVolumeAccessSecretIfRequired(onmetalVolume, getSecret)
	if err != nil {
		return nil, fmt.Errorf("error getting onmetal volume access secret: %w", err)
	}

	return &AggregateOnmetalVolume{
		Volume:           onmetalVolume,
		EncryptionSecret: onmetalEncryptionSecret,
		AccessSecret:     accessSecret,
	}, nil
}

func (s *Server) getAggregateOnmetalVolume(ctx context.Context, id string) (*AggregateOnmetalVolume, error) {
	onmetalVolume := &storagev1alpha1.Volume{}
	if err := s.getManagedAndCreated(ctx, id, onmetalVolume); err != nil {
		if !apierrors.IsNotFound(err) {
			return nil, fmt.Errorf("error getting onmetal volume %s: %w", id, err)
		}
		return nil, status.Errorf(codes.NotFound, "volume %s not found", id)
	}

	onmetalEncryptionSecret := &corev1.Secret{}
	if secret := onmetalVolume.Spec.Encryption; secret != nil {
		if err := s.getManagedAndCreated(ctx, secret.SecretRef.Name, onmetalEncryptionSecret); err != nil {
			if !apierrors.IsNotFound(err) {
				return nil, fmt.Errorf("error getting onmetal encryption secret %s: %w", secret.SecretRef.Name, err)
			}
			return nil, status.Errorf(codes.NotFound, "encryption secret %s not found", secret.SecretRef.Name)
		}
	}

	return s.aggregateOnmetalVolume(onmetalVolume, onmetalEncryptionSecret, s.clientGetSecretFunc(ctx))
}

func (s *Server) listVolumes(ctx context.Context) ([]*ori.Volume, error) {
	onmetalVolumes, err := s.listAggregateOnmetalVolumes(ctx)
	if err != nil {
		return nil, fmt.Errorf("error listing volumes: %w", err)
	}

	var res []*ori.Volume
	for _, onmetalVolume := range onmetalVolumes {
		volume, err := s.convertAggregateOnmetalVolume(&onmetalVolume)
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
		if !sel.Matches(labels.Set(oriVolume.Metadata.Labels)) {
			continue
		}

		res = append(res, oriVolume)
	}
	return res
}

func (s *Server) getVolume(ctx context.Context, id string) (*ori.Volume, error) {
	onmetalVolume, err := s.getAggregateOnmetalVolume(ctx, id)
	if err != nil {
		return nil, err
	}

	return s.convertAggregateOnmetalVolume(onmetalVolume)
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
