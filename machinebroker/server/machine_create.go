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
	commonv1alpha1 "github.com/onmetal/onmetal-api/apis/common/v1alpha1"
	computev1alpha1 "github.com/onmetal/onmetal-api/apis/compute/v1alpha1"
	networkingv1alpha1 "github.com/onmetal/onmetal-api/apis/networking/v1alpha1"
	storagev1alpha1 "github.com/onmetal/onmetal-api/apis/storage/v1alpha1"
	onmetalapiannotations "github.com/onmetal/onmetal-api/apiutils/annotations"
	machinebrokerv1alpha1 "github.com/onmetal/onmetal-api/machinebroker/api/v1alpha1"
	"github.com/onmetal/onmetal-api/machinebroker/apiutils"
	"github.com/onmetal/onmetal-api/machinebroker/cleaner"
	ori "github.com/onmetal/onmetal-api/ori/apis/runtime/v1alpha1"
	corev1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/api/resource"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	quotav1 "k8s.io/apiserver/pkg/quota/v1"
	"sigs.k8s.io/controller-runtime/pkg/client"
)

type OnmetalVolumeConfig struct {
	Name   string
	Secret *corev1.Secret
	Access *storagev1alpha1.VolumeAccess
	Volume *storagev1alpha1.Volume
}

type OnmetalNetworkConfig struct {
	Network *networkingv1alpha1.Network
}

type OnmetalVirtualIPConfig struct {
	IP        commonv1alpha1.IP
	VirtualIP *networkingv1alpha1.VirtualIP
}

type OnmetalNetworkInterfaceConfig struct {
	Name             string
	Network          *OnmetalNetworkConfig
	VirtualIP        *OnmetalVirtualIPConfig
	NetworkInterface *networkingv1alpha1.NetworkInterface
}

type OnmetalMachineConfig struct {
	ID                string
	IgnitionSecret    *corev1.Secret
	NetworkInterfaces []OnmetalNetworkInterfaceConfig
	Volumes           []OnmetalVolumeConfig
	Machine           *computev1alpha1.Machine
}

func (s *Server) onmetalNetworkInterfaceName(id, networkInterfaceName string) string {
	return s.hashID(id, networkInterfaceName)
}

func (s *Server) onmetalVirtualIPName(id, networkInterfaceName string, ipFamily corev1.IPFamily) string {
	return s.hashID(id, networkInterfaceName, string(ipFamily))
}

func (s *Server) onmetalVolumeName(id, volumeName string) string {
	return s.hashID(id, volumeName)
}

func (s *Server) onmetalIgnitionSecretName(id string) string {
	return s.hashID(id, "ignition")
}

func (s *Server) getOnmetalResources(resources *ori.MachineResources) (corev1.ResourceList, error) {
	return corev1.ResourceList{
		corev1.ResourceCPU:    *resource.NewQuantity(int64(resources.CpuCount), resource.DecimalSI),
		corev1.ResourceMemory: *resource.NewQuantity(int64(resources.MemoryBytes), resource.DecimalSI),
	}, nil
}

func (s *Server) findOnmetalMachineClass(ctx context.Context, resources *ori.MachineResources) (*computev1alpha1.MachineClass, error) {
	machineClassList := &computev1alpha1.MachineClassList{}
	if err := s.client.List(ctx, machineClassList); err != nil {
		return nil, fmt.Errorf("error listing machine classes: %w", err)
	}

	expectedResources := corev1.ResourceList{
		corev1.ResourceCPU:    *resource.NewQuantity(int64(resources.CpuCount), resource.DecimalSI),
		corev1.ResourceMemory: *resource.NewQuantity(int64(resources.MemoryBytes), resource.DecimalSI),
	}

	var matches []computev1alpha1.MachineClass
	for _, machineClass := range machineClassList.Items {
		actualResources := quotav1.Mask(machineClass.Capabilities, []corev1.ResourceName{corev1.ResourceCPU, corev1.ResourceMemory})
		if !quotav1.Equals(actualResources, expectedResources) {
			continue
		}

		matches = append(matches, machineClass)
	}

	switch len(matches) {
	case 0:
		return nil, fmt.Errorf("no machine class found satisfying resource requirements")
	case 1:
		machineClass := matches[0]
		return &machineClass, nil
	default:
		return nil, fmt.Errorf("ambiguous matches for requirements")
	}
}

