// Copyright 2023 OnMetal authors
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

package core_test

import (
	"crypto/x509"
	"crypto/x509/pkix"

	computev1alpha1 "github.com/onmetal/onmetal-api/api/compute/v1alpha1"
	networkingv1alpha1 "github.com/onmetal/onmetal-api/api/networking/v1alpha1"
	storagev1alpha1 "github.com/onmetal/onmetal-api/api/storage/v1alpha1"
	utilcertificate "github.com/onmetal/onmetal-api/utils/certificate"
	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
	certificatesv1 "k8s.io/api/certificates/v1"
	corev1 "k8s.io/api/core/v1"
	. "sigs.k8s.io/controller-runtime/pkg/envtest/komega"
)

var _ = Describe("CertificateApprovalController", func() {
	DescribeTable("certificate approval",
		func(ctx SpecContext, commonName, organization string) {
			By("creating a certificate signing request")
			csr, _, _, err := utilcertificate.GenerateAndCreateCertificateSigningRequest(
				ctx,
				k8sClient,
				certificatesv1.KubeAPIServerClientSignerName,
				&x509.CertificateRequest{
					Subject: pkix.Name{
						CommonName:   commonName,
						Organization: []string{organization},
					},
				},
				utilcertificate.DefaultKubeAPIServerClientGetUsages,
				nil,
			)
			Expect(err).NotTo(HaveOccurred())

			By("waiting for the csr to be approved and a certificate to be available")
			Eventually(ctx, Object(csr)).Should(
				HaveField("Status.Conditions", ContainElement(SatisfyAll(
					HaveField("Type", certificatesv1.CertificateApproved),
					HaveField("Status", corev1.ConditionTrue),
				))),
			)
		},
		Entry("machine pool",
			computev1alpha1.MachinePoolCommonName("my-pool"),
			computev1alpha1.MachinePoolsGroup,
		),
		Entry("volume pool",
			storagev1alpha1.VolumePoolCommonName("my-pool"),
			storagev1alpha1.VolumePoolsGroup,
		),
		Entry("bucket pool",
			storagev1alpha1.BucketPoolCommonName("my-pool"),
			storagev1alpha1.BucketPoolsGroup,
		),
		Entry("network plugin",
			networkingv1alpha1.NetworkPluginCommonName("my-plugin"),
			networkingv1alpha1.NetworkPluginsGroup,
		),
	)
})
