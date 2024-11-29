// SPDX-FileCopyrightText: 2023 SAP SE or an SAP affiliate company and IronCore contributors
// SPDX-License-Identifier: Apache-2.0

package server_test

import (
	computev1alpha1 "github.com/ironcore-dev/ironcore/api/compute/v1alpha1"
	corev1alpha1 "github.com/ironcore-dev/ironcore/api/core/v1alpha1"
	iri "github.com/ironcore-dev/ironcore/iri/apis/machine/v1alpha1"
	irimeta "github.com/ironcore-dev/ironcore/iri/apis/meta/v1alpha1"
	machinepoolletv1alpha1 "github.com/ironcore-dev/ironcore/poollet/machinepoollet/api/v1alpha1"
	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
	"sigs.k8s.io/controller-runtime/pkg/client"
)

var _ = Describe("DeleteReservation", func() {
	ns, srv := SetupTest()

	It("should correctly delete a reservation", func(ctx SpecContext) {
		By("creating a reservation")

		//corev1alpha1.ResourceCPU:    resource.MustParse("1"),
		//corev1alpha1.ResourceMemory: resource.MustParse("1Gi"),
		res, err := srv.CreateReservation(ctx, &iri.CreateReservationRequest{Reservation: &iri.Reservation{
			Metadata: &irimeta.ObjectMetadata{
				Labels: map[string]string{
					machinepoolletv1alpha1.ReservationUIDLabel: "foobar",
				},
			},
			Spec: &iri.ReservationSpec{
				Resources: map[string][]byte{
					string(corev1alpha1.ResourceCPU):    nil,
					string(corev1alpha1.ResourceMemory): nil,
				},
			},
		}})
		Expect(err).NotTo(HaveOccurred())
		Expect(res).NotTo(BeNil())

		By("deleting the reservation")
		reservationID := res.Reservation.Metadata.Id
		Expect(srv.DeleteReservation(ctx, &iri.DeleteReservationRequest{
			ReservationId: reservationID,
		})).Error().NotTo(HaveOccurred())

		By("listing for ironcore reservations in the namespace")
		ironcoreReservationList := &computev1alpha1.ReservationList{}
		Expect(k8sClient.List(ctx, ironcoreReservationList, client.InNamespace(ns.Name))).To(Succeed())

		By("asserting there are no ironcore reservations in the returned list")
		Expect(ironcoreReservationList.Items).To(BeEmpty())
	})
})
