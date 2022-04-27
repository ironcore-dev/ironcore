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
	"time"

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
	corev1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/api/resource"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"sigs.k8s.io/controller-runtime/pkg/client"

	commonv1alpha1 "github.com/onmetal/onmetal-api/apis/common/v1alpha1"
	storagev1alpha1 "github.com/onmetal/onmetal-api/apis/storage/v1alpha1"
)

var _ = Describe("VolumeScheduler", func() {
	ns := SetupTest(ctx)

	It("should schedule volumes on volume pools", func() {
		By("creating a volume pool")
		volumePool := &storagev1alpha1.VolumePool{
			ObjectMeta: metav1.ObjectMeta{
				GenerateName: "test-pool-",
			},
		}
		Expect(k8sClient.Create(ctx, volumePool)).To(Succeed(), "failed to create volume pool")

		By("patching the volume pool status to contain a volume class")
		volumePoolBase := volumePool.DeepCopy()
		volumePool.Status.AvailableVolumeClasses = []corev1.LocalObjectReference{{Name: "my-volumeclass"}}
		Expect(k8sClient.Status().Patch(ctx, volumePool, client.MergeFrom(volumePoolBase))).
			To(Succeed(), "failed to patch volume pool status")

		By("creating a volume w/ the requested volume class")
		volume := &storagev1alpha1.Volume{
			ObjectMeta: metav1.ObjectMeta{
				Namespace:    ns.Name,
				GenerateName: "test-volume-",
			},
			Spec: storagev1alpha1.VolumeSpec{
				VolumeClassRef: corev1.LocalObjectReference{
					Name: "my-volumeclass",
				},
				Resources: corev1.ResourceList{
					corev1.ResourceStorage: resource.MustParse("1Gi"),
				},
			},
		}
		Expect(k8sClient.Create(ctx, volume)).To(Succeed(), "failed to create volume")

		By("waiting for the volume to be scheduled onto the volume pool")
		volumeKey := client.ObjectKeyFromObject(volume)
		Eventually(func(g Gomega) {
			Expect(k8sClient.Get(ctx, volumeKey, volume)).To(Succeed(), "failed to get volume")
			g.Expect(volume.Spec.VolumePoolRef.Name).To(Equal(volumePool.Name))
			g.Expect(volume.Status.State).To(Equal(storagev1alpha1.VolumeStatePending))
		}).Should(Succeed())
	})

	It("should schedule schedule volumes onto volume pools if the pool becomes available later than the volume", func() {
		By("creating a volume w/ the requested volume class")
		volume := &storagev1alpha1.Volume{
			ObjectMeta: metav1.ObjectMeta{
				Namespace:    ns.Name,
				GenerateName: "test-volume-",
			},
			Spec: storagev1alpha1.VolumeSpec{
				VolumeClassRef: corev1.LocalObjectReference{
					Name: "my-volumeclass",
				},
				Resources: corev1.ResourceList{
					corev1.ResourceStorage: resource.MustParse("1Gi"),
				},
			},
		}
		Expect(k8sClient.Create(ctx, volume)).To(Succeed(), "failed to create volume")

		By("waiting for the volume to indicate it is pending")
		volumeKey := client.ObjectKeyFromObject(volume)
		Eventually(func(g Gomega) {
			Expect(k8sClient.Get(ctx, volumeKey, volume)).To(Succeed())
			g.Expect(volume.Spec.VolumePoolRef.Name).To(BeEmpty())
			g.Expect(volume.Status.State).To(Equal(storagev1alpha1.VolumeStatePending))
		}).Should(Succeed())

		By("creating a volume pool")
		volumePool := &storagev1alpha1.VolumePool{
			ObjectMeta: metav1.ObjectMeta{
				GenerateName: "test-pool-",
			},
		}
		Expect(k8sClient.Create(ctx, volumePool)).To(Succeed(), "failed to create volume pool")

		By("patching the volume pool status to contain a volume class")
		volumePoolBase := volumePool.DeepCopy()
		volumePool.Status.AvailableVolumeClasses = []corev1.LocalObjectReference{{Name: "my-volumeclass"}}
		Expect(k8sClient.Status().Patch(ctx, volumePool, client.MergeFrom(volumePoolBase))).
			To(Succeed(), "failed to patch volume pool status")

		By("waiting for the volume to be scheduled onto the volume pool")
		Eventually(func() string {
			Expect(k8sClient.Get(ctx, volumeKey, volume)).To(Succeed(), "failed to get volume")
			return volume.Spec.VolumePoolRef.Name
		}).Should(Equal(volumePool.Name))
	})

	It("should schedule onto volume pools with matching labels", func() {
		By("creating a volume pool w/o matching labels")
		volumePoolNoMatchingLabels := &storagev1alpha1.VolumePool{
			ObjectMeta: metav1.ObjectMeta{
				GenerateName: "test-pool-",
			},
		}
		Expect(k8sClient.Create(ctx, volumePoolNoMatchingLabels)).To(Succeed(), "failed to create volume pool")

		By("patching the volume pool status to contain a volume class")
		volumePoolNoMatchingLabelsBase := volumePoolNoMatchingLabels.DeepCopy()
		volumePoolNoMatchingLabels.Status.AvailableVolumeClasses = []corev1.LocalObjectReference{{Name: "my-volumeclass"}}
		Expect(k8sClient.Status().Patch(ctx, volumePoolNoMatchingLabels, client.MergeFrom(volumePoolNoMatchingLabelsBase))).
			To(Succeed(), "failed to patch volume pool status")

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
		volumePoolMatchingLabelsBase := volumePoolMatchingLabels.DeepCopy()
		volumePoolMatchingLabels.Status.AvailableVolumeClasses = []corev1.LocalObjectReference{{Name: "my-volumeclass"}}
		Expect(k8sClient.Status().Patch(ctx, volumePoolMatchingLabels, client.MergeFrom(volumePoolMatchingLabelsBase))).
			To(Succeed(), "failed to patch volume pool status")

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
				VolumeClassRef: corev1.LocalObjectReference{
					Name: "my-volumeclass",
				},
				Resources: corev1.ResourceList{
					corev1.ResourceStorage: resource.MustParse("1Gi"),
				},
			},
		}
		Expect(k8sClient.Create(ctx, volume)).To(Succeed(), "failed to create volume")

		By("waiting for the volume to be scheduled onto the volume pool")
		volumeKey := client.ObjectKeyFromObject(volume)
		Eventually(func(g Gomega) {
			Expect(k8sClient.Get(ctx, volumeKey, volume)).To(Succeed(), "failed to get volume")
			g.Expect(volume.Spec.VolumePoolRef.Name).To(Equal(volumePoolMatchingLabels.Name))
			g.Expect(volume.Status.State).To(Equal(storagev1alpha1.VolumeStatePending))
		}).Should(Succeed())
	})

	It("should schedule a volume with corresponding tolerations onto a volume pool with taints", func() {
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
		volumePoolBase := taintedVolumePool.DeepCopy()
		taintedVolumePool.Status.AvailableVolumeClasses = []corev1.LocalObjectReference{{Name: "my-volumeclass"}}
		Expect(k8sClient.Status().Patch(ctx, taintedVolumePool, client.MergeFrom(volumePoolBase))).
			To(Succeed(), "failed to patch the volume pool status")

		By("creating a volume")
		volume := &storagev1alpha1.Volume{
			ObjectMeta: metav1.ObjectMeta{
				Namespace:    ns.Name,
				GenerateName: "test-volume-",
			},
			Spec: storagev1alpha1.VolumeSpec{
				VolumeClassRef: corev1.LocalObjectReference{
					Name: "my-volumeclass",
				},
				Resources: corev1.ResourceList{
					corev1.ResourceStorage: resource.MustParse("1Gi"),
				},
			},
		}
		Expect(k8sClient.Create(ctx, volume)).To(Succeed(), "failed to create the volume")

		By("observing the volume isn't scheduled onto the volume pool")
		volumeKey := client.ObjectKeyFromObject(volume)
		Consistently(func() string {
			Expect(k8sClient.Get(ctx, volumeKey, volume)).To(Succeed())
			return volume.Spec.VolumePoolRef.Name
		}, 1*time.Second, interval).Should(BeEmpty())

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
		Consistently(func() string {
			Expect(k8sClient.Get(ctx, volumeKey, volume)).To(Succeed())
			return volume.Spec.VolumePoolRef.Name
		}, 1*time.Second, interval).Should(BeEmpty())

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
			g.Expect(volume.Spec.VolumePoolRef.Name).To(Equal(taintedVolumePool.Name))
		}, timeout, interval).Should(Succeed())
	})
})
