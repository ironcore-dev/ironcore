// Copyright 2022 OnMetal authors
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
	computev1alpha1 "github.com/onmetal/onmetal-api/api/compute/v1alpha1"
	machinebrokerv1alpha1 "github.com/onmetal/onmetal-api/broker/machinebroker/api/v1alpha1"
	ori "github.com/onmetal/onmetal-api/ori/apis/machine/v1alpha1"
	orimeta "github.com/onmetal/onmetal-api/ori/apis/meta/v1alpha1"
	machinepoolletv1alpha1 "github.com/onmetal/onmetal-api/poollet/machinepoollet/api/v1alpha1"
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
						machinepoolletv1alpha1.MachineUIDLabel: "root-uid",
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

		machine := &computev1alpha1.Machine{}
		Expect(k8sClient.Get(ctx, client.ObjectKey{Namespace: ns.Name, Name: res.Machine.Metadata.Id}, machine)).To(Succeed())
		Expect(machine.Labels).To(HaveKeyWithValue(machinebrokerv1alpha1.DownwardAPIRootMachineUIDLabel, "root-uid"))
	})
})
