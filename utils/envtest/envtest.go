// Copyright 2021 OnMetal authors
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

package envtest

import (
	"context"
	"fmt"
	"os"
	"path/filepath"
	"reflect"
	"sort"
	"strings"
	"time"

	"github.com/onmetal/controller-utils/conditionutils"
	"github.com/onmetal/onmetal-api/utils/envtest/internal/testing/addr"
	"github.com/onmetal/onmetal-api/utils/envtest/internal/testing/certs"
	corev1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/api/meta"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apimachinery/pkg/runtime/schema"
	"k8s.io/apimachinery/pkg/util/wait"
	"k8s.io/client-go/rest"
	apiregistrationv1 "k8s.io/kube-aggregator/pkg/apis/apiregistration/v1"
	"k8s.io/utils/pointer"
	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/client/apiutil"
	"sigs.k8s.io/controller-runtime/pkg/envtest"
)

const (
	envUseExistingCluster = "USE_EXISTING_CLUSTER"
)

type APIServiceInstallOptions struct {
	Paths              []string
	ErrorIfPathMissing bool

	ClientCertDir string
	ClientCAData  []byte

	// LocalServingCertDir is the allocated directory for serving certificates.
	// it will be automatically populated by the local temp dir
	LocalServingCertDir string

	LocalServingPort int

	LocalServingHost             string
	LocalServingHostExternalName string
	LocalServingCAData           []byte

	APIServices []*apiregistrationv1.APIService

	ServiceNamespace string
	ServiceName      string
}

func (o *APIServiceInstallOptions) tlsCertPath() string {
	return filepath.Join(o.LocalServingCertDir, "tls.crt")
}

func (o *APIServiceInstallOptions) tlsKeyPath() string {
	return filepath.Join(o.LocalServingCertDir, "tls.key")
}

func (o *APIServiceInstallOptions) caCertPath() string {
	return filepath.Join(o.LocalServingCertDir, "ca.crt")
}

func (o *APIServiceInstallOptions) clientCertPath() string {
	return filepath.Join(o.ClientCertDir, "client.crt")
}

func (o *APIServiceInstallOptions) clientKeyPath() string {
	return filepath.Join(o.ClientCertDir, "client.key")
}

func (o *APIServiceInstallOptions) clientCACertPath() string {
	return filepath.Join(o.ClientCertDir, "client-ca.crt")
}

func (o *APIServiceInstallOptions) setupCA() error {
	apiServiceCA, err := certs.NewTinyCA()
	if err != nil {
		return fmt.Errorf("unable to set up api service CA: %v", err)
	}

	names := []string{"localhost", o.LocalServingHost, o.LocalServingHostExternalName}
	apiServiceCert, err := apiServiceCA.NewServingCert(names, []string{"aggregated-apiserver.kube-system.svc"})
	if err != nil {
		return fmt.Errorf("unable to set up api service serving certs: %v", err)
	}

	localServingCertsDir, err := os.MkdirTemp("", "envtest-apiservice-certs-")
	o.LocalServingCertDir = localServingCertsDir
	if err != nil {
		return fmt.Errorf("unable to create directory for apiservice serving certs: %v", err)
	}

	certData, keyData, err := apiServiceCert.AsBytes()
	if err != nil {
		return fmt.Errorf("unable to marshal api service serving certs: %v", err)
	}

	if err := os.WriteFile(o.caCertPath(), apiServiceCA.CA.CertBytes(), 0640); err != nil {
		return fmt.Errorf("unable to write api service ca to disk: %w", err)
	}
	if err := os.WriteFile(o.tlsCertPath(), certData, 0640); err != nil {
		return fmt.Errorf("unable to write api service serving cert to disk: %w", err)
	}
	if err := os.WriteFile(o.tlsKeyPath(), keyData, 0640); err != nil {
		return fmt.Errorf("unable to write api service serving key to disk: %w", err)
	}

	o.LocalServingCAData = apiServiceCA.CA.CertBytes()
	return err
}

