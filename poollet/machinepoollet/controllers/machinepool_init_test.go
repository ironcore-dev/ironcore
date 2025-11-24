// SPDX-FileCopyrightText: 2025 SAP SE or an SAP affiliate company and IronCore contributors
// SPDX-License-Identifier: Apache-2.0

package controllers_test

import (
	"context"

	commonv1alpha1 "github.com/ironcore-dev/ironcore/api/common/v1alpha1"
	computev1alpha1 "github.com/ironcore-dev/ironcore/api/compute/v1alpha1"
	"github.com/ironcore-dev/ironcore/poollet/machinepoollet/controllers"
	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	. "sigs.k8s.io/controller-runtime/pkg/envtest/komega"
)

var _ = Describe("MachinePoolInit", func() {
	It("should set topology labels", func(ctx SpecContext) {
		initializedCalled := false

		mpi := &controllers.MachinePoolInit{
			Client:          k8sClient,
			MachinePoolName: "test-pool",
			ProviderID:      "provider-123",
			TopologyLabels: map[commonv1alpha1.TopologyLabel]string{
				commonv1alpha1.TopologyLabelRegion: "foo-region-1",
				commonv1alpha1.TopologyLabelZone:   "foo-zone-1",
			},
			OnInitialized: func(ctx context.Context) error {
				initializedCalled = true
				return nil
			},
		}

		Expect(mpi.Start(ctx)).ToNot(HaveOccurred())
		Expect(initializedCalled).To(BeTrue())

		pool := &computev1alpha1.MachinePool{
			ObjectMeta: metav1.ObjectMeta{
				Name: "test-pool",
			},
		}

		Eventually(Object(pool)).Should(SatisfyAll(
			HaveField("ObjectMeta.Labels", HaveKeyWithValue("topology.ironcore.dev/region", "foo-region-1")),
			HaveField("ObjectMeta.Labels", HaveKeyWithValue("topology.ironcore.dev/zone", "foo-zone-1")),
		))
	})
})
