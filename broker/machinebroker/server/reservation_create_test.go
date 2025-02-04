// SPDX-FileCopyrightText: 2023 SAP SE or an SAP affiliate company and IronCore contributors
// SPDX-License-Identifier: Apache-2.0

package server_test

import (
	computev1alpha1 "github.com/ironcore-dev/ironcore/api/compute/v1alpha1"
	corev1alpha1 "github.com/ironcore-dev/ironcore/api/core/v1alpha1"
	machinebrokerv1alpha1 "github.com/ironcore-dev/ironcore/broker/machinebroker/api/v1alpha1"
	"github.com/ironcore-dev/ironcore/broker/machinebroker/apiutils"
	iri "github.com/ironcore-dev/ironcore/iri/apis/machine/v1alpha1"
	irimeta "github.com/ironcore-dev/ironcore/iri/apis/meta/v1alpha1"
	machinepoolletv1alpha1 "github.com/ironcore-dev/ironcore/poollet/machinepoollet/api/v1alpha1"
	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
	"k8s.io/apimachinery/pkg/api/resource"
	"sigs.k8s.io/controller-runtime/pkg/client"
)

var _ = Describe("CreateReservation", func() {
	ns, srv := SetupTest()

	cpuResource := resource.MustParse("1")
	cpu, err := cpuResource.Marshal()
	Expect(err).NotTo(HaveOccurred())

	It("should correctly create a reservation", func(ctx SpecContext) {
		By("creating a reservation")
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

		By("getting the ironcore reservation")
		ironcoreReservation := &computev1alpha1.Reservation{}
		ironcoreReservationKey := client.ObjectKey{Namespace: ns.Name, Name: res.Reservation.Metadata.Id}
		Expect(k8sClient.Get(ctx, ironcoreReservationKey, ironcoreReservation)).To(Succeed())

		By("inspecting the ironcore reservation")
		Expect(ironcoreReservation.Labels).To(Equal(map[string]string{
			machinepoolletv1alpha1.DownwardAPILabel("root-reservation-uid"): "foobar",
			machinebrokerv1alpha1.CreatedLabel:                              "true",
			machinebrokerv1alpha1.ManagerLabel:                              machinebrokerv1alpha1.ReservationBrokerManager,
		}))
		encodedIRIAnnotations, err := apiutils.EncodeAnnotationsAnnotation(nil)
		Expect(err).NotTo(HaveOccurred())
		encodedIRILabels, err := apiutils.EncodeLabelsAnnotation(map[string]string{
			machinepoolletv1alpha1.ReservationUIDLabel: "foobar",
		})
		Expect(err).NotTo(HaveOccurred())
		Expect(ironcoreReservation.Annotations).To(Equal(map[string]string{
			machinebrokerv1alpha1.AnnotationsAnnotation: encodedIRIAnnotations,
			machinebrokerv1alpha1.LabelsAnnotation:      encodedIRILabels,
		}))

		Expect(ironcoreReservation.Spec.Resources).To(Equal(corev1alpha1.ResourceList{
			corev1alpha1.ResourceCPU: resource.MustParse("1"),
		}))
	})
})