func (s *Server) getOnmetalVolumeData(
	machineID string,
	volume *ori.VolumeConfig,
) (*computev1alpha1.Volume, *OnmetalVolumeConfig, error) {
	var (
		src                 computev1alpha1.VolumeSource
		onmetalVolumeConfig *OnmetalVolumeConfig
	)
	switch {
	case volume.EmptyDisk != nil:
		src.EmptyDisk = &computev1alpha1.EmptyDiskVolumeSource{}
		if sizeLimit := volume.EmptyDisk.SizeLimitBytes; sizeLimit != 0 {
			src.EmptyDisk.SizeLimit = resource.NewQuantity(int64(sizeLimit), resource.DecimalSI)
		}
	case volume.Access != nil:
		onmetalVolumeName := s.onmetalVolumeName(machineID, volume.Name)

		var onmetalVolumeSecret *corev1.Secret
		if secretData := volume.Access.SecretData; secretData != nil {
			onmetalVolumeSecret = &corev1.Secret{
				ObjectMeta: metav1.ObjectMeta{
					Namespace: s.namespace,
					Name:      onmetalVolumeName,
				},
				Type: corev1.SecretTypeOpaque,
				Data: secretData,
			}
			apiutils.SetMachineIDLabel(onmetalVolumeSecret, machineID)
			apiutils.SetVolumeNameLabel(onmetalVolumeSecret, volume.Name)
		}

		onmetalVolume := &storagev1alpha1.Volume{
			ObjectMeta: metav1.ObjectMeta{
				Namespace: s.namespace,
				Name:      onmetalVolumeName,
			},
			Spec: storagev1alpha1.VolumeSpec{},
		}
		apiutils.SetMachineIDLabel(onmetalVolume, machineID)
		apiutils.SetVolumeNameLabel(onmetalVolume, volume.Name)
		onmetalapiannotations.SetExternallyMangedBy(onmetalVolume, machinebrokerv1alpha1.MachineBrokerManager)

		onmetalVolumeAccess := &storagev1alpha1.VolumeAccess{
			Driver:           volume.Access.Driver,
			Handle:           volume.Access.Handle,
			VolumeAttributes: volume.Access.Attributes,
		}
		if onmetalVolumeSecret != nil {
			onmetalVolumeAccess.SecretRef = &corev1.LocalObjectReference{Name: onmetalVolumeSecret.Name}
		}

		onmetalVolumeConfig = &OnmetalVolumeConfig{
			Name:   volume.Name,
			Secret: onmetalVolumeSecret,
			Access: onmetalVolumeAccess,
			Volume: onmetalVolume,
		}
		src.VolumeRef = &corev1.LocalObjectReference{Name: onmetalVolumeName}
	}

	onmetalMachineVolume := &computev1alpha1.Volume{
		Name:         volume.Name,
		Device:       volume.Device,
		VolumeSource: src,
	}
	return onmetalMachineVolume, onmetalVolumeConfig, nil
}

func (s *Server) getOnmetalVirtualIPConfig(
	machineID, networkInterfaceName string,
	virtualIP *ori.VirtualIPConfig,
) (*OnmetalVirtualIPConfig, error) {
	ip, err := commonv1alpha1.ParseIP(virtualIP.Ip)
	if err != nil {
		return nil, fmt.Errorf("error parsing virtual ip: %w", err)
	}

	onmetalVirtualIP := &networkingv1alpha1.VirtualIP{
		ObjectMeta: metav1.ObjectMeta{
			Namespace: s.namespace,
			Name:      s.onmetalVirtualIPName(machineID, networkInterfaceName, ip.Family()),
		},
		Spec: networkingv1alpha1.VirtualIPSpec{
			Type:     networkingv1alpha1.VirtualIPTypePublic,
			IPFamily: ip.Family(),
		},
	}
	apiutils.SetMachineIDLabel(onmetalVirtualIP, machineID)
	apiutils.SetNetworkInterfaceNameLabel(onmetalVirtualIP, networkInterfaceName)
	apiutils.SetIPFamilyLabel(onmetalVirtualIP, ip.Family())
	onmetalapiannotations.SetExternallyMangedBy(onmetalVirtualIP, machinebrokerv1alpha1.MachineBrokerManager)

	return &OnmetalVirtualIPConfig{
		IP:        ip,
		VirtualIP: onmetalVirtualIP,
	}, nil
}

