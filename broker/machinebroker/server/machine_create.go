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
	commonv1alpha1 "github.com/onmetal/onmetal-api/api/common/v1alpha1"
	computev1alpha1 "github.com/onmetal/onmetal-api/api/compute/v1alpha1"
	"github.com/onmetal/onmetal-api/broker/common/cleaner"
	machinebrokerv1alpha1 "github.com/onmetal/onmetal-api/broker/machinebroker/api/v1alpha1"
	"github.com/onmetal/onmetal-api/broker/machinebroker/apiutils"
	ori "github.com/onmetal/onmetal-api/ori/apis/machine/v1alpha1"
	machinepoolletv1alpha1 "github.com/onmetal/onmetal-api/poollet/machinepoollet/api/v1alpha1"
	"github.com/onmetal/onmetal-api/utils/maps"
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

type OnmetalMachineConfig struct {
	Labels                  map[string]string
	Annotations             map[string]string
	Power                   computev1alpha1.Power
	MachineClassName        string
	Image                   string
	IgnitionData            []byte
	NetworkInterfaceConfigs []*OnmetalNetworkInterfaceConfig
	VolumeConfigs           []*OnmetalVolumeConfig
}

func (s *Server) onmetalMachinePoolRef() *corev1.LocalObjectReference {
	if s.cluster.MachinePoolName() == "" {
		return nil
	}
	return &corev1.LocalObjectReference{Name: s.cluster.MachinePoolName()}
}

func (s *Server) prepareOnmetalMachinePower(power ori.Power) (computev1alpha1.Power, error) {
	switch power {
	case ori.Power_POWER_ON:
		return computev1alpha1.PowerOn, nil
	case ori.Power_POWER_OFF:
		return computev1alpha1.PowerOff, nil
	default:
		return "", fmt.Errorf("unknown power state %v", power)
	}
}

func (s *Server) prepareOnmetalMachineLabels(machine *ori.Machine) (map[string]string, error) {
	labels := make(map[string]string)

	for downwardAPILabelName, defaultLabelName := range s.brokerDownwardAPILabels {
		value := machine.GetMetadata().GetLabels()[machinepoolletv1alpha1.DownwardAPILabel(downwardAPILabelName)]
		if value == "" {
			value = machine.GetMetadata().GetLabels()[defaultLabelName]
		}
		if value != "" {
			labels[machinepoolletv1alpha1.DownwardAPILabel(downwardAPILabelName)] = value
		}
	}

	return labels, nil
}

func (s *Server) prepareOnmetalMachineAnnotations(machine *ori.Machine) (map[string]string, error) {
	annotationsValue, err := apiutils.EncodeAnnotationsAnnotation(machine.GetMetadata().GetAnnotations())
	if err != nil {
		return nil, fmt.Errorf("error encoding annotations: %w", err)
	}

	labelsValue, err := apiutils.EncodeLabelsAnnotation(machine.GetMetadata().GetLabels())
	if err != nil {
		return nil, fmt.Errorf("error encoding labels: %w", err)
	}

	return map[string]string{
		machinebrokerv1alpha1.AnnotationsAnnotation: annotationsValue,
		machinebrokerv1alpha1.LabelsAnnotation:      labelsValue,
	}, nil
}

func (s *Server) getOnmetalMachineConfig(machine *ori.Machine) (*OnmetalMachineConfig, error) {
	onmetalPower, err := s.prepareOnmetalMachinePower(machine.Spec.Power)
	if err != nil {
		return nil, err
	}

	var onmetalImage string
	if image := machine.Spec.Image; image != nil {
		onmetalImage = image.Image
	}

	onmetalNicCfgs := make([]*OnmetalNetworkInterfaceConfig, len(machine.Spec.NetworkInterfaces))
	for i, nic := range machine.Spec.NetworkInterfaces {
		onmetalNicCfg, err := s.getOnmetalNetworkInterfaceConfig(nic)
		if err != nil {
			return nil, fmt.Errorf("[network interface %s] %w", nic.Name, err)
		}

		onmetalNicCfgs[i] = onmetalNicCfg
	}

	onmetalVolumeCfgs := make([]*OnmetalVolumeConfig, len(machine.Spec.Volumes))
	for i, volume := range machine.Spec.Volumes {
		onmetalVolumeCfg, err := s.getOnmetalVolumeConfig(volume)
		if err != nil {
			return nil, fmt.Errorf("[volume %s]: %w", volume.Name, err)
		}

		onmetalVolumeCfgs[i] = onmetalVolumeCfg
	}

	labels, err := s.prepareOnmetalMachineLabels(machine)
	if err != nil {
		return nil, fmt.Errorf("error preparing onmetal machine labels: %w", err)
	}

	annotations, err := s.prepareOnmetalMachineAnnotations(machine)
	if err != nil {
		return nil, fmt.Errorf("error preparing onmetal machine annotations: %w", err)
	}

	return &OnmetalMachineConfig{
		Labels:                  labels,
		Annotations:             annotations,
		Power:                   onmetalPower,
		MachineClassName:        machine.Spec.Class,
		Image:                   onmetalImage,
		IgnitionData:            machine.Spec.IgnitionData,
		NetworkInterfaceConfigs: onmetalNicCfgs,
		VolumeConfigs:           onmetalVolumeCfgs,
	}, nil
}