func (o *APIServiceInstallOptions) generateHostPort() (string, int, error) {
	if o.LocalServingPort == 0 {
		port, host, err := addr.Suggest(o.LocalServingHost)
		if err != nil {
			return "", 0, fmt.Errorf("unable to grab random port for serving api services on: %v", err)
		}
		o.LocalServingPort = port
		o.LocalServingHost = host
	}
	host := o.LocalServingHostExternalName
	if host == "" {
		host = o.LocalServingHost
	}
	return host, o.LocalServingPort, nil
}

func (o *APIServiceInstallOptions) generateService(cfg *rest.Config) (namespace, name string, err error) {
	ctx := context.TODO()
	c, err := client.New(cfg, client.Options{})
	if err != nil {
		return "", "", fmt.Errorf("error creating client: %w", err)
	}

	namespace = o.ServiceNamespace
	if namespace == "" {
		namespace = metav1.NamespaceSystem
	}

	host, _, err := o.generateHostPort()
	if err != nil {
		return "", "", fmt.Errorf("error generating host port: %w", err)
	}

	name = o.ServiceName
	if name == "" {
		name = "aggregated-apiserver"
	}

	service := &corev1.Service{
		TypeMeta: metav1.TypeMeta{
			APIVersion: corev1.SchemeGroupVersion.String(),
			Kind:       "Service",
		},
		ObjectMeta: metav1.ObjectMeta{
			Namespace: namespace,
			Name:      name,
		},
		Spec: corev1.ServiceSpec{
			Type:         corev1.ServiceTypeExternalName,
			ExternalName: host,
		},
	}
	if err := c.Patch(ctx, service, client.Apply, client.ForceOwnership, fieldOwner); err != nil {
		return "", "", fmt.Errorf("error applying service: %w", err)
	}

	o.ServiceNamespace = namespace
	o.ServiceName = name
	return namespace, name, nil
}

// ModifyAPIServiceDefinitions modifies APIService definitions by:
// - applying CABundle based on the provided tinyca
// - applying service based on the created service
func (o *APIServiceInstallOptions) ModifyAPIServiceDefinitions(cfg *rest.Config) error {
	// generate host port.
	_, port, err := o.generateHostPort()
	if err != nil {
		return fmt.Errorf("error generating host port: %w", err)
	}

	svcNamespace, svcName, err := o.generateService(cfg)
	if err != nil {
		return fmt.Errorf("error generating service: %w", err)
	}

	for _, apiService := range o.APIServices {
		apiService.Spec.CABundle = o.LocalServingCAData
		apiService.Spec.Service = &apiregistrationv1.ServiceReference{
			Namespace: svcNamespace,
			Name:      svcName,
			Port:      pointer.Int32(int32(port)),
		}
	}
	return nil
}

func (o *APIServiceInstallOptions) SetupClientCA() error {
	clientCA, err := certs.NewTinyCA()
	if err != nil {
		return fmt.Errorf("unable to set up apiserver client CA: %v", err)
	}

	clientCert, err := clientCA.NewClientCert(certs.ClientInfo{
		Name:   "envtest-environment",
		Groups: []string{"system:masters"},
	})
	if err != nil {
		return fmt.Errorf("unable to set up apiserver client serving certs: %v", err)
	}

	clientCertDir, err := os.MkdirTemp("", "envtest-apiserver-client-certs-")
	o.ClientCertDir = clientCertDir
	if err != nil {
		return fmt.Errorf("unable to create directory for apiserver client certs: %v", err)
	}

	certData, keyData, err := clientCert.AsBytes()
	if err != nil {
		return fmt.Errorf("unable to marshal apiserver client certs: %v", err)
	}

	if err := os.WriteFile(o.clientCACertPath(), clientCA.CA.CertBytes(), 0640); err != nil {
		return fmt.Errorf("unable to write apiserver client ca to disk: %w", err)
	}
	if err := os.WriteFile(o.clientCertPath(), certData, 0640); err != nil {
		return fmt.Errorf("unable to write apiserver client cert to disk: %w", err)
	}
	if err := os.WriteFile(o.clientKeyPath(), keyData, 0640); err != nil {
		return fmt.Errorf("unable to write apiserver client key to disk: %w", err)
	}

	o.ClientCAData = clientCA.CA.CertBytes()
	return err
}

