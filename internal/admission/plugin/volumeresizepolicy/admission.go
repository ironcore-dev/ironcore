// SPDX-FileCopyrightText: 2023 SAP SE or an SAP affiliate company and IronCore contributors
// SPDX-License-Identifier: Apache-2.0

package volumeresizepolicy

import (
	"context"
	"fmt"
	"io"

	"github.com/ironcore-dev/ironcore/api/storage/v1alpha1"
	"github.com/ironcore-dev/ironcore/client-go/ironcore"
	"github.com/ironcore-dev/ironcore/internal/apis/core"
	"github.com/ironcore-dev/ironcore/internal/apis/storage"
	apierrors "k8s.io/apimachinery/pkg/api/errors"
	v1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apiserver/pkg/admission"
)

const PluginName = "VolumeExpansion"

func Register(plugins *admission.Plugins) {
	plugins.Register(PluginName, func(config io.Reader) (admission.Interface, error) {
		return NewVolumeExpansion(), nil
	})
}

type VolumeExpansion struct {
	client ironcore.Interface
	*admission.Handler
}

func NewVolumeExpansion() admission.Interface {
	return &VolumeExpansion{
		Handler: admission.NewHandler(admission.Update),
	}
}

func (v *VolumeExpansion) Validate(ctx context.Context, a admission.Attributes, o admission.ObjectInterfaces) error {
	if shouldIgnore(a) {
		return nil
	}

	volume, ok := a.GetObject().(*storage.Volume)
	if !ok {
		return apierrors.NewBadRequest("Resource was marked with kind Volume but was unable to be converted")
	}

	oldVolume, ok := a.GetOldObject().(*storage.Volume)
	if !ok {
		return apierrors.NewBadRequest("Resource was marked with kind Volume but was unable to be converted")
	}

	volumeSize, ok := volume.Spec.Resources[core.ResourceStorage]
	if !ok {
		return apierrors.NewBadRequest("Volume does not contain any capacity information")
	}

	oldVolumeSize, ok := oldVolume.Spec.Resources[core.ResourceStorage]
	if !ok {
		return apierrors.NewBadRequest("Old Volume does not contain any capacity information")
	}

	if volumeSize.Equal(oldVolumeSize) {
		return nil
	}

	// Volume size changed, therefore we need to check whether the VolumeClass supports Volume expansion
	volumeClass, err := v.client.StorageV1alpha1().VolumeClasses().Get(ctx, volume.Spec.VolumeClassRef.Name, v1.GetOptions{})
	if err != nil {
		return apierrors.NewBadRequest(fmt.Sprintf("Could not get VolumeClass for Volume: %v", err))
	}

	switch volumeClass.ResizePolicy {
	case v1alpha1.ResizePolicyStatic:
		if volumeSize.Value() != oldVolumeSize.Value() {
			return apierrors.NewBadRequest("VolumeClass ResizePolicy does not allow resizing")
		}
	case v1alpha1.ResizePolicyExpandOnly:
		if volumeSize.Value() < oldVolumeSize.Value() {
			return apierrors.NewBadRequest("VolumeClass ResizePolicy does not allow shrinking")
		}
	default:
		return nil
	}

	return nil
}

func (v *VolumeExpansion) SetExternalIronCoreClientSet(client ironcore.Interface) {
	v.client = client
}

func (v *VolumeExpansion) ValidateInitialization() error {
	if v.client == nil {
		return fmt.Errorf("missing client")
	}
	return nil
}

func shouldIgnore(a admission.Attributes) bool {
	if a.GetKind().GroupKind() != storage.Kind("Volume") {
		return true
	}

	volume, ok := a.GetObject().(*storage.Volume)
	if !ok {
		return true
	}

	if volume.Spec.VolumeClassRef == nil {
		return true
	}
	return false
}
