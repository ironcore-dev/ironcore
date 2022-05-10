// Copyright 2022 OnMetal authors
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
//      http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.

package validation_test

import (
	"github.com/onmetal/onmetal-api/apis/storage"
	. "github.com/onmetal/onmetal-api/apis/storage/validation"
	. "github.com/onmetal/onmetal-api/testutils/validation"
	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
	"github.com/onsi/gomega/types"
	corev1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/api/resource"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

var _ = Describe("VolumeClaim", func() {
	DescribeTable("ValidateVolume",
		func(volume *storage.VolumeClaim, match types.GomegaMatcher) {
			errList := ValidateVolumeClaim(volume)
			Expect(errList).To(match)
		},
		Entry("missing name",
			&storage.VolumeClaim{},
			ContainElement(RequiredField("metadata.name")),
		),
		Entry("missing namespace",
			&storage.VolumeClaim{ObjectMeta: metav1.ObjectMeta{Name: "foo"}},
			ContainElement(RequiredField("metadata.namespace")),
		),
		Entry("bad name",
			&storage.VolumeClaim{ObjectMeta: metav1.ObjectMeta{Name: "foo*"}},
			ContainElement(InvalidField("metadata.name")),
		),
		Entry("no volume class ref",
			&storage.VolumeClaim{},
			ContainElement(RequiredField("spec.volumeClassRef")),
		),
		Entry("invalid volume ref name",
			&storage.VolumeClaim{
				Spec: storage.VolumeClaimSpec{
					VolumeClassRef: corev1.LocalObjectReference{Name: "foo*"},
				},
			},
			ContainElement(InvalidField("spec.volumeClassRef.name")),
		),
		Entry("invalid volume ref name",
			&storage.VolumeClaim{
				Spec: storage.VolumeClaimSpec{
					VolumeRef: &corev1.LocalObjectReference{Name: "foo*"},
				},
			},
			ContainElement(InvalidField("spec.volumeRef.name")),
		),
		Entry("no resources[storage]",
			&storage.VolumeClaim{},
			ContainElement(RequiredField("spec.resources[storage]")),
		),
		Entry("negative resources[storage]",
			&storage.VolumeClaim{
				Spec: storage.VolumeClaimSpec{
					Resources: map[corev1.ResourceName]resource.Quantity{
						corev1.ResourceStorage: resource.MustParse("-1"),
					},
				},
			},
			ContainElement(InvalidField("spec.resources[storage]")),
		),
	)

	DescribeTable("ValidateVolumeClaimUpdate",
		func(newVolumeClaim, oldVolumeClaim *storage.VolumeClaim, match types.GomegaMatcher) {
			errList := ValidateVolumeClaimUpdate(newVolumeClaim, oldVolumeClaim)
			Expect(errList).To(match)
		},
		Entry("immutable volumeClassRef",
			&storage.VolumeClaim{
				Spec: storage.VolumeClaimSpec{
					VolumeClassRef: corev1.LocalObjectReference{Name: "foo"},
				},
			},
			&storage.VolumeClaim{
				Spec: storage.VolumeClaimSpec{
					VolumeClassRef: corev1.LocalObjectReference{Name: "bar"},
				},
			},
			ContainElement(ImmutableField("spec")),
		),
		Entry("immutable volumeRef if set",
			&storage.VolumeClaim{
				Spec: storage.VolumeClaimSpec{
					VolumeRef: &corev1.LocalObjectReference{Name: "foo"},
				},
			},
			&storage.VolumeClaim{
				Spec: storage.VolumeClaimSpec{
					VolumeRef: &corev1.LocalObjectReference{Name: "bar"},
				},
			},
			ContainElement(ImmutableField("spec")),
		),
		Entry("mutable volumeRef if not set",
			&storage.VolumeClaim{
				Spec: storage.VolumeClaimSpec{
					VolumeRef: &corev1.LocalObjectReference{Name: "foo"},
				},
			},
			&storage.VolumeClaim{},
			Not(ContainElement(ImmutableField("spec"))),
		),
	)
})
