// SPDX-FileCopyrightText: SAP SE or an SAP affiliate company and IronCore contributors
// SPDX-License-Identifier: Apache-2.0

package compute_test

import (
	"crypto/x509"
	"crypto/x509/pkix"
	"net"
	"net/url"
	"testing"

	certificatesv1 "k8s.io/api/certificates/v1"
	"k8s.io/apimachinery/pkg/util/sets"

	"github.com/ironcore-dev/ironcore/internal/controllers/core/certificate/compute"
)

var (
	clientUsages = sets.New(
		certificatesv1.UsageDigitalSignature,
		certificatesv1.UsageKeyEncipherment,
		certificatesv1.UsageClientAuth,
	)
	ecdsaUsages = sets.New(
		certificatesv1.UsageDigitalSignature,
		certificatesv1.UsageClientAuth,
	)
)

func validAPIServerMachinePoolletClientCSR() *x509.CertificateRequest {
	return &x509.CertificateRequest{
		Subject: pkix.Name{
			CommonName:   "ironcore-apiserver-machinepoollet-client",
			Organization: []string{"system:masters"},
		},
	}
}

func TestValidateAPIServerMachinePoolletClientCSR(t *testing.T) {
	tests := []struct {
		name    string
		req     *x509.CertificateRequest
		usages  sets.Set[certificatesv1.KeyUsage]
		wantErr bool
	}{
		{
			name:   "accepts a well-formed RSA apiserver client CSR",
			req:    validAPIServerMachinePoolletClientCSR(),
			usages: clientUsages,
		},
		{
			name:   "accepts a well-formed ECDSA apiserver client CSR",
			req:    validAPIServerMachinePoolletClientCSR(),
			usages: ecdsaUsages,
		},
		{
			name: "rejects a wrong organization",
			req: func() *x509.CertificateRequest {
				r := validAPIServerMachinePoolletClientCSR()
				r.Subject.Organization = []string{"system:nodes"}
				return r
			}(),
			usages:  clientUsages,
			wantErr: true,
		},
		{
			name: "rejects a wrong common name",
			req: func() *x509.CertificateRequest {
				r := validAPIServerMachinePoolletClientCSR()
				r.Subject.CommonName = "someone-else"
				return r
			}(),
			usages:  clientUsages,
			wantErr: true,
		},
		{
			name: "rejects DNS SANs",
			req: func() *x509.CertificateRequest {
				r := validAPIServerMachinePoolletClientCSR()
				r.DNSNames = []string{"nope"}
				return r
			}(),
			usages:  clientUsages,
			wantErr: true,
		},
		{
			name: "rejects email SANs",
			req: func() *x509.CertificateRequest {
				r := validAPIServerMachinePoolletClientCSR()
				r.EmailAddresses = []string{"nope@example.com"}
				return r
			}(),
			usages:  clientUsages,
			wantErr: true,
		},
		{
			name: "rejects IP SANs",
			req: func() *x509.CertificateRequest {
				r := validAPIServerMachinePoolletClientCSR()
				r.IPAddresses = []net.IP{net.ParseIP("10.0.0.1")}
				return r
			}(),
			usages:  clientUsages,
			wantErr: true,
		},
		{
			name: "rejects URI SANs",
			req: func() *x509.CertificateRequest {
				r := validAPIServerMachinePoolletClientCSR()
				r.URIs = []*url.URL{{Scheme: "https", Host: "nope"}}
				return r
			}(),
			usages:  clientUsages,
			wantErr: true,
		},
		{
			name: "rejects server-auth usages",
			req:  validAPIServerMachinePoolletClientCSR(),
			usages: sets.New(
				certificatesv1.UsageDigitalSignature,
				certificatesv1.UsageServerAuth,
			),
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := compute.ValidateAPIServerMachinePoolletClientCSR(tt.req, tt.usages)
			if tt.wantErr && err == nil {
				t.Error("expected an error, got nil")
			}
			if !tt.wantErr && err != nil {
				t.Errorf("expected no error, got %v", err)
			}
		})
	}
}

func TestIsAPIServerMachinePoolletClientCert(t *testing.T) {
	csrFor := func(signerName string, usages sets.Set[certificatesv1.KeyUsage]) *certificatesv1.CertificateSigningRequest {
		return &certificatesv1.CertificateSigningRequest{
			Spec: certificatesv1.CertificateSigningRequestSpec{
				SignerName: signerName,
				Usages:     sets.List(usages),
			},
		}
	}

	if !compute.IsAPIServerMachinePoolletClientCert(
		csrFor(certificatesv1.KubeAPIServerClientSignerName, ecdsaUsages),
		validAPIServerMachinePoolletClientCSR(),
	) {
		t.Error("expected a well-formed kube-apiserver-client CSR to be recognized")
	}

	if compute.IsAPIServerMachinePoolletClientCert(
		csrFor(certificatesv1.KubeletServingSignerName, ecdsaUsages),
		validAPIServerMachinePoolletClientCSR(),
	) {
		t.Error("expected a non kube-apiserver-client signer CSR not to be recognized")
	}
}
