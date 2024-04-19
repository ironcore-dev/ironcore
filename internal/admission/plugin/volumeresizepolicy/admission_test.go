// SPDX-FileCopyrightText: 2023 SAP SE or an SAP affiliate company and IronCore contributors
// SPDX-License-Identifier: Apache-2.0

package volumeresizepolicy_test

import (
	"context"

	corev1alpha1 "github.com/ironcore-dev/ironcore/api/core/v1alpha1"
	storagev1alpha1 "github.com/ironcore-dev/ironcore/api/storage/v1alpha1"
	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
	corev1 "k8s.io/api/core/v1"
	apierrors "k8s.io/apimachinery/pkg/api/errors"
	"k8s.io/apimachinery/pkg/api/resource"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"sigs.k8s.io/controller-runtime/pkg/client"
	. "sigs.k8s.io/controller-runtime/pkg/envtest/komega"
)

var _ = Describe("Admission", func() {
	ns, volumePool := SetupTest()

	var (
		volumeClassStatic     = &storagev1alpha1.VolumeClass{}
		volumeClassExpandOnly = &storagev1alpha1.VolumeClass{}
	)

	BeforeEach(func(ctx SpecContext) {
		By("creating an expand only VolumeClass")
		volumeClassExpandOnly = &storagev1alpha1.VolumeClass{
			ObjectMeta: metav1.ObjectMeta{
				Name: "expand-only",
			},
			Capabilities: corev1alpha1.ResourceList{
				corev1alpha1.ResourceIOPS: resource.MustParse("100"),
				corev1alpha1.ResourceTPS:  resource.MustParse("100"),
			},
			ResizePolicy: storagev1alpha1.ResizePolicyExpandOnly,
		}
		Expect(k8sClient.Create(ctx, volumeClassExpandOnly)).To(Succeed())
		DeferCleanup(func(ctx context.Context) error {
			return client.IgnoreNotFound(k8sClient.Delete(ctx, volumeClassExpandOnly))
		})

		By("creating a static VolumeClass")
		volumeClassStatic = &storagev1alpha1.VolumeClass{
			ObjectMeta: metav1.ObjectMeta{
				Name: "static",
			},
			Capabilities: corev1alpha1.ResourceList{
				corev1alpha1.ResourceIOPS: resource.MustParse("100"),
				corev1alpha1.ResourceTPS:  resource.MustParse("100"),
			},
			ResizePolicy: storagev1alpha1.ResizePolicyStatic,
		}
		Expect(k8sClient.Create(ctx, volumeClassStatic)).To(Succeed())
		DeferCleanup(func(ctx context.Context) error {
			return client.IgnoreNotFound(k8sClient.Delete(ctx, volumeClassStatic))
		})
	})

	It("should allow the resizing of a Volume if the VolumeClass supports it", func(ctx SpecContext) {
		By("creating a Volume")
		volume := &storagev1alpha1.Volume{
			ObjectMeta: metav1.ObjectMeta{
				Namespace:    ns.Name,
				GenerateName: "volume-",
			},
			Spec: storagev1alpha1.VolumeSpec{
				VolumeClassRef: &corev1.LocalObjectReference{Name: volumeClassExpandOnly.Name},
				VolumePoolRef:  &corev1.LocalObjectReference{Name: volumePool.Name},
				Resources: corev1alpha1.ResourceList{
					corev1alpha1.ResourceStorage: resource.MustParse("1Gi"),
				},
			},
		}
		Expect(k8sClient.Create(ctx, volume)).To(Succeed())

		By("patching the Volume to increase the Volume size")
		volumeBase := volume.DeepCopy()
		volume.Spec.Resources[corev1alpha1.ResourceStorage] = resource.MustParse("2Gi")
		Expect(k8sClient.Patch(ctx, volume, client.MergeFrom(volumeBase))).To(Succeed())

		By("ensuring that Volume has been resized")
		Consistently(Object(volume)).Should(SatisfyAll(
			HaveField("Spec.Resources", Equal(corev1alpha1.ResourceList{
				corev1alpha1.ResourceStorage: resource.MustParse("2Gi"),
			})),
		))
	})

	It("should not allow the resizing of a Volume if the VolumeClass does not support it", func(ctx SpecContext) {
		By("creating a Volume")
		volume := &storagev1alpha1.Volume{
			ObjectMeta: metav1.ObjectMeta{
				Namespace:    ns.Name,
				GenerateName: "volume-",
			},
			Spec: storagev1alpha1.VolumeSpec{
				VolumeClassRef: &corev1.LocalObjectReference{Name: volumeClassStatic.Name},
				VolumePoolRef:  &corev1.LocalObjectReference{Name: volumePool.Name},
				Resources: corev1alpha1.ResourceList{
					corev1alpha1.ResourceStorage: resource.MustParse("1Gi"),
				},
			},
		}
		Expect(k8sClient.Create(ctx, volume)).To(Succeed())

		By("patching the Volume to increase the Volume size")
		volumeBase := volume.DeepCopy()
		volume.Spec.Resources[corev1alpha1.ResourceStorage] = resource.MustParse("2Gi")
		Expect(k8sClient.Patch(ctx, volume, client.MergeFrom(volumeBase))).To(
			MatchError(apierrors.NewBadRequest("VolumeClass ResizePolicy does not allow resizing").Error()))

		By("ensuring that Volume has been resized")
		Consistently(Object(volume)).Should(SatisfyAll(
			HaveField("Spec.Resources", Equal(corev1alpha1.ResourceList{
				corev1alpha1.ResourceStorage: resource.MustParse("1Gi"),
			})),
		))
	})

	It("should not allow the shrinking of a Volume if the VolumeClass does not support it", func(ctx SpecContext) {
		By("creating a Volume")
		volume := &storagev1alpha1.Volume{
			ObjectMeta: metav1.ObjectMeta{
				Namespace:    ns.Name,
				GenerateName: "volume-",
			},
			Spec: storagev1alpha1.VolumeSpec{
				VolumeClassRef: &corev1.LocalObjectReference{Name: volumeClassExpandOnly.Name},
				VolumePoolRef:  &corev1.LocalObjectReference{Name: volumePool.Name},
				Resources: corev1alpha1.ResourceList{
					corev1alpha1.ResourceStorage: resource.MustParse("2Gi"),
				},
			},
		}
		Expect(k8sClient.Create(ctx, volume)).To(Succeed())

		By("patching the Volume to decrease the Volume size")
		volumeBase := volume.DeepCopy()
		volume.Spec.Resources[corev1alpha1.ResourceStorage] = resource.MustParse("1Gi")
		Expect(k8sClient.Patch(ctx, volume, client.MergeFrom(volumeBase))).To(
			MatchError(apierrors.NewBadRequest("VolumeClass ResizePolicy does not allow shrinking").Error()))

		By("ensuring that Volume has been resized")
		Consistently(Object(volume)).Should(SatisfyAll(
			HaveField("Spec.Resources", Equal(corev1alpha1.ResourceList{
				corev1alpha1.ResourceStorage: resource.MustParse("2Gi"),
			})),
		))
	})
})
