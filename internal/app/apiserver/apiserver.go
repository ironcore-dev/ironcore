// SPDX-FileCopyrightText: SAP SE or an SAP affiliate company and IronCore contributors
// SPDX-License-Identifier: Apache-2.0

package apiserver

import (
	"context"
	"crypto/tls"
	"crypto/x509"
	"crypto/x509/pkix"
	"fmt"
	"net"
	"time"

	certificatesv1 "k8s.io/api/certificates/v1"
	"k8s.io/client-go/rest"
	"k8s.io/component-base/compatibility"

	computev1alpha1 "github.com/ironcore-dev/ironcore/api/compute/v1alpha1"
	corev1alpha1 "github.com/ironcore-dev/ironcore/api/core/v1alpha1"
	ipamv1alpha1 "github.com/ironcore-dev/ironcore/api/ipam/v1alpha1"
	networkingv1alpha1 "github.com/ironcore-dev/ironcore/api/networking/v1alpha1"
	storagev1alpha1 "github.com/ironcore-dev/ironcore/api/storage/v1alpha1"
	informers "github.com/ironcore-dev/ironcore/client-go/informers/externalversions"
	clientset "github.com/ironcore-dev/ironcore/client-go/ironcore/versioned"
	ironcoreopenapi "github.com/ironcore-dev/ironcore/client-go/openapi"
	ironcoreinitializer "github.com/ironcore-dev/ironcore/internal/admission/initializer"
	"github.com/ironcore-dev/ironcore/internal/admission/plugin/machinevolumedevices"
	"github.com/ironcore-dev/ironcore/internal/admission/plugin/resourcequota"
	"github.com/ironcore-dev/ironcore/internal/admission/plugin/volumeresizepolicy"
	"github.com/ironcore-dev/ironcore/internal/api"
	"github.com/ironcore-dev/ironcore/internal/apis/compute"
	"github.com/ironcore-dev/ironcore/internal/apiserver"
	"github.com/ironcore-dev/ironcore/internal/machinepoollet/client"
	"github.com/ironcore-dev/ironcore/internal/quota/evaluator/ironcore"
	utilcertificate "github.com/ironcore-dev/ironcore/utils/certificate"
	config "github.com/ironcore-dev/ironcore/utils/client/config"
	apiequality "github.com/ironcore-dev/ironcore/utils/equality"
	"github.com/ironcore-dev/ironcore/utils/quota"
	utilrest "github.com/ironcore-dev/ironcore/utils/rest"
	"github.com/spf13/cobra"
	"github.com/spf13/pflag"
	"k8s.io/apimachinery/pkg/api/equality"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apimachinery/pkg/runtime/schema"
	utilerrors "k8s.io/apimachinery/pkg/util/errors"
	utilruntime "k8s.io/apimachinery/pkg/util/runtime"
	"k8s.io/apiserver/pkg/admission"
	"k8s.io/apiserver/pkg/endpoints/openapi"
	genericapiserver "k8s.io/apiserver/pkg/server"
	"k8s.io/apiserver/pkg/server/healthz"
	genericoptions "k8s.io/apiserver/pkg/server/options"
	serverstorage "k8s.io/apiserver/pkg/server/storage"
	utilfeature "k8s.io/apiserver/pkg/util/feature"
	netutils "k8s.io/utils/net"
)

const defaultEtcdPathPrefix = "/registry/ironcore.dev"

func init() {
	utilruntime.Must(apiequality.AddFuncs(equality.Semantic))
}

func NewResourceConfig() *serverstorage.ResourceConfig {
	cfg := serverstorage.NewResourceConfig()
	cfg.EnableVersions(
		computev1alpha1.SchemeGroupVersion,
		corev1alpha1.SchemeGroupVersion,
		storagev1alpha1.SchemeGroupVersion,
		networkingv1alpha1.SchemeGroupVersion,
		ipamv1alpha1.SchemeGroupVersion,
	)
	return cfg
}

