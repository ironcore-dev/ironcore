// SPDX-FileCopyrightText: 2023 SAP SE or an SAP affiliate company and IronCore contributors
// SPDX-License-Identifier: Apache-2.0

package server

import (
	"context"
	"fmt"

	"github.com/go-logr/logr"
	corev1alpha1 "github.com/ironcore-dev/ironcore/api/core/v1alpha1"
	storagev1alpha1 "github.com/ironcore-dev/ironcore/api/storage/v1alpha1"
	brokerutils "github.com/ironcore-dev/ironcore/broker/common/utils"
	volumebrokerv1alpha1 "github.com/ironcore-dev/ironcore/broker/volumebroker/api/v1alpha1"

	"github.com/ironcore-dev/ironcore/broker/volumebroker/apiutils"
	iri "github.com/ironcore-dev/ironcore/iri/apis/volume/v1alpha1"
	volumepoolletv1alpha1 "github.com/ironcore-dev/ironcore/poollet/volumepoollet/api/v1alpha1"

	utilsmaps "github.com/ironcore-dev/ironcore/utils/maps"
	corev1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/api/resource"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"sigs.k8s.io/controller-runtime/pkg/client"
)

type AggregateIronCoreVolume struct {
	Volume           *storagev1alpha1.Volume
	EncryptionSecret *corev1.Secret
	AccessSecret     *corev1.Secret
}

func (s *Server) getIronCoreVolumeConfig(_ context.Context, volume *iri.Volume) (*AggregateIronCoreVolume, error) {
	var volumePoolRef *corev1.LocalObjectReference
	if s.volumePoolName != "" {
		volumePoolRef = &corev1.LocalObjectReference{
			Name: s.volumePoolName,
		}
	}

	var encryptionSecret *corev1.Secret
	if encryption := volume.Spec.Encryption; encryption != nil {
		encryptionSecret = &corev1.Secret{
			ObjectMeta: metav1.ObjectMeta{
				Namespace: s.namespace,
				Name:      s.idGen.Generate(),
			},
			Type: corev1.SecretTypeOpaque,
			Data: encryption.SecretData,
		}
		apiutils.SetPurpose(encryptionSecret, volumebrokerv1alpha1.VolumeEncryptionPurpose)
	}

	var encryption *storagev1alpha1.VolumeEncryption
	if encryptionSecret != nil {
		encryption = &storagev1alpha1.VolumeEncryption{
			SecretRef: corev1.LocalObjectReference{
				Name: encryptionSecret.Name,
			},
		}
	}

	labels := brokerutils.PrepareDownwardAPILabels(
		volume.GetMetadata().GetLabels(),
		s.brokerDownwardAPILabels,
		volumepoolletv1alpha1.VolumeDownwardAPIPrefix,
	)

	var image string
	var volumeSnapshotRef *corev1.LocalObjectReference

	image = volume.Spec.Image // TODO: Remove this once volume.Spec.Image is deprecated

	if dataSource := volume.Spec.VolumeDataSource; dataSource != nil {
		switch {
		case dataSource.SnapshotDataSource != nil:
			volumeSnapshotRef = &corev1.LocalObjectReference{Name: dataSource.SnapshotDataSource.SnapshotId}
			image = "" // TODO: Remove this once volume.Spec.Image is deprecated
		case dataSource.ImageDataSource != nil:
			image = dataSource.ImageDataSource.Image
		}
	}

	ironcoreVolume := &storagev1alpha1.Volume{
		ObjectMeta: metav1.ObjectMeta{
			Namespace: s.namespace,
			Name:      s.idGen.Generate(),
			Labels: utilsmaps.AppendMap(labels, map[string]string{
				volumebrokerv1alpha1.ManagerLabel: volumebrokerv1alpha1.VolumeBrokerManager,
			}),
		},
		Spec: storagev1alpha1.VolumeSpec{
			VolumeClassRef:     &corev1.LocalObjectReference{Name: volume.Spec.Class},
			VolumePoolSelector: s.volumePoolSelector,
			VolumePoolRef:      volumePoolRef,
			Resources: corev1alpha1.ResourceList{
				corev1alpha1.ResourceStorage: *resource.NewQuantity(volume.Spec.Resources.StorageBytes, resource.DecimalSI),
			},
			Image:              image, // TODO: Remove this once volume.Spec.Image is deprecated
			ImagePullSecretRef: nil,   // TODO: Fill if necessary
			Encryption:         encryption,
			VolumeDataSource: storagev1alpha1.VolumeDataSource{
				VolumeSnapshotRef: volumeSnapshotRef,
				OSImage:           getOSImageIfPresent(image),
			},
		},
	}
	if err := apiutils.SetObjectMetadata(ironcoreVolume, volume.Metadata); err != nil {
		return nil, err
	}

	return &AggregateIronCoreVolume{
		Volume:           ironcoreVolume,
		EncryptionSecret: encryptionSecret,
	}, nil
}

