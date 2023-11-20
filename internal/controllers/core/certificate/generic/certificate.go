// SPDX-FileCopyrightText: 2023 SAP SE or an SAP affiliate company and IronCore contributors
// SPDX-License-Identifier: Apache-2.0

package generic

import (
	"crypto/x509"
	"encoding/pem"
	"fmt"

	authv1 "k8s.io/api/authorization/v1"
	certificatesv1 "k8s.io/api/certificates/v1"
)

type CertificateSigningRequestRecognizer interface {
	Recognize(csr *certificatesv1.CertificateSigningRequest, x509CR *x509.CertificateRequest) bool
	Permission() authv1.ResourceAttributes
	SuccessMessage() string
}

type certificateSigningRequestRecognizer struct {
	recognize      func(csr *certificatesv1.CertificateSigningRequest, x509CR *x509.CertificateRequest) bool
	permission     authv1.ResourceAttributes
	successMessage string
}

func NewCertificateSigningRequestRecognizer(
	recognize func(csr *certificatesv1.CertificateSigningRequest, x509CR *x509.CertificateRequest) bool,
	permission authv1.ResourceAttributes,
	successMessage string,
) CertificateSigningRequestRecognizer {
	return &certificateSigningRequestRecognizer{
		recognize:      recognize,
		permission:     permission,
		successMessage: successMessage,
	}
}

func (r *certificateSigningRequestRecognizer) Recognize(csr *certificatesv1.CertificateSigningRequest, x509CR *x509.CertificateRequest) bool {
	return r.recognize(csr, x509CR)
}

func (r *certificateSigningRequestRecognizer) Permission() authv1.ResourceAttributes {
	return r.permission
}

func (r *certificateSigningRequestRecognizer) SuccessMessage() string {
	return r.successMessage
}

const (
	CertificateRequestPEMBlockType = "CERTIFICATE REQUEST"
)

func ParseCertificateRequest(pemBytes []byte) (*x509.CertificateRequest, error) {
	block, _ := pem.Decode(pemBytes)
	if block == nil || block.Type != CertificateRequestPEMBlockType {
		return nil, fmt.Errorf("pem block type must be %s", CertificateRequestPEMBlockType)
	}

	csr, err := x509.ParseCertificateRequest(block.Bytes)
	if err != nil {
		return nil, err
	}

	return csr, nil
}

func GetCertificateSigningRequestApprovalCondition(status *certificatesv1.CertificateSigningRequestStatus) (approved, denied bool) {
	for _, c := range status.Conditions {
		if c.Type == certificatesv1.CertificateApproved {
			approved = true
		}
		if c.Type == certificatesv1.CertificateDenied {
			denied = true
		}
	}
	return approved, denied
}

func IsCertificateSigningRequestApproved(csr *certificatesv1.CertificateSigningRequest) bool {
	approved, denied := GetCertificateSigningRequestApprovalCondition(&csr.Status)
	return approved && !denied
}
