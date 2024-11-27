// SPDX-FileCopyrightText: 2023 SAP SE or an SAP affiliate company and IronCore contributors
// SPDX-License-Identifier: Apache-2.0

package config

import (
	"fmt"

	"github.com/spf13/pflag"

	"github.com/ironcore-dev/ironcore/utils/generic"
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

	// QPS specifies the queries per second allowed for the client.
	QPS float32

	// Burst specified the burst rate allowed for the client.
	Burst int
}

// BindFlagOptions are options for GetConfigOptions.BindFlags.
type BindFlagOptions struct {
	// NameFunc can modify the flag names if non-nil.
	NameFunc func(string) string
}

// WithNamePrefix adds a flag name prefix to all flags.
func WithNamePrefix(prefix string) func(*BindFlagOptions) {
	return func(options *BindFlagOptions) {
		options.NameFunc = func(name string) string {
			return fmt.Sprintf("%s%s", prefix, name)
		}
	}
}

// WithNameSuffix adds a flag name suffix to all flags.
func WithNameSuffix(suffix string) func(*BindFlagOptions) {
	return func(options *BindFlagOptions) {
		options.NameFunc = func(name string) string {
			return fmt.Sprintf("%s%s", name, suffix)
		}
	}
}

// BindFlags binds values of GetConfigOptions as flags to the given flag set.
func (o *GetConfigOptions) BindFlags(fs *pflag.FlagSet, opts ...func(*BindFlagOptions)) {
	bo := &BindFlagOptions{}
	for _, opt := range opts {
		opt(bo)
	}
	if bo.NameFunc == nil {
		bo.NameFunc = generic.Identity[string]
	}

	fs.StringVar(&o.Kubeconfig, bo.NameFunc(KubeconfigFlagName), "", "Paths to a kubeconfig. Only required if out-of-cluster.")
	fs.StringVar(&o.KubeconfigSecretName, bo.NameFunc(KubeconfigSecretNameFlagName), "", "Name of a kubeconfig secret to use.")
	fs.StringVar(&o.KubeconfigSecretNamespace, bo.NameFunc(KubeconfigSecretNamespaceFlagName), "", "Namespace of the kubeconfig secret to use. If empty, use in-cluster namespace.")
	fs.StringVar(&o.BootstrapKubeconfig, bo.NameFunc(BootstrapKubeconfigFlagName), "", "Path to a bootstrap kubeconfig.")
	fs.BoolVar(&o.RotateCertificates, bo.NameFunc(RotateCertificatesFlagName), false, "Whether to use automatic certificate / config rotation.")
	fs.StringVar(&o.EgressSelectorConfig, bo.NameFunc(EgressSelectorConfigFlagName), "", "Path pointing to an egress selector config to use.")
	fs.Float32Var(&o.QPS, "qps", QPS, "Kubernetes client qps.")
	fs.IntVar(&o.Burst, "burst", Burst, "Kubernetes client burst.")
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
	if o.QPS != 0 {
		o2.QPS = o.QPS
	}
	if o.Burst != 0 {
		o2.Burst = o.Burst
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

// WithQPS sets GetConfigOptions.QPS to the specified value.
type WithQPS float32

// ApplyToGetConfig implements GetConfigOption.
func (c WithQPS) ApplyToGetConfig(o *GetConfigOptions) {
	o.QPS = float32(c)
}

// WithBurst sets GetConfigOptions.Burst to the specified value.
type WithBurst int

// ApplyToGetConfig implements GetConfigOption.
func (c WithBurst) ApplyToGetConfig(o *GetConfigOptions) {
	o.Burst = int(c)
}