func (s *Server) getOnmetalNetworkConfig(
	machineID, networkInterfaceName string,
	network *ori.NetworkConfig,
) (*OnmetalNetworkConfig, error) {
	onmetalNetwork := &networkingv1alpha1.Network{
		ObjectMeta: metav1.ObjectMeta{
			Namespace: s.namespace,
			Name:      s.onmetalNetworkInterfaceName(machineID, networkInterfaceName),
		},
		Spec: networkingv1alpha1.NetworkSpec{
			ProviderID: network.Handle,
		},
	}
	apiutils.SetMachineIDLabel(onmetalNetwork, machineID)
	apiutils.SetNetworkInterfaceNameLabel(onmetalNetwork, networkInterfaceName)
	onmetalapiannotations.SetExternallyMangedBy(onmetalNetwork, machinebrokerv1alpha1.MachineBrokerManager)

	return &OnmetalNetworkConfig{
		Network: onmetalNetwork,
	}, nil
}

func (s *Server) getOnmetalNetworkInterfaceData(
	machineID string,
	networkInterface *ori.NetworkInterfaceConfig,
) (*computev1alpha1.NetworkInterface, *OnmetalNetworkInterfaceConfig, error) {
	onmetalNetworkInterfaceName := s.onmetalNetworkInterfaceName(machineID, networkInterface.Name)

	onmetalNetworkConfig, err := s.getOnmetalNetworkConfig(machineID, networkInterface.Name, networkInterface.Network)
	if err != nil {
		return nil, nil, err
	}

	var onmetalVirtualIPConfig *OnmetalVirtualIPConfig
	if virtualIP := networkInterface.VirtualIp; virtualIP != nil {
		var err error
		onmetalVirtualIPConfig, err = s.getOnmetalVirtualIPConfig(machineID, networkInterface.Name, virtualIP)
		if err != nil {
			return nil, nil, err
		}
	}

	ips, err := s.parseIPs(networkInterface.Ips)
	if err != nil {
		return nil, nil, fmt.Errorf("error parsing network interface ips: %w", err)
	}

	onmetalNetworkInterface := &networkingv1alpha1.NetworkInterface{
		ObjectMeta: metav1.ObjectMeta{
			Namespace: s.namespace,
			Name:      onmetalNetworkInterfaceName,
		},
		Spec: networkingv1alpha1.NetworkInterfaceSpec{
			NetworkRef: corev1.LocalObjectReference{Name: onmetalNetworkConfig.Network.Name},
			IPFamilies: s.getOnmetalIPsIPFamilies(ips),
			IPs:        s.onmetalIPsToOnmetalIPSources(ips),
		},
	}
	apiutils.SetMachineIDLabel(onmetalNetworkInterface, machineID)
	apiutils.SetNetworkInterfaceNameLabel(onmetalNetworkInterface, networkInterface.Name)
	if onmetalVirtualIPConfig != nil {
		onmetalNetworkInterface.Spec.VirtualIP = &networkingv1alpha1.VirtualIPSource{
			VirtualIPRef: &corev1.LocalObjectReference{Name: onmetalVirtualIPConfig.VirtualIP.Name},
		}
	}

	computeNetworkInterface := &computev1alpha1.NetworkInterface{
		Name: networkInterface.Name,
		NetworkInterfaceSource: computev1alpha1.NetworkInterfaceSource{
			NetworkInterfaceRef: &corev1.LocalObjectReference{Name: onmetalNetworkInterface.Name},
		},
	}
	onmetalNetworkInterfaceConfig := &OnmetalNetworkInterfaceConfig{
		Name:             networkInterface.Name,
		Network:          onmetalNetworkConfig,
		VirtualIP:        onmetalVirtualIPConfig,
		NetworkInterface: onmetalNetworkInterface,
	}
	return computeNetworkInterface, onmetalNetworkInterfaceConfig, nil
}

