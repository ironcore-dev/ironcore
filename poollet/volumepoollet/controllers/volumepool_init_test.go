// SPDX-FileCopyrightText: 2023 SAP SE or an SAP affiliate company and IronCore contributors
// SPDX-License-Identifier: Apache-2.0

package controllers_test

import (
	"context"

	commonv1alpha1 "github.com/ironcore-dev/ironcore/api/common/v1alpha1"
	storagev1alpha1 "github.com/ironcore-dev/ironcore/api/storage/v1alpha1"
	"github.com/ironcore-dev/ironcore/poollet/volumepoollet/controllers"
	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"sigs.k8s.io/controller-runtime/pkg/client"
	. "sigs.k8s.io/controller-runtime/pkg/envtest/komega"
)

var _ = Describe("VolumePoolInit", func() {
	It("should set topology labels", func(ctx SpecContext) {
		initializedCalled := false

		vpi := &controllers.VolumePoolInit{
			Client:         k8sClient,
			VolumePoolName: "test-pool",
			ProviderID:     "provider-123",
			TopologyLabels: map[commonv1alpha1.TopologyLabel]string{
				commonv1alpha1.TopologyLabelRegion: "foo-region-1",
				commonv1alpha1.TopologyLabelZone:   "foo-zone-1",
			},
			OnInitialized: func(ctx context.Context) error {
				initializedCalled = true
				return nil
			},
		}

		Expect(vpi.Start(ctx)).To(Succeed())

		pool := &storagev1alpha1.VolumePool{
			ObjectMeta: metav1.ObjectMeta{
				Name: "test-pool",
			},
		}
		Expect(k8sClient.Get(ctx, client.ObjectKeyFromObject(pool), pool)).To(Succeed())
		DeferCleanup(k8sClient.Delete, pool)

		Expect(initializedCalled).To(BeTrue(), "OnInitialized should have been called")

		Eventually(Object(pool)).Should(SatisfyAll(
			HaveField("ObjectMeta.Labels", HaveKeyWithValue("topology.ironcore.dev/region", "foo-region-1")),
			HaveField("ObjectMeta.Labels", HaveKeyWithValue("topology.ironcore.dev/zone", "foo-zone-1")),
		))
	})
})
