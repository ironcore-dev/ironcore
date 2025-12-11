// SPDX-FileCopyrightText: 2023 SAP SE or an SAP affiliate company and IronCore contributors
// SPDX-License-Identifier: Apache-2.0

package server

import (
	"context"
	"fmt"

	storagev1alpha1 "github.com/ironcore-dev/ironcore/api/storage/v1alpha1"
	"github.com/ironcore-dev/ironcore/broker/common"
	volumebrokerv1alpha1 "github.com/ironcore-dev/ironcore/broker/volumebroker/api/v1alpha1"
	iri "github.com/ironcore-dev/ironcore/iri/apis/volume/v1alpha1"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
	corev1 "k8s.io/api/core/v1"
	apierrors "k8s.io/apimachinery/pkg/api/errors"
	"sigs.k8s.io/controller-runtime/pkg/client"
)

func (s *Server) listManagedAndCreated(ctx context.Context, ironcoreVolumeList *storagev1alpha1.VolumeList, filter *iri.VolumeFilter) error {
	matchingLabels := client.MatchingLabels{
		volumebrokerv1alpha1.ManagerLabel: volumebrokerv1alpha1.VolumeBrokerManager,
		volumebrokerv1alpha1.CreatedLabel: "true",
	}

	if filter != nil && filter.LabelSelector != nil {
		for k := range filter.LabelSelector {
			matchingLabels[k] = filter.LabelSelector[k]
		}
	}

	return s.client.List(ctx, ironcoreVolumeList,
		client.InNamespace(s.namespace),
		matchingLabels,
	)
}

func (s *Server) listAggregateIronCoreVolumes(ctx context.Context, filter *iri.VolumeFilter) ([]AggregateIronCoreVolume, error) {
	ironcoreVolumeList := &storagev1alpha1.VolumeList{}
	if err := s.listManagedAndCreated(ctx, ironcoreVolumeList, filter); err != nil {
		return nil, fmt.Errorf("error listing ironcore volumes: %w", err)
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

	var res []AggregateIronCoreVolume
	for i := range ironcoreVolumeList.Items {
		ironcoreVolume := &ironcoreVolumeList.Items[i]

		aggregateIronCoreVolume, err := s.aggregateIronCoreVolume(ironcoreVolume, secretByNameGetter.Get)
		if err != nil {
			return nil, fmt.Errorf("error aggregating ironcore volume %s: %w", ironcoreVolume.Name, err)
		}

		res = append(res, *aggregateIronCoreVolume)
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

func (s *Server) getIronCoreVolumeAccessSecretIfRequired(
	ironcoreVolume *storagev1alpha1.Volume,
	getSecret func(string) (*corev1.Secret, error),
) (*corev1.Secret, error) {
	if ironcoreVolume.Status.State != storagev1alpha1.VolumeStateAvailable {
		return nil, nil
	}

	access := ironcoreVolume.Status.Access
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

func (s *Server) getIronCoreVolumeEncryptionSecretIfRequired(
	ironcoreVolume *storagev1alpha1.Volume,
	getSecret func(string) (*corev1.Secret, error),
) (*corev1.Secret, error) {
	if ironcoreVolume.Spec.Encryption == nil {
		return nil, nil
	}

	secretName := ironcoreVolume.Spec.Encryption.SecretRef.Name
	return getSecret(secretName)
}

func (s *Server) aggregateIronCoreVolume(
	ironcoreVolume *storagev1alpha1.Volume,
	getSecret func(string) (*corev1.Secret, error),
) (*AggregateIronCoreVolume, error) {
	accessSecret, err := s.getIronCoreVolumeAccessSecretIfRequired(ironcoreVolume, getSecret)
	if err != nil {
		return nil, fmt.Errorf("error getting ironcore volume access secret: %w", err)
	}

	encryptionSecret, err := s.getIronCoreVolumeEncryptionSecretIfRequired(ironcoreVolume, getSecret)
	if err != nil {
		return nil, fmt.Errorf("error getting ironcore volume access secret: %w", err)
	}

	return &AggregateIronCoreVolume{
		Volume:           ironcoreVolume,
		EncryptionSecret: encryptionSecret,
		AccessSecret:     accessSecret,
	}, nil
}

func (s *Server) getAggregateIronCoreVolume(ctx context.Context, id string) (*AggregateIronCoreVolume, error) {
	ironcoreVolume := &storagev1alpha1.Volume{}
	if err := s.getManagedAndCreated(ctx, id, ironcoreVolume); err != nil {
		if !apierrors.IsNotFound(err) {
			return nil, fmt.Errorf("error getting ironcore volume %s: %w", id, err)
		}
		return nil, status.Errorf(codes.NotFound, "volume %s not found", id)
	}

	return s.aggregateIronCoreVolume(ironcoreVolume, s.clientGetSecretFunc(ctx))
}

func (s *Server) listVolumes(ctx context.Context, filter *iri.VolumeFilter) ([]*iri.Volume, error) {
	ironcoreVolumes, err := s.listAggregateIronCoreVolumes(ctx, filter)
	if err != nil {
		return nil, fmt.Errorf("error listing volumes: %w", err)
	}

	var res []*iri.Volume
	for _, ironcoreVolume := range ironcoreVolumes {
		volume, err := s.convertAggregateIronCoreVolume(&ironcoreVolume)
		if err != nil {
			return nil, err
		}

		res = append(res, volume)
	}
	return res, nil
}

func (s *Server) getVolume(ctx context.Context, id string) (*iri.Volume, error) {
	ironcoreVolume, err := s.getAggregateIronCoreVolume(ctx, id)
	if err != nil {
		return nil, err
	}

	return s.convertAggregateIronCoreVolume(ironcoreVolume)
}

func (s *Server) ListVolumes(ctx context.Context, req *iri.ListVolumesRequest) (*iri.ListVolumesResponse, error) {
	if filter := req.Filter; filter != nil && filter.Id != "" {
		volume, err := s.getVolume(ctx, filter.Id)
		if err != nil {
			if status.Code(err) != codes.NotFound {
				return nil, err
			}
			return &iri.ListVolumesResponse{
				Volumes: []*iri.Volume{},
			}, nil
		}

		return &iri.ListVolumesResponse{
			Volumes: []*iri.Volume{volume},
		}, nil
	}

	volumes, err := s.listVolumes(ctx, req.Filter)
	if err != nil {
		return nil, err
	}

	return &iri.ListVolumesResponse{
		Volumes: volumes,
	}, nil
}