func (s *Server) getOnmetalMachineConfig(ctx context.Context, cfg *ori.MachineConfig, machineID string) (*OnmetalMachineConfig, error) {
	onmetalMachineClass, err := s.findOnmetalMachineClass(ctx, cfg.Resources)
	if err != nil {
		return nil, fmt.Errorf("error finding onmetal machine class: %w", err)
	}

	var ignitionSecret *corev1.Secret
	if ignition := cfg.Ignition; ignition != nil {
		ignitionSecret = &corev1.Secret{
			ObjectMeta: metav1.ObjectMeta{
				Namespace: s.namespace,
				Name:      s.onmetalIgnitionSecretName(machineID),
			},
			Type: corev1.SecretTypeOpaque,
			Data: map[string][]byte{
				computev1alpha1.DefaultIgnitionKey: ignition.Data,
			},
		}
		apiutils.SetMachineIDLabel(ignitionSecret, machineID)
	}

	var (
		onmetalMachineVolumes []computev1alpha1.Volume
		onmetalVolumeConfigs  []OnmetalVolumeConfig
	)
	for _, volume := range cfg.Volumes {
		onmetalMachineVolume, onmetalVolumeConfig, err := s.getOnmetalVolumeData(machineID, volume)
		if err != nil {
			return nil, err
		}

		onmetalMachineVolumes = append(onmetalMachineVolumes, *onmetalMachineVolume)
		if onmetalVolumeConfig != nil {
			onmetalVolumeConfigs = append(onmetalVolumeConfigs, *onmetalVolumeConfig)
		}
	}

	var (
		onmetalMachineNetworkInterfaces []computev1alpha1.NetworkInterface
		onmetalNetworkInterfaceConfigs  []OnmetalNetworkInterfaceConfig
	)
	for _, networkInterface := range cfg.NetworkInterfaces {
		onmetalMachineNetworkInterface, onmetalNetworkInterfaceConfig, err := s.getOnmetalNetworkInterfaceData(machineID, networkInterface)
		if err != nil {
			return nil, err
		}

		onmetalMachineNetworkInterfaces = append(onmetalMachineNetworkInterfaces, *onmetalMachineNetworkInterface)
		onmetalNetworkInterfaceConfigs = append(onmetalNetworkInterfaceConfigs, *onmetalNetworkInterfaceConfig)
	}

	onmetalMachine := &computev1alpha1.Machine{
		ObjectMeta: metav1.ObjectMeta{
			Namespace: s.namespace,
			Name:      machineID,
		},
		Spec: computev1alpha1.MachineSpec{
			MachineClassRef:     corev1.LocalObjectReference{Name: onmetalMachineClass.Name},
			MachinePoolSelector: s.machinePoolSelector,
			Image:               cfg.Image,
			ImagePullSecretRef:  nil, // TODO: Fill if necessary.
			NetworkInterfaces:   onmetalMachineNetworkInterfaces,
			Volumes:             onmetalMachineVolumes,
		},
	}
	apiutils.SetMachineIDLabel(onmetalMachine, machineID)
	apiutils.SetMachineManagerLabel(onmetalMachine, machinebrokerv1alpha1.MachineBrokerManager)
	if err := apiutils.SetMetadataAnnotation(onmetalMachine, cfg.Metadata); err != nil {
		return nil, err
	}
	if err := apiutils.SetAnnotationsAnnotation(onmetalMachine, cfg.Annotations); err != nil {
		return nil, err
	}
	if err := apiutils.SetLabelsAnnotation(onmetalMachine, cfg.Labels); err != nil {
		return nil, err
	}
	if s.machinePoolName != "" {
		onmetalMachine.Spec.MachinePoolRef = &corev1.LocalObjectReference{Name: s.machinePoolName}
	}
	if ignitionSecret != nil {
		onmetalMachine.Spec.IgnitionRef = &commonv1alpha1.SecretKeySelector{
			LocalObjectReference: corev1.LocalObjectReference{Name: s.onmetalIgnitionSecretName(machineID)},
		}
	}

	return &OnmetalMachineConfig{
		ID:                machineID,
		IgnitionSecret:    ignitionSecret,
		NetworkInterfaces: onmetalNetworkInterfaceConfigs,
		Volumes:           onmetalVolumeConfigs,
		Machine:           onmetalMachine,
	}, nil
}

func (s *Server) setupCleaner(ctx context.Context, log logr.Logger, retErr *error) (c *cleaner.Cleaner, cleanup func()) {
	c = cleaner.New()
	cleanup = func() {
		if *retErr != nil {
			select {
			case <-ctx.Done():
				log.Info("Cannot do cleanup since context expired")
				return
			default:
				if err := c.Cleanup(ctx); err != nil {
					log.Error(err, "Error cleaning up")
				}
			}
		}
	}
	return c, cleanup
}

