// Copyright 2023 IronCore authors
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
