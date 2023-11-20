// SPDX-FileCopyrightText: 2023 SAP SE or an SAP affiliate company and IronCore contributors
// SPDX-License-Identifier: Apache-2.0

package validation_test

import (
	commonv1alpha1 "github.com/ironcore-dev/ironcore/api/common/v1alpha1"
	"github.com/ironcore-dev/ironcore/internal/apis/core"
	"github.com/ironcore-dev/ironcore/internal/apis/storage"
	. "github.com/ironcore-dev/ironcore/internal/apis/storage/validation"
	. "github.com/ironcore-dev/ironcore/internal/testutils/validation"
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
			Not(ContainElement(RequiredField("spec.volumeClassRef"))),
		),
		Entry("invalid volume class ref name",
			&storage.Volume{
				Spec: storage.VolumeSpec{
					VolumeClassRef: &corev1.LocalObjectReference{Name: "foo*"},
				},
			},
			ContainElement(InvalidField("spec.volumeClassRef.name")),
		),
		Entry("valid volume pool ref name subdomain",
			&storage.Volume{
				Spec: storage.VolumeSpec{
					VolumePoolRef: &corev1.LocalObjectReference{Name: "foo.bar.baz"},
				},
			},
			Not(ContainElement(InvalidField("spec.volumePoolRef.name"))),
		),
		Entry("invalid claim ref name",
			&storage.Volume{
				Spec: storage.VolumeSpec{
					ClaimRef: &commonv1alpha1.LocalUIDReference{Name: "foo*"},
				},
			},
			ContainElement(InvalidField("spec.claimRef.name")),
		),
		Entry("unclaimable and claim ref",
			&storage.Volume{
				Spec: storage.VolumeSpec{
					Unclaimable: true,
					ClaimRef:    &commonv1alpha1.LocalUIDReference{Name: "foo"},
				},
			},
			ContainElement(ForbiddenField("spec.claimRef")),
		),
		Entry("classless: image pull secret ref",
			&storage.Volume{
				Spec: storage.VolumeSpec{
					ImagePullSecretRef: &corev1.LocalObjectReference{Name: "foo"},
				},
			},
			ContainElement(ForbiddenField("spec.imagePullSecretRef")),
		),
		Entry("classful: invalid image pull secret ref name",
			&storage.Volume{
				Spec: storage.VolumeSpec{
					VolumeClassRef:     &corev1.LocalObjectReference{Name: "foo"},
					ImagePullSecretRef: &corev1.LocalObjectReference{Name: "foo*"},
				},
			},
			ContainElement(InvalidField("spec.imagePullSecretRef.name")),
		),
		Entry("classless: no resources[storage]",
			&storage.Volume{},
			Not(ContainElement(RequiredField("spec.resources[storage]"))),
		),
		Entry("classless: any resources",
			&storage.Volume{
				Spec: storage.VolumeSpec{
					Resources: core.ResourceList{
						core.ResourceStorage: resource.MustParse("1Gi"),
					},
				},
			},
			ContainElement(ForbiddenField("spec.resources")),
		),
		Entry("classful: no resources[storage]",
			&storage.Volume{
				Spec: storage.VolumeSpec{
					VolumeClassRef: &corev1.LocalObjectReference{Name: "foo"},
				},
			},
			ContainElement(RequiredField("spec.resources[storage]")),
		),
		Entry("classful: negative resources[storage]",
			&storage.Volume{
				Spec: storage.VolumeSpec{
					VolumeClassRef: &corev1.LocalObjectReference{Name: "foo"},
					Resources: core.ResourceList{
						core.ResourceStorage: resource.MustParse("-1"),
					},
				},
			},
			ContainElement(InvalidField("spec.resources[storage]")),
		),
		Entry("valid encryption secret ref name",
			&storage.Volume{
				Spec: storage.VolumeSpec{
					VolumeClassRef: &corev1.LocalObjectReference{Name: "foo"},
					Encryption:     &storage.VolumeEncryption{SecretRef: corev1.LocalObjectReference{Name: "foo"}},
				},
			},
			Not(ContainElement(InvalidField("spec.encryption.secretRef.name"))),
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
					VolumeClassRef: &corev1.LocalObjectReference{Name: "foo"},
				},
			},
			&storage.Volume{
				Spec: storage.VolumeSpec{
					VolumeClassRef: &corev1.LocalObjectReference{Name: "bar"},
				},
			},
			ContainElement(ImmutableField("spec.volumeClassRef")),
		),
		Entry("classful: immutable volumePoolRef if set",
			&storage.Volume{
				Spec: storage.VolumeSpec{
					VolumeClassRef: &corev1.LocalObjectReference{Name: "foo"},
					VolumePoolRef:  &corev1.LocalObjectReference{Name: "foo"},
				},
			},
			&storage.Volume{
				Spec: storage.VolumeSpec{
					VolumeClassRef: &corev1.LocalObjectReference{Name: "foo"},
					VolumePoolRef:  &corev1.LocalObjectReference{Name: "bar"},
				},
			},
			ContainElement(ImmutableField("spec.volumePoolRef")),
		),
		Entry("classful: mutable volumePoolRef if not set",
			&storage.Volume{
				Spec: storage.VolumeSpec{
					VolumeClassRef: &corev1.LocalObjectReference{Name: "foo"},
					VolumePoolRef:  &corev1.LocalObjectReference{Name: "foo"},
				},
			},
			&storage.Volume{
				Spec: storage.VolumeSpec{
					VolumeClassRef: &corev1.LocalObjectReference{Name: "foo"},
				},
			},
			Not(ContainElement(ImmutableField("spec.volumePoolRef"))),
		),
		Entry("immutable encryption: modify encryption field",
			&storage.Volume{
				Spec: storage.VolumeSpec{
					VolumeClassRef: &corev1.LocalObjectReference{Name: "foo"},
					Encryption:     &storage.VolumeEncryption{SecretRef: corev1.LocalObjectReference{Name: "foo"}},
				},
			},
			&storage.Volume{
				Spec: storage.VolumeSpec{
					VolumeClassRef: &corev1.LocalObjectReference{Name: "foo"},
					Encryption:     &storage.VolumeEncryption{SecretRef: corev1.LocalObjectReference{Name: "bar"}},
				},
			},
			ContainElement(ImmutableField("spec.encryption")),
		),
		Entry("immutable encryption: add encryption field",
			&storage.Volume{
				Spec: storage.VolumeSpec{
					VolumeClassRef: &corev1.LocalObjectReference{Name: "foo"},
				},
			},
			&storage.Volume{
				Spec: storage.VolumeSpec{
					VolumeClassRef: &corev1.LocalObjectReference{Name: "foo"},
					Encryption:     &storage.VolumeEncryption{SecretRef: corev1.LocalObjectReference{Name: "foo"}},
				},
			},
			ContainElement(ImmutableField("spec.encryption")),
		),
		Entry("immutable encryption: remove encryption field",
			&storage.Volume{
				Spec: storage.VolumeSpec{
					VolumeClassRef: &corev1.LocalObjectReference{Name: "foo"},
					Encryption:     &storage.VolumeEncryption{SecretRef: corev1.LocalObjectReference{Name: "foo"}},
				},
			},
			&storage.Volume{
				Spec: storage.VolumeSpec{
					VolumeClassRef: &corev1.LocalObjectReference{Name: "foo"},
				},
			},
			ContainElement(ImmutableField("spec.encryption")),
		),
	)
})