func (s *Server) CreateMachine(ctx context.Context, req *ori.CreateMachineRequest) (res *ori.CreateMachineResponse, retErr error) {
	log := s.loggerFrom(ctx)

	log.V(1).Info("Generating machine id")
	machineID := s.generateID()
	log = log.WithValues("MachineID", machineID)
	log.V(1).Info("Generated machine id")

	cfg := req.Config

	log.V(1).Info("Getting machine configuration")
	onmetalMachineCfg, err := s.getOnmetalMachineConfig(ctx, cfg, machineID)
	if err != nil {
		return nil, fmt.Errorf("error getting onmetal machine config: %w", err)
	}

	cleaner, cleanup := s.setupCleaner(ctx, log, &retErr)
	defer cleanup()

	if ignitionSecret := onmetalMachineCfg.IgnitionSecret; ignitionSecret != nil {
		if err := s.createOnmetalIgnitionSecret(ctx, log, cleaner, ignitionSecret); err != nil {
			return nil, err
		}
	}

	for _, onmetalNetworkInterfaceCfg := range onmetalMachineCfg.NetworkInterfaces {
		if err := s.createOnmetalNetworkInterface(ctx, log, cleaner, onmetalNetworkInterfaceCfg); err != nil {
			return nil, err
		}
	}

	for _, onmetalVolumeCfg := range onmetalMachineCfg.Volumes {
		if err := s.createOnmetalVolume(ctx, log, cleaner, onmetalVolumeCfg); err != nil {
			return nil, err
		}
	}

	if err := s.createOnmetalMachine(ctx, log, cleaner, onmetalMachineCfg); err != nil {
		return nil, err
	}

	return &ori.CreateMachineResponse{
		Machine: &ori.Machine{
			Id:          onmetalMachineCfg.ID,
			Metadata:    req.Config.Metadata,
			Annotations: req.Config.Annotations,
			Labels:      req.Config.Labels,
		},
	}, nil
}

func (s *Server) createOnmetalMachine(ctx context.Context, log logr.Logger, cleaner *cleaner.Cleaner, onmetalMachineCfg *OnmetalMachineConfig) error {
	log.V(1).Info("Creating machine")
	onmetalMachine := onmetalMachineCfg.Machine
	if err := s.client.Create(ctx, onmetalMachine); err != nil {
		return fmt.Errorf("error creating machine: %w", err)
	}
	cleaner.Add(func(ctx context.Context) error {
		if err := s.client.Delete(ctx, onmetalMachine); client.IgnoreNotFound(err) != nil {
			return fmt.Errorf("error deleting machine: %w", err)
		}
		return nil
	})
	return nil
}

func (s *Server) createOnmetalNetwork(ctx context.Context, log logr.Logger, cleaner *cleaner.Cleaner, onmetalNetworkCfg *OnmetalNetworkConfig) error {
	log.V(1).Info("Creating network")
	onmetalNetwork := onmetalNetworkCfg.Network
	if err := s.client.Create(ctx, onmetalNetwork); err != nil {
		return fmt.Errorf("error creating network: %w", err)
	}

	baseOnmetalNetwork := onmetalNetwork.DeepCopy()
	onmetalNetwork.Status.State = networkingv1alpha1.NetworkStateAvailable
	if err := s.client.Status().Patch(ctx, onmetalNetwork, client.MergeFrom(baseOnmetalNetwork)); err != nil {
		return fmt.Errorf("error patching onmetal network status state to available: %w", err)
	}

	cleaner.Add(func(ctx context.Context) error {
		if err := s.client.Delete(ctx, onmetalNetwork); client.IgnoreNotFound(err) != nil {
			return fmt.Errorf("error deleting network: %w", err)
		}
		return nil
	})
	return nil
}

func (s *Server) createOnmetalIgnitionSecret(ctx context.Context, log logr.Logger, cleaner *cleaner.Cleaner, ignitionSecret *corev1.Secret) error {
	log.V(1).Info("Creating ignition secret")
	if err := s.client.Create(ctx, ignitionSecret); err != nil {
		return fmt.Errorf("error creating ignition secret: %w", err)
	}

	cleaner.Add(func(ctx context.Context) error {
		if err := s.client.Delete(ctx, ignitionSecret); client.IgnoreNotFound(err) != nil {
			return fmt.Errorf("error deleting ignition secret: %w", err)
		}
		return nil
	})
	return nil
}

func (s *Server) createOnmetalVirtualIP(ctx context.Context, log logr.Logger, cleaner *cleaner.Cleaner, onmetalVirtualIPCfg *OnmetalVirtualIPConfig) (*networkingv1alpha1.VirtualIP, error) {
	log.V(1).Info("Creating virtual ip")
	onmetalVirtualIP := onmetalVirtualIPCfg.VirtualIP
	if err := s.client.Create(ctx, onmetalVirtualIP); err != nil {
		return nil, fmt.Errorf("error creating virtual ip: %w", err)
	}

	cleaner.Add(func(ctx context.Context) error {
		if err := s.client.Delete(ctx, onmetalVirtualIP); client.IgnoreNotFound(err) != nil {
			return fmt.Errorf("error deleting virtual ip: %w", err)
		}
		return nil
	})
	return onmetalVirtualIP, nil
}