func (o *IronCoreAPIServerOptions) AddFlags(fs *pflag.FlagSet) {
	o.RecommendedOptions.AddFlags(fs)

	// machinepoollet related flags:
	fs.StringSliceVar(&o.MachinePoolletConfig.PreferredAddressTypes, "machinepoollet-preferred-address-types", o.MachinePoolletConfig.PreferredAddressTypes,
		"List of the preferred MachinePoolAddressTypes to use for machinepoollet connections.")

	fs.DurationVar(&o.MachinePoolletConfig.HTTPTimeout, "machinepoollet-timeout", o.MachinePoolletConfig.HTTPTimeout,
		"Timeout for machinepoollet operations.")

	fs.StringVar(&o.MachinePoolletConfig.CertFile, "machinepoollet-client-certificate", o.MachinePoolletConfig.CertFile,
		"Path to a client cert file for TLS.")

	fs.StringVar(&o.MachinePoolletConfig.KeyFile, "machinepoollet-client-key", o.MachinePoolletConfig.KeyFile,
		"Path to a client key file for TLS.")

	fs.StringVar(&o.MachinePoolletConfig.CAFile, "machinepoollet-certificate-authority", o.MachinePoolletConfig.CAFile,
		"Path to a cert file for the certificate authority.")

	o.MachinePoolletClientGetConfig.BindFlags(fs, config.WithNamePrefix("machinepoollet-client-"))
}

func NewIronCoreAPIServerOptions() *IronCoreAPIServerOptions {
	o := &IronCoreAPIServerOptions{
		RecommendedOptions: genericoptions.NewRecommendedOptions(
			defaultEtcdPathPrefix,
			api.Codecs.LegacyCodec(
				computev1alpha1.SchemeGroupVersion,
				corev1alpha1.SchemeGroupVersion,
				storagev1alpha1.SchemeGroupVersion,
				networkingv1alpha1.SchemeGroupVersion,
				ipamv1alpha1.SchemeGroupVersion,
			),
		),
		MachinePoolletConfig: client.MachinePoolletClientConfig{
			Port:         12319,
			ReadOnlyPort: 12320,
			PreferredAddressTypes: []string{
				string(compute.MachinePoolHostName),

				// internal, preferring DNS if reported
				string(compute.MachinePoolInternalDNS),
				string(compute.MachinePoolInternalIP),

				// external, preferring DNS if reported
				string(compute.MachinePoolExternalDNS),
				string(compute.MachinePoolExternalIP),
			},
			HTTPTimeout: time.Duration(5) * time.Second,
		},
	}
	o.RecommendedOptions.Etcd.StorageConfig.EncodeVersioner = runtime.NewMultiGroupVersioner(
		computev1alpha1.SchemeGroupVersion,
		schema.GroupKind{Group: computev1alpha1.SchemeGroupVersion.Group},
		schema.GroupKind{Group: corev1alpha1.SchemeGroupVersion.Group},
		schema.GroupKind{Group: storagev1alpha1.SchemeGroupVersion.Group},
		schema.GroupKind{Group: networkingv1alpha1.SchemeGroupVersion.Group},
		schema.GroupKind{Group: ipamv1alpha1.SchemeGroupVersion.Group},
	)
	return o
}

func NewCommandStartIronCoreAPIServer(ctx context.Context, defaults *IronCoreAPIServerOptions) *cobra.Command {
	o := *defaults
	cmd := &cobra.Command{
		Short: "Launch an ironcoreAPI server",
		Long:  "Launch an ironcoreAPI server",
		RunE: func(cmd *cobra.Command, args []string) error {
			if err := o.Complete(); err != nil {
				return err
			}
			if err := o.Validate(args); err != nil {
				return err
			}
			if err := o.Run(ctx); err != nil {
				return err
			}
			return nil
		},
	}

	o.AddFlags(cmd.Flags())
	utilfeature.DefaultMutableFeatureGate.AddFlag(cmd.Flags())

	return cmd
}

func (o *IronCoreAPIServerOptions) Validate(args []string) error {
	var errors []error
	errors = append(errors, o.RecommendedOptions.Validate()...)
	return utilerrors.NewAggregate(errors)
}

