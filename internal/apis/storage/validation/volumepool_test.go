// SPDX-FileCopyrightText: 2023 SAP SE or an SAP affiliate company and IronCore contributors
// SPDX-License-Identifier: Apache-2.0

package validation_test

import (
	"github.com/ironcore-dev/ironcore/internal/apis/storage"
	. "github.com/ironcore-dev/ironcore/internal/apis/storage/validation"
	. "github.com/ironcore-dev/ironcore/internal/testutils/validation"
	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
	"github.com/onsi/gomega/types"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

var _ = Describe("VolumePool", func() {
	DescribeTable("ValidateVolumePool",
		func(volumePool *storage.VolumePool, match types.GomegaMatcher) {
			errList := ValidateVolumePool(volumePool)
			Expect(errList).To(match)
		},
		Entry("missing name",
			&storage.VolumePool{},
			ContainElement(RequiredField("metadata.name")),
		),
		Entry("bad name",
			&storage.VolumePool{ObjectMeta: metav1.ObjectMeta{Name: "foo*"}},
			ContainElement(InvalidField("metadata.name")),
		),
		Entry("dns subdomain name",
			&storage.VolumePool{ObjectMeta: metav1.ObjectMeta{Name: "foo.bar.baz"}},
			Not(ContainElement(InvalidField("metadata.name"))),
		),
	)

	DescribeTable("ValidateVolumeUpdate",
		func(newVolumePool, oldVolumePool *storage.VolumePool, match types.GomegaMatcher) {
			errList := ValidateVolumePoolUpdate(newVolumePool, oldVolumePool)
			Expect(errList).To(match)
		},
		Entry("immutable providerID if set",
			&storage.VolumePool{
				Spec: storage.VolumePoolSpec{
					ProviderID: "foo",
				},
			},
			&storage.VolumePool{
				Spec: storage.VolumePoolSpec{
					ProviderID: "bar",
				},
			},
			ContainElement(ImmutableField("spec.providerID")),
		),
		Entry("mutable providerID if not set",
			&storage.VolumePool{
				Spec: storage.VolumePoolSpec{
					ProviderID: "foo",
				},
			},
			&storage.VolumePool{},
			Not(ContainElement(ImmutableField("spec.providerID"))),
		),
	)
})
