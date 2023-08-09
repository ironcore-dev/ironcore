// Copyright 2021 OnMetal authors
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

package storage

import (
	"fmt"
	"math"

	corev1alpha1 "github.com/onmetal/onmetal-api/api/core/v1alpha1"
	. "github.com/onmetal/onmetal-api/utils/testing"
	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
	corev1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/api/resource"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"sigs.k8s.io/controller-runtime/pkg/client"
	. "sigs.k8s.io/controller-runtime/pkg/envtest/komega"

	commonv1alpha1 "github.com/onmetal/onmetal-api/api/common/v1alpha1"
	storagev1alpha1 "github.com/onmetal/onmetal-api/api/storage/v1alpha1"
)

var _ = Describe("VolumeScheduler", func() {
	ns := SetupNamespace(&k8sClient)
	volumeClass := SetupVolumeClass()

	BeforeEach(func(ctx SpecContext) {
		By("waiting for the cached client to report no volume pools")
		Eventually(New(cacheK8sClient).ObjectList(&storagev1alpha1.VolumePoolList{})).Should(HaveField("Items", BeEmpty()))
	})

	AfterEach(func(ctx SpecContext) {
		By("deleting all volume pools")
		Expect(k8sClient.DeleteAllOf(ctx, &storagev1alpha1.VolumePool{})).To(Succeed())
	})

	It("should schedule volumes on volume pools", func(ctx SpecContext) {
		By("creating a volume pool")
		volumePool := &storagev1alpha1.VolumePool{
			ObjectMeta: metav1.ObjectMeta{
				GenerateName: "test-pool-",
			},
		}
		Expect(k8sClient.Create(ctx, volumePool)).To(Succeed(), "failed to create volume pool")

		By("patching the volume pool status to contain a volume class")
		Eventually(UpdateStatus(volumePool, func() {
			volumePool.Status.AvailableVolumeClasses = []corev1.LocalObjectReference{{Name: volumeClass.Name}}
			volumePool.Status.Allocatable = corev1alpha1.ResourceList{
				corev1alpha1.ClassCountFor(corev1alpha1.ClassTypeVolumeClass, volumeClass.Name): resource.MustParse("100Gi"),
			}
		})).Should(Succeed())

		By("creating a volume w/ the requested volume class")
		volume := &storagev1alpha1.Volume{
			ObjectMeta: metav1.ObjectMeta{
				Namespace:    ns.Name,
				GenerateName: "test-volume-",
			},
			Spec: storagev1alpha1.VolumeSpec{
				VolumeClassRef: &corev1.LocalObjectReference{Name: volumeClass.Name},
				Resources: corev1alpha1.ResourceList{
					corev1alpha1.ResourceStorage: resource.MustParse("1Gi"),
				},
			},
		}
		Expect(k8sClient.Create(ctx, volume)).To(Succeed(), "failed to create volume")

		By("waiting for the volume to be scheduled onto the volume pool")
		volumeKey := client.ObjectKeyFromObject(volume)
		Eventually(func(g Gomega) {
			Expect(k8sClient.Get(ctx, volumeKey, volume)).To(Succeed(), "failed to get volume")
			g.Expect(volume.Spec.VolumePoolRef).To(Equal(&corev1.LocalObjectReference{Name: volumePool.Name}))
		}).Should(Succeed())
	})

	It("should schedule schedule volumes onto volume pools if the pool becomes available later than the volume", func(ctx SpecContext) {
		By("creating a volume w/ the requested volume class")
		volume := &storagev1alpha1.Volume{
			ObjectMeta: metav1.ObjectMeta{
				Namespace:    ns.Name,
				GenerateName: "test-volume-",
			},
			Spec: storagev1alpha1.VolumeSpec{
				VolumeClassRef: &corev1.LocalObjectReference{Name: volumeClass.Name},
				Resources: corev1alpha1.ResourceList{
					corev1alpha1.ResourceStorage: resource.MustParse("1Gi"),
				},
			},
		}
		Expect(k8sClient.Create(ctx, volume)).To(Succeed(), "failed to create volume")

		By("waiting for the volume to indicate it is pending")
		volumeKey := client.ObjectKeyFromObject(volume)
		Eventually(func(g Gomega) {
			Expect(k8sClient.Get(ctx, volumeKey, volume)).To(Succeed())
			g.Expect(volume.Spec.VolumePoolRef).To(BeNil())
		}).Should(Succeed())

		By("creating a volume pool")
		volumePool := &storagev1alpha1.VolumePool{
			ObjectMeta: metav1.ObjectMeta{
				GenerateName: "test-pool-",
			},
		}
		Expect(k8sClient.Create(ctx, volumePool)).To(Succeed(), "failed to create volume pool")

		By("patching the volume pool status to contain a volume class")
		Eventually(UpdateStatus(volumePool, func() {
			volumePool.Status.AvailableVolumeClasses = []corev1.LocalObjectReference{{Name: volumeClass.Name}}
			volumePool.Status.Allocatable = corev1alpha1.ResourceList{
				corev1alpha1.ClassCountFor(corev1alpha1.ClassTypeVolumeClass, volumeClass.Name): resource.MustParse("100Gi"),
			}
		})).Should(Succeed())

		By("waiting for the volume to be scheduled onto the volume pool")
		Eventually(func() *corev1.LocalObjectReference {
			Expect(k8sClient.Get(ctx, volumeKey, volume)).To(Succeed(), "failed to get volume")
			return volume.Spec.VolumePoolRef
		}).Should(Equal(&corev1.LocalObjectReference{Name: volumePool.Name}))
	})

	It("should schedule onto volume pools with matching labels", func(ctx SpecContext) {
		By("creating a volume pool w/o matching labels")
		volumePoolNoMatchingLabels := &storagev1alpha1.VolumePool{
			ObjectMeta: metav1.ObjectMeta{
				GenerateName: "test-pool-",
			},
		}
		Expect(k8sClient.Create(ctx, volumePoolNoMatchingLabels)).To(Succeed(), "failed to create volume pool")

		By("patching the volume pool status to contain a volume class")
		Eventually(UpdateStatus(volumePoolNoMatchingLabels, func() {
			volumePoolNoMatchingLabels.Status.AvailableVolumeClasses = []corev1.LocalObjectReference{{Name: volumeClass.Name}}
			volumePoolNoMatchingLabels.Status.Allocatable = corev1alpha1.ResourceList{
				corev1alpha1.ClassCountFor(corev1alpha1.ClassTypeVolumeClass, volumeClass.Name): resource.MustParse("100Gi"),
			}
		})).Should(Succeed())

		By("creating a volume pool w/ matching labels")
		volumePoolMatchingLabels := &storagev1alpha1.VolumePool{
			ObjectMeta: metav1.ObjectMeta{
				GenerateName: "test-pool-",
				Labels: map[string]string{
					"foo": "bar",
				},
			},
		}
		Expect(k8sClient.Create(ctx, volumePoolMatchingLabels)).To(Succeed(), "failed to create volume pool")

		By("patching the volume pool status to contain a volume class")
		Eventually(UpdateStatus(volumePoolMatchingLabels, func() {
			volumePoolMatchingLabels.Status.AvailableVolumeClasses = []corev1.LocalObjectReference{{Name: volumeClass.Name}}
			volumePoolMatchingLabels.Status.Allocatable = corev1alpha1.ResourceList{
				corev1alpha1.ClassCountFor(corev1alpha1.ClassTypeVolumeClass, volumeClass.Name): resource.MustParse("100Gi"),
			}
		})).Should(Succeed())

		By("creating a volume w/ the requested volume class")
		volume := &storagev1alpha1.Volume{
			ObjectMeta: metav1.ObjectMeta{
				Namespace:    ns.Name,
				GenerateName: "test-volume-",
			},
			Spec: storagev1alpha1.VolumeSpec{
				VolumePoolSelector: map[string]string{
					"foo": "bar",
				},
				VolumeClassRef: &corev1.LocalObjectReference{Name: volumeClass.Name},
				Resources: corev1alpha1.ResourceList{
					corev1alpha1.ResourceStorage: resource.MustParse("1Gi"),
				},
			},
		}
		Expect(k8sClient.Create(ctx, volume)).To(Succeed(), "failed to create volume")

		By("waiting for the volume to be scheduled onto the volume pool")
		volumeKey := client.ObjectKeyFromObject(volume)
		Eventually(func(g Gomega) {
			Expect(k8sClient.Get(ctx, volumeKey, volume)).To(Succeed(), "failed to get volume")
			g.Expect(volume.Spec.VolumePoolRef).To(Equal(&corev1.LocalObjectReference{Name: volumePoolMatchingLabels.Name}))
		}).Should(Succeed())
	})

	It("should schedule a volume with corresponding tolerations onto a volume pool with taints", func(ctx SpecContext) {
		By("creating a volume pool w/ taints")
		taintedVolumePool := &storagev1alpha1.VolumePool{
			ObjectMeta: metav1.ObjectMeta{
				GenerateName: "test-pool-",
			},
			Spec: storagev1alpha1.VolumePoolSpec{
				Taints: []commonv1alpha1.Taint{
					{
						Key:    "key",
						Value:  "value",
						Effect: commonv1alpha1.TaintEffectNoSchedule,
					},
					{
						Key:    "key1",
						Effect: commonv1alpha1.TaintEffectNoSchedule,
					},
				},
			},
		}
		Expect(k8sClient.Create(ctx, taintedVolumePool)).To(Succeed(), "failed to create the volume pool")

		By("patching the volume pool status to contain a volume class")
		Eventually(UpdateStatus(taintedVolumePool, func() {
			taintedVolumePool.Status.AvailableVolumeClasses = []corev1.LocalObjectReference{{Name: volumeClass.Name}}
			taintedVolumePool.Status.Allocatable = corev1alpha1.ResourceList{
				corev1alpha1.ClassCountFor(corev1alpha1.ClassTypeVolumeClass, volumeClass.Name): resource.MustParse("100Gi"),
			}
		})).Should(Succeed())

		By("creating a volume")
		volume := &storagev1alpha1.Volume{
			ObjectMeta: metav1.ObjectMeta{
				Namespace:    ns.Name,
				GenerateName: "test-volume-",
			},
			Spec: storagev1alpha1.VolumeSpec{
				VolumeClassRef: &corev1.LocalObjectReference{Name: volumeClass.Name},
				Resources: corev1alpha1.ResourceList{
					corev1alpha1.ResourceStorage: resource.MustParse("1Gi"),
				},
			},
		}
		Expect(k8sClient.Create(ctx, volume)).To(Succeed(), "failed to create the volume")

		By("observing the volume isn't scheduled onto the volume pool")
		volumeKey := client.ObjectKeyFromObject(volume)
		Consistently(func() *corev1.LocalObjectReference {
			Expect(k8sClient.Get(ctx, volumeKey, volume)).To(Succeed())
			return volume.Spec.VolumePoolRef
		}).Should(BeNil())

		By("patching the volume to contain only one of the corresponding tolerations")
		volumeBase := volume.DeepCopy()
		volume.Spec.Tolerations = append(volume.Spec.Tolerations, commonv1alpha1.Toleration{
			Key:      "key",
			Value:    "value",
			Effect:   commonv1alpha1.TaintEffectNoSchedule,
			Operator: commonv1alpha1.TolerationOpEqual,
		})
		Expect(k8sClient.Patch(ctx, volume, client.MergeFrom(volumeBase))).To(Succeed(), "failed to patch the volume's spec")

		By("observing the volume isn't scheduled onto the volume pool")
		Consistently(func() *corev1.LocalObjectReference {
			Expect(k8sClient.Get(ctx, volumeKey, volume)).To(Succeed())
			return volume.Spec.VolumePoolRef
		}).Should(BeNil())

		By("patching the volume to contain all of the corresponding tolerations")
		volumeBase = volume.DeepCopy()
		volume.Spec.Tolerations = append(volume.Spec.Tolerations, commonv1alpha1.Toleration{
			Key:      "key1",
			Effect:   commonv1alpha1.TaintEffectNoSchedule,
			Operator: commonv1alpha1.TolerationOpExists,
		})
		Expect(k8sClient.Patch(ctx, volume, client.MergeFrom(volumeBase))).To(Succeed(), "failed to patch the volume's spec")

		By("observing the volume is scheduled onto the volume pool")
		Eventually(func(g Gomega) {
			Expect(k8sClient.Get(ctx, volumeKey, volume)).To(Succeed(), "failed to get the volume")
			g.Expect(volume.Spec.VolumePoolRef).To(Equal(&corev1.LocalObjectReference{Name: taintedVolumePool.Name}))
		}).Should(Succeed())
	})

	It("should schedule volume on pool with most allocatable resources", func(ctx SpecContext) {
		By("creating a volume pool")
		volumePool := &storagev1alpha1.VolumePool{
			ObjectMeta: metav1.ObjectMeta{
				GenerateName: "test-pool-",
			},
		}
		Expect(k8sClient.Create(ctx, volumePool)).To(Succeed(), "failed to create volume pool")

		By("patching the volume pool status to contain a volume class")
		Eventually(UpdateStatus(volumePool, func() {
			volumePool.Status.AvailableVolumeClasses = []corev1.LocalObjectReference{{Name: volumeClass.Name}}
			volumePool.Status.Allocatable = corev1alpha1.ResourceList{
				corev1alpha1.ClassCountFor(corev1alpha1.ClassTypeVolumeClass, volumeClass.Name): resource.MustParse("100Gi"),
			}
		})).Should(Succeed())

		By("creating a second volume pool")
		secondVolumePool := &storagev1alpha1.VolumePool{
			ObjectMeta: metav1.ObjectMeta{
				GenerateName: "second-test-pool-",
			},
		}
		Expect(k8sClient.Create(ctx, secondVolumePool)).To(Succeed(), "failed to create the second volume pool")

		By("creating a second volume class")
		secondVolumeClass := &storagev1alpha1.VolumeClass{
			ObjectMeta: metav1.ObjectMeta{
				GenerateName: "second-volume-class-",
			},
			Capabilities: corev1alpha1.ResourceList{
				corev1alpha1.ResourceTPS:  resource.MustParse("50Mi"),
				corev1alpha1.ResourceIOPS: resource.MustParse("5000"),
			},
		}
		Expect(k8sClient.Create(ctx, secondVolumeClass)).To(Succeed(), "failed to create second volume class")

		By("patching the second volume pool status to contain a both volume classes")
		Eventually(UpdateStatus(secondVolumePool, func() {
			secondVolumePool.Status.AvailableVolumeClasses = []corev1.LocalObjectReference{
				{Name: volumeClass.Name},
				{Name: secondVolumeClass.Name},
			}
			secondVolumePool.Status.Allocatable = corev1alpha1.ResourceList{
				corev1alpha1.ClassCountFor(corev1alpha1.ClassTypeVolumeClass, volumeClass.Name):       resource.MustParse("50Gi"),
				corev1alpha1.ClassCountFor(corev1alpha1.ClassTypeVolumeClass, secondVolumeClass.Name): resource.MustParse("200Gi"),
			}
		})).Should(Succeed())

		By("creating a volume")
		volume := &storagev1alpha1.Volume{
			ObjectMeta: metav1.ObjectMeta{
				Namespace:    ns.Name,
				GenerateName: "test-volume-",
			},
			Spec: storagev1alpha1.VolumeSpec{
				VolumeClassRef: &corev1.LocalObjectReference{
					Name: volumeClass.Name,
				},
				Resources: corev1alpha1.ResourceList{
					corev1alpha1.ResourceStorage: resource.MustParse("5Gi"),
				},
			},
		}
		Expect(k8sClient.Create(ctx, volume)).To(Succeed(), "failed to create the volume")

		By("checking that the volume is scheduled onto the volume pool")
		Eventually(Object(volume)).Should(SatisfyAll(
			HaveField("Spec.VolumePoolRef.Name", Equal(volumePool.Name)),
		))
	})

	It("should schedule volumes evenly on pools", func(ctx SpecContext) {
		By("creating a volume pool")
		volumePool := &storagev1alpha1.VolumePool{
			ObjectMeta: metav1.ObjectMeta{
				GenerateName: "test-pool-",
			},
		}
		Expect(k8sClient.Create(ctx, volumePool)).To(Succeed(), "failed to create volume pool")

		By("patching the volume pool status to contain a volume class")
		Eventually(UpdateStatus(volumePool, func() {
			volumePool.Status.AvailableVolumeClasses = []corev1.LocalObjectReference{{Name: volumeClass.Name}}
			volumePool.Status.Allocatable = corev1alpha1.ResourceList{
				corev1alpha1.ClassCountFor(corev1alpha1.ClassTypeVolumeClass, volumeClass.Name): resource.MustParse("100Gi"),
			}
		})).Should(Succeed())

		By("creating a second volume pool")
		secondVolumePool := &storagev1alpha1.VolumePool{
			ObjectMeta: metav1.ObjectMeta{
				GenerateName: "second-test-pool-",
			},
		}
		Expect(k8sClient.Create(ctx, secondVolumePool)).To(Succeed(), "failed to create the second volume pool")

		By("patching the second volume pool status to contain a both volume classes")
		Eventually(UpdateStatus(secondVolumePool, func() {
			secondVolumePool.Status.AvailableVolumeClasses = []corev1.LocalObjectReference{
				{Name: volumeClass.Name},
			}
			secondVolumePool.Status.Allocatable = corev1alpha1.ResourceList{
				corev1alpha1.ClassCountFor(corev1alpha1.ClassTypeVolumeClass, volumeClass.Name): resource.MustParse("100Gi"),
			}
		})).Should(Succeed())

		By("creating volumes")
		var volumes []*storagev1alpha1.Volume
		for i := 0; i < 50; i++ {
			volume := &storagev1alpha1.Volume{
				ObjectMeta: metav1.ObjectMeta{
					Namespace:    ns.Name,
					GenerateName: fmt.Sprintf("test-volume-%d-", i),
				},
				Spec: storagev1alpha1.VolumeSpec{
					VolumeClassRef: &corev1.LocalObjectReference{
						Name: volumeClass.Name,
					},
					Resources: corev1alpha1.ResourceList{
						corev1alpha1.ResourceStorage: resource.MustParse("1Gi"),
					},
				},
			}
			Expect(k8sClient.Create(ctx, volume)).To(Succeed(), "failed to create the volume")
			volumes = append(volumes, volume)
		}

		By("checking that every volume is scheduled onto a volume pool")
		var numInstancesPool1, numInstancesPool2 int64
		for i := 0; i < 50; i++ {
			Eventually(Object(volumes[i])).Should(SatisfyAll(
				HaveField("Spec.VolumePoolRef", Not(BeNil())),
			))

			switch volumes[i].Spec.VolumePoolRef.Name {
			case volumePool.Name:
				numInstancesPool1++
			case secondVolumePool.Name:
				numInstancesPool2++
			}
		}

		By("checking that volume are roughly distributed")
		Expect(math.Abs(float64(numInstancesPool1 - numInstancesPool2))).To(BeNumerically("<", 5))
	})

	It("should schedule a volumes once the capacity is sufficient", func(ctx SpecContext) {
		By("creating a volume pool")
		volumePool := &storagev1alpha1.VolumePool{
			ObjectMeta: metav1.ObjectMeta{
				GenerateName: "test-pool-",
			},
		}
		Expect(k8sClient.Create(ctx, volumePool)).To(Succeed(), "failed to create volume pool")
		By("patching the volume pool status to contain a volume class")
		Eventually(UpdateStatus(volumePool, func() {
			volumePool.Status.AvailableVolumeClasses = []corev1.LocalObjectReference{{Name: volumeClass.Name}}
			volumePool.Status.Allocatable = corev1alpha1.ResourceList{
				corev1alpha1.ClassCountFor(corev1alpha1.ClassTypeVolumeClass, volumeClass.Name): resource.MustParse("5Gi"),
			}
		})).Should(Succeed())

		By("creating a volume")
		volume := &storagev1alpha1.Volume{
			ObjectMeta: metav1.ObjectMeta{
				Namespace:    ns.Name,
				GenerateName: "test-volume-",
			},
			Spec: storagev1alpha1.VolumeSpec{
				VolumeClassRef: &corev1.LocalObjectReference{
					Name: volumeClass.Name,
				},
				Resources: corev1alpha1.ResourceList{
					corev1alpha1.ResourceStorage: resource.MustParse("10Gi"),
				},
			},
		}
		Expect(k8sClient.Create(ctx, volume)).To(Succeed(), "failed to create the volume")

		By("checking that the volume is scheduled onto the volume pool")
		Consistently(Object(volume)).Should(SatisfyAll(
			HaveField("Spec.VolumePoolRef", BeNil()),
		))

		By("patching the volume pool status to contain a volume class")
		Eventually(UpdateStatus(volumePool, func() {
			volumePool.Status.Allocatable = corev1alpha1.ResourceList{
				corev1alpha1.ClassCountFor(corev1alpha1.ClassTypeVolumeClass, volumeClass.Name): resource.MustParse("20Gi"),
			}
		})).Should(Succeed())

		By("checking that the volume is scheduled onto the volume pool")
		Eventually(Object(volume)).Should(SatisfyAll(
			HaveField("Spec.VolumePoolRef", Equal(&corev1.LocalObjectReference{Name: volumePool.Name})),
		))
	})
})