func (o *APIServiceInstallOptions) PrepWithoutInstalling(cfg *rest.Config) error {
	if err := o.setupCA(); err != nil {
		return fmt.Errorf("error setting up ca: %w", err)
	}

	if err := o.ModifyAPIServiceDefinitions(cfg); err != nil {
		return fmt.Errorf("error modifying api service definitions: %w", err)
	}

	if o.LocalServingHost == "" {
		host, port, err := o.generateHostPort()
		if err != nil {
			return fmt.Errorf("error generating host port: %w", err)
		}

		o.LocalServingHost = host
		o.LocalServingPort = port
	}

	return nil
}

func (o *APIServiceInstallOptions) Install(cfg *rest.Config) error {
	if err := readAPIServiceFiles(o); err != nil {
		return fmt.Errorf("error reading api services: %w", err)
	}

	if err := o.PrepWithoutInstalling(cfg); err != nil {
		return fmt.Errorf("[aggregated api server] error preparing: %w", err)
	}

	if err := o.ApplyAPIServices(cfg); err != nil {
		return fmt.Errorf("[aggregated api server] error installing: %w", err)
	}

	return nil
}

func (o *APIServiceInstallOptions) Stop() error {
	return nil
}

const fieldOwner = client.FieldOwner("envtest.onmetal-api.onmetal.de")

func (o *APIServiceInstallOptions) ApplyAPIServices(cfg *rest.Config) error {
	ctx := context.TODO()
	c, err := client.New(cfg, client.Options{})
	if err != nil {
		return fmt.Errorf("error creating client: %w", err)
	}

	for _, apiService := range o.APIServices {
		desired := apiService.DeepCopy()
		apiService.TypeMeta = metav1.TypeMeta{
			APIVersion: apiregistrationv1.SchemeGroupVersion.String(),
			Kind:       "APIService",
		}

		if _, err := ctrl.CreateOrUpdate(ctx, c, apiService, func() error {
			apiService.Spec = desired.Spec
			return nil
		}); err != nil {
			return fmt.Errorf("error applying api service: %w", err)
		}
	}
	return nil
}

type AdditionalService struct {
	Name string
	Host string
	Port int
}

type EnvironmentExtensions struct {
	APIServiceInstallOptions APIServiceInstallOptions
	APIServices              []*apiregistrationv1.APIService

	// APIServiceDirectoryPaths is a list of paths containing APIService yaml or json configs.
	// If both this field and Paths field in APIServiceInstallOptions are specified, the
	// values are merged.
	APIServiceDirectoryPaths []string

	ErrorIfAPIServicePathIsMissing bool

	AdditionalServices []AdditionalService
}

func (e *EnvironmentExtensions) GetAdditionalServiceHost(name string) string {
	for _, svc := range e.AdditionalServices {
		if svc.Name == name {
			return svc.Host
		}
	}
	return ""
}

func (e *EnvironmentExtensions) GetAdditionalServicePort(name string) int {
	for _, svc := range e.AdditionalServices {
		if svc.Name == name {
			return svc.Port
		}
	}
	return 0
}

func (e *EnvironmentExtensions) GetAdditionalServiceHostPort(name string) (string, int) {
	for _, svc := range e.AdditionalServices {
		if svc.Name == name {
			return svc.Host, svc.Port
		}
	}
	return "", 0
}

func envUsesExistingCluster(env *envtest.Environment) bool {
	if env.UseExistingCluster == nil {
		return strings.ToLower(os.Getenv(envUseExistingCluster)) == "true"
	}
	return *env.UseExistingCluster
}

func configureAPIServerAggregation(env *envtest.Environment, ext *EnvironmentExtensions) {
	env.ControlPlane.GetAPIServer().Configure().
		Set("proxy-client-key-file", ext.APIServiceInstallOptions.clientKeyPath()).
		Set("proxy-client-cert-file", ext.APIServiceInstallOptions.clientCertPath()).
		Set("requestheader-client-ca-file", ext.APIServiceInstallOptions.clientCACertPath()).
		Set("requestheader-allowed-names", "envtest-environment,localhost").
		Set("requestheader-extra-headers-prefix", "X-Remote-Extra-").
		Set("requestheader-group-headers", "X-Remote-Group").
		Set("requestheader-username-headers", "X-Remote-User")
}

