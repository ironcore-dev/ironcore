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
	"context"
	"crypto/x509"
	"fmt"
	"os"
	"os/user"
	"path/filepath"
	"strings"
	"time"

	"github.com/go-logr/logr"
	utilrest "github.com/onmetal/onmetal-api/utils/rest"
	certificatesv1 "k8s.io/api/certificates/v1"
	corev1 "k8s.io/api/core/v1"
	utilnet "k8s.io/apimachinery/pkg/util/net"
	"k8s.io/apiserver/pkg/server/egressselector"
	"k8s.io/client-go/rest"
	"k8s.io/client-go/tools/clientcmd"
	clientcmdapi "k8s.io/client-go/tools/clientcmd/api"
	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/client"
)

const (
	// KubeconfigFlagName is the name of the kubeconfig flag.
	KubeconfigFlagName = "kubeconfig"

	// KubeconfigSecretNameFlagName is the name of the kubeconfig-secret-name flag.
	KubeconfigSecretNameFlagName = "kubeconfig-secret-name"

	// KubeconfigSecretNamespaceFlagName is the name of the kubeconfig-secret-namespace flag.
	KubeconfigSecretNamespaceFlagName = "kubeconfig-secret-namespace"

	// BootstrapKubeconfigFlagName is the name of the bootstrap-kubeconfig flag.
	BootstrapKubeconfigFlagName = "bootstrap-kubeconfig"

	// RotateFlagName is the name of the rotate flag.
	RotateFlagName = "rotate"

	// EgressSelectorConfigFlagName is the name of the egress-selector-config flag.
	EgressSelectorConfigFlagName = "egress-selector-config"
)

var (
	log = ctrl.Log.WithName("client").WithName("config")
)

// EgressSelectionName is the name of the egress configuration to use.
type EgressSelectionName string

const (
	// EgressSelectionNameControlPlane instructs to use the controlplane egress selection.
	EgressSelectionNameControlPlane EgressSelectionName = "controlplane"
	// EgressSelectionNameEtcd instructs to use the etcd egress selection.
	EgressSelectionNameEtcd EgressSelectionName = "etcd"
	// EgressSelectionNameCluster instructs to use the cluster egress selection.
	EgressSelectionNameCluster EgressSelectionName = "cluster"
)

// NetworkContext returns the corresponding network context of the egress selection.
func (n EgressSelectionName) NetworkContext() (egressselector.NetworkContext, error) {
	switch n {
	case EgressSelectionNameControlPlane:
		return egressselector.ControlPlane.AsNetworkContext(), nil
	case EgressSelectionNameEtcd:
		return egressselector.Etcd.AsNetworkContext(), nil
	case EgressSelectionNameCluster:
		return egressselector.Cluster.AsNetworkContext(), nil
	default:
		return egressselector.NetworkContext{}, fmt.Errorf("unknown egress selection name %q", n)
	}
}

func setGetConfigOptionsDefaults(o *GetConfigOptions) {
	if o.EgressSelectionName == "" {
		o.EgressSelectionName = EgressSelectionNameControlPlane
	}
}

type NewControllerFunc func(ctx context.Context, store Store, bootstrapCfg *rest.Config, opts ControllerOptions) (Controller, error)

type GetterOptions struct {
	Name              string
	SignerName        string
	Template          *x509.CertificateRequest
	GetUsages         func(privateKey any) []certificatesv1.KeyUsage
	RequestedDuration *time.Duration
	LogConstructor    func() logr.Logger
	NewController     NewControllerFunc
	ForceInitial      bool
}

func setGetterOptionsDefaults(o *GetterOptions) {
	if o.LogConstructor == nil {
		log := ctrl.Log.WithName("client").WithName("config").WithValues("getter", o.Name)
		o.LogConstructor = func() logr.Logger {
			return log
		}
	}
	if o.NewController == nil {
		o.NewController = NewController
	}
}

type Getter struct {
	name              string
	signerName        string
	template          *x509.CertificateRequest
	getUsages         func(privateKey any) []certificatesv1.KeyUsage
	requestedDuration *time.Duration
	logConstructor    func() logr.Logger
	newController     NewControllerFunc
	forceInitial      bool
}

