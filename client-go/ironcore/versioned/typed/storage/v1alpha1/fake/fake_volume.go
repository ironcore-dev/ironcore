// SPDX-FileCopyrightText: 2025 SAP SE or an SAP affiliate company and IronCore contributors
// SPDX-License-Identifier: Apache-2.0

// Code generated by client-gen. DO NOT EDIT.

package fake

import (
	v1alpha1 "github.com/ironcore-dev/ironcore/api/storage/v1alpha1"
	storagev1alpha1 "github.com/ironcore-dev/ironcore/client-go/applyconfigurations/storage/v1alpha1"
	typedstoragev1alpha1 "github.com/ironcore-dev/ironcore/client-go/ironcore/versioned/typed/storage/v1alpha1"
	gentype "k8s.io/client-go/gentype"
)

// fakeVolumes implements VolumeInterface
type fakeVolumes struct {
	*gentype.FakeClientWithListAndApply[*v1alpha1.Volume, *v1alpha1.VolumeList, *storagev1alpha1.VolumeApplyConfiguration]
	Fake *FakeStorageV1alpha1
}

func newFakeVolumes(fake *FakeStorageV1alpha1, namespace string) typedstoragev1alpha1.VolumeInterface {
	return &fakeVolumes{
		gentype.NewFakeClientWithListAndApply[*v1alpha1.Volume, *v1alpha1.VolumeList, *storagev1alpha1.VolumeApplyConfiguration](
			fake.Fake,
			namespace,
			v1alpha1.SchemeGroupVersion.WithResource("volumes"),
			v1alpha1.SchemeGroupVersion.WithKind("Volume"),
			func() *v1alpha1.Volume { return &v1alpha1.Volume{} },
			func() *v1alpha1.VolumeList { return &v1alpha1.VolumeList{} },
			func(dst, src *v1alpha1.VolumeList) { dst.ListMeta = src.ListMeta },
			func(list *v1alpha1.VolumeList) []*v1alpha1.Volume { return gentype.ToPointerSlice(list.Items) },
			func(list *v1alpha1.VolumeList, items []*v1alpha1.Volume) {
				list.Items = gentype.FromPointerSlice(items)
			},
		),
		fake,
	}
}