func getOSImageIfPresent(image string) *string {
	if image == "" {
		return nil
	}
	return &image
}

func (s *Server) createIronCoreVolume(ctx context.Context, log logr.Logger, volume *AggregateIronCoreVolume) (retErr error) {
	c, cleanup := s.setupCleaner(ctx, log, &retErr)
	defer cleanup()

	if volume.EncryptionSecret != nil {
		log.V(1).Info("Creating ironcore encryption secret")
		if err := s.client.Create(ctx, volume.EncryptionSecret); err != nil {
			return fmt.Errorf("error creating ironcore encryption secret: %w", err)
		}
		c.Add(func(ctx context.Context) error {
			if err := s.client.Delete(ctx, volume.EncryptionSecret); client.IgnoreNotFound(err) != nil {
				return fmt.Errorf("error deleting ironcore encryption secret: %w", err)
			}
			return nil
		})
	}

	log.V(1).Info("Creating ironcore volume")
	if err := s.client.Create(ctx, volume.Volume); err != nil {
		return fmt.Errorf("error creating ironcore volume: %w", err)
	}
	c.Add(func(ctx context.Context) error {
		if err := s.client.Delete(ctx, volume.Volume); client.IgnoreNotFound(err) != nil {
			return fmt.Errorf("error deleting ironcore volume: %w", err)
		}
		return nil
	})

	if volume.EncryptionSecret != nil {
		log.V(1).Info("Patching encryption secret to be controlled by ironcore volume")
		if err := apiutils.PatchControlledBy(ctx, s.client, volume.Volume, volume.EncryptionSecret); err != nil {
			return fmt.Errorf("error patching encryption secret to be controlled by ironcore volume: %w", err)
		}
	}

	log.V(1).Info("Patching ironcore volume as created")
	if err := apiutils.PatchCreated(ctx, s.client, volume.Volume); err != nil {
		return fmt.Errorf("error patching ironcore volume as created: %w", err)
	}

	// Reset cleaner since everything from now on operates on a consistent volume
	c.Reset()

	accessSecret, err := s.getIronCoreVolumeAccessSecretIfRequired(volume.Volume, s.clientGetSecretFunc(ctx))
	if err != nil {
		return err
	}

	volume.AccessSecret = accessSecret
	return nil
}

func (s *Server) CreateVolume(ctx context.Context, req *iri.CreateVolumeRequest) (res *iri.CreateVolumeResponse, retErr error) {
	log := s.loggerFrom(ctx)

	log.V(1).Info("Getting volume configuration")
	cfg, err := s.getIronCoreVolumeConfig(ctx, req.Volume)
	if err != nil {
		return nil, fmt.Errorf("error getting ironcore volume config: %w", err)
	}

	if err := s.createIronCoreVolume(ctx, log, cfg); err != nil {
		return nil, fmt.Errorf("error creating ironcore volume: %w", err)
	}

	v, err := s.convertAggregateIronCoreVolume(cfg)
	if err != nil {
		return nil, err
	}

	return &iri.CreateVolumeResponse{
		Volume: v,
	}, nil
}
