// Copyright 2022 OnMetal authors
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

package server

import (
	"context"
	"crypto/tls"
	"crypto/x509"
	"errors"
	"fmt"
	"net"
	"net/http"
	"os"
	"path/filepath"
	"strings"
	"sync"
	"time"

	"github.com/go-chi/chi/v5"
	"github.com/go-logr/logr"
	computev1alpha1 "github.com/onmetal/onmetal-api/api/compute/v1alpha1"
	orimachine "github.com/onmetal/onmetal-api/ori/apis/machine"
	utilshttp "github.com/onmetal/onmetal-api/utils/http"
	"k8s.io/apiserver/pkg/authentication/authenticator"
	"k8s.io/apiserver/pkg/authentication/authenticatorfactory"
	"k8s.io/apiserver/pkg/authentication/user"
	"k8s.io/apiserver/pkg/authorization/authorizer"
	"k8s.io/apiserver/pkg/authorization/authorizerfactory"
	"k8s.io/apiserver/pkg/server/dynamiccertificates"
	"k8s.io/apiserver/pkg/server/options"
	"k8s.io/client-go/kubernetes"
	"k8s.io/client-go/rest"
	certutil "k8s.io/client-go/util/cert"
	"k8s.io/client-go/util/keyutil"
	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/healthz"
	logf "sigs.k8s.io/controller-runtime/pkg/log"
)

var log = logf.Log.WithName("machinepoollet").WithName("server")

type Options struct {
	// MachineRuntime is the ori-machine runtime service.
	MachineRuntime orimachine.RuntimeService

	// Log is the logger to use in the server.
	// If unset, a package-global router will be used.
	Log logr.Logger

	// HostnameOverride is an optional hostname override to supply for self-signed certificate generation.
	HostnameOverride string

	// Address is the address to listen on.
	// Leave empty to auto-determine host / ephemeral port.
	Address string

	// CertDir is the directory that contains the server key and certificate.
	// If not set, the server would look up the key and certificate in
	// {TempDir}/onmetal-api-machinepool-server/serving-certs. The key and certificate
	// must be named tls.key and tls.crt, respectively.
	// If the files don't exist, a self-signed certificate is generated.
	CertDir string

	// Auth are options for authentication.
	Auth AuthOptions
	// DisableAuth will turn off authN/Z.
	DisableAuth bool

	StreamCreationTimeout time.Duration
	StreamIdleTimeout     time.Duration
	ShutdownTimeout       time.Duration
	CacheTTL              time.Duration
}

type AuthOptions struct {
	MachinePoolName string
	Authentication  AuthenticationOptions
	Authorization   AuthorizationOptions
}

type AuthenticationOptions struct {
	ClientCAFile string
}

type AuthorizationOptions struct {
	Anonymous bool
}

type Server struct {
	mu sync.RWMutex

	started bool

	log logr.Logger

	auth Auth

	machineRuntime orimachine.RuntimeService

	cacheTTL time.Duration

	address string

	certDir                 string
	clientCACertificateFile string

	streamCreationTimeout time.Duration
	streamIdleTimeout     time.Duration
	shutdownTimeout       time.Duration
}

func setOptionsDefaults(opts *Options) {
	if opts.Log.GetSink() == nil {
		opts.Log = log
	}

	if opts.CertDir == "" {
		opts.CertDir = filepath.Join(os.TempDir(), "machinepoollet-server", "serving-certs")
	}

	if opts.StreamCreationTimeout == 0 {
		opts.StreamCreationTimeout = 30 * time.Second
	}
	if opts.StreamIdleTimeout == 0 {
		opts.StreamIdleTimeout = 30 * time.Second
	}
	if opts.ShutdownTimeout == 0 {
		opts.ShutdownTimeout = 3 * time.Second
	}
	if opts.CacheTTL == 0 {
		opts.CacheTTL = 1 * time.Minute
	}
}

type Auth interface {
	authenticator.Request
	authorizer.RequestAttributesGetter
	authorizer.Authorizer
}

type authWrapper struct {
	authenticator.Request
	authorizer.RequestAttributesGetter
	authorizer.Authorizer
}

type MachinePoolRequestAttr struct {
	MachinePoolName string
}

func (m MachinePoolRequestAttr) GetRequestAttributes(u user.Info, req *http.Request) authorizer.Attributes {
	return authorizer.AttributesRecord{
		User:            u,
		Verb:            getAPIVerb(req.Method),
		Namespace:       "",
		APIGroup:        computev1alpha1.SchemeGroupVersion.Group,
		APIVersion:      computev1alpha1.SchemeGroupVersion.Version,
		Resource:        "machinepools",
		Subresource:     "proxy",
		Name:            m.MachinePoolName,
		ResourceRequest: true,
		Path:            req.URL.Path,
	}
}

