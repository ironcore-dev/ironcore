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

	"github.com/go-logr/logr"
	corev1alpha1 "github.com/onmetal/onmetal-api/api/core/v1alpha1"
	storagev1alpha1 "github.com/onmetal/onmetal-api/api/storage/v1alpha1"
	volumebrokerv1alpha1 "github.com/onmetal/onmetal-api/broker/volumebroker/api/v1alpha1"
	"github.com/onmetal/onmetal-api/broker/volumebroker/apiutils"
	ori "github.com/onmetal/onmetal-api/ori/apis/volume/v1alpha1"
	corev1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/api/resource"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"sigs.k8s.io/controller-runtime/pkg/client"
)

type AggregateOnmetalVolume struct {
	Volume           *storagev1alpha1.Volume
	EncryptionSecret *corev1.Secret
	AccessSecret     *corev1.Secret
}

func (s *Server) getOnmetalVolumeConfig(_ context.Context, volume *ori.Volume) (*AggregateOnmetalVolume, error) {
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
	}

	var encryption *storagev1alpha1.VolumeEncryption
	if encryptionSecret != nil {
		encryption = &storagev1alpha1.VolumeEncryption{
			SecretRef: corev1.LocalObjectReference{
				Name: encryptionSecret.Name,
			},
		}
	}

	onmetalVolume := &storagev1alpha1.Volume{
		ObjectMeta: metav1.ObjectMeta{
			Namespace: s.namespace,
			Name:      s.idGen.Generate(),
		},
		Spec: storagev1alpha1.VolumeSpec{
			VolumeClassRef:     &corev1.LocalObjectReference{Name: volume.Spec.Class},
			VolumePoolSelector: s.volumePoolSelector,
			VolumePoolRef:      volumePoolRef,
			Resources: corev1alpha1.ResourceList{
				corev1alpha1.ResourceStorage: *resource.NewQuantity(int64(volume.Spec.Resources.StorageBytes), resource.DecimalSI),
			},
			Image:              volume.Spec.Image,
			ImagePullSecretRef: nil, // TODO: Fill if necessary
			Encryption:         encryption,
		},
	}
	if err := apiutils.SetObjectMetadata(onmetalVolume, volume.Metadata); err != nil {
		return nil, err
	}
	apiutils.SetVolumeManagerLabel(onmetalVolume, volumebrokerv1alpha1.VolumeBrokerManager)

	return &AggregateOnmetalVolume{
		Volume:           onmetalVolume,
		EncryptionSecret: encryptionSecret,
	}, nil
}

func (s *Server) createOnmetalVolume(ctx context.Context, log logr.Logger, volume *AggregateOnmetalVolume) (retErr error) {
	c, cleanup := s.setupCleaner(ctx, log, &retErr)
	defer cleanup()

	if volume.EncryptionSecret != nil {
		log.V(1).Info("Creating onmetal encryption secret")
		if err := s.client.Create(ctx, volume.EncryptionSecret); err != nil {
			return fmt.Errorf("error creating onmetal encryption secret: %w", err)
		}
		c.Add(func(ctx context.Context) error {
			if err := s.client.Delete(ctx, volume.EncryptionSecret); client.IgnoreNotFound(err) != nil {
				return fmt.Errorf("error deleting onmetal encryption secret: %w", err)
			}
			return nil
		})
	}

	log.V(1).Info("Creating onmetal volume")
	if err := s.client.Create(ctx, volume.Volume); err != nil {
		return fmt.Errorf("error creating onmetal volume: %w", err)
	}
	c.Add(func(ctx context.Context) error {
		if err := s.client.Delete(ctx, volume.Volume); client.IgnoreNotFound(err) != nil {
			return fmt.Errorf("error deleting onmetal volume: %w", err)
		}
		return nil
	})

	log.V(1).Info("Patching onmetal volume as created")
	if err := apiutils.PatchCreated(ctx, s.client, volume.Volume); err != nil {
		return fmt.Errorf("error patching onmetal machine as created: %w", err)
	}

	// Reset cleaner since everything from now on operates on a consistent volume
	c.Reset()

	accessSecret, err := s.getOnmetalVolumeAccessSecretIfRequired(volume.Volume, s.clientGetSecretFunc(ctx))
	if err != nil {
		return err
	}

	volume.AccessSecret = accessSecret
	return nil
}

func (s *Server) CreateVolume(ctx context.Context, req *ori.CreateVolumeRequest) (res *ori.CreateVolumeResponse, retErr error) {
	log := s.loggerFrom(ctx)

	log.V(1).Info("Getting volume configuration")
	cfg, err := s.getOnmetalVolumeConfig(ctx, req.Volume)
	if err != nil {
		return nil, fmt.Errorf("error getting onmetal volume config: %w", err)
	}

	if err := s.createOnmetalVolume(ctx, log, cfg); err != nil {
		return nil, fmt.Errorf("error creating onmetal volume: %w", err)
	}

	v, err := s.convertAggregateOnmetalVolume(cfg)
	if err != nil {
		return nil, err
	}

	return &ori.CreateVolumeResponse{
		Volume: v,
	}, nil
}
