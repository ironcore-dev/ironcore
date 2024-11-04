package controllers_test

import (
	"github.com/gogo/protobuf/proto"
	_ "github.com/ironcore-dev/ironcore/api/common/v1alpha1"
	computev1alpha1 "github.com/ironcore-dev/ironcore/api/compute/v1alpha1"
	_ "github.com/ironcore-dev/ironcore/api/networking/v1alpha1"
	iri "github.com/ironcore-dev/ironcore/iri/apis/machine/v1alpha1"
	testingmachine "github.com/ironcore-dev/ironcore/iri/testing/machine"
	machinepoolletv1alpha1 "github.com/ironcore-dev/ironcore/poollet/machinepoollet/api/v1alpha1"
	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
	_ "k8s.io/api/core/v1"
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	. "sigs.k8s.io/controller-runtime/pkg/envtest/komega"
)

var _ = FDescribe("MachineController", func() {
	ns, mp, _, srv := SetupTest()

	It("Should create a machine with an ephemeral NIC and ensure claimed networkInterfaceRef matches the ephemeral NIC", func(ctx SpecContext) {

		By("creating a reservation")
		const fooAnnotationValue = "bar"
		reservation := &computev1alpha1.Reservation{
			ObjectMeta: metav1.ObjectMeta{
				Namespace:    ns.Name,
				GenerateName: "reservation-",
				Annotations: map[string]string{
					fooAnnotation: fooAnnotationValue,
				},
			},
			Spec: computev1alpha1.ReservationSpec{
				Pools: []corev1.LocalObjectReference{
					{Name: mp.Name},
				},
				Resources: nil,
			},
		}
		Expect(k8sClient.Create(ctx, reservation)).To(Succeed())

		By("waiting for the runtime to report the machine, volume and network interface")
		Eventually(srv).Should(SatisfyAll(
			HaveField("Reservations", HaveLen(1)),
		))

		_, iriReservation := GetSingleMapEntry(srv.Reservations)

		By("inspecting the iri machine")
		Expect(iriReservation.Metadata.Labels).To(HaveKeyWithValue(machinepoolletv1alpha1.DownwardAPILabel(fooDownwardAPILabel), fooAnnotationValue))

		By("waiting for the ironcore machine status to be up-to-date")
		Eventually(Object(reservation)).Should(SatisfyAll(
			HaveField("Status.ObservedGeneration", reservation.Status.Pools),
		))

		By("setting the network interface id in the machine status")
		iriReservation = &testingmachine.FakeReservation{Reservation: *proto.Clone(&iriReservation.Reservation).(*iri.Reservation)}
		iriReservation.Metadata.Generation = 1

		srv.SetReservations([]*testingmachine.FakeReservation{iriReservation})
	})
})
