// SPDX-FileCopyrightText: 2026 SAP SE or an SAP affiliate company and IronCore contributors
// SPDX-License-Identifier: Apache-2.0

package compute

import (
	"crypto/x509"
	"fmt"

	computev1alpha1 "github.com/ironcore-dev/ironcore/api/compute/v1alpha1"
	"github.com/ironcore-dev/ironcore/internal/controllers/core/certificate/generic"
	"golang.org/x/exp/slices"
	authorizationv1 "k8s.io/api/authorization/v1"
	certificatesv1 "k8s.io/api/certificates/v1"
	"k8s.io/apimachinery/pkg/util/sets"
)

var PoolLifecycleControllerRequiredUsages = sets.New[certificatesv1.KeyUsage](
	certificatesv1.UsageDigitalSignature,
	certificatesv1.UsageKeyEncipherment,
	certificatesv1.UsageClientAuth,
)

func IsPoolLifecycleControllerClientCert(csr *certificatesv1.CertificateSigningRequest, x509cr *x509.CertificateRequest) bool {
	if csr.Spec.SignerName != certificatesv1.KubeAPIServerClientSignerName {
		return false
	}

	return ValidatePoolLifecycleControllerClientCSR(x509cr, sets.New(csr.Spec.Usages...)) == nil
}

func ValidatePoolLifecycleControllerClientCSR(req *x509.CertificateRequest, usages sets.Set[certificatesv1.KeyUsage]) error {
	if !slices.Equal([]string{computev1alpha1.PoolLifecycleControllersGroup}, req.Subject.Organization) {
		return fmt.Errorf("organization is not %s", computev1alpha1.PoolLifecycleControllersGroup)
	}

	if len(req.DNSNames) > 0 {
		return fmt.Errorf("dns subject alternative names are not allowed")
	}
	if len(req.EmailAddresses) > 0 {
		return fmt.Errorf("email subject alternative names are not allowed")
	}
	if len(req.IPAddresses) > 0 {
		return fmt.Errorf("ip subject alternative names are not allowed")
	}
	if len(req.URIs) > 0 {
		return fmt.Errorf("uri subject alternative names are not allowed")
	}

	if req.Subject.CommonName != computev1alpha1.PoolLifecycleControllerCommonName {
		return fmt.Errorf("subject common name is not %s", computev1alpha1.PoolLifecycleControllerCommonName)
	}

	if !PoolLifecycleControllerRequiredUsages.Equal(usages) {
		return fmt.Errorf("usages did not match %v", sets.List(PoolLifecycleControllerRequiredUsages))
	}

	return nil
}

var (
	PoolLifecycleControllerRecognizer = generic.NewCertificateSigningRequestRecognizer(
		IsPoolLifecycleControllerClientCert,
		authorizationv1.ResourceAttributes{
			Group:       certificatesv1.GroupName,
			Resource:    "certificatesigningrequests",
			Verb:        "create",
			Subresource: "poollifecycleclient",
		},
		"Auto approving pool-lifecycle-controller client certificate after SubjectAccessReview.",
	)
)

func init() {
	Recognizers = append(Recognizers, PoolLifecycleControllerRecognizer)
}