func (o *IronCoreAPIServerOptions) Complete() error {
	machinevolumedevices.Register(o.RecommendedOptions.Admission.Plugins)
	resourcequota.Register(o.RecommendedOptions.Admission.Plugins)
	volumeresizepolicy.Register(o.RecommendedOptions.Admission.Plugins)

	o.RecommendedOptions.Admission.RecommendedPluginOrder = append(
		o.RecommendedOptions.Admission.RecommendedPluginOrder,
		machinevolumedevices.PluginName,
		resourcequota.PluginName,
		volumeresizepolicy.PluginName,
	)

	return nil
}

func (o *IronCoreAPIServerOptions) Config(ctx context.Context) (*apiserver.Config, error) {
	if err := o.RecommendedOptions.SecureServing.MaybeDefaultWithSelfSignedCerts("localhost", nil, []net.IP{netutils.ParseIPSloppy("127.0.0.1")}); err != nil {
		return nil, fmt.Errorf("error creating self-signed certificates: %w", err)
	}

	if err := o.setupMachinePoolletClientRotation(ctx); err != nil {
		return nil, fmt.Errorf("error setting up machinepoollet client certificate rotation: %w", err)
	}

	o.RecommendedOptions.ExtraAdmissionInitializers = func(c *genericapiserver.RecommendedConfig) ([]admission.PluginInitializer, error) {
		ironcoreClient, err := clientset.NewForConfig(c.LoopbackClientConfig)
		if err != nil {
			return nil, err
		}

		informerFactory := informers.NewSharedInformerFactory(ironcoreClient, c.LoopbackClientConfig.Timeout)
		o.SharedInformerFactory = informerFactory

		quotaRegistry := quota.NewRegistry(api.Scheme)
		if err := quota.AddAllToRegistry(quotaRegistry, ironcore.NewEvaluatorsForAdmission(ironcoreClient, informerFactory)); err != nil {
			return nil, fmt.Errorf("error initializing quota registry: %w", err)
		}

		genericInitializer := ironcoreinitializer.New(ironcoreClient, informerFactory, quotaRegistry)

		return []admission.PluginInitializer{
			genericInitializer,
		}, nil
	}

	serverConfig := genericapiserver.NewRecommendedConfig(api.Codecs)

	serverConfig.EffectiveVersion = compatibility.NewEffectiveVersionFromString("1.0", "", "")

	serverConfig.OpenAPIConfig = genericapiserver.DefaultOpenAPIConfig(ironcoreopenapi.GetOpenAPIDefinitions, openapi.NewDefinitionNamer(api.Scheme))
	serverConfig.OpenAPIConfig.Info.Title = "ironcore-api"
	serverConfig.OpenAPIConfig.Info.Version = "0.1"

	serverConfig.OpenAPIV3Config = genericapiserver.DefaultOpenAPIV3Config(ironcoreopenapi.GetOpenAPIDefinitions, openapi.NewDefinitionNamer(api.Scheme))
	serverConfig.OpenAPIV3Config.Info.Title = "ironcore-api"
	serverConfig.OpenAPIV3Config.Info.Version = "0.1"

	if err := o.RecommendedOptions.ApplyTo(serverConfig); err != nil {
		return nil, err
	}

	apiResourceConfig := NewResourceConfig()

	if o.machinePoolletClientCtrl != nil {
		serverConfig.AddHealthzChecks(healthz.NamedCheck("machinepoollet-client-rotator", o.machinePoolletClientCtrl.Check))
	}

	config := &apiserver.Config{
		GenericConfig: serverConfig,
		ExtraConfig: apiserver.ExtraConfig{
			APIResourceConfigSource: apiResourceConfig,
			MachinePoolletConfig:    o.MachinePoolletConfig,
		},
	}

	if config.GenericConfig.EgressSelector != nil {
		// Use the config.GenericConfig.EgressSelector lookup to find the dialer to connect to the machinepoollet
		config.ExtraConfig.MachinePoolletConfig.Lookup = config.GenericConfig.EgressSelector.Lookup
	}

	return config, nil
}

