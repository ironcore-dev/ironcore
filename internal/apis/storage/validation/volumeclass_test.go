// SPDX-FileCopyrightText: 2023 SAP SE or an SAP affiliate company and IronCore contributors
// SPDX-License-Identifier: Apache-2.0

package validation_test

import (
	"github.com/ironcore-dev/ironcore/internal/apis/core"
	"github.com/ironcore-dev/ironcore/internal/apis/storage"
	. "github.com/ironcore-dev/ironcore/internal/apis/storage/validation"
	. "github.com/ironcore-dev/ironcore/internal/testutils/validation"
	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
	"github.com/onsi/gomega/types"
	"k8s.io/apimachinery/pkg/api/resource"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

var _ = Describe("VolumeClass", func() {
	DescribeTable("ValidateVolumeClass",
		func(volumeClass *storage.VolumeClass, match types.GomegaMatcher) {
			errList := ValidateVolumeClass(volumeClass)
			Expect(errList).To(match)
		},
		Entry("missing name",
			&storage.VolumeClass{},
			ContainElement(RequiredField("metadata.name")),
		),
		Entry("bad name",
			&storage.VolumeClass{ObjectMeta: metav1.ObjectMeta{Name: "foo*"}},
			ContainElement(InvalidField("metadata.name")),
		),
		Entry("missing required capabilities",
			&storage.VolumeClass{},
			ContainElements(
				InvalidField("capabilities[tps]"),
				InvalidField("capabilities[iops]"),
			),
		),
		Entry("invalid capabilities",
			&storage.VolumeClass{
				Capabilities: core.ResourceList{
					core.ResourceTPS:  resource.MustParse("-1"),
					core.ResourceIOPS: resource.MustParse("-1"),
				},
			},
			ContainElements(
				InvalidField("capabilities[tps]"),
				InvalidField("capabilities[iops]"),
			),
		),
		Entry("valid capabilities",
			&storage.VolumeClass{
				Capabilities: core.ResourceList{
					core.ResourceTPS:  resource.MustParse("500Mi"),
					core.ResourceIOPS: resource.MustParse("100"),
				},
			},
			Not(ContainElements(
				InvalidField("capabilities[tps]"),
				InvalidField("capabilities[iops]"),
			)),
		),
		Entry("valid resizePolicy",
			&storage.VolumeClass{
				ResizePolicy: storage.ResizePolicyStatic,
			},
			Not(ContainElements(
				InvalidField("resizePolicy"),
			)),
		),
		Entry("invalid resizePolicy",
			&storage.VolumeClass{
				ResizePolicy: "foo",
			},
			ContainElements(
				NotSupportedField("resizePolicy"),
			),
		),
	)

	DescribeTable("ValidateVolumeClassUpdate",
		func(newVolumeClass, oldVolumeClass *storage.VolumeClass, match types.GomegaMatcher) {
			errList := ValidateVolumeClassUpdate(newVolumeClass, oldVolumeClass)
			Expect(errList).To(match)
		},
		Entry("immutable capabilities",
			&storage.VolumeClass{
				Capabilities: core.ResourceList{
					core.ResourceIOPS: resource.MustParse("1000"),
				},
			},
			&storage.VolumeClass{
				Capabilities: core.ResourceList{},
			},
			ContainElement(ImmutableField("capabilities")),
		),
	)
})
