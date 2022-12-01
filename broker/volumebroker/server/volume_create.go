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
	storagev1alpha1 "github.com/onmetal/onmetal-api/api/storage/v1alpha1"
	volumebrokerv1alpha1 "github.com/onmetal/onmetal-api/broker/volumebroker/api/v1alpha1"
	"github.com/onmetal/onmetal-api/broker/volumebroker/apiutils"
	"github.com/onmetal/onmetal-api/broker/volumebroker/cleaner"
	ori "github.com/onmetal/onmetal-api/ori/apis/volume/v1alpha1"
	corev1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/api/resource"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"sigs.k8s.io/controller-runtime/pkg/client"
)

type OnmetalVolumeConfig struct {
	Volume *storagev1alpha1.Volume
}

func (s *Server) getOnmetalVolumeConfig(_ context.Context, cfg *ori.VolumeConfig, volumeID string) (*OnmetalVolumeConfig, error) {
	onmetalVolume := &storagev1alpha1.Volume{
		ObjectMeta: metav1.ObjectMeta{
			Namespace: s.namespace,
			Name:      volumeID,
		},
		Spec: storagev1alpha1.VolumeSpec{
			VolumeClassRef:     &corev1.LocalObjectReference{Name: cfg.Class},
			VolumePoolSelector: s.volumePoolSelector,
			Resources: corev1.ResourceList{
				corev1.ResourceStorage: *resource.NewQuantity(int64(cfg.Resources.StorageBytes), resource.DecimalSI),
			},
			Image:              cfg.Image,
			ImagePullSecretRef: nil, // TODO: Fill if necessary
		},
	}
	apiutils.SetVolumeIDLabel(onmetalVolume, volumeID)
	apiutils.SetVolumeManagerLabel(onmetalVolume, volumebrokerv1alpha1.VolumeBrokerManager)
	if err := apiutils.SetMetadataAnnotation(onmetalVolume, cfg.Metadata); err != nil {
		return nil, err
	}
	if err := apiutils.SetAnnotationsAnnotation(onmetalVolume, cfg.Annotations); err != nil {
		return nil, err
	}
	if err := apiutils.SetLabelsAnnotation(onmetalVolume, cfg.Labels); err != nil {
		return nil, err
	}
	if s.volumePoolName != "" {
		onmetalVolume.Spec.VolumePoolRef = &corev1.LocalObjectReference{Name: s.volumePoolName}
	}

	return &OnmetalVolumeConfig{
		Volume: onmetalVolume,
	}, nil
}

func (s *Server) createOnmetalVolume(ctx context.Context, log logr.Logger, cleaner *cleaner.Cleaner, cfg *OnmetalVolumeConfig) (ori.VolumeState, error) {
	log.V(1).Info("Creating volume")
	onmetalVolume := cfg.Volume
	if err := s.client.Create(ctx, onmetalVolume); err != nil {
		return 0, fmt.Errorf("error creating volume: %w", err)
	}

	cleaner.Add(func(ctx context.Context) error {
		if err := s.client.Delete(ctx, onmetalVolume); client.IgnoreNotFound(err) != nil {
			return fmt.Errorf("error deleting volume: %w", err)
		}
		return nil
	})
	return s.convertOnmetalVolumeState(onmetalVolume.Status.State), nil
}

func (s *Server) CreateVolume(ctx context.Context, req *ori.CreateVolumeRequest) (res *ori.CreateVolumeResponse, retErr error) {
	log := s.loggerFrom(ctx)

	log.V(1).Info("Generating volume id")
	volumeID := s.generateID()
	log = log.WithValues("VolumeID", volumeID)
	log.V(1).Info("Generated volume id")

	cfg := req.Config

	log.V(1).Info("Getting volume configuration")
	onmetalVolumeCfg, err := s.getOnmetalVolumeConfig(ctx, cfg, volumeID)
	if err != nil {
		return nil, fmt.Errorf("error getting onmetal volume config: %w", err)
	}

	c, cleanup := s.setupCleaner(ctx, log, &retErr)
	defer cleanup()

	state, err := s.createOnmetalVolume(ctx, log, c, onmetalVolumeCfg)
	if err != nil {
		return nil, err
	}

	return &ori.CreateVolumeResponse{
		Volume: &ori.Volume{
			Id:          volumeID,
			Metadata:    req.Config.Metadata,
			State:       state,
			Annotations: req.Config.Annotations,
			Labels:      req.Config.Labels,
		},
	}, nil
}
