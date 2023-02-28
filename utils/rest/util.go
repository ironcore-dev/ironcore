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

package rest

import (
	"crypto/tls"
	"fmt"
	"net/http"
	"os"
	"time"

	"github.com/onmetal/onmetal-api/utils/certificate"
	utilnet "k8s.io/apimachinery/pkg/util/net"
	"k8s.io/client-go/rest"
	certutil "k8s.io/client-go/util/cert"
	"k8s.io/client-go/util/connrotation"
)

func ConfigWithCertificate(cfg *rest.Config, cert *tls.Certificate) (*rest.Config, error) {
	certData, keyData, err := certificate.Marshal(cert)
	if err != nil {
		return nil, fmt.Errorf("error marshalling tls certificate: %w", err)
	}

	certCfg := rest.AnonymousClientConfig(cfg)
	certCfg.CertData = certData
	certCfg.KeyData = keyData
	return certCfg, nil
}

func CertificateFromConfig(cfg *rest.Config) (*tls.Certificate, error) {
	if cfg.CertData == nil && cfg.CertFile != "" {
		certData, err := os.ReadFile(cfg.CertFile)
		if err != nil {
			return nil, fmt.Errorf("error reading certificate file %q: %w", cfg.CertFile, err)
		}

		cfg.CertData = certData
	}

	if cfg.KeyData == nil && cfg.KeyFile != "" {
		keyData, err := os.ReadFile(cfg.KeyFile)
		if err != nil {
			return nil, fmt.Errorf("error reading key file %q: %w", cfg.KeyFile, err)
		}

		cfg.KeyData = keyData
	}

	if cfg.CertData == nil && cfg.KeyData == nil {
		return nil, nil
	}

	cert, err := tls.X509KeyPair(cfg.CertData, cfg.KeyData)
	if err != nil {
		return nil, fmt.Errorf("error parsing key pair: %w", err)
	}

	return &cert, nil
}

func DynamicCertificateConfig(
	cfg *rest.Config,
	getCertificate func() *tls.Certificate,
	dialFunc utilnet.DialFunc,
) (*rest.Config, func(), error) {
	cfg = rest.AnonymousClientConfig(cfg)
	tlsConfig, err := rest.TLSConfigFor(cfg)
	if err != nil {
		return nil, nil, fmt.Errorf("error getting tls config: %w", err)
	}
	if tlsConfig == nil {
		tlsConfig = &tls.Config{}
	}

	tlsConfig.Certificates = nil
	tlsConfig.GetClientCertificate = func(info *tls.CertificateRequestInfo) (*tls.Certificate, error) {
		cert := getCertificate()
		if cert == nil {
			return &tls.Certificate{Certificate: nil}, nil
		}
		return cert, nil
	}

	d := connrotation.NewDialer(connrotation.DialFunc(dialFunc))
	cfg.Transport = utilnet.SetTransportDefaults(&http.Transport{
		Proxy:               http.ProxyFromEnvironment,
		TLSHandshakeTimeout: 10 * time.Second,
		TLSClientConfig:     tlsConfig,
		MaxIdleConnsPerHost: 25,
		DialContext:         d.DialContext,
	})

	// Zero out all existing TLS options since our new transport enforces them.
	cfg.CertData = nil
	cfg.KeyData = nil
	cfg.CertFile = ""
	cfg.KeyFile = ""
	cfg.CAData = nil
	cfg.CAFile = ""
	cfg.Insecure = false
	cfg.NextProtos = nil

	return cfg, d.CloseAll, nil
}

func IsConfigValid(cfg *rest.Config) bool {
	if cfg == nil {
		return false
	}

	transportCfg, err := cfg.TransportConfig()
	if err != nil {
		return false
	}

	certs, err := certutil.ParseCertsPEM(transportCfg.TLS.CertData)
	if err != nil {
		return false
	}
	if len(certs) == 0 {
		return false
	}

	now := time.Now()
	for _, cert := range certs {
		if now.After(cert.NotAfter) {
			return false
		}
	}
	return true
}
