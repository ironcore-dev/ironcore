// SPDX-FileCopyrightText: 2023 SAP SE or an SAP affiliate company and IronCore contributors
// SPDX-License-Identifier: Apache-2.0

package controllers_test

import (
	corev1alpha1 "github.com/ironcore-dev/ironcore/api/core/v1alpha1"
	storagev1alpha1 "github.com/ironcore-dev/ironcore/api/storage/v1alpha1"
	iri "github.com/ironcore-dev/ironcore/iri/apis/volume/v1alpha1"
	"github.com/ironcore-dev/ironcore/iri/testing/volume"
	"github.com/ironcore-dev/ironcore/utils/quota"
	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
	corev1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/api/resource"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	. "sigs.k8s.io/controller-runtime/pkg/envtest/komega"
)

var _ = Describe("VolumePoolController", func() {
	ns, volumePool, volumeClass, expandableVolumeClass, srv := SetupTest()

	It("should add volume classes to pool", func(ctx SpecContext) {
		By("checking if the default volume classes are present")
		Eventually(Object(volumePool)).Should(SatisfyAll(
			HaveField("Status.AvailableVolumeClasses", ContainElements([]corev1.LocalObjectReference{
				{
					Name: volumeClass.Name,
				},
				{
					Name: expandableVolumeClass.Name,
				},
			}))),
		)

		By("creating a volume class")
		testVolumeClass := &storagev1alpha1.VolumeClass{
			ObjectMeta: metav1.ObjectMeta{
				GenerateName: "test-vc-1-",
			},
			Capabilities: corev1alpha1.ResourceList{
				corev1alpha1.ResourceTPS:  resource.MustParse("250Mi"),
				corev1alpha1.ResourceIOPS: resource.MustParse("100"),
			},
		}
		Expect(k8sClient.Create(ctx, testVolumeClass)).To(Succeed(), "failed to create test volume class")
		DeferCleanup(k8sClient.Delete, testVolumeClass)

		srv.SetVolumeClasses([]*volume.FakeVolumeClassStatus{
			{
				VolumeClassStatus: iri.VolumeClassStatus{
					VolumeClass: &iri.VolumeClass{
						Name: volumeClass.Name,
						Capabilities: &iri.VolumeClassCapabilities{
							Tps:  262144000,
							Iops: 15000,
						},
					},
				},
			},
			{
				VolumeClassStatus: iri.VolumeClassStatus{
					VolumeClass: &iri.VolumeClass{
						Name: expandableVolumeClass.Name,
						Capabilities: &iri.VolumeClassCapabilities{
							Tps:  262144000,
							Iops: 1000,
						},
					},
				},
			},
			{
				VolumeClassStatus: iri.VolumeClassStatus{
					VolumeClass: &iri.VolumeClass{
						Name: testVolumeClass.Name,
						Capabilities: &iri.VolumeClassCapabilities{
							Tps:  262144000,
							Iops: 100,
						},
					},
				},
			},
		})

		By("checking if the test volume class is present")
		Eventually(Object(volumePool)).Should(SatisfyAll(
			HaveField("Status.AvailableVolumeClasses", ContainElements([]corev1.LocalObjectReference{
				{
					Name: volumeClass.Name,
				},
				{
					Name: expandableVolumeClass.Name,
				},
				{
					Name: testVolumeClass.Name,
				},
			}))),
		)
	})

	It("should add volume classes to pool", func(ctx SpecContext) {
		By("checking if the default volume classes are present")
		Eventually(Object(volumePool)).Should(SatisfyAll(
			HaveField("Status.AvailableVolumeClasses", ContainElements([]corev1.LocalObjectReference{
				{
					Name: volumeClass.Name,
				},
				{
					Name: expandableVolumeClass.Name,
				},
			}))),
		)

		By("creating a volume class")
		srv.SetVolumeClasses([]*volume.FakeVolumeClassStatus{
			{
				VolumeClassStatus: iri.VolumeClassStatus{
					VolumeClass: &iri.VolumeClass{
						Name: volumeClass.Name,
						Capabilities: &iri.VolumeClassCapabilities{
							Tps:  262144000,
							Iops: 15000,
						},
					},
				},
			},
			{
				VolumeClassStatus: iri.VolumeClassStatus{
					VolumeClass: &iri.VolumeClass{
						Name: expandableVolumeClass.Name,
						Capabilities: &iri.VolumeClassCapabilities{
							Tps:  262144000,
							Iops: 1000,
						},
					},
				},
			},
			{
				VolumeClassStatus: iri.VolumeClassStatus{
					VolumeClass: &iri.VolumeClass{
						Name: "testVolumeClass.Name",
						Capabilities: &iri.VolumeClassCapabilities{
							Tps:  262144000,
							Iops: 100,
						},
					},
				},
			},
		})

		Eventually(Object(volumePool)).Should(SatisfyAll(
			HaveField("Status.AvailableVolumeClasses", ContainElements([]corev1.LocalObjectReference{
				{
					Name: volumeClass.Name,
				},
				{
					Name: expandableVolumeClass.Name,
				},
			}))),
		)

		testVolumeClass := &storagev1alpha1.VolumeClass{
			ObjectMeta: metav1.ObjectMeta{
				GenerateName: "test-vc-1-",
			},
			Capabilities: corev1alpha1.ResourceList{
				corev1alpha1.ResourceTPS:  resource.MustParse("250Mi"),
				corev1alpha1.ResourceIOPS: resource.MustParse("100"),
			},
		}
		Expect(k8sClient.Create(ctx, testVolumeClass)).To(Succeed(), "failed to create test volume class")
		DeferCleanup(k8sClient.Delete, testVolumeClass)

		By("checking if the test volume class is present")
		Eventually(Object(volumePool)).Should(SatisfyAll(
			HaveField("Status.AvailableVolumeClasses", ContainElements([]corev1.LocalObjectReference{
				{
					Name: volumeClass.Name,
				},
				{
					Name: expandableVolumeClass.Name,
				},
				{
					Name: testVolumeClass.Name,
				},
			}))),
		)
	})

	It("should calculate pool capacity", func(ctx SpecContext) {
		var (
			volumeClassCapacity, expandableVolumeClassCapacity = resource.MustParse("12Gi"), resource.MustParse("50Gi")
		)

		By("announcing the capacity")
		srv.SetVolumeClasses([]*volume.FakeVolumeClassStatus{
			{
				VolumeClassStatus: iri.VolumeClassStatus{
					VolumeClass: &iri.VolumeClass{
						Name: volumeClass.Name,
						Capabilities: &iri.VolumeClassCapabilities{
							Tps:  262144000,
							Iops: 15000,
						},
					},
					Quantity: volumeClassCapacity.Value(),
				},
			},
			{
				VolumeClassStatus: iri.VolumeClassStatus{
					VolumeClass: &iri.VolumeClass{
						Name: expandableVolumeClass.Name,
						Capabilities: &iri.VolumeClassCapabilities{
							Tps:  262144000,
							Iops: 1000,
						},
					},
					Quantity: expandableVolumeClassCapacity.Value(),
				},
			},
		})

		By("checking if the capacity is correct")
		Eventually(Object(volumePool)).Should(SatisfyAll(
			HaveField("Status.Capacity", Satisfy(func(capacity corev1alpha1.ResourceList) bool {
				return quota.Equals(capacity, corev1alpha1.ResourceList{
					corev1alpha1.ClassCountFor(corev1alpha1.ClassTypeVolumeClass, volumeClass.Name):           volumeClassCapacity,
					corev1alpha1.ClassCountFor(corev1alpha1.ClassTypeVolumeClass, expandableVolumeClass.Name): expandableVolumeClassCapacity,
				})
			})),
		))

		By("creating a volume")
		volume := &storagev1alpha1.Volume{
			ObjectMeta: metav1.ObjectMeta{
				Namespace:    ns.Name,
				GenerateName: "volume-",
			},
			Spec: storagev1alpha1.VolumeSpec{
				VolumeClassRef: &corev1.LocalObjectReference{Name: volumeClass.Name},
				VolumePoolRef:  &corev1.LocalObjectReference{Name: volumePool.Name},
				Resources: corev1alpha1.ResourceList{
					corev1alpha1.ResourceStorage: resource.MustParse("10Gi"),
				},
			},
		}
		Expect(k8sClient.Create(ctx, volume)).To(Succeed(), "failed to create volume")
		DeferCleanup(expectVolumeDeleted, volume)

		By("checking if the allocatable resources are correct")
		Eventually(Object(volumePool)).Should(SatisfyAll(
			HaveField("Status.Allocatable", Satisfy(func(allocatable corev1alpha1.ResourceList) bool {
				return quota.Equals(allocatable, corev1alpha1.ResourceList{
					corev1alpha1.ClassCountFor(corev1alpha1.ClassTypeVolumeClass, volumeClass.Name):           resource.MustParse("2Gi"),
					corev1alpha1.ClassCountFor(corev1alpha1.ClassTypeVolumeClass, expandableVolumeClass.Name): expandableVolumeClassCapacity,
				})
			})),
		))

		By("creating a second volume")
		secondVolume := &storagev1alpha1.Volume{
			ObjectMeta: metav1.ObjectMeta{
				Namespace:    ns.Name,
				GenerateName: "volume-",
			},
			Spec: storagev1alpha1.VolumeSpec{
				VolumeClassRef: &corev1.LocalObjectReference{Name: expandableVolumeClass.Name},
				VolumePoolRef:  &corev1.LocalObjectReference{Name: volumePool.Name},
				Resources: corev1alpha1.ResourceList{
					corev1alpha1.ResourceStorage: resource.MustParse("10Gi"),
				},
			},
		}
		Expect(k8sClient.Create(ctx, secondVolume)).To(Succeed(), "failed to create second volume")
		DeferCleanup(expectVolumeDeleted, secondVolume)

		By("checking if the allocatable resources are correct")
		Eventually(Object(volumePool)).Should(SatisfyAll(
			HaveField("Status.Allocatable", Satisfy(func(allocatable corev1alpha1.ResourceList) bool {
				return quota.Equals(allocatable, corev1alpha1.ResourceList{
					corev1alpha1.ClassCountFor(corev1alpha1.ClassTypeVolumeClass, volumeClass.Name):           resource.MustParse("2Gi"),
					corev1alpha1.ClassCountFor(corev1alpha1.ClassTypeVolumeClass, expandableVolumeClass.Name): resource.MustParse("40Gi"),
				})
			})),
		))
	})
})
