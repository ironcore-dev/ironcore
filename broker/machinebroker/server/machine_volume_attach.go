// SPDX-FileCopyrightText: 2023 SAP SE or an SAP affiliate company and IronCore contributors
// SPDX-License-Identifier: Apache-2.0

package server

import (
	"context"
	"fmt"

	"github.com/go-logr/logr"
	commonv1alpha1 "github.com/ironcore-dev/ironcore/api/common/v1alpha1"
	computev1alpha1 "github.com/ironcore-dev/ironcore/api/compute/v1alpha1"
	storagev1alpha1 "github.com/ironcore-dev/ironcore/api/storage/v1alpha1"
	"github.com/ironcore-dev/ironcore/broker/common/cleaner"
	machinebrokerv1alpha1 "github.com/ironcore-dev/ironcore/broker/machinebroker/api/v1alpha1"
	iri "github.com/ironcore-dev/ironcore/iri/apis/machine/v1alpha1"
	metautils "github.com/ironcore-dev/ironcore/utils/meta"
	corev1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/api/resource"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"sigs.k8s.io/controller-runtime/pkg/client"
)

type IronCoreVolumeConfig struct {
	Name      string
	Device    string
	EmptyDisk *IronCoreVolumeEmptyDiskConfig
	Remote    *IronCoreVolumeRemoteConfig
}

type IronCoreVolumeEmptyDiskConfig struct {
	SizeLimit *resource.Quantity
}

type IronCoreVolumeRemoteConfig struct {
	Driver     string
	Handle     string
	Attributes map[string]string
	SecretData map[string][]byte
}

func (s *Server) getIronCoreVolumeConfig(volume *iri.Volume) (*IronCoreVolumeConfig, error) {
	var (
		emptyDisk *IronCoreVolumeEmptyDiskConfig
		remote    *IronCoreVolumeRemoteConfig
	)
	switch {
	case volume.EmptyDisk != nil:
		var sizeLimit *resource.Quantity
		if sizeBytes := volume.EmptyDisk.SizeBytes; sizeBytes > 0 {
			sizeLimit = resource.NewQuantity(int64(sizeBytes), resource.DecimalSI)
		}
		emptyDisk = &IronCoreVolumeEmptyDiskConfig{
			SizeLimit: sizeLimit,
		}
	case volume.Connection != nil:
		remote = &IronCoreVolumeRemoteConfig{
			Driver:     volume.Connection.Driver,
			Handle:     volume.Connection.Handle,
			Attributes: volume.Connection.Attributes,
			SecretData: volume.Connection.SecretData,
		}
	default:
		return nil, fmt.Errorf("unrecognized volume %#v", volume)
	}

	return &IronCoreVolumeConfig{
		Name:      volume.Name,
		Device:    volume.Device,
		EmptyDisk: emptyDisk,
		Remote:    remote,
	}, nil
}

