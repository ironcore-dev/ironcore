// SPDX-FileCopyrightText: 2023 SAP SE or an SAP affiliate company and IronCore contributors
// SPDX-License-Identifier: Apache-2.0

package validation

import (
	"github.com/ironcore-dev/ironcore/internal/apis/compute"
	"github.com/ironcore-dev/ironcore/internal/apis/core"
	. "github.com/ironcore-dev/ironcore/internal/testutils/validation"
	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
	"github.com/onsi/gomega/types"
	"k8s.io/apimachinery/pkg/api/resource"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

var _ = Describe("Reservation", func() {
	DescribeTable("ValidateReservation",
		func(machine *compute.Reservation, match types.GomegaMatcher) {
			errList := ValidateReservation(machine)
			Expect(errList).To(match)
		},
		Entry("missing name",
			&compute.Reservation{},
			ContainElement(RequiredField("metadata.name")),
		),
		Entry("missing namespace",
			&compute.Reservation{ObjectMeta: metav1.ObjectMeta{Name: "foo"}},
			ContainElement(RequiredField("metadata.namespace")),
		),
		Entry("bad name",
			&compute.Reservation{ObjectMeta: metav1.ObjectMeta{Name: "foo*"}},
			ContainElement(InvalidField("metadata.name")),
		),
		Entry("bad resources",
			&compute.Reservation{
				ObjectMeta: metav1.ObjectMeta{Name: "foo"},
				Spec: compute.ReservationSpec{
					Resources: map[core.ResourceName]resource.Quantity{
						core.ResourceCPU: resource.MustParse("-1"),
					},
				},
			},
			ContainElement(InvalidField("spec.resources[cpu]")),
		),
	)

	DescribeTable("ValidateReservationUpdate",
		func(newReservation, oldReservation *compute.Reservation, match types.GomegaMatcher) {
			errList := ValidateReservationUpdate(newReservation, oldReservation)
			Expect(errList).To(match)
		},
		Entry("immutable resources",
			&compute.Reservation{
				Spec: compute.ReservationSpec{
					Resources: map[core.ResourceName]resource.Quantity{
						core.ResourceCPU: resource.MustParse("1"),
					},
				},
			},
			&compute.Reservation{
				Spec: compute.ReservationSpec{
					Resources: map[core.ResourceName]resource.Quantity{
						core.ResourceCPU: resource.MustParse("2"),
					},
				},
			},
			ContainElement(ImmutableField("spec.resources")),
		),
	)
})
