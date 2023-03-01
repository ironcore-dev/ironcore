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

package networking

import (
	"crypto/x509"
	"fmt"
	"strings"

	networkingv1alpha1 "github.com/onmetal/onmetal-api/api/networking/v1alpha1"
	"github.com/onmetal/onmetal-api/internal/controllers/core/certificate/generic"
	"golang.org/x/exp/slices"
	authv1 "k8s.io/api/authorization/v1"
	certificatesv1 "k8s.io/api/certificates/v1"
	"k8s.io/apimachinery/pkg/util/sets"
)

var (
	NetworkPluginRequiredUsages = sets.New[certificatesv1.KeyUsage](
		certificatesv1.UsageDigitalSignature,
		certificatesv1.UsageKeyEncipherment,
		certificatesv1.UsageClientAuth,
	)
)

func IsNetworkPluginClientCert(csr *certificatesv1.CertificateSigningRequest, x509cr *x509.CertificateRequest) bool {
	if csr.Spec.SignerName != certificatesv1.KubeAPIServerClientSignerName {
		return false
	}

	return ValidateNetworkPluginClientCSR(x509cr, sets.New(csr.Spec.Usages...)) == nil
}

func ValidateNetworkPluginClientCSR(req *x509.CertificateRequest, usages sets.Set[certificatesv1.KeyUsage]) error {
	if !slices.Equal([]string{networkingv1alpha1.NetworkPluginsGroup}, req.Subject.Organization) {
		return fmt.Errorf("organization is not %s", networkingv1alpha1.NetworkPluginsGroup)
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

	if !strings.HasPrefix(req.Subject.CommonName, networkingv1alpha1.NetworkPluginUserNamePrefix) {
		return fmt.Errorf("subject common name does not begin with %s", networkingv1alpha1.NetworkPluginUserNamePrefix)
	}

	if !NetworkPluginRequiredUsages.Equal(usages) {
		return fmt.Errorf("usages did not match %v", sets.List(NetworkPluginRequiredUsages))
	}

	return nil
}

var (
	NetworkPluginRecognizer = generic.NewCertificateSigningRequestRecognizer(
		IsNetworkPluginClientCert,
		authv1.ResourceAttributes{
			Group:       certificatesv1.GroupName,
			Resource:    "certificatesigningrequests",
			Verb:        "create",
			Subresource: "networkpluginclient",
		},
		"Auto approving network plugin client certificate after SubjectAccessReview.",
	)
)

func init() {
	Recognizers = append(Recognizers, NetworkPluginRecognizer)
}