func NewGetter(opts GetterOptions) (*Getter, error) {
	if opts.Name == "" {
		return nil, fmt.Errorf("must specify Name")
	}
	if opts.SignerName == "" {
		return nil, fmt.Errorf("must specify SignerName")
	}
	if opts.Template == nil {
		return nil, fmt.Errorf("must specify Template")
	}
	if opts.GetUsages == nil {
		return nil, fmt.Errorf("must specify GetUsages")
	}
	setGetterOptionsDefaults(&opts)

	return &Getter{
		name:              opts.Name,
		signerName:        opts.SignerName,
		template:          opts.Template,
		getUsages:         opts.GetUsages,
		requestedDuration: opts.RequestedDuration,
		logConstructor:    opts.LogConstructor,
		newController:     opts.NewController,
		forceInitial:      opts.ForceInitial,
	}, nil
}

func NewGetterOrDie(opts GetterOptions) *Getter {
	getter, err := NewGetter(opts)
	if err != nil {
		log.Error(err, "unable to create getter")
		os.Exit(1)
	}
	return getter
}

func StoreFromOptions(o *GetConfigOptions) (Store, error) {
	switch {
	case o.Kubeconfig != "" && o.KubeconfigSecretName != "":
		return nil, fmt.Errorf("cannot specify both kubeconfig and kubeconfig-secret-name")
	case o.Kubeconfig != "":
		return FileStore(o.Kubeconfig), nil
	case o.KubeconfigSecretName != "":
		namespace := o.KubeconfigSecretNamespace
		if namespace == "" {
			namespace = LoadDefaultNamespace()
		}

		defaultCfg, err := LoadDefaultConfig("")
		if err != nil {
			return nil, fmt.Errorf("error loading default config: %w", err)
		}

		c, err := client.New(defaultCfg, client.Options{})
		if err != nil {
			return nil, fmt.Errorf("error creating client: %w", err)
		}

		key := client.ObjectKey{Namespace: namespace, Name: o.KubeconfigSecretName}
		return NewSecretStore(c, key, &SecretStoreOptions{
			Field: o.KubeconfigSecretField,
		}), nil
	default:
		return nil, nil
	}
}

func LoaderFromOptions(o *GetConfigOptions) (Loader, error) {
	switch {
	case o.Kubeconfig != "" && o.KubeconfigSecretName != "":
		return nil, fmt.Errorf("cannot specify both kubeconfig and kubeconfig-secret-name")
	case o.Kubeconfig != "":
		return FileLoader(o.Kubeconfig), nil
	case o.KubeconfigSecretName != "":
		namespace := o.KubeconfigSecretNamespace
		if namespace == "" {
			namespace = LoadDefaultNamespace()
		}

		defaultCfg, err := LoadDefaultConfig("")
		if err != nil {
			return nil, fmt.Errorf("error loading default config: %w", err)
		}

		c, err := client.New(defaultCfg, client.Options{})
		if err != nil {
			return nil, fmt.Errorf("error creating client: %w", err)
		}

		key := client.ObjectKey{Namespace: namespace, Name: o.KubeconfigSecretName}
		return NewSecretLoader(c, key, &SecretLoaderOptions{
			Field: o.KubeconfigSecretField,
		}), nil
	default:
		return nil, nil
	}
}

// GetConfig creates a *rest.Config for talking to a Kubernetes API server.
// Kubeconfig / the '--kubeconfig' flag instruct to use the kubeconfig file at that location.
// Otherwise, will assume running in cluster and use the cluster provided kubeconfig.
//
// It also applies saner defaults for QPS and burst based on the Kubernetes
// controller manager defaults (20 QPS, 30 burst)
//
// # Config precedence
//
// * Kubeconfig / --kubeconfig value / flag pointing at a file
//
// * KUBECONFIG environment variable pointing at a file
//
// * In-cluster config if running in cluster
//
// * $HOME/.kube/config if exists.
func (g *Getter) GetConfig(ctx context.Context, opts ...GetConfigOption) (*rest.Config, Controller, error) {
	o := &GetConfigOptions{}
	o.ApplyOptions(opts)
	setGetConfigOptionsDefaults(o)

	if o.Kubeconfig != "" && o.KubeconfigSecretName != "" {
		return nil, nil, fmt.Errorf("cannot specify kubeconfig and kubeconfig-secret-name")
	}

	switch {
	case o.Rotate:
		g.logConstructor().Info("Getting config and initializing rotation")
		return g.getConfigAndController(ctx, o)
	case o.BootstrapKubeconfig != "":
		g.logConstructor().Info("Getting / bootstrapping config if necessary")
		return g.getAndBootstrapConfigIfNecessary(ctx, o)
	default:
		g.logConstructor().Info("Getting config")
		return g.getConfig(ctx, o)
	}
}