type IronCoreAPIServerOptions struct {
	RecommendedOptions   *genericoptions.RecommendedOptions
	MachinePoolletConfig client.MachinePoolletClientConfig

	// MachinePoolletClientGetConfig holds the options for optionally
	// bootstrapping / CSR-rotating the machinepoollet client certificate.
	// When unset, the static --machinepoollet-client-certificate/-key path is used.
	MachinePoolletClientGetConfig config.GetConfigOptions

	// machinePoolletClientCtrl is the running rotator, set when rotation is
	// enabled; used to register a health check on the apiserver.
	machinePoolletClientCtrl config.Controller

	SharedInformerFactory informers.SharedInformerFactory
}

// setupMachinePoolletClientRotation, when rotation is requested via the
// machinepoollet-client-* flags, builds and starts a client certificate
// rotator and feeds the rotating certificate into the machinepoollet dialer
// via a dynamic tls.Config.GetClientCertificate callback. It must run before
// NewMachinePoolConnectionInfoGetter captures the transport tls.Config.
func (o *IronCoreAPIServerOptions) setupMachinePoolletClientRotation(ctx context.Context) error {
	if !o.MachinePoolletClientGetConfig.RotateCertificates {
		if o.MachinePoolletClientGetConfig.BootstrapKubeconfig != "" {
			return fmt.Errorf("--machinepoollet-client-bootstrap-kubeconfig requires --machinepoollet-client-rotate-certificates (a static bootstrap config cannot be rotated)")
		}
		return nil
	}

	getter, err := config.NewGetter(config.GetterOptions{
		Name:       "apiserver-machinepoollet-client",
		SignerName: certificatesv1.KubeAPIServerClientSignerName,
		Template: &x509.CertificateRequest{
			Subject: pkix.Name{
				CommonName:   "ironcore-apiserver-machinepoollet-client",
				Organization: []string{"system:masters"},
			},
		},
		GetUsages: utilcertificate.DefaultKubeAPIServerClientGetUsages,
	})
	if err != nil {
		return err
	}

	// GetConfig already initializes the rotator (Init(ctx, false)); start it
	// on the long-lived Run context so rotation continues for the apiserver's
	// lifetime.
	_, ctrl, err := getter.GetConfig(ctx, &o.MachinePoolletClientGetConfig)
	if err != nil {
		return err
	}
	o.machinePoolletClientCtrl = ctrl
	go func() {
		if err := ctrl.Start(ctx); err != nil {
			utilruntime.HandleError(fmt.Errorf("error running machinepoollet client certificate rotator: %w", err))
		}
	}()

	o.MachinePoolletConfig.GetClientCertificate = func(*tls.CertificateRequestInfo) (*tls.Certificate, error) {
		return clientCertFromConfig(ctrl.ClientConfig())
	}
	return nil
}

// clientCertFromConfig extracts a tls.Certificate from a rest.Config produced
// by the rotator, which materializes CertData/KeyData. It copies the config
// before reading so the rotator's cached config is not mutated.
func clientCertFromConfig(cfg *rest.Config) (*tls.Certificate, error) {
	if cfg == nil {
		return nil, nil
	}
	shallow := *cfg
	return utilrest.CertificateFromConfig(&shallow)
}

func (o *IronCoreAPIServerOptions) Run(ctx context.Context) error {
	config, err := o.Config(ctx)
	if err != nil {
		return err
	}

	server, err := config.Complete().New()
	if err != nil {
		return err
	}

	server.GenericAPIServer.AddPostStartHookOrDie("start-ironcore-api-server-informers", func(hookContext genericapiserver.PostStartHookContext) error {
		config.GenericConfig.SharedInformerFactory.Start(hookContext.Done())
		o.SharedInformerFactory.Start(hookContext.Done())
		return nil
	})

	return server.GenericAPIServer.PrepareRun().RunWithContext(ctx)
}
