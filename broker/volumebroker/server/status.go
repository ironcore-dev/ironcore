// SPDX-FileCopyrightText: 2023 SAP SE or an SAP affiliate company and IronCore contributors
// SPDX-License-Identifier: Apache-2.0

package server

import (
	"context"
	"fmt"

	corev1alpha1 "github.com/ironcore-dev/ironcore/api/core/v1alpha1"
	storagev1alpha1 "github.com/ironcore-dev/ironcore/api/storage/v1alpha1"
	iri "github.com/ironcore-dev/ironcore/iri/apis/volume/v1alpha1"
	apierrors "k8s.io/apimachinery/pkg/api/errors"
	"k8s.io/apimachinery/pkg/api/resource"
	"k8s.io/apimachinery/pkg/util/sets"
	"sigs.k8s.io/controller-runtime/pkg/client"
)

func (s *Server) getTargetIronCoreVolumePools(ctx context.Context) ([]storagev1alpha1.VolumePool, error) {
	if s.volumePoolName != "" {
		ironcoreVolumePool := &storagev1alpha1.VolumePool{}
		ironcoreVolumePoolKey := client.ObjectKey{Name: s.volumePoolName}
		if err := s.client.Get(ctx, ironcoreVolumePoolKey, ironcoreVolumePool); err != nil {
			if !apierrors.IsNotFound(err) {
				return nil, fmt.Errorf("error getting volume pool %s: %w", s.volumePoolName, err)
			}
			return nil, nil
		}
	}

	volumePoolList := &storagev1alpha1.VolumePoolList{}
	if err := s.client.List(ctx, volumePoolList,
		client.MatchingLabels(s.volumePoolSelector),
	); err != nil {
		return nil, fmt.Errorf("error listing volume pools: %w", err)
	}
	return volumePoolList.Items, nil
}

func (s *Server) gatherAvailableVolumeClassNames(ironcoreVolumePools []storagev1alpha1.VolumePool) sets.Set[string] {
	res := sets.New[string]()
	for _, ironcoreVolumePool := range ironcoreVolumePools {
		for _, availableVolumeClass := range ironcoreVolumePool.Status.AvailableVolumeClasses {
			res.Insert(availableVolumeClass.Name)
		}
	}
	return res
}

func (s *Server) gatherVolumeClassQuantity(ironcoreVolumePools []storagev1alpha1.VolumePool) map[string]*resource.Quantity {
	res := map[string]*resource.Quantity{}
	for _, ironcoreVolumePool := range ironcoreVolumePools {
		for resourceName, resourceQuantity := range ironcoreVolumePool.Status.Capacity {
			if corev1alpha1.IsClassCountResource(resourceName) {
				if _, ok := res[string(resourceName)]; !ok {
					res[string(resourceName)] = resource.NewQuantity(0, resource.BinarySI)
				}
				res[string(resourceName)].Add(resourceQuantity)
			}
		}
	}
	return res
}

func (s *Server) filterIronCoreVolumeClasses(
	availableVolumeClassNames sets.Set[string],
	volumeClasses []storagev1alpha1.VolumeClass,
) []storagev1alpha1.VolumeClass {
	var filtered []storagev1alpha1.VolumeClass
	for _, volumeClass := range volumeClasses {
		if !availableVolumeClassNames.Has(volumeClass.Name) {
			continue
		}

		filtered = append(filtered, volumeClass)
	}
	return filtered
}

func (s *Server) convertIronCoreVolumeClassStatus(volumeClass *storagev1alpha1.VolumeClass, quantity *resource.Quantity) (*iri.VolumeClassStatus, error) {
	tps := volumeClass.Capabilities.TPS()
	iops := volumeClass.Capabilities.IOPS()

	return &iri.VolumeClassStatus{
		VolumeClass: &iri.VolumeClass{
			Name: volumeClass.Name,
			Capabilities: &iri.VolumeClassCapabilities{
				Tps:  tps.Value(),
				Iops: iops.Value(),
			},
		},
		Quantity: quantity.Value(),
	}, nil
}

func (s *Server) Status(ctx context.Context, req *iri.StatusRequest) (*iri.StatusResponse, error) {
	log := s.loggerFrom(ctx)

	log.V(1).Info("Getting target ironcore volume pools")
	ironcoreVolumePools, err := s.getTargetIronCoreVolumePools(ctx)
	if err != nil {
		return nil, fmt.Errorf("error getting target ironcore volume pools: %w", err)
	}

	log.V(1).Info("Gathering available volume class names")
	availableIronCoreVolumeClassNames := s.gatherAvailableVolumeClassNames(ironcoreVolumePools)

	if len(availableIronCoreVolumeClassNames) == 0 {
		log.V(1).Info("No available volume classes")
		return &iri.StatusResponse{VolumeClassStatus: []*iri.VolumeClassStatus{}}, nil
	}

	log.V(1).Info("Gathering volume class quantity")
	volumeClassQuantity := s.gatherVolumeClassQuantity(ironcoreVolumePools)

	log.V(1).Info("Listing ironcore volume classes")
	ironcoreVolumeClassList := &storagev1alpha1.VolumeClassList{}
	if err := s.client.List(ctx, ironcoreVolumeClassList); err != nil {
		return nil, fmt.Errorf("error listing ironcore volume classes: %w", err)
	}

	availableIronCoreVolumeClasses := s.filterIronCoreVolumeClasses(availableIronCoreVolumeClassNames, ironcoreVolumeClassList.Items)
	volumeClassStatus := make([]*iri.VolumeClassStatus, 0, len(availableIronCoreVolumeClasses))
	for _, ironcoreVolumeClass := range availableIronCoreVolumeClasses {
		quantity, ok := volumeClassQuantity[string(corev1alpha1.ClassCountFor(corev1alpha1.ClassTypeVolumeClass, ironcoreVolumeClass.Name))]
		if !ok {
			log.V(1).Info("Ignored class - missing quantity", "VolumeClass", ironcoreVolumeClass.Name)
			continue
		}

		volumeClass, err := s.convertIronCoreVolumeClassStatus(&ironcoreVolumeClass, quantity)
		if err != nil {
			return nil, fmt.Errorf("error converting ironcore volume class %s: %w", ironcoreVolumeClass.Name, err)
		}

		volumeClassStatus = append(volumeClassStatus, volumeClass)
	}

	log.V(1).Info("Returning volume classes")
	return &iri.StatusResponse{
		VolumeClassStatus: volumeClassStatus,
	}, nil
}
