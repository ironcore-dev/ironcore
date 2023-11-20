// SPDX-FileCopyrightText: 2023 SAP SE or an SAP affiliate company and IronCore contributors
// SPDX-License-Identifier: Apache-2.0

package addresses_test

import (
	"os"

	computev1alpha1 "github.com/ironcore-dev/ironcore/api/compute/v1alpha1"
	. "github.com/ironcore-dev/ironcore/poollet/machinepoollet/addresses"
	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
)

var _ = Describe("Addresses", func() {
	Describe("Get", func() {
		BeforeEach(func() {
			envVars := []string{
				KubernetesServiceName,
				KubernetesPodNamespaceEnvVar,
				KubernetesClusterDomainEnvVar,
			}

			By("registering cleanup for all relevant env vars")
			for _, envVar := range envVars {
				DeferCleanup(os.Setenv, envVar, os.Getenv(envVar))
			}

			By("unsetting all relevant env vars")
			for _, envVar := range envVars {
				Expect(os.Unsetenv(envVar)).To(Succeed())
			}
		})

		It("should return the addresses from the specified file if specified", func() {
			addresses, err := Get(&GetOptions{Filename: "./testdata/addresses.yaml"})
			Expect(err).NotTo(HaveOccurred())

			Expect(addresses).To(Equal([]computev1alpha1.MachinePoolAddress{
				{
					Type:    computev1alpha1.MachinePoolHostName,
					Address: "foo.bar",
				},
				{
					Type:    computev1alpha1.MachinePoolInternalIP,
					Address: "10.0.0.1",
				},
			}))
		})

		It("should return an error if nothing is specified and the in-cluster env vars are not present", func() {
			Expect(os.Setenv(KubernetesServiceName, "10.0.0.1")).To(Succeed())
			_, err := Get()
			Expect(err).To(MatchError(ErrNotInCluster))
		})

		It("should return the values extracted from the in-cluster env vars", func() {
			Expect(os.Setenv(KubernetesServiceName, "10.0.0.1")).To(Succeed())
			Expect(os.Setenv(KubernetesPodNamespaceEnvVar, "foo")).To(Succeed())

			Expect(Get()).To(ConsistOf(
				computev1alpha1.MachinePoolAddress{
					Type:    computev1alpha1.MachinePoolInternalDNS,
					Address: "10.0.0.1.foo.svc.cluster.local",
				},
				computev1alpha1.MachinePoolAddress{
					Type:    computev1alpha1.MachinePoolInternalIP,
					Address: "10.0.0.1",
				},
			))
		})
	})
})