func prepareAdditionalServices(additionalServices []AdditionalService) error {
	for i := range additionalServices {
		additionalService := &additionalServices[i]
		port, host, err := addr.Suggest(additionalService.Host)
		if err != nil {
			return fmt.Errorf("[additional service %s] error suggesting host / port: %w", additionalService.Name, err)
		}

		additionalService.Host = host
		additionalService.Port = port
	}
	return nil
}

func StartWithExtensions(env *envtest.Environment, ext *EnvironmentExtensions) (*rest.Config, error) {
	ext.APIServiceInstallOptions.APIServices = mergeAPIServices(ext.APIServiceInstallOptions.APIServices, ext.APIServices)
	ext.APIServiceInstallOptions.Paths = mergePaths(ext.APIServiceInstallOptions.Paths, ext.APIServiceDirectoryPaths)
	ext.APIServiceInstallOptions.ErrorIfPathMissing = ext.ErrorIfAPIServicePathIsMissing

	if err := ext.APIServiceInstallOptions.SetupClientCA(); err != nil {
		return nil, fmt.Errorf("error setting up client ca: %w", err)
	}

	if !envUsesExistingCluster(env) {
		configureAPIServerAggregation(env, ext)
	}

	cfg, err := env.Start()
	if err != nil {
		return nil, err
	}

	if err := ext.APIServiceInstallOptions.Install(cfg); err != nil {
		if err := env.Stop(); err != nil {
			log.Error(err, "Error stopping test-env")
		}
		return nil, err
	}

	if err := prepareAdditionalServices(ext.AdditionalServices); err != nil {
		if err := env.Stop(); err != nil {
			log.Error(err, "Error stopping test-env")
		}
		return nil, err
	}

	return cfg, nil
}

func StopWithExtensions(env *envtest.Environment, ext *EnvironmentExtensions) error {
	if err := ext.APIServiceInstallOptions.Stop(); err != nil {
		return fmt.Errorf("error stopping aggregated api server: %w", err)
	}

	if err := env.Stop(); err != nil {
		return fmt.Errorf("error stopping environment: %w", err)
	}

	return nil
}

func WaitUntilTypesDiscoverableWithTimeout(timeout time.Duration, c client.Client, objs ...client.Object) error {
	ctx, cancel := context.WithTimeout(context.Background(), timeout)
	defer cancel()

	return WaitUntilTypesDiscoverable(ctx, c, objs...)
}

func waitUntilGVKsDiscoverable(ctx context.Context, c client.Client, gvks map[schema.GroupVersionKind]struct{}) error {
	if err := wait.PollUntilContextCancel(ctx, 50*time.Millisecond, true, func(ctx context.Context) (done bool, err error) {
		for gvk := range gvks {
			mappings, err := c.RESTMapper().RESTMappings(gvk.GroupKind(), gvk.Version)
			if err != nil {
				if !meta.IsNoMatchError(err) {
					return false, fmt.Errorf("error getting rest mappings for %s: %w", gvk, err)
				}
				continue
			}

			switch n := len(mappings); n {
			case 0:
				continue
			case 1:
				delete(gvks, gvk)
			default:
				return false, fmt.Errorf("unexpected state: multiple rest mappings for %s: %v", gvk, mappings)
			}
		}

		return len(gvks) == 0, nil
	}); err != nil {
		if len(gvks) == 0 {
			return err
		}
		unavailableGVKs := make([]string, 0, len(gvks))
		for gvk := range gvks {
			unavailableGVKs = append(unavailableGVKs, gvk.String())
		}
		sort.Strings(unavailableGVKs)
		return fmt.Errorf("%w, unavailable gvks: %v", err, unavailableGVKs)
	}
	return nil
}

func WaitUntilTypesDiscoverable(ctx context.Context, c client.Client, objs ...client.Object) error {
	gvks := make(map[schema.GroupVersionKind]struct{})
	for _, obj := range objs {
		gvk, err := apiutil.GVKForObject(obj, c.Scheme())
		if err != nil {
			return fmt.Errorf("error getting gvk for %T: %w", obj, err)
		}

		gvks[gvk] = struct{}{}
	}

	return waitUntilGVKsDiscoverable(ctx, c, gvks)
}

