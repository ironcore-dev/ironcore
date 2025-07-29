// SPDX-FileCopyrightText: 2023 SAP SE or an SAP affiliate company and IronCore contributors
// SPDX-License-Identifier: Apache-2.0

package validation_test

import (
	commonv1alpha1 "github.com/ironcore-dev/ironcore/api/common/v1alpha1"
	"github.com/ironcore-dev/ironcore/internal/apis/storage"
	. "github.com/ironcore-dev/ironcore/internal/apis/storage/validation"
	. "github.com/ironcore-dev/ironcore/internal/testutils/validation"
	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
	"github.com/onsi/gomega/types"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

var _ = Describe("VolumeSnapshotContent", func() {
	DescribeTable("ValidateVolumeSnapshotContent",
		func(volumeSnapshotContent *storage.VolumeSnapshotContent, match types.GomegaMatcher) {
			errList := ValidateVolumeSnapshotContent(volumeSnapshotContent)
			Expect(errList).To(match)
		},
		Entry("missing name",
			&storage.VolumeSnapshotContent{},
			ContainElement(RequiredField("metadata.name")),
		),
		Entry("missing namespace",
			&storage.VolumeSnapshotContent{ObjectMeta: metav1.ObjectMeta{Name: "foo"}},
			ContainElement(RequiredField("metadata.namespace")),
		),
		Entry("bad name",
			&storage.VolumeSnapshotContent{ObjectMeta: metav1.ObjectMeta{Name: "foo*"}},
			ContainElement(InvalidField("metadata.name")),
		),
		Entry("no source",
			&storage.VolumeSnapshotContent{},
			Not(ContainElement(RequiredField("spec.source"))),
		),
		Entry("no source with volume snapshot handle",
			&storage.VolumeSnapshotContent{},
			Not(ContainElement(RequiredField("spec.source.snapshotHandle"))),
		),
		Entry("no volume snapshot ref",
			&storage.VolumeSnapshotContent{},
			Not(ContainElement(RequiredField("spec.volumeSnapshotRef"))),
		),
		Entry("invalid volume snapshot ref name",
			&storage.VolumeSnapshotContent{
				Spec: storage.VolumeSnapshotContentSpec{
					VolumeSnapshotRef: &commonv1alpha1.UIDReference{Name: "foo*"},
				},
			},
			ContainElement(InvalidField("spec.volumeSnapshotRef.name")),
		),
	)

	DescribeTable("ValidateVolumeSnapshotContentUpdate",
		func(newVolumeSnapshotContent, oldVolumeSnapshotContent *storage.VolumeSnapshotContent, match types.GomegaMatcher) {
			errList := ValidateVolumeSnapshotContentUpdate(newVolumeSnapshotContent, oldVolumeSnapshotContent)
			Expect(errList).To(match)
		},
		Entry("immutable source",
			&storage.VolumeSnapshotContent{
				Spec: storage.VolumeSnapshotContentSpec{
					Source: &storage.VolumeSnapshotContentSource{VolumeSnapshotHandle: "snapshotHandle4987"},
				},
			},
			&storage.VolumeSnapshotContent{
				Spec: storage.VolumeSnapshotContentSpec{
					Source: &storage.VolumeSnapshotContentSource{VolumeSnapshotHandle: "snapshotHandle9847"},
				},
			},
			ContainElement(ImmutableField("spec.source")),
		),
		Entry("immutable volumeSnapshotRef",
			&storage.VolumeSnapshotContent{
				Spec: storage.VolumeSnapshotContentSpec{
					VolumeSnapshotRef: &commonv1alpha1.UIDReference{Name: "foo"},
				},
			},
			&storage.VolumeSnapshotContent{
				Spec: storage.VolumeSnapshotContentSpec{
					VolumeSnapshotRef: &commonv1alpha1.UIDReference{Name: "bar"},
				},
			},
			ContainElement(ImmutableField("spec.volumeSnapshotRef")),
		),
	)
})
