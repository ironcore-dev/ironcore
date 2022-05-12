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
	commonv1alpha1 "github.com/onmetal/onmetal-api/apis/common/v1alpha1"
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

var _ = Describe("Volume", func() {
	DescribeTable("ValidateVolume",
		func(volume *storage.Volume, match types.GomegaMatcher) {
			errList := ValidateVolume(volume)
			Expect(errList).To(match)
		},
		Entry("missing name",
			&storage.Volume{},
			ContainElement(RequiredField("metadata.name")),
		),
		Entry("missing namespace",
			&storage.Volume{ObjectMeta: metav1.ObjectMeta{Name: "foo"}},
			ContainElement(RequiredField("metadata.namespace")),
		),
		Entry("bad name",
			&storage.Volume{ObjectMeta: metav1.ObjectMeta{Name: "foo*"}},
			ContainElement(InvalidField("metadata.name")),
		),
		Entry("no volume class ref",
			&storage.Volume{},
			ContainElement(RequiredField("spec.volumeClassRef")),
		),
		Entry("invalid volume ref name",
			&storage.Volume{
				Spec: storage.VolumeSpec{
					VolumeClassRef: corev1.LocalObjectReference{Name: "foo*"},
				},
			},
			ContainElement(InvalidField("spec.volumeClassRef.name")),
		),
		Entry("invalid claim ref name",
			&storage.Volume{
				Spec: storage.VolumeSpec{
					ClaimRef: &commonv1alpha1.LocalUIDReference{Name: "foo*"},
				},
			},
			ContainElement(InvalidField("spec.claimRef.name")),
		),
		Entry("invalid image pull secret ref name",
			&storage.Volume{
				Spec: storage.VolumeSpec{
					ImagePullSecretRef: &corev1.LocalObjectReference{Name: "foo*"},
				},
			},
			ContainElement(InvalidField("spec.imagePullSecretRef.name")),
		),
		Entry("no resources[storage]",
			&storage.Volume{},
			ContainElement(RequiredField("spec.resources[storage]")),
		),
		Entry("negative resources[storage]",
			&storage.Volume{
				Spec: storage.VolumeSpec{
					Resources: map[corev1.ResourceName]resource.Quantity{
						corev1.ResourceStorage: resource.MustParse("-1"),
					},
				},
			},
			ContainElement(InvalidField("spec.resources[storage]")),
		),
	)

	DescribeTable("ValidateVolumeUpdate",
		func(newVolume, oldVolume *storage.Volume, match types.GomegaMatcher) {
			errList := ValidateVolumeUpdate(newVolume, oldVolume)
			Expect(errList).To(match)
		},
		Entry("immutable volumeClassRef",
			&storage.Volume{
				Spec: storage.VolumeSpec{
					VolumeClassRef: corev1.LocalObjectReference{Name: "foo"},
				},
			},
			&storage.Volume{
				Spec: storage.VolumeSpec{
					VolumeClassRef: corev1.LocalObjectReference{Name: "bar"},
				},
			},
			ContainElement(ImmutableField("spec.volumeClassRef")),
		),
		Entry("immutable volumePoolRef if set",
			&storage.Volume{
				Spec: storage.VolumeSpec{
					VolumePoolRef: &corev1.LocalObjectReference{Name: "foo"},
				},
			},
			&storage.Volume{
				Spec: storage.VolumeSpec{
					VolumePoolRef: &corev1.LocalObjectReference{Name: "bar"},
				},
			},
			ContainElement(ImmutableField("spec.volumePoolRef")),
		),
		Entry("mutable volumePoolRef if not set",
			&storage.Volume{
				Spec: storage.VolumeSpec{
					VolumePoolRef: &corev1.LocalObjectReference{Name: "foo"},
				},
			},
			&storage.Volume{},
			Not(ContainElement(ImmutableField("spec.volumePoolRef"))),
		),
	)
})
