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

package controllers_test

import (
	corev1alpha1 "github.com/onmetal/onmetal-api/api/core/v1alpha1"
	storagev1alpha1 "github.com/onmetal/onmetal-api/api/storage/v1alpha1"
	ori "github.com/onmetal/onmetal-api/ori/apis/volume/v1alpha1"
	"github.com/onmetal/onmetal-api/ori/testing/volume"
	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
	corev1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/api/resource"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	. "sigs.k8s.io/controller-runtime/pkg/envtest/komega"
)

const (
	TestVolumePool = "test-volumepool"
)

var _ = Describe("VolumePoolController", func() {
	_, _, _, srv := SetupTest()

	It("should add volume classes to pool", func(ctx SpecContext) {
		srv.SetVolumeClasses([]*volume.FakeVolumeClass{})

		By("creating a volume class")
		vc := &storagev1alpha1.VolumeClass{
			ObjectMeta: metav1.ObjectMeta{
				GenerateName: "test-vc-1-",
			},
			Capabilities: corev1alpha1.ResourceList{
				corev1alpha1.ResourceTPS:  resource.MustParse("250Mi"),
				corev1alpha1.ResourceIOPS: resource.MustParse("1"),
			},
		}
		Expect(k8sClient.Create(ctx, vc)).To(Succeed(), "failed to create test volume class")

		srv.SetVolumeClasses([]*volume.FakeVolumeClass{
			{
				VolumeClass: ori.VolumeClass{
					Name: vc.Name,
					Capabilities: &ori.VolumeClassCapabilities{
						Tps:  262144000,
						Iops: 1,
					},
				},
			},
		})

		By("creating a volume pool")
		volumePool := &storagev1alpha1.VolumePool{
			ObjectMeta: metav1.ObjectMeta{
				Name: TestVolumePool,
			},
		}
		Expect(k8sClient.Create(ctx, volumePool)).To(Succeed(), "failed to create test volume pool")

		By("checking if the default volume class is present")
		Eventually(Object(volumePool)).Should(SatisfyAll(
			HaveField("Status.AvailableVolumeClasses", Equal([]corev1.LocalObjectReference{
				{
					Name: vc.Name,
				},
			}))),
		)

		By("creating a second volume class")
		vc2 := &storagev1alpha1.VolumeClass{
			ObjectMeta: metav1.ObjectMeta{
				GenerateName: "test-vc-2-",
			},
			Capabilities: corev1alpha1.ResourceList{
				corev1alpha1.ResourceTPS:  resource.MustParse("250Mi"),
				corev1alpha1.ResourceIOPS: resource.MustParse("2"),
			},
		}
		Expect(k8sClient.Create(ctx, vc2)).To(Succeed(), "failed to create test volume class")

		Eventually(Object(volumePool)).Should(SatisfyAll(
			HaveField("Status.AvailableVolumeClasses", HaveLen(1))),
		)

		srv.SetVolumeClasses([]*volume.FakeVolumeClass{
			{
				VolumeClass: ori.VolumeClass{
					Name: vc.Name,
					Capabilities: &ori.VolumeClassCapabilities{
						Tps:  262144000,
						Iops: 1,
					},
				},
			},
			{
				VolumeClass: ori.VolumeClass{
					Name: vc2.Name,
					Capabilities: &ori.VolumeClassCapabilities{
						Tps:  262144000,
						Iops: 2,
					},
				},
			},
		})

		By("checking if the second volume class is present")
		Eventually(Object(volumePool)).Should(SatisfyAll(
			HaveField("Status.AvailableVolumeClasses", Equal([]corev1.LocalObjectReference{
				{
					Name: vc.Name,
				},
				{
					Name: vc2.Name,
				},
			}))),
		)
	})
})
