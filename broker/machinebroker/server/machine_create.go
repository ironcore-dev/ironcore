// SPDX-FileCopyrightText: 2023 SAP SE or an SAP affiliate company and IronCore contributors
// SPDX-License-Identifier: Apache-2.0

package server

import (
	"context"
	"fmt"

	"github.com/go-logr/logr"
	commonv1alpha1 "github.com/ironcore-dev/ironcore/api/common/v1alpha1"
	computev1alpha1 "github.com/ironcore-dev/ironcore/api/compute/v1alpha1"
	"github.com/ironcore-dev/ironcore/broker/common/cleaner"
	machinebrokerv1alpha1 "github.com/ironcore-dev/ironcore/broker/machinebroker/api/v1alpha1"
	"github.com/ironcore-dev/ironcore/broker/machinebroker/apiutils"
	machinepoolletv1alpha1 "github.com/ironcore-dev/ironcore/poollet/machinepoollet/api/v1alpha1"

	poolletutils "github.com/ironcore-dev/ironcore/utils/poollet"

	iri "github.com/ironcore-dev/ironcore/iri/apis/machine/v1alpha1"
	"github.com/ironcore-dev/ironcore/utils/maps"
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

type IronCoreMachineConfig struct {
	Labels                  map[string]string
	Annotations             map[string]string
	Power                   computev1alpha1.Power
	MachineClassName        string
	Image                   string
	IgnitionData            []byte
	NetworkInterfaceConfigs []*IronCoreNetworkInterfaceConfig
	VolumeConfigs           []*IronCoreVolumeConfig
}

func (s *Server) ironcoreMachinePoolRef() *corev1.LocalObjectReference {
	if s.cluster.MachinePoolName() == "" {
		return nil
	}
	return &corev1.LocalObjectReference{Name: s.cluster.MachinePoolName()}
}

func (s *Server) prepareIronCoreMachinePower(power iri.Power) (computev1alpha1.Power, error) {
	switch power {
	case iri.Power_POWER_ON:
		return computev1alpha1.PowerOn, nil
	case iri.Power_POWER_OFF:
		return computev1alpha1.PowerOff, nil
	default:
		return "", fmt.Errorf("unknown power state %v", power)
	}
}

func (s *Server) prepareIronCoreMachineLabels(machine *iri.Machine) map[string]string {
	labels := make(map[string]string)

	for downwardAPILabelName, defaultLabelName := range s.brokerDownwardAPILabels {
		value := machine.GetMetadata().GetLabels()[poolletutils.DownwardAPILabel(machinepoolletv1alpha1.MachineDownwardAPIPrefix, downwardAPILabelName)]
		if value == "" {
			value = machine.GetMetadata().GetLabels()[defaultLabelName]
		}
		if value != "" {
			labels[poolletutils.DownwardAPILabel(machinepoolletv1alpha1.MachineDownwardAPIPrefix, downwardAPILabelName)] = value
		}
	}

	return labels
}

func (s *Server) prepareIronCoreMachineAnnotations(machine *iri.Machine) (map[string]string, error) {
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

func (s *Server) getIronCoreMachineConfig(machine *iri.Machine) (*IronCoreMachineConfig, error) {
	ironcorePower, err := s.prepareIronCoreMachinePower(machine.Spec.Power)
	if err != nil {
		return nil, err
	}

	var ironcoreImage string
	if image := machine.Spec.Image; image != nil {
		ironcoreImage = image.Image
	}

	ironcoreNicCfgs := make([]*IronCoreNetworkInterfaceConfig, len(machine.Spec.NetworkInterfaces))
	for i, nic := range machine.Spec.NetworkInterfaces {
		ironcoreNicCfg, err := s.getIronCoreNetworkInterfaceConfig(nic)
		if err != nil {
			return nil, fmt.Errorf("[network interface %s] %w", nic.Name, err)
		}

		ironcoreNicCfgs[i] = ironcoreNicCfg
	}

	ironcoreVolumeCfgs := make([]*IronCoreVolumeConfig, len(machine.Spec.Volumes))
	for i, volume := range machine.Spec.Volumes {
		ironcoreVolumeCfg, err := s.getIronCoreVolumeConfig(volume)
		if err != nil {
			return nil, fmt.Errorf("[volume %s]: %w", volume.Name, err)
		}

		ironcoreVolumeCfgs[i] = ironcoreVolumeCfg
	}

	labels := s.prepareIronCoreMachineLabels(machine)
	annotations, err := s.prepareIronCoreMachineAnnotations(machine)
	if err != nil {
		return nil, fmt.Errorf("error preparing ironcore machine annotations: %w", err)
	}

	return &IronCoreMachineConfig{
		Labels:                  labels,
		Annotations:             annotations,
		Power:                   ironcorePower,
		MachineClassName:        machine.Spec.Class,
		Image:                   ironcoreImage,
		IgnitionData:            machine.Spec.IgnitionData,
		NetworkInterfaceConfigs: ironcoreNicCfgs,
		VolumeConfigs:           ironcoreVolumeCfgs,
	}, nil
}

func (s *Server) createIronCoreMachine(
	ctx context.Context,
	log logr.Logger,
	cfg *IronCoreMachineConfig,
) (res *AggregateIronCoreMachine, retErr error) {
	c, cleanup := s.setupCleaner(ctx, log, &retErr)
	defer cleanup()

	var (
		ignitionRef    *commonv1alpha1.SecretKeySelector
		ignitionSecret *corev1.Secret
	)
	if ignitionData := cfg.IgnitionData; len(ignitionData) > 0 {
		log.V(1).Info("Creating ironcore ignition secret")
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
			return nil, fmt.Errorf("error creating ironcore ignition secret: %w", err)
		}
		c.Add(cleaner.CleanupObject(s.cluster.Client(), ignitionSecret))

		ignitionRef = &commonv1alpha1.SecretKeySelector{Name: ignitionSecret.Name}
	}

	var (
		ironcoreMachineNics []computev1alpha1.NetworkInterface
		aggIronCoreNics     = make(map[string]*AggregateIronCoreNetworkInterface)
	)
	for _, nicCfg := range cfg.NetworkInterfaceConfigs {
		ironcoreMachineNic, aggIronCoreNic, err := s.createIronCoreNetworkInterface(ctx, log, c, nil, nicCfg)
		if err != nil {
			return nil, fmt.Errorf("[network interface %s] error creating: %w", nicCfg.Name, err)
		}

		ironcoreMachineNics = append(ironcoreMachineNics, *ironcoreMachineNic)
		aggIronCoreNics[nicCfg.Name] = aggIronCoreNic
	}

	var (
		ironcoreMachineVolumes []computev1alpha1.Volume
		aggIronCoreVolumes     = make(map[string]*AggregateIronCoreVolume)
	)
	for _, volumeCfg := range cfg.VolumeConfigs {
		ironcoreMachineVolume, aggIronCoreVolume, err := s.createIronCoreVolume(ctx, log, c, nil, volumeCfg)
		if err != nil {
			return nil, fmt.Errorf("[volume %s] error creating: %w", volumeCfg.Name, err)
		}

		ironcoreMachineVolumes = append(ironcoreMachineVolumes, *ironcoreMachineVolume)
		if aggIronCoreVolume != nil {
			aggIronCoreVolumes[volumeCfg.Name] = aggIronCoreVolume
		}
	}

	ironcoreMachine := &computev1alpha1.Machine{
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
			MachinePoolRef:      s.ironcoreMachinePoolRef(),
			Power:               cfg.Power,
			Image:               cfg.Image,
			ImagePullSecretRef:  nil, // TODO: Specify as soon as available.
			NetworkInterfaces:   ironcoreMachineNics,
			Volumes:             ironcoreMachineVolumes,
			IgnitionRef:         ignitionRef,
		},
	}
	log.V(1).Info("Creating ironcore machine")
	if err := s.cluster.Client().Create(ctx, ironcoreMachine); err != nil {
		return nil, fmt.Errorf("error creating ironcore machine: %w", err)
	}
	c.Add(cleaner.CleanupObject(s.cluster.Client(), ironcoreMachine))

	if ignitionSecret != nil {
		log.V(1).Info("Patching ignition secret to be controlled by ironcore machine")
		if err := apiutils.PatchControlledBy(ctx, s.cluster.Client(), ironcoreMachine, ignitionSecret); err != nil {
			return nil, fmt.Errorf("error patching ignition secret to be controlled by ironcore machine: %w", err)
		}
	}
	for _, aggIronCoreNic := range aggIronCoreNics {
		if err := s.bindIronCoreMachineNetworkInterface(ctx, ironcoreMachine, aggIronCoreNic.NetworkInterface); err != nil {
			return nil, fmt.Errorf("error binding ironcore network interface to ironcore machine: %w", err)
		}
	}
	for _, aggIronCoreVolume := range aggIronCoreVolumes {
		if err := s.bindIronCoreMachineVolume(ctx, ironcoreMachine, aggIronCoreVolume.Volume); err != nil {
			return nil, fmt.Errorf("error binding ironcore volume to ironcore machine: %w", err)
		}
	}

	log.V(1).Info("Patching ironcore machine as created")
	if err := apiutils.PatchCreated(ctx, s.cluster.Client(), ironcoreMachine); err != nil {
		return nil, fmt.Errorf("error patching ironcore machine as created: %w", err)
	}

	return &AggregateIronCoreMachine{
		IgnitionSecret:    ignitionSecret,
		Machine:           ironcoreMachine,
		NetworkInterfaces: aggIronCoreNics,
		Volumes:           aggIronCoreVolumes,
	}, nil
}

func (s *Server) CreateMachine(ctx context.Context, req *iri.CreateMachineRequest) (res *iri.CreateMachineResponse, retErr error) {
	log := s.loggerFrom(ctx)

	log.V(1).Info("Getting ironcore machine config")
	cfg, err := s.getIronCoreMachineConfig(req.Machine)
	if err != nil {
		return nil, fmt.Errorf("error getting ironcore machine config: %w", err)
	}

	log.V(1).Info("Creating ironcore machine")
	ironcoreMachine, err := s.createIronCoreMachine(ctx, log, cfg)
	if err != nil {
		return nil, fmt.Errorf("error creating ironcore machine: %w", err)
	}

	m, err := s.convertAggregateIronCoreMachine(ironcoreMachine)
	if err != nil {
		return nil, fmt.Errorf("error converting ironcore machine: %w", err)
	}

	return &iri.CreateMachineResponse{
		Machine: m,
	}, nil
}