var clientObjectType = reflect.TypeOf((*client.Object)(nil)).Elem()

func WaitUntilGroupVersionsDiscoverable(ctx context.Context, c client.Client, scheme *runtime.Scheme, gvs ...schema.GroupVersion) error {
	gvks := make(map[schema.GroupVersionKind]struct{})
	for _, gv := range gvs {
		types := scheme.KnownTypes(gv)
		for _, typ := range types {
			if reflect.PtrTo(typ).Implements(clientObjectType) {
				obj := reflect.New(typ).Interface().(client.Object)
				kinds, unversioned, err := scheme.ObjectKinds(obj)
				if err != nil {
					return err
				}
				if unversioned {
					continue
				}

				for _, kind := range kinds {
					gvks[kind] = struct{}{}
				}
			}
		}
	}

	return waitUntilGVKsDiscoverable(ctx, c, gvks)
}

func WaitUntilAPIServicesAvailable(ctx context.Context, c client.Client, services ...*apiregistrationv1.APIService) error {
	apiServices := make([]*apiregistrationv1.APIService, 0, len(services))
	for _, service := range services {
		apiServices = append(apiServices, service.DeepCopy())
	}

	if err := wait.PollUntilContextCancel(ctx, 50*time.Millisecond, true, func(ctx context.Context) (done bool, err error) {
		for i := len(apiServices) - 1; i >= 0; i-- {
			apiService := apiServices[i]
			key := client.ObjectKeyFromObject(apiService)
			if err := c.Get(ctx, client.ObjectKeyFromObject(apiService), apiService); err != nil {
				return false, fmt.Errorf("error getting api service %s: %w", key, err)
			}

			status := conditionutils.MustFindSliceStatus(apiService.Status.Conditions, string(apiregistrationv1.Available))
			if status == corev1.ConditionTrue {
				apiServices = append(apiServices[:i], apiServices[i+1:]...)
			}
		}
		return len(apiServices) == 0, nil
	}); err != nil {
		if len(apiServices) == 0 {
			return err
		}
		unavailableAPIServices := make([]string, 0, len(apiServices))
		for _, apiService := range apiServices {
			unavailableAPIServices = append(unavailableAPIServices, apiService.Name)
		}
		sort.Strings(unavailableAPIServices)
		return fmt.Errorf("%w, unavailable api serivces: %v", err, unavailableAPIServices)
	}
	return nil
}

func WaitUntilGroupVersionsDiscoverableWithTimeout(timeout time.Duration, c client.Client, scheme *runtime.Scheme, gvs ...schema.GroupVersion) error {
	ctx, cancel := context.WithTimeout(context.Background(), timeout)
	defer cancel()

	return WaitUntilGroupVersionsDiscoverable(ctx, c, scheme, gvs...)
}

func WaitUntilAPIServicesReady(ctx context.Context, ext *EnvironmentExtensions, c client.Client, scheme *runtime.Scheme) error {
	apiServices := ext.APIServiceInstallOptions.APIServices

	if err := WaitUntilAPIServicesAvailable(ctx, c, apiServices...); err != nil {
		return fmt.Errorf("error waiting for api services to be available: %w", err)
	}

	groupVersions := make([]schema.GroupVersion, 0, len(apiServices))
	for _, apiService := range apiServices {
		groupVersions = append(groupVersions, schema.GroupVersion{Group: apiService.Spec.Group, Version: apiService.Spec.Version})
	}
	if err := WaitUntilGroupVersionsDiscoverable(ctx, c, scheme, groupVersions...); err != nil {
		return fmt.Errorf("error waiting for group versions to be available: %w", err)
	}
	return nil
}

func WaitUntilAPIServicesReadyWithTimeout(timeout time.Duration, ext *EnvironmentExtensions, c client.Client, scheme *runtime.Scheme) error {
	ctx, cancel := context.WithTimeout(context.Background(), timeout)
	defer cancel()

	return WaitUntilAPIServicesReady(ctx, ext, c, scheme)
}
