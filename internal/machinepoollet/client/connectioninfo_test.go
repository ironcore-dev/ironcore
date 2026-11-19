// SPDX-FileCopyrightText: 2024 SAP SE or an SAP affiliate company and IronCore contributors
// SPDX-License-Identifier: Apache-2.0

package client

import (
	"crypto/tls"
	"testing"
)

func TestClientTLSConfigUsesGetClientCertificate(t *testing.T) {
	called := false
	cfg := &MachinePoolletClientConfig{
		GetClientCertificate: func(*tls.CertificateRequestInfo) (*tls.Certificate, error) {
			called = true
			return &tls.Certificate{}, nil
		},
	}

	tlsCfg, err := cfg.clientTLSConfig()
	if err != nil {
		t.Fatalf("clientTLSConfig() error: %v", err)
	}
	if tlsCfg.GetClientCertificate == nil {
		t.Fatal("expected tls.Config.GetClientCertificate to be set")
	}
	if _, err := tlsCfg.GetClientCertificate(&tls.CertificateRequestInfo{}); err != nil {
		t.Fatalf("GetClientCertificate error: %v", err)
	}
	if !called {
		t.Fatal("expected the configured GetClientCertificate to be invoked")
	}
}

func TestClientTLSConfigNilCallbackIsHarmless(t *testing.T) {
	cfg := &MachinePoolletClientConfig{}
	tlsCfg, err := cfg.clientTLSConfig()
	if err != nil {
		t.Fatalf("clientTLSConfig() error: %v", err)
	}
	if tlsCfg != nil && tlsCfg.GetClientCertificate != nil {
		t.Fatal("expected no GetClientCertificate when callback is unset")
	}
}
