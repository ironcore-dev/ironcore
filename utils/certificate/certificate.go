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

package certificate

import (
	"bytes"
	"context"
	"crypto/ecdsa"
	"crypto/elliptic"
	"crypto/rand"
	"crypto/rsa"
	"crypto/tls"
	"crypto/x509"
	"encoding/pem"
	"fmt"
	"time"

	certificatesv1 "k8s.io/api/certificates/v1"
	apierrors "k8s.io/apimachinery/pkg/api/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/fields"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apimachinery/pkg/types"
	"k8s.io/apimachinery/pkg/watch"
	"k8s.io/client-go/tools/cache"
	watchtools "k8s.io/client-go/tools/watch"
	certutil "k8s.io/client-go/util/cert"
	"k8s.io/client-go/util/certificate/csr"
	"k8s.io/client-go/util/keyutil"
	"sigs.k8s.io/controller-runtime/pkg/client"
)

var (
	kubeAPIServerClientUsagesWithEncipherment = []certificatesv1.KeyUsage{
		// https://tools.ietf.org/html/rfc5280#section-4.2.1.3
		//
		// Digital signature allows the certificate to be used to verify
		// digital signatures used during TLS negotiation.
		certificatesv1.UsageDigitalSignature,
		// KeyEncipherment allows the cert/key pair to be used to encrypt
		// keys, including the symmetric keys negotiated during TLS setup
		// and used for data transfer.
		certificatesv1.UsageKeyEncipherment,
		// ClientAuth allows the cert to be used by a TLS client to
		// authenticate itself to the TLS server.
		certificatesv1.UsageClientAuth,
	}
	kubeAPIServerClientUsagesNoEncipherment = []certificatesv1.KeyUsage{
		// https://tools.ietf.org/html/rfc5280#section-4.2.1.3
		//
		// Digital signature allows the certificate to be used to verify
		// digital signatures used during TLS negotiation.
		certificatesv1.UsageDigitalSignature,
		// ClientAuth allows the cert to be used by a TLS client to
		// authenticate itself to the TLS server.
		certificatesv1.UsageClientAuth,
	}
)

func DefaultKubeAPIServerClientGetUsages(privateKey any) []certificatesv1.KeyUsage {
	switch privateKey.(type) {
	case *rsa.PrivateKey:
		return kubeAPIServerClientUsagesWithEncipherment
	default:
		return kubeAPIServerClientUsagesNoEncipherment
	}
}

func GenerateCertificateSigningRequestData(template *x509.CertificateRequest) (csrPEM, keyPEM []byte, key interface{}, err error) {
	privateKey, err := ecdsa.GenerateKey(elliptic.P256(), rand.Reader)
	if err != nil {
		return nil, nil, nil, fmt.Errorf("error generating client certificate private key: %w", err)
	}

	der, err := x509.MarshalECPrivateKey(privateKey)
	if err != nil {
		return nil, nil, nil, fmt.Errorf("error marshalling client certificate private key to DER: %w", err)
	}

	keyPEM = pem.EncodeToMemory(&pem.Block{Type: keyutil.ECPrivateKeyBlockType, Bytes: der})

	csrPEM, err = certutil.MakeCSRFromTemplate(privateKey, template)
	if err != nil {
		return nil, nil, nil, fmt.Errorf("error creating a csr using the certificate private key and template: %w", err)
	}

	return csrPEM, keyPEM, privateKey, nil
}

func MakeCertificatesCertificateSigningRequest(
	signerName string,
	csrPem []byte,
	usages []certificatesv1.KeyUsage,
	requestedDuration *time.Duration,
) *certificatesv1.CertificateSigningRequest {
	var expirationSeconds *int32
	if requestedDuration != nil {
		expirationSeconds = csr.DurationToExpirationSeconds(*requestedDuration)
	}

	return &certificatesv1.CertificateSigningRequest{
		ObjectMeta: metav1.ObjectMeta{
			GenerateName: "csr-",
		},
		Spec: certificatesv1.CertificateSigningRequestSpec{
			Request:           csrPem,
			Usages:            usages,
			SignerName:        signerName,
			ExpirationSeconds: expirationSeconds,
		},
	}
}

func createOrUseCSR(
	ctx context.Context,
	c client.Client,
	csrObj *certificatesv1.CertificateSigningRequest,
	privateKey any,
) error {
	newCSRObj := csrObj.DeepCopy()

	if err := c.Create(ctx, csrObj); err != nil {
		if !apierrors.IsAlreadyExists(err) {
			return fmt.Errorf("error creating certificate signing request: %w", err)
		}

		csrObjKey := client.ObjectKeyFromObject(csrObj)
		if err := c.Get(ctx, csrObjKey, csrObj); err != nil {
			return fmt.Errorf("error getting existing certificate signing request: %w", err)
		}

		if err := ensureCompatible(csrObj, newCSRObj, privateKey); err != nil {
			return fmt.Errorf("existing certificate signing request is not compatible: %w", err)
		}
	}
	return nil
}