func (g *Getter) getConfig(ctx context.Context, o *GetConfigOptions) (*rest.Config, Controller, error) {
	loader, err := LoaderFromOptions(o)
	if err != nil {
		return nil, nil, fmt.Errorf("error getting loader: %w", err)
	}

	var cfg *rest.Config
	if loader != nil {
		cfg, err = loader.Load(ctx, &clientcmd.ConfigOverrides{
			CurrentContext: o.Context,
		})
	} else {
		cfg, err = LoadDefaultConfig(o.Context)
	}
	if err != nil {
		return nil, nil, err
	}

	dialFunc, err := GetEgressSelectorDial(o.EgressSelectorConfig, o.EgressSelectionName)
	if err != nil {
		return nil, nil, err
	}

	cfg.Dial = dialFunc
	return cfg, nil, nil
}

func (g *Getter) getAndBootstrapConfigIfNecessary(ctx context.Context, o *GetConfigOptions) (*rest.Config, Controller, error) {
	store, err := StoreFromOptions(o)
	if err != nil {
		return nil, nil, fmt.Errorf("error getting store: %w", err)
	}
	if store == nil {
		return nil, nil, fmt.Errorf("must specify either kubeconfig or kubeconfig-secret-name when bootstrap is enabled")
	}

	dialFunc, err := GetEgressSelectorDial(o.EgressSelectorConfig, o.EgressSelectionName)
	if err != nil {
		return nil, nil, err
	}

	cfg, err := store.Get(ctx)
	if IgnoreErrConfigNotFound(err) != nil {
		return nil, nil, fmt.Errorf("error getting kubeconfig: %w", err)
	}
	if cfg != nil {
		cfg.Dial = dialFunc
	}

	bootstrapCfg, err := FileLoader(o.BootstrapKubeconfig).Load(ctx, nil)
	if IgnoreErrConfigNotFound(err) != nil {
		return nil, nil, fmt.Errorf("error loading bootstrap kubeconfig: %w", err)
	}
	if bootstrapCfg != nil {
		cfg.Dial = dialFunc
	}

	if cfg == nil && bootstrapCfg == nil {
		return nil, nil, fmt.Errorf("must specify either valid kubeconfig or bootstrap-kubeconfig")
	}

	cfg, newCfg, err := utilrest.UseOrRequestConfig(ctx, cfg, bootstrapCfg, g.signerName, g.template, g.getUsages, g.requestedDuration)
	if err != nil {
		return nil, nil, fmt.Errorf("error using / requesting config: %w", err)
	}
	if newCfg {
		if err := store.Set(ctx, cfg); err != nil {
			return nil, nil, fmt.Errorf("error persisting new config: %w", err)
		}
	}

	cfg.Dial = dialFunc
	return cfg, nil, nil
}

func (g *Getter) getConfigAndController(ctx context.Context, o *GetConfigOptions) (*rest.Config, Controller, error) {
	store, err := StoreFromOptions(o)
	if err != nil {
		return nil, nil, fmt.Errorf("error getting store: %w", err)
	}
	if store == nil {
		return nil, nil, fmt.Errorf("must specify either kubeconfig or kubeconfig-secret-name when rotate is enabled")
	}

	dialFunc, err := GetEgressSelectorDial(o.EgressSelectorConfig, o.EgressSelectionName)
	if err != nil {
		return nil, nil, err
	}

	bootstrapCfg, err := FileLoader(o.BootstrapKubeconfig).Load(ctx, nil)
	if IgnoreErrConfigNotFound(err) != nil {
		return nil, nil, fmt.Errorf("error loading bootstrap kubeconfig: %w", err)
	}
	if bootstrapCfg != nil {
		bootstrapCfg.Dial = dialFunc
	}

	c, err := g.newController(ctx, store, bootstrapCfg, ControllerOptions{
		Name:              g.name,
		SignerName:        g.signerName,
		Template:          g.template,
		GetUsages:         g.getUsages,
		RequestedDuration: g.requestedDuration,
		LogConstructor: func() logr.Logger {
			return g.logConstructor().WithValues("controller", g.name)
		},
		DialFunc:     dialFunc,
		ForceInitial: g.forceInitial,
	})
	if err != nil {
		return nil, nil, fmt.Errorf("error creating rotator: %w", err)
	}
	if err := c.Init(ctx, false); err != nil {
		return nil, nil, fmt.Errorf("error running rotator once: %w", err)
	}
	return c.TransportConfig(), c, nil
}

