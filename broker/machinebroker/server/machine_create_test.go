// Copyright 2022 IronCore authors
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
//      http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.

package server_test

import (
	computev1alpha1 "github.com/ironcore-dev/ironcore/api/compute/v1alpha1"
	machinebrokerv1alpha1 "github.com/ironcore-dev/ironcore/broker/machinebroker/api/v1alpha1"
	"github.com/ironcore-dev/ironcore/broker/machinebroker/apiutils"
	ori "github.com/ironcore-dev/ironcore/ori/apis/machine/v1alpha1"
	orimeta "github.com/ironcore-dev/ironcore/ori/apis/meta/v1alpha1"
	machinepoolletv1alpha1 "github.com/ironcore-dev/ironcore/poollet/machinepoollet/api/v1alpha1"
	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
	"sigs.k8s.io/controller-runtime/pkg/client"
)

var _ = Describe("CreateMachine", func() {
	ns, srv := SetupTest()
	machineClass := SetupMachineClass()

	It("should correctly create a machine", func(ctx SpecContext) {
		By("creating a machine")
		res, err := srv.CreateMachine(ctx, &ori.CreateMachineRequest{
			Machine: &ori.Machine{
				Metadata: &orimeta.ObjectMetadata{
					Labels: map[string]string{
						machinepoolletv1alpha1.MachineUIDLabel: "foobar",
					},
				},
				Spec: &ori.MachineSpec{
					Power: ori.Power_POWER_ON,
					Image: &ori.ImageSpec{
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
			machinepoolletv1alpha1.DownwardAPILabel("root-machine-uid"): "foobar",
			machinebrokerv1alpha1.CreatedLabel:                          "true",
			machinebrokerv1alpha1.ManagerLabel:                          machinebrokerv1alpha1.MachineBrokerManager,
		}))
		encodedORIAnnotations, err := apiutils.EncodeAnnotationsAnnotation(nil)
		Expect(err).NotTo(HaveOccurred())
		encodedORILabels, err := apiutils.EncodeLabelsAnnotation(map[string]string{
			machinepoolletv1alpha1.MachineUIDLabel: "foobar",
		})
		Expect(err).NotTo(HaveOccurred())
		Expect(ironcoreMachine.Annotations).To(Equal(map[string]string{
			machinebrokerv1alpha1.AnnotationsAnnotation: encodedORIAnnotations,
			machinebrokerv1alpha1.LabelsAnnotation:      encodedORILabels,
		}))
		Expect(ironcoreMachine.Spec.Power).To(Equal(computev1alpha1.PowerOn))
		Expect(ironcoreMachine.Spec.Image).To(Equal("example.org/foo:latest"))
		Expect(ironcoreMachine.Spec.MachineClassRef.Name).To(Equal(machineClass.Name))
	})
})
