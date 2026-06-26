// SPDX-FileCopyrightText: 2026 SAP SE or an SAP affiliate company and IronCore contributors
// SPDX-License-Identifier: Apache-2.0

package compute

import (
	"context"
	"time"

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/client-go/kubernetes/scheme"
	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/client/fake"

	computev1alpha1 "github.com/ironcore-dev/ironcore/api/compute/v1alpha1"
)

var _ = Describe("MachinePoolLifecycleReconciler healthData cleanup", func() {
	It("removes the healthData entry when the MachinePool no longer exists", func() {
		const poolName = "deleted-pool"

		fakeClient := fake.NewClientBuilder().WithScheme(scheme.Scheme).Build()

		r := &MachinePoolLifecycleReconciler{
			Client:      fakeClient,
			GracePeriod: 30 * time.Second,
		}
		r.setMachinePoolHealth(poolName, &MachinePoolHealth{
			lastChangeDetectedTime: time.Now(),
		})
		Expect(r.getMachinePoolHealth(poolName)).NotTo(BeNil())

		result, err := r.Reconcile(context.Background(), ctrl.Request{
			NamespacedName: client.ObjectKey{Name: poolName},
		})

		Expect(err).NotTo(HaveOccurred())
		Expect(result).To(Equal(ctrl.Result{}))
		Expect(r.getMachinePoolHealth(poolName)).To(BeNil(), "expected healthData entry to be cleaned up after NotFound")
	})

	It("keeps unrelated healthData entries intact when one MachinePool is deleted", func() {
		const deletedPool = "gone-pool"
		const otherPool = "other-pool"

		fakeClient := fake.NewClientBuilder().
			WithScheme(scheme.Scheme).
			WithObjects(&computev1alpha1.MachinePool{
				ObjectMeta: metav1.ObjectMeta{Name: otherPool},
			}).
			Build()

		r := &MachinePoolLifecycleReconciler{
			Client:      fakeClient,
			GracePeriod: 30 * time.Second,
		}
		other := &MachinePoolHealth{lastChangeDetectedTime: time.Now()}
		r.setMachinePoolHealth(deletedPool, &MachinePoolHealth{lastChangeDetectedTime: time.Now()})
		r.setMachinePoolHealth(otherPool, other)

		_, err := r.Reconcile(context.Background(), ctrl.Request{
			NamespacedName: client.ObjectKey{Name: deletedPool},
		})

		Expect(err).NotTo(HaveOccurred())
		Expect(r.getMachinePoolHealth(deletedPool)).To(BeNil())
		Expect(r.getMachinePoolHealth(otherPool)).To(Equal(other))
	})
})
