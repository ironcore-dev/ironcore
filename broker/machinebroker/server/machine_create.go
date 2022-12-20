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
	"math/big"

	"github.com/go-logr/logr"
	commonv1alpha1 "github.com/onmetal/onmetal-api/api/common/v1alpha1"
	computev1alpha1 "github.com/onmetal/onmetal-api/api/compute/v1alpha1"
	machinebrokerv1alpha1 "github.com/onmetal/onmetal-api/broker/machinebroker/api/v1alpha1"
	"github.com/onmetal/onmetal-api/broker/machinebroker/apiutils"
	ori "github.com/onmetal/onmetal-api/ori/apis/machine/v1alpha1"
	"gopkg.in/inf.v0"
	corev1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/api/resource"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"sigs.k8s.io/controller-runtime/pkg/client"
)

type OnmetalMachineConfig struct {
	IgnitionSecret *corev1.Secret
	Machine        *computev1alpha1.Machine
}

func (s *Server) onmetalMachinePoolRef() *corev1.LocalObjectReference {
	if s.cluster.MachinePoolName() == "" {
		return nil
	}
	return &corev1.LocalObjectReference{Name: s.cluster.MachinePoolName()}
}

func (s *Server) prepareOnmetalVolumeAttachment(volumeSpec *ori.VolumeAttachment) (computev1alpha1.Volume, error) {
	var src computev1alpha1.VolumeSource
	switch {
	case volumeSpec.VolumeId != "":
		src = computev1alpha1.VolumeSource{
			VolumeRef: &corev1.LocalObjectReference{Name: volumeSpec.VolumeId},
		}
	case volumeSpec.EmptyDisk != nil:
		var sizeLimit *resource.Quantity
		if sizeBytes := volumeSpec.EmptyDisk.SizeBytes; sizeBytes > 0 {
			n := new(big.Int).SetUint64(sizeBytes)
			b := inf.NewDecBig(n, 0)
			sizeLimit = resource.NewDecimalQuantity(*b, resource.DecimalSI)
		}

		src = computev1alpha1.VolumeSource{
			EmptyDisk: &computev1alpha1.EmptyDiskVolumeSource{
				SizeLimit: sizeLimit,
			},
		}
	default:
		return computev1alpha1.Volume{}, fmt.Errorf("volume does neither specify empty disk nor volume id")
	}
	return computev1alpha1.Volume{
		Name:         volumeSpec.Name,
		Device:       &volumeSpec.Device,
		VolumeSource: src,
	}, nil
}

func (s *Server) prepareOnmetalNetworkInterfaceAttachment(networkInterfaceSpec *ori.NetworkInterfaceAttachment) (computev1alpha1.NetworkInterface, error) {
	var src computev1alpha1.NetworkInterfaceSource
	switch {
	case networkInterfaceSpec.NetworkInterfaceId != "":
		src = computev1alpha1.NetworkInterfaceSource{
			NetworkInterfaceRef: &corev1.LocalObjectReference{Name: networkInterfaceSpec.NetworkInterfaceId},
		}
	default:
		return computev1alpha1.NetworkInterface{}, fmt.Errorf("network interface does not specify network interface id")
	}
	return computev1alpha1.NetworkInterface{
		Name:                   networkInterfaceSpec.Name,
		NetworkInterfaceSource: src,
	}, nil
}

