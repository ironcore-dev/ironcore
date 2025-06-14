// SPDX-FileCopyrightText: 2025 SAP SE or an SAP affiliate company and IronCore contributors
// SPDX-License-Identifier: Apache-2.0

// Code generated by client-gen. DO NOT EDIT.

package fake

import (
	v1alpha1 "github.com/ironcore-dev/ironcore/api/compute/v1alpha1"
	computev1alpha1 "github.com/ironcore-dev/ironcore/client-go/applyconfigurations/compute/v1alpha1"
	typedcomputev1alpha1 "github.com/ironcore-dev/ironcore/client-go/ironcore/versioned/typed/compute/v1alpha1"
	gentype "k8s.io/client-go/gentype"
)

// fakeMachinePools implements MachinePoolInterface
type fakeMachinePools struct {
	*gentype.FakeClientWithListAndApply[*v1alpha1.MachinePool, *v1alpha1.MachinePoolList, *computev1alpha1.MachinePoolApplyConfiguration]
	Fake *FakeComputeV1alpha1
}

func newFakeMachinePools(fake *FakeComputeV1alpha1) typedcomputev1alpha1.MachinePoolInterface {
	return &fakeMachinePools{
		gentype.NewFakeClientWithListAndApply[*v1alpha1.MachinePool, *v1alpha1.MachinePoolList, *computev1alpha1.MachinePoolApplyConfiguration](
			fake.Fake,
			"",
			v1alpha1.SchemeGroupVersion.WithResource("machinepools"),
			v1alpha1.SchemeGroupVersion.WithKind("MachinePool"),
			func() *v1alpha1.MachinePool { return &v1alpha1.MachinePool{} },
			func() *v1alpha1.MachinePoolList { return &v1alpha1.MachinePoolList{} },
			func(dst, src *v1alpha1.MachinePoolList) { dst.ListMeta = src.ListMeta },
			func(list *v1alpha1.MachinePoolList) []*v1alpha1.MachinePool {
				return gentype.ToPointerSlice(list.Items)
			},
			func(list *v1alpha1.MachinePoolList, items []*v1alpha1.MachinePool) {
				list.Items = gentype.FromPointerSlice(items)
			},
		),
		fake,
	}
}