func (s *Server) createIronCoreVolume(
	ctx context.Context,
	log logr.Logger,
	c *cleaner.Cleaner,
	optIronCoreMachine client.Object,
	cfg *IronCoreVolumeConfig,
) (ironcoreMachineVolume *computev1alpha1.Volume, aggIronCoreVolume *AggregateIronCoreVolume, retErr error) {
	var ironcoreVolumeSrc computev1alpha1.VolumeSource
	switch {
	case cfg.Remote != nil:
		log.V(1).Info("Creating ironcore volume")
		remote := cfg.Remote
		ironcoreVolume := &storagev1alpha1.Volume{
			ObjectMeta: metav1.ObjectMeta{
				Namespace:       s.cluster.Namespace(),
				Name:            s.cluster.IDGen().Generate(),
				OwnerReferences: s.optionalOwnerReferences(ironcoreMachineGVK, optIronCoreMachine),
				Annotations: map[string]string{
					commonv1alpha1.ManagedByAnnotation: machinebrokerv1alpha1.MachineBrokerManager,
				},
				Labels: map[string]string{
					machinebrokerv1alpha1.ManagerLabel: machinebrokerv1alpha1.MachineBrokerManager,
				},
			},
			Spec: storagev1alpha1.VolumeSpec{
				ClaimRef: s.optionalLocalUIDReference(optIronCoreMachine),
			},
		}
		if err := s.cluster.Client().Create(ctx, ironcoreVolume); err != nil {
			return nil, nil, fmt.Errorf("error creating ironcore volume: %w", err)
		}
		c.Add(cleaner.CleanupObject(s.cluster.Client(), ironcoreVolume))

		var (
			secretRef    *corev1.LocalObjectReference
			accessSecret *corev1.Secret
		)
		if secretData := remote.SecretData; secretData != nil {
			log.V(1).Info("Creating ironcore volume secret")
			accessSecret = &corev1.Secret{
				ObjectMeta: metav1.ObjectMeta{
					Namespace: s.cluster.Namespace(),
					Name:      s.cluster.IDGen().Generate(),
					OwnerReferences: []metav1.OwnerReference{
						metautils.MakeControllerRef(
							storagev1alpha1.SchemeGroupVersion.WithKind("Volume"),
							ironcoreVolume,
						),
					},
					Labels: map[string]string{
						machinebrokerv1alpha1.ManagerLabel: machinebrokerv1alpha1.MachineBrokerManager,
					},
				},
				Type: storagev1alpha1.SecretTypeVolumeAuth,
				Data: secretData,
			}
			if err := s.cluster.Client().Create(ctx, accessSecret); err != nil {
				return nil, nil, fmt.Errorf("error creating ironcore volume secret: %w", err)
			}
			c.Add(cleaner.CleanupObject(s.cluster.Client(), accessSecret))
			secretRef = &corev1.LocalObjectReference{Name: accessSecret.Name}
		}

		log.V(1).Info("Patching ironcore volume status")
		baseIronCoreVolume := ironcoreVolume.DeepCopy()
		ironcoreVolume.Status.State = storagev1alpha1.VolumeStateAvailable
		ironcoreVolume.Status.Access = &storagev1alpha1.VolumeAccess{
			SecretRef:        secretRef,
			Driver:           remote.Driver,
			Handle:           remote.Handle,
			VolumeAttributes: remote.Attributes,
		}
		if err := s.cluster.Client().Status().Patch(ctx, ironcoreVolume, client.MergeFrom(baseIronCoreVolume)); err != nil {
			return nil, nil, fmt.Errorf("error patching ironcore volume status: %w", err)
		}

		aggIronCoreVolume = &AggregateIronCoreVolume{
			Volume:       ironcoreVolume,
			AccessSecret: accessSecret,
		}
		ironcoreVolumeSrc.VolumeRef = &corev1.LocalObjectReference{Name: ironcoreVolume.Name}
	case cfg.EmptyDisk != nil:
		ironcoreVolumeSrc.EmptyDisk = &computev1alpha1.EmptyDiskVolumeSource{
			SizeLimit: cfg.EmptyDisk.SizeLimit,
		}
	}
	return &computev1alpha1.Volume{
		Name:         cfg.Name,
		Device:       &cfg.Device,
		VolumeSource: ironcoreVolumeSrc,
	}, aggIronCoreVolume, nil
}

func (s *Server) attachIronCoreVolume(
	ctx context.Context,
	log logr.Logger,
	ironcoreMachine *computev1alpha1.Machine,
	ironcoreMachineVolume *computev1alpha1.Volume,
) error {
	baseIronCoreMachine := ironcoreMachine.DeepCopy()
	ironcoreMachine.Spec.Volumes = append(ironcoreMachine.Spec.Volumes, *ironcoreMachineVolume)
	if err := s.cluster.Client().Patch(ctx, ironcoreMachine, client.StrategicMergeFrom(baseIronCoreMachine)); err != nil {
		return fmt.Errorf("error patching ironcore machine volumes: %w", err)
	}
	return nil
}

func (s *Server) AttachVolume(ctx context.Context, req *iri.AttachVolumeRequest) (res *iri.AttachVolumeResponse, retErr error) {
	machineID := req.MachineId
	volumeName := req.Volume.Name
	log := s.loggerFrom(ctx, "MachineID", machineID, "VolumeName", volumeName)

	log.V(1).Info("Getting ironcore machine")
	ironcoreMachine, err := s.getIronCoreMachine(ctx, machineID)
	if err != nil {
		return nil, err
	}

	log.V(1).Info("Getting ironcore volume config")
	cfg, err := s.getIronCoreVolumeConfig(req.Volume)
	if err != nil {
		return nil, err
	}

	c, cleanup := s.setupCleaner(ctx, log, &retErr)
	defer cleanup()

	log.V(1).Info("Creating ironcore volume")
	ironcoreMachineVolume, _, err := s.createIronCoreVolume(ctx, log, c, ironcoreMachine, cfg)
	if err != nil {
		return nil, err
	}

	log.V(1).Info("Attaching ironcore volume")
	if err := s.attachIronCoreVolume(ctx, log, ironcoreMachine, ironcoreMachineVolume); err != nil {
		return nil, err
	}

	return &iri.AttachVolumeResponse{}, nil
}
