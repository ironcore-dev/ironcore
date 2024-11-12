package controllers_test

import (
	"github.com/gogo/protobuf/proto"
	_ "github.com/ironcore-dev/ironcore/api/common/v1alpha1"
	computev1alpha1 "github.com/ironcore-dev/ironcore/api/compute/v1alpha1"
	corev1alpha1 "github.com/ironcore-dev/ironcore/api/core/v1alpha1"
	_ "github.com/ironcore-dev/ironcore/api/networking/v1alpha1"
	iri "github.com/ironcore-dev/ironcore/iri/apis/machine/v1alpha1"
	testingmachine "github.com/ironcore-dev/ironcore/iri/testing/machine"
	machinepoolletv1alpha1 "github.com/ironcore-dev/ironcore/poollet/machinepoollet/api/v1alpha1"
	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
	_ "k8s.io/api/core/v1"
	corev1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/api/resource"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	. "sigs.k8s.io/controller-runtime/pkg/envtest/komega"
)

var _ = FDescribe("MachineController", func() {
	ns, mp, _, srv := SetupTest()

	It("Should create a reservation on a matching pool", func(ctx SpecContext) {

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
				Resources: corev1alpha1.ResourceList{
					corev1alpha1.ResourceCPU: resource.MustParse("1"),
				},
			},
		}
		Expect(k8sClient.Create(ctx, reservation)).To(Succeed())

		By("ensuring the ironcore reservation status is pending")
		Eventually(Object(reservation)).Should(HaveField("Status.Pools", ConsistOf(computev1alpha1.ReservationPoolStatus{
			Name:  mp.Name,
			State: computev1alpha1.ReservationStatePending,
		})))

		By("waiting for the runtime to report the reservation")
		Eventually(srv).Should(SatisfyAll(
			HaveField("Reservations", HaveLen(1)),
		))

		_, iriReservation := GetSingleMapEntry(srv.Reservations)

		By("inspecting the iri reservation")
		Expect(iriReservation.Metadata.Labels).To(HaveKeyWithValue(machinepoolletv1alpha1.DownwardAPILabel(fooDownwardAPILabel), fooAnnotationValue))

		By("setting the reservation state to accepted")
		iriReservation = &testingmachine.FakeReservation{Reservation: *proto.Clone(&iriReservation.Reservation).(*iri.Reservation)}
		iriReservation.Status.State = iri.ReservationState_RESERVATION_STATE_ACCEPTED
		srv.SetReservations([]*testingmachine.FakeReservation{iriReservation})

		By("ensuring the ironcore reservation status is pending accepted")
		Eventually(Object(reservation)).Should(HaveField("Status.Pools", ConsistOf(computev1alpha1.ReservationPoolStatus{
			Name:  mp.Name,
			State: computev1alpha1.ReservationStatePending,
		})))
	})
})