func WaitForCertificate(ctx context.Context, c client.WithWatch, name string, uid types.UID) ([]byte, error) {
	fieldSelector := fields.OneTermEqualSelector("metadata.name", name).String()

	lw := &cache.ListWatch{
		ListFunc: func(options metav1.ListOptions) (runtime.Object, error) {
			options.FieldSelector = fieldSelector
			list := &certificatesv1.CertificateSigningRequestList{}
			return list, c.List(ctx, list, &client.ListOptions{Raw: &options})
		},
		WatchFunc: func(options metav1.ListOptions) (watch.Interface, error) {
			options.FieldSelector = fieldSelector
			return c.Watch(ctx, &certificatesv1.CertificateSigningRequestList{}, &client.ListOptions{Raw: &options})
		},
	}

	var certData []byte
	if _, err := watchtools.UntilWithSync(ctx, lw, &certificatesv1.CertificateSigningRequest{}, nil,
		func(event watch.Event) (bool, error) {
			switch event.Type {
			case watch.Modified, watch.Added:
			case watch.Deleted:
				return false, fmt.Errorf("certificate signing request %q was deleted", name)
			default:
				return false, nil
			}

			csrObj, ok := event.Object.(*certificatesv1.CertificateSigningRequest)
			if !ok {
				return false, fmt.Errorf("unexpected type received: %T", event.Object)
			}

			if csrObj.UID != uid {
				return false, fmt.Errorf("certificate signing request %q changed UIDs", name)
			}

			var approved bool
			for _, c := range csrObj.Status.Conditions {
				switch c.Type {
				case certificatesv1.CertificateDenied:
					return false, fmt.Errorf("certificate signing request is denied, reason: %v, message: %v", c.Reason, c.Message)
				case certificatesv1.CertificateFailed:
					return false, fmt.Errorf("certificate signing request failed, reason: %v, message: %v", c.Reason, c.Message)
				case certificatesv1.CertificateApproved:
					approved = true
				}
			}
			if approved {
				if csrCertData := csrObj.Status.Certificate; len(csrCertData) > 0 {
					certData = csrCertData
					return true, nil
				}
			}
			return false, nil
		},
	); err != nil {
		return nil, fmt.Errorf("error waiting for certificate: %w", err)
	}
	return certData, nil
}

func GenerateAndCreateCertificateSigningRequest(
	ctx context.Context,
	c client.Client,
	signerName string,
	template *x509.CertificateRequest,
	getUsages func(privateKey any) []certificatesv1.KeyUsage,
	requestedDuration *time.Duration,
) (csrObj *certificatesv1.CertificateSigningRequest, keyPEM []byte, privateKey any, err error) {
	if signerName == "" {
		return nil, nil, nil, fmt.Errorf("must specify signerName")
	}
	if template == nil {
		return nil, nil, nil, fmt.Errorf("must specify template")
	}
	if getUsages == nil {
		return nil, nil, nil, fmt.Errorf("must specify getUsages")
	}

	csrPEM, keyPEM, privateKey, err := GenerateCertificateSigningRequestData(template)
	if err != nil {
		return nil, nil, nil, fmt.Errorf("error generating csr data: %w", err)
	}

	usages := getUsages(privateKey)
	csrObj = MakeCertificatesCertificateSigningRequest(signerName, csrPEM, usages, requestedDuration)

	if err := createOrUseCSR(ctx, c, csrObj, privateKey); err != nil {
		return nil, nil, nil, fmt.Errorf("error creating / using certificate signing request: %w", err)
	}

	return csrObj, keyPEM, privateKey, nil
}

func RequestCertificate(
	ctx context.Context,
	c client.WithWatch,
	signerName string,
	template *x509.CertificateRequest,
	getUsages func(privateKey any) []certificatesv1.KeyUsage,
	requestedDuration *time.Duration,
) (*tls.Certificate, error) {
	if signerName == "" {
		return nil, fmt.Errorf("must specify signerName")
	}
	if template == nil {
		return nil, fmt.Errorf("must specify template")
	}
	if getUsages == nil {
		return nil, fmt.Errorf("must specify getUsages")
	}

	csrObj, keyPEM, _, err := GenerateAndCreateCertificateSigningRequest(ctx, c, signerName, template, getUsages, requestedDuration)
	if err != nil {
		return nil, err
	}

	certPEM, err := WaitForCertificate(ctx, c, csrObj.Name, csrObj.UID)
	if err != nil {
		return nil, fmt.Errorf("error waiting for certificate: %w", err)
	}

	newCert, err := tls.X509KeyPair(certPEM, keyPEM)
	if err != nil {
		return nil, fmt.Errorf("error creating key pair: %w", err)
	}

	if err := setCertificateLeaf(&newCert); err != nil {
		return nil, err
	}
	return &newCert, nil
}

func Marshal(cert *tls.Certificate) (certPEM, keyPEM []byte, err error) {
	if len(cert.Certificate) == 0 {
		return nil, nil, fmt.Errorf("no certificates in certificate chain")
	}

	var buf bytes.Buffer
	for _, certDER := range cert.Certificate {
		if err := pem.Encode(&buf, &pem.Block{
			Type:  "CERTIFICATE",
			Bytes: certDER,
		}); err != nil {
			return nil, nil, fmt.Errorf("error encodign certificate to pem: %w", err)
		}
	}

	certPEM = buf.Bytes()

	keyDER, err := x509.MarshalPKCS8PrivateKey(cert.PrivateKey)
	if err != nil {
		return nil, nil, fmt.Errorf("error marshaling private key: %w", err)
	}

	keyPEM = pem.EncodeToMemory(&pem.Block{
		Type:  "PRIVATE KEY",
		Bytes: keyDER,
	})

	return certPEM, keyPEM, nil
}