func NewAuth(cfg *rest.Config, opts AuthOptions) (Auth, error) {
	c, err := kubernetes.NewForConfig(cfg)
	if err != nil {
		return nil, err
	}

	var clientCAProvider dynamiccertificates.CAContentProvider
	if clientCAFile := opts.Authentication.ClientCAFile; clientCAFile != "" {
		clientCAProvider, err = dynamiccertificates.NewDynamicCAContentFromFile("ca-cert-bundle", clientCAFile)
		if err != nil {
			return nil, err
		}
	}

	authN, _, err := authenticatorfactory.DelegatingAuthenticatorConfig{
		CacheTTL:                           2 * time.Minute,
		TokenAccessReviewClient:            c.AuthenticationV1(),
		WebhookRetryBackoff:                options.DefaultAuthWebhookRetryBackoff(),
		ClientCertificateCAContentProvider: clientCAProvider,
	}.New()
	if err != nil {
		return nil, err
	}

	authZ, err := authorizerfactory.DelegatingAuthorizerConfig{
		AllowCacheTTL:             5 * time.Minute,
		DenyCacheTTL:              30 * time.Second,
		SubjectAccessReviewClient: c.AuthorizationV1(),
		WebhookRetryBackoff:       options.DefaultAuthWebhookRetryBackoff(),
	}.New()
	if err != nil {
		return nil, err
	}

	return authWrapper{
		Request:                 authN,
		RequestAttributesGetter: MachinePoolRequestAttr{opts.MachinePoolName},
		Authorizer:              authZ,
	}, nil
}

// GetHostname returns OS's hostname if 'hostnameOverride' is empty; otherwise, return 'hostnameOverride'.
// Copied from Kubernetes' nodeutil to avoid a dependency to kubernetes/kubernetes.
func GetHostname(hostnameOverride string) (string, error) {
	hostName := hostnameOverride
	if len(hostName) == 0 {
		nodeName, err := os.Hostname()
		if err != nil {
			return "", fmt.Errorf("couldn't determine hostname: %v", err)
		}
		hostName = nodeName
	}

	// Trim whitespaces first to avoid getting an empty hostname
	// For linux, the hostname is read from file /proc/sys/kernel/hostname directly
	hostName = strings.TrimSpace(hostName)
	if len(hostName) == 0 {
		return "", fmt.Errorf("empty hostname is invalid")
	}
	return strings.ToLower(hostName), nil
}

func New(cfg *rest.Config, opts Options) (*Server, error) {
	if cfg == nil {
		return nil, fmt.Errorf("must specify config")
	}
	if opts.MachineRuntime == nil {
		return nil, fmt.Errorf("must specify MachineRuntime")
	}

	setOptionsDefaults(&opts)

	auth, caCertificateFile, err := initializeAuth(cfg, opts)
	if err != nil {
		return nil, err
	}

	if err := initializeTLS(opts); err != nil {
		return nil, err
	}

	return &Server{
		log:                     opts.Log,
		auth:                    auth,
		machineRuntime:          opts.MachineRuntime,
		address:                 opts.Address,
		certDir:                 opts.CertDir,
		clientCACertificateFile: caCertificateFile,
		streamCreationTimeout:   opts.StreamCreationTimeout,
		streamIdleTimeout:       opts.StreamIdleTimeout,
		shutdownTimeout:         opts.ShutdownTimeout,
		cacheTTL:                opts.CacheTTL,
	}, nil
}

func initializeAuth(cfg *rest.Config, opts Options) (auth Auth, caCertificateFile string, err error) {
	if !opts.DisableAuth {
		auth, err = NewAuth(cfg, opts.Auth)
		if err != nil {
			return nil, "", err
		}

		caCertificateFile = opts.Auth.Authentication.ClientCAFile
	}
	return auth, caCertificateFile, nil
}

func initializeTLS(opts Options) error {
	ok, err := certutil.CanReadCertAndKey(
		getCertPath(opts.CertDir),
		getKeyPath(opts.CertDir),
	)
	if err != nil || ok {
		return err
	}
	hostName, err := GetHostname(opts.HostnameOverride)
	if err != nil {
		return err
	}

	cert, key, err := certutil.GenerateSelfSignedCertKey(hostName, nil, nil)
	if err != nil {
		return fmt.Errorf("unable to generate self signed cert: %w", err)
	}

	if err := certutil.WriteCert(getCertPath(opts.CertDir), cert); err != nil {
		return err
	}

	if err := keyutil.WriteKey(getKeyPath(opts.CertDir), key); err != nil {
		return err
	}
	return nil
}

func getCertPath(certDir string) string {
	return filepath.Join(certDir, "tls.crt")
}

func getKeyPath(certDir string) string {
	return filepath.Join(certDir, "tls.key")
}

