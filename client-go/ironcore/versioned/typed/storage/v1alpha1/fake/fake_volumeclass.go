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

// fakeVolumeClasses implements VolumeClassInterface
type fakeVolumeClasses struct {
	*gentype.FakeClientWithListAndApply[*v1alpha1.VolumeClass, *v1alpha1.VolumeClassList, *storagev1alpha1.VolumeClassApplyConfiguration]
	Fake *FakeStorageV1alpha1
}

func newFakeVolumeClasses(fake *FakeStorageV1alpha1) typedstoragev1alpha1.VolumeClassInterface {
	return &fakeVolumeClasses{
		gentype.NewFakeClientWithListAndApply[*v1alpha1.VolumeClass, *v1alpha1.VolumeClassList, *storagev1alpha1.VolumeClassApplyConfiguration](
			fake.Fake,
			"",
			v1alpha1.SchemeGroupVersion.WithResource("volumeclasses"),
			v1alpha1.SchemeGroupVersion.WithKind("VolumeClass"),
			func() *v1alpha1.VolumeClass { return &v1alpha1.VolumeClass{} },
			func() *v1alpha1.VolumeClassList { return &v1alpha1.VolumeClassList{} },
			func(dst, src *v1alpha1.VolumeClassList) { dst.ListMeta = src.ListMeta },
			func(list *v1alpha1.VolumeClassList) []*v1alpha1.VolumeClass {
				return gentype.ToPointerSlice(list.Items)
			},
			func(list *v1alpha1.VolumeClassList, items []*v1alpha1.VolumeClass) {
				list.Items = gentype.FromPointerSlice(items)
			},
		),
		fake,
	}
}
