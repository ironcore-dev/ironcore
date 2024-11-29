// SPDX-FileCopyrightText: 2023 SAP SE or an SAP affiliate company and IronCore contributors
// SPDX-License-Identifier: Apache-2.0

package server_test

import (
	corev1alpha1 "github.com/ironcore-dev/ironcore/api/core/v1alpha1"
	iri "github.com/ironcore-dev/ironcore/iri/apis/machine/v1alpha1"
	irimeta "github.com/ironcore-dev/ironcore/iri/apis/meta/v1alpha1"
	machinepoolletv1alpha1 "github.com/ironcore-dev/ironcore/poollet/machinepoollet/api/v1alpha1"
	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
	"k8s.io/apimachinery/pkg/api/resource"
)

var _ = Describe("ListReservations", func() {
	_, srv := SetupTest()

	cpuResource := resource.MustParse("1")
	cpu, err := cpuResource.Marshal()
	Expect(err).NotTo(HaveOccurred())

	It("should correctly list reservations", func(ctx SpecContext) {
		By("creating multiple reservations")
		const noOfReservations = 4

		reservations := make([]any, noOfReservations)
		for i := 0; i < noOfReservations; i++ {
			res, err := srv.CreateReservation(ctx, &iri.CreateReservationRequest{
				Reservation: &iri.Reservation{
					Metadata: &irimeta.ObjectMetadata{
						Labels: map[string]string{
							machinepoolletv1alpha1.ReservationUIDLabel: "foobar",
						},
					},
					Spec: &iri.ReservationSpec{
						Resources: map[string][]byte{
							string(corev1alpha1.ResourceCPU): cpu,
						},
					},
				},
			})
			Expect(err).NotTo(HaveOccurred())
			Expect(res).NotTo(BeNil())
			reservations[i] = res.Reservation
		}

		By("listing the reservations")
		Expect(srv.ListReservations(ctx, &iri.ListReservationsRequest{})).To(HaveField("Reservations", ConsistOf(reservations...)))
	})
})
