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
	machinebrokerv1alpha1 "github.com/onmetal/onmetal-api/broker/machinebroker/api/v1alpha1"
	"github.com/onmetal/onmetal-api/broker/machinebroker/apiutils"
	ori "github.com/onmetal/onmetal-api/ori/apis/machine/v1alpha1"
	onmetalapiannotations "github.com/onmetal/onmetal-api/utils/annotations"
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/client"
)

type OnmetalVolumeConfig struct {
	AccessSecret *corev1.Secret
	Volume       *storagev1alpha1.Volume
	Access       *storagev1alpha1.VolumeAccess
}

func (s *Server) getOnmetalVolumeConfig(volume *ori.Volume) (*OnmetalVolumeConfig, error) {
	var onmetalVolumeSecret *corev1.Secret
	if secretData := volume.Spec.SecretData; secretData != nil {
		onmetalVolumeSecret = &corev1.Secret{
			ObjectMeta: metav1.ObjectMeta{
				Namespace: s.cluster.Namespace(),
				Name:      s.cluster.IDGen().Generate(),
			},
			Type: corev1.SecretTypeOpaque,
			Data: secretData,
		}
		apiutils.SetPurpose(onmetalVolumeSecret, machinebrokerv1alpha1.VolumeAccessPurpose)
	}

	onmetalVolume := &storagev1alpha1.Volume{
		ObjectMeta: metav1.ObjectMeta{
			Namespace: s.cluster.Namespace(),
			Name:      s.cluster.IDGen().Generate(),
		},
		Spec: storagev1alpha1.VolumeSpec{},
	}
	if err := apiutils.SetObjectMetadata(onmetalVolume, volume.Metadata); err != nil {
		return nil, err
	}
	apiutils.SetManagerLabel(onmetalVolume, machinebrokerv1alpha1.MachineBrokerManager)
	onmetalapiannotations.SetExternallyMangedBy(onmetalVolume, machinebrokerv1alpha1.MachineBrokerManager) // TODO: Is this really required?

	var onmetalVolumeAccessSecretRef *corev1.LocalObjectReference
	if onmetalVolumeSecret != nil {
		onmetalVolumeAccessSecretRef = &corev1.LocalObjectReference{Name: onmetalVolumeSecret.Name}
	}
	onmetalVolumeAccess := &storagev1alpha1.VolumeAccess{
		Driver:           volume.Spec.Driver,
		Handle:           volume.Spec.Handle,
		VolumeAttributes: volume.Spec.Attributes,
		SecretRef:        onmetalVolumeAccessSecretRef,
	}

	return &OnmetalVolumeConfig{
		AccessSecret: onmetalVolumeSecret,
		Volume:       onmetalVolume,
		Access:       onmetalVolumeAccess,
	}, nil
}

func (s *Server) createOnmetalVolume(ctx context.Context, log logr.Logger, cfg *OnmetalVolumeConfig) (res *AggreagateOnmetalVolume, retErr error) {
	c, cleanup := s.setupCleaner(ctx, log, &retErr)
	defer cleanup()

	log.V(1).Info("Creating onmetal volume")
	if err := s.cluster.Client().Create(ctx, cfg.Volume); err != nil {
		return nil, fmt.Errorf("error creating onmetal volume: %w", err)
	}
	c.Add(func(ctx context.Context) error {
		if err := s.cluster.Client().Delete(ctx, cfg.Volume); client.IgnoreNotFound(err) != nil {
			return fmt.Errorf("error deleting onmetal volume: %w", err)
		}
		return nil
	})

	if cfg.AccessSecret != nil {
		log.V(1).Info("Creating onmetal volume secret")
		_ = ctrl.SetControllerReference(cfg.Volume, cfg.AccessSecret, s.cluster.Client().Scheme())
		if err := s.cluster.Client().Create(ctx, cfg.AccessSecret); err != nil {
			return nil, fmt.Errorf("error creating onmetal volume secret: %w", err)
		}
	}

	log.V(1).Info("Patching onmetal volume access")
	baseVolume := cfg.Volume.DeepCopy()
	cfg.Volume.Status.State = storagev1alpha1.VolumeStateAvailable
	cfg.Volume.Status.Access = cfg.Access
	if err := s.cluster.Client().Status().Patch(ctx, cfg.Volume, client.MergeFrom(baseVolume)); err != nil {
		return nil, fmt.Errorf("error patching onmetal volume access: %w", err)
	}

	log.V(1).Info("Patching onmetal volume to created")
	if err := apiutils.PatchCreated(ctx, s.cluster.Client(), cfg.Volume); err != nil {
		return nil, fmt.Errorf("error patching onmetal volume to created: %w", err)
	}

	return &AggreagateOnmetalVolume{
		Volume:       cfg.Volume,
		AccessSecret: cfg.AccessSecret,
	}, nil
}

func (s *Server) CreateVolume(ctx context.Context, req *ori.CreateVolumeRequest) (*ori.CreateVolumeResponse, error) {
	log := s.loggerFrom(ctx)

	cfg, err := s.getOnmetalVolumeConfig(req.Volume)
	if err != nil {
		return nil, fmt.Errorf("error getting onmetal volume config: %w", err)
	}

	onmetalVolume, err := s.createOnmetalVolume(ctx, log, cfg)
	if err != nil {
		return nil, fmt.Errorf("error creating onmetal volume: %w", err)
	}

	v, err := s.convertOnmetalVolume(onmetalVolume)
	if err != nil {
		return nil, fmt.Errorf("error converting onmetal volume: %w", err)
	}

	return &ori.CreateVolumeResponse{
		Volume: v,
	}, nil
}