func (s *Server) getOnmetalMachineConfig(machine *ori.Machine) (*OnmetalMachineConfig, error) {
	var onmetalIgnitionSecret *corev1.Secret
	if ignition := machine.Spec.Ignition; ignition != nil {
		onmetalIgnitionSecret = &corev1.Secret{
			ObjectMeta: metav1.ObjectMeta{
				Namespace: s.cluster.Namespace(),
				Name:      s.cluster.IDGen().Generate(),
			},
			Type: corev1.SecretTypeOpaque,
			Data: map[string][]byte{
				computev1alpha1.DefaultIgnitionKey: ignition.Data,
			},
		}
		apiutils.SetPurpose(onmetalIgnitionSecret, machinebrokerv1alpha1.IgnitionPurpose)
	}

	var onmetalImage string
	if image := machine.Spec.Image; image != nil {
		onmetalImage = image.Image
	}

	onmetalNetworkInterfaceAttachments := make([]computev1alpha1.NetworkInterface, len(machine.Spec.NetworkInterfaces))
	for i, networkInterface := range machine.Spec.NetworkInterfaces {
		onmetalNetworkInterfaceAttachments[i] = computev1alpha1.NetworkInterface{
			Name: networkInterface.Name,
			NetworkInterfaceSource: computev1alpha1.NetworkInterfaceSource{
				NetworkInterfaceRef: &corev1.LocalObjectReference{Name: networkInterface.NetworkInterfaceId},
			},
		}
	}

	onmetalVolumeAttachments := make([]computev1alpha1.Volume, len(machine.Spec.Volumes))
	for i, volume := range machine.Spec.Volumes {
		onmetalVolumeAttachment, err := s.prepareOnmetalVolumeAttachment(volume)
		if err != nil {
			return nil, fmt.Errorf("error preparing onmetal machine volume %s: %w", volume.Device, err)
		}

		onmetalVolumeAttachments[i] = onmetalVolumeAttachment
	}

	var onmetalMachineIgnitionRef *commonv1alpha1.SecretKeySelector
	if onmetalIgnitionSecret != nil {
		onmetalMachineIgnitionRef = &commonv1alpha1.SecretKeySelector{
			Name: onmetalIgnitionSecret.Name,
			Key:  computev1alpha1.DefaultIgnitionKey,
		}
	}

	onmetalMachine := &computev1alpha1.Machine{
		ObjectMeta: metav1.ObjectMeta{
			Namespace: s.cluster.Namespace(),
			Name:      s.cluster.IDGen().Generate(),
		},
		Spec: computev1alpha1.MachineSpec{
			MachineClassRef:     corev1.LocalObjectReference{Name: machine.Spec.Class},
			MachinePoolSelector: s.cluster.MachinePoolSelector(),
			MachinePoolRef:      s.onmetalMachinePoolRef(),
			Image:               onmetalImage,
			ImagePullSecretRef:  nil, // TODO: Specify if required.
			NetworkInterfaces:   onmetalNetworkInterfaceAttachments,
			Volumes:             onmetalVolumeAttachments,
			IgnitionRef:         onmetalMachineIgnitionRef,
		},
	}
	if err := apiutils.SetObjectMetadata(onmetalMachine, machine.Metadata); err != nil {
		return nil, err
	}
	apiutils.SetManagerLabel(onmetalMachine, machinebrokerv1alpha1.MachineBrokerManager)

	return &OnmetalMachineConfig{
		IgnitionSecret: onmetalIgnitionSecret,
		Machine:        onmetalMachine,
	}, nil
}

func (s *Server) createOnmetalMachine(ctx context.Context, log logr.Logger, cfg *OnmetalMachineConfig) (res *AggregateOnmetalMachine, retErr error) {
	c, cleanup := s.setupCleaner(ctx, log, &retErr)
	defer cleanup()

	if cfg.IgnitionSecret != nil {
		log.V(1).Info("Creating onmetal ignition secret")
		if err := s.cluster.Client().Create(ctx, cfg.IgnitionSecret); err != nil {
			return nil, fmt.Errorf("error onmetal creating ignition secret: %w", err)
		}

		c.Add(func(ctx context.Context) error {
			if err := s.cluster.Client().Delete(ctx, cfg.IgnitionSecret); client.IgnoreNotFound(err) != nil {
				return fmt.Errorf("error onmetal deleting ignition secret: %w", err)
			}
			return nil
		})
	}

	log.V(1).Info("Creating onmetal machine")
	if err := s.cluster.Client().Create(ctx, cfg.Machine); err != nil {
		return nil, fmt.Errorf("error creating onmetal machine: %w", err)
	}
	c.Add(func(ctx context.Context) error {
		if err := s.cluster.Client().Delete(ctx, cfg.Machine); client.IgnoreNotFound(err) != nil {
			return fmt.Errorf("error deleting onmetal machine: %w", err)
		}
		return nil
	})

	if cfg.IgnitionSecret != nil {
		log.V(1).Info("Patching ignition secret to be controlled by onmetal machine")
		if err := apiutils.PatchControlledBy(ctx, s.cluster.Client(), cfg.Machine, cfg.IgnitionSecret); err != nil {
			return nil, fmt.Errorf("error patching ignition secret to be controlled by onmetal machine: %w", err)
		}
	}

	log.V(1).Info("Patching onmetal machine as created")
	if err := apiutils.PatchCreated(ctx, s.cluster.Client(), cfg.Machine); err != nil {
		return nil, fmt.Errorf("error patching onmetal machine as created: %w", err)
	}

	return &AggregateOnmetalMachine{
		IgnitionSecret: cfg.IgnitionSecret,
		Machine:        cfg.Machine,
	}, nil
}

func (s *Server) CreateMachine(ctx context.Context, req *ori.CreateMachineRequest) (res *ori.CreateMachineResponse, retErr error) {
	log := s.loggerFrom(ctx)

	log.V(1).Info("Getting onmetal machine config")
	cfg, err := s.getOnmetalMachineConfig(req.Machine)
	if err != nil {
		return nil, fmt.Errorf("error getting onmetal machine config: %w", err)
	}

	log.V(1).Info("Creating onmetal machine")
	onmetalMachine, err := s.createOnmetalMachine(ctx, log, cfg)
	if err != nil {
		return nil, fmt.Errorf("error creating onmetal machine: %w", err)
	}

	m, err := s.convertAggregateOnmetalMachine(onmetalMachine)
	if err != nil {
		return nil, fmt.Errorf("error converting onmetal machine: %w", err)
	}

	return &ori.CreateMachineResponse{
		Machine: m,
	}, nil
}
