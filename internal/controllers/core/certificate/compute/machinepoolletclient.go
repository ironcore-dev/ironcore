// SPDX-FileCopyrightText: SAP SE or an SAP affiliate company and IronCore contributors
// SPDX-License-Identifier: Apache-2.0

package compute

import (
	"crypto/x509"
	"fmt"

	"github.com/ironcore-dev/ironcore/internal/controllers/core/certificate/generic"
	"golang.org/x/exp/slices"
	authorizationv1 "k8s.io/api/authorization/v1"
	certificatesv1 "k8s.io/api/certificates/v1"
	"k8s.io/apimachinery/pkg/util/sets"
)

const (
	apiServerMachinePoolletClientCommonName = "ironcore-apiserver-machinepoollet-client"
	apiServerMachinePoolletClientOrg        = "system:masters"
)

var (
	// apiServerClientRequiredUsages is the usage set for RSA keys.
	apiServerClientRequiredUsages = sets.New[certificatesv1.KeyUsage](
		certificatesv1.UsageDigitalSignature,
		certificatesv1.UsageKeyEncipherment,
		certificatesv1.UsageClientAuth,
	)

	// apiServerClientRequiredUsagesNoEncipherment is the usage set for EC keys,
	// which cannot perform key encipherment. The rotator toolkit generates
	// ECDSA keys, so its CSRs always carry this usage set.
	apiServerClientRequiredUsagesNoEncipherment = sets.New[certificatesv1.KeyUsage](
		certificatesv1.UsageDigitalSignature,
		certificatesv1.UsageClientAuth,
	)
)

func IsAPIServerMachinePoolletClientCert(csr *certificatesv1.CertificateSigningRequest, x509cr *x509.CertificateRequest) bool {
	if csr.Spec.SignerName != certificatesv1.KubeAPIServerClientSignerName {
		return false
	}

	return ValidateAPIServerMachinePoolletClientCSR(x509cr, sets.New(csr.Spec.Usages...)) == nil
}

func ValidateAPIServerMachinePoolletClientCSR(req *x509.CertificateRequest, usages sets.Set[certificatesv1.KeyUsage]) error {
	if !slices.Equal([]string{apiServerMachinePoolletClientOrg}, req.Subject.Organization) {
		return fmt.Errorf("organization is not %s", apiServerMachinePoolletClientOrg)
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

	if req.Subject.CommonName != apiServerMachinePoolletClientCommonName {
		return fmt.Errorf("subject common name is not %s", apiServerMachinePoolletClientCommonName)
	}

	if !apiServerClientRequiredUsages.Equal(usages) && !apiServerClientRequiredUsagesNoEncipherment.Equal(usages) {
		return fmt.Errorf("usages did not match %v", sets.List(apiServerClientRequiredUsages))
	}

	return nil
}

var (
	MachinePoolletClientRecognizer = generic.NewCertificateSigningRequestRecognizer(
		IsAPIServerMachinePoolletClientCert,
		authorizationv1.ResourceAttributes{
			Group:       certificatesv1.GroupName,
			Resource:    "certificatesigningrequests",
			Verb:        "create",
			Subresource: "machinepoolletclient",
		},
		"Auto approving apiserver machinepoollet client certificate after SubjectAccessReview.",
	)
)

func init() {
	Recognizers = append(Recognizers, MachinePoolletClientRecognizer)
}
