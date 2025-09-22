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
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

var _ = Describe("VolumeSnapshot", func() {
	DescribeTable("ValidateVolumeSnapshot",
		func(volumeSnapshot *storage.VolumeSnapshot, match types.GomegaMatcher) {
			errList := ValidateVolumeSnapshot(volumeSnapshot)
			Expect(errList).To(match)
		},
		Entry("missing name",
			&storage.VolumeSnapshot{},
			ContainElement(RequiredField("metadata.name")),
		),
		Entry("missing namespace",
			&storage.VolumeSnapshot{ObjectMeta: metav1.ObjectMeta{Name: "foo"}},
			ContainElement(RequiredField("metadata.namespace")),
		),
		Entry("bad name",
			&storage.VolumeSnapshot{ObjectMeta: metav1.ObjectMeta{Name: "foo*"}},
			ContainElement(InvalidField("metadata.name")),
		),
		Entry("no volume ref",
			&storage.VolumeSnapshot{},
			Not(ContainElement(RequiredField("spec.volumeRef"))),
		),
		Entry("invalid volume ref name",
			&storage.VolumeSnapshot{
				Spec: storage.VolumeSnapshotSpec{
					VolumeRef: &corev1.LocalObjectReference{Name: "foo*"},
				},
			},
			ContainElement(InvalidField("spec.volumeRef.name")),
		),
	)

	DescribeTable("ValidateVolumeSnapshotUpdate",
		func(newVolumeSnapshot, oldVolumeSnapshot *storage.VolumeSnapshot, match types.GomegaMatcher) {
			errList := ValidateVolumeSnapshotUpdate(newVolumeSnapshot, oldVolumeSnapshot)
			Expect(errList).To(match)
		},
		Entry("immutable volumeRef",
			&storage.VolumeSnapshot{
				Spec: storage.VolumeSnapshotSpec{
					VolumeRef: &corev1.LocalObjectReference{Name: "foo"},
				},
			},
			&storage.VolumeSnapshot{
				Spec: storage.VolumeSnapshotSpec{
					VolumeRef: &corev1.LocalObjectReference{Name: "bar"},
				},
			},
			ContainElement(ImmutableField("spec.volumeRef")),
		),
	)
})
