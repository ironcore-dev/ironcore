// SPDX-FileCopyrightText: 2023 SAP SE or an SAP affiliate company and IronCore contributors
// SPDX-License-Identifier: Apache-2.0

package clientcmd

import (
	"k8s.io/client-go/rest"
	clientcmdapi "k8s.io/client-go/tools/clientcmd/api"
)

// RESTConfigToConfig transforms a rest.Config to a clientcmdapi.Config.
// Some properties like e.g. a predefined namespace / proxy cannot be translated and might be lost.
// Same applies for custom dials / transports that cannot be serialized.
func RESTConfigToConfig(cfg *rest.Config) (*clientcmdapi.Config, error) {
	return &clientcmdapi.Config{
		Clusters: map[string]*clientcmdapi.Cluster{"default-cluster": {
			Server:                   cfg.Host,
			TLSServerName:            cfg.ServerName,
			InsecureSkipTLSVerify:    cfg.Insecure,
			CertificateAuthority:     cfg.CAFile,
			CertificateAuthorityData: cfg.CAData,
			DisableCompression:       cfg.DisableCompression,
		}},
		AuthInfos: map[string]*clientcmdapi.AuthInfo{"default-auth": {
			ClientCertificate:     cfg.CertFile,
			ClientCertificateData: cfg.CertData,
			ClientKey:             cfg.KeyFile,
			ClientKeyData:         cfg.KeyData,
			Token:                 cfg.BearerToken,
			TokenFile:             cfg.BearerTokenFile,
			Impersonate:           cfg.Impersonate.UserName,
			ImpersonateUID:        cfg.Impersonate.UID,
			ImpersonateGroups:     cfg.Impersonate.Groups,
			ImpersonateUserExtra:  cfg.Impersonate.Extra,
			Username:              cfg.Username,
			Password:              cfg.Password,
		}},
		Contexts: map[string]*clientcmdapi.Context{"default-context": {
			Cluster:   "default-cluster",
			AuthInfo:  "default-auth",
			Namespace: "default",
		}},
		CurrentContext: "default-context",
	}, nil
}
