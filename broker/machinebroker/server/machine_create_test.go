// SPDX-FileCopyrightText: 2023 SAP SE or an SAP affiliate company and IronCore contributors
// SPDX-License-Identifier: Apache-2.0

package server_test

import (
	computev1alpha1 "github.com/ironcore-dev/ironcore/api/compute/v1alpha1"
	machinebrokerv1alpha1 "github.com/ironcore-dev/ironcore/broker/machinebroker/api/v1alpha1"
	"github.com/ironcore-dev/ironcore/broker/machinebroker/apiutils"
	iri "github.com/ironcore-dev/ironcore/iri/apis/machine/v1alpha1"
	irimeta "github.com/ironcore-dev/ironcore/iri/apis/meta/v1alpha1"
	machinepoolletv1alpha1 "github.com/ironcore-dev/ironcore/poollet/machinepoollet/api/v1alpha1"
	poolletutils "github.com/ironcore-dev/ironcore/utils/poollet"
	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
	"sigs.k8s.io/controller-runtime/pkg/client"
)

var _ = Describe("CreateMachine", func() {
	ns, srv := SetupTest()
	machineClass := SetupMachineClass()

	It("should correctly create a machine", func(ctx SpecContext) {
		By("creating a machine")
		res, err := srv.CreateMachine(ctx, &iri.CreateMachineRequest{
			Machine: &iri.Machine{
				Metadata: &irimeta.ObjectMetadata{
					Labels: map[string]string{
						machinepoolletv1alpha1.MachineUIDLabel: "foobar",
					},
				},
				Spec: &iri.MachineSpec{
					Power: iri.Power_POWER_ON,
					Image: &iri.ImageSpec{
						Image: "example.org/foo:latest",
					},
					Class: machineClass.Name,
				},
			},
		})
		Expect(err).NotTo(HaveOccurred())
		Expect(res).NotTo(BeNil())

		By("getting the ironcore machine")
		ironcoreMachine := &computev1alpha1.Machine{}
		ironcoreMachineKey := client.ObjectKey{Namespace: ns.Name, Name: res.Machine.Metadata.Id}
		Expect(k8sClient.Get(ctx, ironcoreMachineKey, ironcoreMachine)).To(Succeed())

		By("inspecting the ironcore machine")
		Expect(ironcoreMachine.Labels).To(Equal(map[string]string{
			poolletutils.DownwardAPILabel(machinepoolletv1alpha1.MachineDownwardAPIPrefix, "root-machine-uid"): "foobar",
			machinebrokerv1alpha1.CreatedLabel: "true",
			machinebrokerv1alpha1.ManagerLabel: machinebrokerv1alpha1.MachineBrokerManager,
		}))
		encodedIRIAnnotations, err := apiutils.EncodeAnnotationsAnnotation(nil)
		Expect(err).NotTo(HaveOccurred())
		encodedIRILabels, err := apiutils.EncodeLabelsAnnotation(map[string]string{
			machinepoolletv1alpha1.MachineUIDLabel: "foobar",
		})
		Expect(err).NotTo(HaveOccurred())
		Expect(ironcoreMachine.Annotations).To(Equal(map[string]string{
			machinebrokerv1alpha1.AnnotationsAnnotation: encodedIRIAnnotations,
			machinebrokerv1alpha1.LabelsAnnotation:      encodedIRILabels,
		}))
		Expect(ironcoreMachine.Spec.Power).To(Equal(computev1alpha1.PowerOn))
		Expect(ironcoreMachine.Spec.Image).To(Equal("example.org/foo:latest"))
		Expect(ironcoreMachine.Spec.MachineClassRef.Name).To(Equal(machineClass.Name))
	})
})