func (s *Server) setOnmetalVirtualIPIP(ctx context.Context, log logr.Logger, onmetalVirtualIP *networkingv1alpha1.VirtualIP, ip commonv1alpha1.IP) error {
	log.V(1).Info("Patching virtual ip status")
	base := onmetalVirtualIP.DeepCopy()
	onmetalVirtualIP.Status.IP = &ip
	if err := s.client.Status().Patch(ctx, onmetalVirtualIP, client.MergeFrom(base)); err != nil {
		return fmt.Errorf("error patching virtual ip status: %w", err)
	}
	return nil
}

func (s *Server) createOnmetalNetworkInterface(ctx context.Context, log logr.Logger, cleaner *cleaner.Cleaner, onmetalNetworkInterfaceCfg OnmetalNetworkInterfaceConfig) error {
	networkInterfaceName := onmetalNetworkInterfaceCfg.Name
	log = log.WithValues("NetworkInterfaceName", networkInterfaceName)

	if err := s.createOnmetalNetwork(ctx, log, cleaner, onmetalNetworkInterfaceCfg.Network); err != nil {
		return err
	}

	if onmetalVirtualIPCfg := onmetalNetworkInterfaceCfg.VirtualIP; onmetalVirtualIPCfg != nil {
		onmetalVirtualIP, err := s.createOnmetalVirtualIP(ctx, log, cleaner, onmetalVirtualIPCfg)
		if err != nil {
			return err
		}

		if err := s.setOnmetalVirtualIPIP(ctx, log, onmetalVirtualIP, onmetalVirtualIPCfg.IP); err != nil {
			return err
		}
	}

	log.V(1).Info("Creating network interface")
	onmetalNetworkInterface := onmetalNetworkInterfaceCfg.NetworkInterface
	if err := s.client.Create(ctx, onmetalNetworkInterface); err != nil {
		return fmt.Errorf("error creating network interface %s: %w", networkInterfaceName, err)
	}

	cleaner.Add(func(ctx context.Context) error {
		if err := s.client.Delete(ctx, onmetalNetworkInterface); client.IgnoreNotFound(err) != nil {
			return fmt.Errorf("error deleting network interface %s: %w", networkInterfaceName, err)
		}
		return nil
	})
	return nil
}

func (s *Server) createOnmetalVolume(ctx context.Context, log logr.Logger, cleaner *cleaner.Cleaner, onmetalVolumeCfg OnmetalVolumeConfig) error {
	volumeName := onmetalVolumeCfg.Name
	log = log.WithValues("VolumeName", volumeName)

	if onmetalVolumeSecret := onmetalVolumeCfg.Secret; onmetalVolumeSecret != nil {
		log.V(1).Info("Creating volume secret")
		if err := s.client.Create(ctx, onmetalVolumeSecret); err != nil {
			return fmt.Errorf("error creating volume %s secret: %w", volumeName, err)
		}

		cleaner.Add(func(ctx context.Context) error {
			if err := s.client.Delete(ctx, onmetalVolumeSecret); client.IgnoreNotFound(err) != nil {
				return fmt.Errorf("error deleting volume %s secret: %w", volumeName, err)
			}
			return nil
		})
	}

	onmetalVolume := onmetalVolumeCfg.Volume
	log.V(1).Info("Creating volume")
	if err := s.client.Create(ctx, onmetalVolume); err != nil {
		return fmt.Errorf("error creating volume %s: %w", volumeName, err)
	}

	cleaner.Add(func(ctx context.Context) error {
		if err := s.client.Delete(ctx, onmetalVolume); client.IgnoreNotFound(err) != nil {
			return fmt.Errorf("error deleting volume %s: %w", volumeName, err)
		}
		return nil
	})

	log.V(1).Info("Patching volume access")
	base := onmetalVolume.DeepCopy()
	onmetalVolume.Status.State = storagev1alpha1.VolumeStateAvailable
	onmetalVolume.Status.Access = onmetalVolumeCfg.Access
	if err := s.client.Status().Patch(ctx, onmetalVolume, client.MergeFrom(base)); err != nil {
		return fmt.Errorf("error patching volume %s access: %w", volumeName, err)
	}
	return nil
}