func (s *Server) router() http.Handler {
	r := chi.NewRouter()

	r.Use(utilshttp.InjectLogger(s.log))
	r.Use(utilshttp.LogRequest)

	r.Route("/apis/compute.api.onmetal.de", func(r chi.Router) {
		if s.auth != nil {
			r.Use(s.authMiddleware)
		}
		s.registerComputeRoutes(r)
	})

	r.Get("/healthz", healthz.CheckHandler{Checker: healthz.Ping}.ServeHTTP)
	r.Get("/readyz", healthz.CheckHandler{Checker: healthz.Ping}.ServeHTTP)

	return r
}

func (s *Server) authMiddleware(handler http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, req *http.Request) {
		ctx := req.Context()
		log := ctrl.LoggerFrom(ctx)

		log.V(2).Info("Authenticating")
		info, ok, err := s.auth.AuthenticateRequest(req)
		if err != nil {
			log.Error(err, "Authentication error")
			http.Error(w, "Unauthorized", http.StatusUnauthorized)
			return
		}
		if !ok {
			log.V(1).Info("Unauthorized")
			return
		}

		log = log.WithValues(
			"user-name", info.User.GetName(),
			"user-id", info.User.GetUID(),
			"user-groups", info.User.GetGroups(),
		)
		log.V(2).Info("Authenticated")

		ctx = ctrl.LoggerInto(ctx, log)
		req = req.WithContext(ctx)

		attributes := s.auth.GetRequestAttributes(info.User, req)
		log.V(2).Info("Authorizing", "Attributes", attributes)
		decision, _, err := s.auth.Authorize(ctx, attributes)
		if err != nil {
			log.Error(err, "Authorization error")
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}

		if decision != authorizer.DecisionAllow {
			log.V(1).Info("Authorization denied", "Decision", decision)
			http.Error(w, http.StatusText(http.StatusForbidden), http.StatusForbidden)
			return
		}

		log.V(2).Info("Authorized")
		handler.ServeHTTP(w, req)
	})
}

func getAPIVerb(method string) string {
	switch method {
	case http.MethodPost:
		return "create"
	case http.MethodGet:
		return "get"
	case http.MethodPut:
		return "update"
	case http.MethodPatch:
		return "patch"
	case http.MethodDelete:
		return "delete"
	default:
		return ""
	}
}

func (s *Server) registerComputeRoutes(r chi.Router) {
	for _, method := range []string{http.MethodGet, http.MethodPost} {
		r.MethodFunc(method, "/namespaces/{namespace}/machines/{name}/exec", func(w http.ResponseWriter, req *http.Request) {
			namespace := chi.URLParam(req, "namespace")
			name := chi.URLParam(req, "name")
			s.serveExec(w, req, namespace, name)
		})
	}
}

func (s *Server) tlsConfig() (*tls.Config, error) {
	cfg := &tls.Config{
		MinVersion: tls.VersionTLS12,
		ClientAuth: tls.RequestClientCert,
	}

	if s.clientCACertificateFile != "" {
		pem, err := os.ReadFile(s.clientCACertificateFile)
		if err != nil {
			return nil, fmt.Errorf("error reading ca certificate %s: %w", s.clientCACertificateFile, err)
		}

		cfg.ClientAuth = tls.RequireAndVerifyClientCert
		if cfg.ClientCAs == nil {
			cfg.ClientCAs = x509.NewCertPool()
		}
		if !cfg.ClientCAs.AppendCertsFromPEM(pem) {
			return nil, fmt.Errorf("invalid ca certificate")
		}
	}

	return cfg, nil
}

func (s *Server) Start(ctx context.Context) error {
	s.mu.Lock()
	if s.started {
		s.mu.Unlock()
		return fmt.Errorf("server already started")
	}

	s.started = true

	tlsConfig, err := s.tlsConfig()
	if err != nil {
		s.mu.Unlock()
		return err
	}

	ln, err := net.Listen("tcp", s.address)
	if err != nil {
		s.mu.Unlock()
		return fmt.Errorf("error listening: %w", err)
	}

	var (
		srvErr  error
		srvDone = make(chan struct{})
	)

	srv := &http.Server{
		Handler:   s.router(),
		TLSConfig: tlsConfig,
	}

	go func() {
		defer close(srvDone)
		defer func() { _ = ln.Close() }()

		s.log.Info("Start serving", "Address", ln.Addr())
		if err := srv.ServeTLS(
			ln,
			getCertPath(s.certDir),
			getKeyPath(s.certDir),
		); err != nil && !errors.Is(err, http.ErrServerClosed) {
			srvErr = err
		}
	}()

	s.mu.Unlock()

	select {
	case <-srvDone:
		return srvErr
	case <-ctx.Done():
		shutdownCtx, cancel := context.WithTimeout(context.Background(), s.shutdownTimeout)
		defer cancel()
		if shutdownErr := srv.Shutdown(shutdownCtx); shutdownErr != nil {
			if srvErr == nil {
				return shutdownErr
			}
			s.log.Error(shutdownErr, "Error shutting down server")
		}

		return srvErr
	}
}
