// SPDX-FileCopyrightText: 2023 SAP SE or an SAP affiliate company and IronCore contributors
// SPDX-License-Identifier: Apache-2.0

// Code generated by client-gen. DO NOT EDIT.

package v1alpha1

import (
	"context"

	v1alpha1 "github.com/ironcore-dev/ironcore/api/storage/v1alpha1"
	storagev1alpha1 "github.com/ironcore-dev/ironcore/client-go/applyconfigurations/storage/v1alpha1"
	scheme "github.com/ironcore-dev/ironcore/client-go/ironcore/versioned/scheme"
	v1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	types "k8s.io/apimachinery/pkg/types"
	watch "k8s.io/apimachinery/pkg/watch"
	gentype "k8s.io/client-go/gentype"
)

// VolumesGetter has a method to return a VolumeInterface.
// A group's client should implement this interface.
type VolumesGetter interface {
	Volumes(namespace string) VolumeInterface
}

// VolumeInterface has methods to work with Volume resources.
type VolumeInterface interface {
	Create(ctx context.Context, volume *v1alpha1.Volume, opts v1.CreateOptions) (*v1alpha1.Volume, error)
	Update(ctx context.Context, volume *v1alpha1.Volume, opts v1.UpdateOptions) (*v1alpha1.Volume, error)
	// Add a +genclient:noStatus comment above the type to avoid generating UpdateStatus().
	UpdateStatus(ctx context.Context, volume *v1alpha1.Volume, opts v1.UpdateOptions) (*v1alpha1.Volume, error)
	Delete(ctx context.Context, name string, opts v1.DeleteOptions) error
	DeleteCollection(ctx context.Context, opts v1.DeleteOptions, listOpts v1.ListOptions) error
	Get(ctx context.Context, name string, opts v1.GetOptions) (*v1alpha1.Volume, error)
	List(ctx context.Context, opts v1.ListOptions) (*v1alpha1.VolumeList, error)
	Watch(ctx context.Context, opts v1.ListOptions) (watch.Interface, error)
	Patch(ctx context.Context, name string, pt types.PatchType, data []byte, opts v1.PatchOptions, subresources ...string) (result *v1alpha1.Volume, err error)
	Apply(ctx context.Context, volume *storagev1alpha1.VolumeApplyConfiguration, opts v1.ApplyOptions) (result *v1alpha1.Volume, err error)
	// Add a +genclient:noStatus comment above the type to avoid generating ApplyStatus().
	ApplyStatus(ctx context.Context, volume *storagev1alpha1.VolumeApplyConfiguration, opts v1.ApplyOptions) (result *v1alpha1.Volume, err error)
	VolumeExpansion
}

// volumes implements VolumeInterface
type volumes struct {
	*gentype.ClientWithListAndApply[*v1alpha1.Volume, *v1alpha1.VolumeList, *storagev1alpha1.VolumeApplyConfiguration]
}

// newVolumes returns a Volumes
func newVolumes(c *StorageV1alpha1Client, namespace string) *volumes {
	return &volumes{
		gentype.NewClientWithListAndApply[*v1alpha1.Volume, *v1alpha1.VolumeList, *storagev1alpha1.VolumeApplyConfiguration](
			"volumes",
			c.RESTClient(),
			scheme.ParameterCodec,
			namespace,
			func() *v1alpha1.Volume { return &v1alpha1.Volume{} },
			func() *v1alpha1.VolumeList { return &v1alpha1.VolumeList{} }),
	}
}
