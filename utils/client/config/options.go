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

package config

import (
	"github.com/spf13/pflag"
)

// GetConfigOptions are options to supply for a GetConfig call.
type GetConfigOptions struct {
	// Context is the kubeconfig context to load.
	Context string

	// Kubeconfig specifies where to get the kubeconfig from.
	Kubeconfig string

	// KubeconfigSecretName instructs to get the kubeconfig from a secret with the given name.
	KubeconfigSecretName string

	// KubeconfigSecretNamespace instructs to get the kubeconfig from a secret within the given namespace.
	// If unset, LoadDefaultNamespace will be used to determine the namespace.
	KubeconfigSecretNamespace string

	// KubeconfigSecretField specifies the field of the secret to get the kubeconfig from.
	// If unset, DefaultSecretKubeconfigField will be used.
	KubeconfigSecretField string

	// BootstrapKubeconfig specifies the path to the bootstrap kubeconfig to load.
	// The bootstrap kubeconfig will be used to request an up-to-date certificate for the kube-apiserver.
	BootstrapKubeconfig string

	// RotateCertificates specifies whether kubeconfig should be automatically rotated.
	RotateCertificates bool

	// EgressSelectorConfig is the path to an egress selector config to load.
	EgressSelectorConfig string
}

func (o *GetConfigOptions) BindFlags(fs *pflag.FlagSet) {
	fs.StringVar(&o.Kubeconfig, KubeconfigFlagName, "", "Paths to a kubeconfig. Only required if out-of-cluster.")
	fs.StringVar(&o.KubeconfigSecretName, KubeconfigSecretNameFlagName, "", "Name of a kubeconfig secret to use.")
	fs.StringVar(&o.KubeconfigSecretNamespace, KubeconfigSecretNamespaceFlagName, "", "Namespace of the kubeconfig secret to use. If empty, use in-cluster namespace.")
	fs.StringVar(&o.BootstrapKubeconfig, BootstrapKubeconfigFlagName, "", "Path to a bootstrap kubeconfig.")
	fs.BoolVar(&o.RotateCertificates, RotateCertificatesFlagName, false, "Whether to use automatic certificate / config rotation.")
	fs.StringVar(&o.EgressSelectorConfig, EgressSelectorConfigFlagName, "", "Path pointing to an egress selector config to use.")
}

// ApplyToGetConfig implements GetConfigOption.
func (o *GetConfigOptions) ApplyToGetConfig(o2 *GetConfigOptions) {
	if o.Context != "" {
		o2.Context = o.Context
	}
	if o.Kubeconfig != "" {
		o2.Kubeconfig = o.Kubeconfig
	}
	if o.KubeconfigSecretName != "" {
		o2.KubeconfigSecretName = o.KubeconfigSecretName
	}
	if o.KubeconfigSecretNamespace != "" {
		o2.KubeconfigSecretNamespace = o.KubeconfigSecretNamespace
	}
	if o.BootstrapKubeconfig != "" {
		o2.BootstrapKubeconfig = o.BootstrapKubeconfig
	}
	if o.RotateCertificates {
		o2.RotateCertificates = o.RotateCertificates
	}
	if o.EgressSelectorConfig != "" {
		o2.EgressSelectorConfig = o.EgressSelectorConfig
	}
}

// ApplyOptions applies all GetConfigOption tro this GetConfigOptions.
func (o *GetConfigOptions) ApplyOptions(opts []GetConfigOption) {
	for _, opt := range opts {
		opt.ApplyToGetConfig(o)
	}
}

// Context allows specifying the context to load.
type Context string

// ApplyToGetConfig implements GetConfigOption.
func (c Context) ApplyToGetConfig(o *GetConfigOptions) {
	o.Context = string(c)
}

// EgressSelectorConfig allows specifying the path to an egress selector config to use.
type EgressSelectorConfig string

// ApplyToGetConfig implements GetConfigOption.
func (c EgressSelectorConfig) ApplyToGetConfig(o *GetConfigOptions) {
	o.EgressSelectorConfig = string(c)
}

// GetConfigOption are options to a GetConfig call.
type GetConfigOption interface {
	// ApplyToGetConfig modifies the underlying GetConfigOptions.
	ApplyToGetConfig(o *GetConfigOptions)
}

// WithRotate sets GetConfigOptions.RotateCertificates to the specified boolean.
type WithRotate bool

// ApplyToGetConfig implements GetConfigOption.
func (w WithRotate) ApplyToGetConfig(o *GetConfigOptions) {
	o.RotateCertificates = bool(w)
}

// RotateCertificates enables certificate rotation.
var RotateCertificates = WithRotate(true)