func LoadDefaultNamespace() string {
	if ns := os.Getenv("POD_NAMESPACE"); ns != "" {
		return ns
	}

	// Fall back to the namespace associated with the service account token, if available
	if data, err := os.ReadFile("/var/run/secrets/kubernetes.io/serviceaccount/namespace"); err == nil {
		if ns := strings.TrimSpace(string(data)); len(ns) > 0 {
			return ns
		}
	}

	return corev1.NamespaceDefault
}

func loadConfigWithContext(apiServerURL string, loader clientcmd.ClientConfigLoader, context string) (*rest.Config, error) {
	return clientcmd.NewNonInteractiveDeferredLoadingClientConfig(
		loader,
		&clientcmd.ConfigOverrides{
			ClusterInfo: clientcmdapi.Cluster{
				Server: apiServerURL,
			},
			CurrentContext: context,
		}).ClientConfig()
}

func LoadDefaultConfig(context string) (*rest.Config, error) {
	// If the recommended kubeconfig env variable is not specified,
	// try the in-cluster config.
	kubeconfigPath := os.Getenv(clientcmd.RecommendedConfigPathEnvVar)
	if len(kubeconfigPath) == 0 {
		if c, err := rest.InClusterConfig(); err == nil {
			return c, nil
		}
	}

	loadingRules := clientcmd.NewDefaultClientConfigLoadingRules()
	if _, ok := os.LookupEnv("HOME"); !ok {
		u, err := user.Current()
		if err != nil {
			return nil, fmt.Errorf("could not get current user: %v", err)
		}
		loadingRules.Precedence = append(
			loadingRules.Precedence,
			filepath.Join(u.HomeDir, clientcmd.RecommendedHomeDir, clientcmd.RecommendedFileName),
		)
	}

	return loadConfigWithContext("", loadingRules, context)
}

func GetEgressSelectorDial(egressSelectorConfig string, egressSelectionName EgressSelectionName) (utilnet.DialFunc, error) {
	if egressSelectorConfig == "" {
		return nil, nil
	}

	networkContext, err := egressSelectionName.NetworkContext()
	if err != nil {
		return nil, fmt.Errorf("error obtaining network context: %w", err)
	}

	egressSelectorCfg, err := egressselector.ReadEgressSelectorConfiguration(egressSelectorConfig)
	if err != nil {
		return nil, fmt.Errorf("error reading egress selector configuration: %w", err)
	}

	egressSelector, err := egressselector.NewEgressSelector(egressSelectorCfg)
	if err != nil {
		return nil, fmt.Errorf("error creating egress selector: %w", err)
	}

	dial, err := egressSelector.Lookup(networkContext)
	if err != nil {
		return nil, fmt.Errorf("error looking up network context %s: %w", networkContext.EgressSelectionName.String(), err)
	}
	if dial == nil {
		return nil, fmt.Errorf("no dialer for network context %s", networkContext.EgressSelectionName.String())
	}

	return dial, nil
}

// GetConfigOrDie creates a *rest.Config for talking to a Kubernetes apiserver.
// If Kubeconfig / --kubeconfig is set, will use the kubeconfig file at that location. Otherwise, will assume running
// in cluster and use the cluster provided kubeconfig.
//
// Will log an error and exit if there is an error creating the rest.Config.
func (g *Getter) GetConfigOrDie(ctx context.Context, opts ...GetConfigOption) (*rest.Config, Controller) {
	config, rotator, err := g.GetConfig(ctx, opts...)
	if err != nil {
		log.Error(err, "unable to get kubeconfig")
		os.Exit(1)
	}
	return config, rotator
}