func (s *Server) createOnmetalMachine(
	ctx context.Context,
	log logr.Logger,
	cfg *OnmetalMachineConfig,
) (res *AggregateOnmetalMachine, retErr error) {
	c, cleanup := s.setupCleaner(ctx, log, &retErr)
	defer cleanup()

	var (
		ignitionRef    *commonv1alpha1.SecretKeySelector
		ignitionSecret *corev1.Secret
	)
	if ignitionData := cfg.IgnitionData; len(ignitionData) > 0 {
		log.V(1).Info("Creating onmetal ignition secret")
		ignitionSecret = &corev1.Secret{
			ObjectMeta: metav1.ObjectMeta{
				Namespace: s.cluster.Namespace(),
				Name:      s.cluster.IDGen().Generate(),
			},
			Type: computev1alpha1.SecretTypeIgnition,
			Data: map[string][]byte{
				computev1alpha1.DefaultIgnitionKey: ignitionData,
			},
		}
		if err := s.cluster.Client().Create(ctx, ignitionSecret); err != nil {
			return nil, fmt.Errorf("error creating onmetal ignition secret: %w", err)
		}
		c.Add(cleaner.CleanupObject(s.cluster.Client(), ignitionSecret))

		ignitionRef = &commonv1alpha1.SecretKeySelector{Name: ignitionSecret.Name}
	}

	var (
		onmetalMachineNics []computev1alpha1.NetworkInterface
		aggOnmetalNics     = make(map[string]*AggregateOnmetalNetworkInterface)
	)
	for _, nicCfg := range cfg.NetworkInterfaceConfigs {
		onmetalMachineNic, aggOnmetalNic, err := s.createOnmetalNetworkInterface(ctx, log, c, nil, nicCfg)
		if err != nil {
			return nil, fmt.Errorf("[network interface %s] error creating: %w", nicCfg.Name, err)
		}

		onmetalMachineNics = append(onmetalMachineNics, *onmetalMachineNic)
		aggOnmetalNics[nicCfg.Name] = aggOnmetalNic
	}

	var (
		onmetalMachineVolumes []computev1alpha1.Volume
		aggOnmetalVolumes     = make(map[string]*AggregateOnmetalVolume)
	)
	for _, volumeCfg := range cfg.VolumeConfigs {
		onmetalMachineVolume, aggOnmetalVolume, err := s.createOnmetalVolume(ctx, log, c, nil, volumeCfg)
		if err != nil {
			return nil, fmt.Errorf("[volume %s] error creating: %w", volumeCfg.Name, err)
		}

		onmetalMachineVolumes = append(onmetalMachineVolumes, *onmetalMachineVolume)
		if aggOnmetalVolume != nil {
			aggOnmetalVolumes[volumeCfg.Name] = aggOnmetalVolume
		}
	}

	onmetalMachine := &computev1alpha1.Machine{
		ObjectMeta: metav1.ObjectMeta{
			Namespace:   s.cluster.Namespace(),
			Name:        s.cluster.IDGen().Generate(),
			Annotations: cfg.Annotations,
			Labels: maps.AppendMap(cfg.Labels, map[string]string{
				machinebrokerv1alpha1.ManagerLabel: machinebrokerv1alpha1.MachineBrokerManager,
			}),
		},
		Spec: computev1alpha1.MachineSpec{
			MachineClassRef:     corev1.LocalObjectReference{Name: cfg.MachineClassName},
			MachinePoolSelector: s.cluster.MachinePoolSelector(),
			MachinePoolRef:      s.onmetalMachinePoolRef(),
			Power:               cfg.Power,
			Image:               cfg.Image,
			ImagePullSecretRef:  nil, // TODO: Specify as soon as available.
			NetworkInterfaces:   onmetalMachineNics,
			Volumes:             onmetalMachineVolumes,
			IgnitionRef:         ignitionRef,
		},
	}
	log.V(1).Info("Creating onmetal machine")
	if err := s.cluster.Client().Create(ctx, onmetalMachine); err != nil {
		return nil, fmt.Errorf("error creating onmetal machine: %w", err)
	}
	c.Add(cleaner.CleanupObject(s.cluster.Client(), onmetalMachine))

	if ignitionSecret != nil {
		log.V(1).Info("Patching ignition secret to be controlled by onmetal machine")
		if err := apiutils.PatchControlledBy(ctx, s.cluster.Client(), onmetalMachine, ignitionSecret); err != nil {
			return nil, fmt.Errorf("error patching ignition secret to be controlled by onmetal machine: %w", err)
		}
	}
	for _, aggOnmetalNic := range aggOnmetalNics {
		if err := s.bindOnmetalMachineNetworkInterface(ctx, onmetalMachine, aggOnmetalNic.NetworkInterface); err != nil {
			return nil, fmt.Errorf("error binding onmetal network interface to onmetal machine: %w", err)
		}
	}
	for _, aggOnmetalVolume := range aggOnmetalVolumes {
		if err := s.bindOnmetalMachineVolume(ctx, onmetalMachine, aggOnmetalVolume.Volume); err != nil {
			return nil, fmt.Errorf("error binding onmetal volume to onmetal machine: %w", err)
		}
	}

	log.V(1).Info("Patching onmetal machine as created")
	if err := apiutils.PatchCreated(ctx, s.cluster.Client(), onmetalMachine); err != nil {
		return nil, fmt.Errorf("error patching onmetal machine as created: %w", err)
	}

	return &AggregateOnmetalMachine{
		IgnitionSecret:    ignitionSecret,
		Machine:           onmetalMachine,
		NetworkInterfaces: aggOnmetalNics,
		Volumes:           aggOnmetalVolumes,
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
