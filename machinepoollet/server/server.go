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
	"fmt"
	"io"
	"net"
	"net/http"
	"os"
	"path/filepath"
	"strconv"
	"strings"
	"sync"
	"time"

	"github.com/gorilla/mux"
	computev1alpha1 "github.com/onmetal/onmetal-api/apis/compute/v1alpha1"
	"github.com/onmetal/onmetal-api/machinepoollet/terminal"
	http2 "github.com/onmetal/onmetal-api/machinepoollet/terminal/http"
	"k8s.io/apiserver/pkg/authentication/authenticator"
	"k8s.io/apiserver/pkg/authentication/authenticatorfactory"
	"k8s.io/apiserver/pkg/authentication/user"
	"k8s.io/apiserver/pkg/authorization/authorizer"
	"k8s.io/apiserver/pkg/authorization/authorizerfactory"
	"k8s.io/apiserver/pkg/server/dynamiccertificates"
	"k8s.io/apiserver/pkg/server/options"
	"k8s.io/client-go/kubernetes"
	"k8s.io/client-go/rest"
	"k8s.io/client-go/tools/remotecommand"
	certutil "k8s.io/client-go/util/cert"
	"k8s.io/client-go/util/keyutil"
	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/healthz"
	logf "sigs.k8s.io/controller-runtime/pkg/log"
	"sigs.k8s.io/controller-runtime/pkg/runtime/inject"
)

var log = logf.Log.WithName("machinepoollet").WithName("server")

// MachineExec is an interface a provider needs to implement in order to provide exec-functionality.
type MachineExec interface {
	// Exec execs into the target machine.
	Exec(ctx context.Context, namespace, name string, in io.Reader, out, err io.WriteCloser, resize <-chan remotecommand.TerminalSize) error
}

type Options struct {
	// MachineExec allows exec-ing onto a Machine.
	MachineExec MachineExec

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

	auth Auth

	machineExec MachineExec

	address string

	host                    string
	port                    int
	certDir                 string
	clientCACertificateFile string

	streamCreationTimeout time.Duration
	streamIdleTimeout     time.Duration
	shutdownTimeout       time.Duration
}

func setOptionsDefaults(opts *Options) {
	if opts.CertDir == "" {
		opts.CertDir = filepath.Join(os.TempDir(), "onmetal-api-machinepool-server", "serving-certs")
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
	if opts.MachineExec == nil {
		return nil, fmt.Errorf("must specify MachineExec")
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
		auth:                    auth,
		machineExec:             opts.MachineExec,
		address:                 opts.Address,
		certDir:                 opts.CertDir,
		clientCACertificateFile: caCertificateFile,
		streamCreationTimeout:   opts.StreamCreationTimeout,
		streamIdleTimeout:       opts.StreamIdleTimeout,
		shutdownTimeout:         opts.ShutdownTimeout,
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

// InjectFunc implements inject.Injector.
func (s *Server) InjectFunc(f inject.Func) error {
	if err := f(s.machineExec); err != nil {
		return err
	}
	return nil
}

func (s *Server) router() http.Handler {
	m := mux.NewRouter()

	m.Use(s.logMiddleware)

	computeRouter := m.PathPrefix("/apis/compute.api.onmetal.de").Subrouter()
	if s.auth != nil {
		computeRouter.Use(s.authMiddleware)
	}
	s.registerComputeRoutes(computeRouter)

	m.Methods(http.MethodGet).Path("/healthz").Handler(healthz.CheckHandler{Checker: healthz.Ping})
	m.Methods(http.MethodGet).Path("/readyz").Handler(healthz.CheckHandler{Checker: healthz.Ping})

	return m
}

func (s *Server) logMiddleware(handler http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, req *http.Request) {
		req = req.WithContext(ctrl.LoggerInto(req.Context(), log))
		handler.ServeHTTP(w, req)
	})
}

func (s *Server) authMiddleware(handler http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, req *http.Request) {
		ctx := req.Context()

		info, ok, err := s.auth.AuthenticateRequest(req)
		if err != nil || !ok {
			if err != nil {
				log.Error(err, "Authorization error")
			}
			http.Error(w, "Unauthorized", http.StatusUnauthorized)
			return
		}

		log := log.WithValues(
			"user-name", info.User.GetName(),
			"user-id", info.User.GetUID(),
		)
		ctx = ctrl.LoggerInto(ctx, log)
		req = req.WithContext(ctx)

		decision, _, err := s.auth.Authorize(ctx, s.auth.GetRequestAttributes(info.User, req))
		if err != nil {
			log.Error(err, "Authorization error")
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}

		if decision != authorizer.DecisionAllow {
			http.Error(w, http.StatusText(http.StatusForbidden), http.StatusForbidden)
			return
		}

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

type machineExecTerminal struct {
	machineExec MachineExec

	ctx       context.Context
	namespace string
	name      string
}

func (m *machineExecTerminal) Run(in io.Reader, out, err io.WriteCloser, resize <-chan remotecommand.TerminalSize) error {
	return m.machineExec.Exec(m.ctx, m.namespace, m.name, in, out, err, resize)
}

func NewMachineExecTerminal(ctx context.Context, exec MachineExec, namespace, name string) terminal.Terminal {
	return &machineExecTerminal{exec, ctx, namespace, name}
}

func (s *Server) registerComputeRoutes(r *mux.Router) {
	r.Methods(http.MethodGet, http.MethodPost).Path("/namespaces/{namespace}/machines/{name}/exec").HandlerFunc(func(w http.ResponseWriter, req *http.Request) {
		ctx := req.Context()
		vars := mux.Vars(req)
		namespace := vars["namespace"]
		name := vars["name"]

		supportedStreamProtocols := strings.Split(req.Header.Get("X-Stream-Protocol-Version"), ",")

		term := NewMachineExecTerminal(ctx, s.machineExec, namespace, name)
		streamOpts := &http2.Options{
			Stdin:  true,
			Stdout: true,
			TTY:    true,
		}
		http2.Serve(w, req, term, streamOpts, s.streamIdleTimeout, s.streamCreationTimeout, supportedStreamProtocols)
	})
}

func (s *Server) Port() int {
	s.mu.RLock()
	defer s.mu.RUnlock()
	return s.port
}

func (s *Server) Host() string {
	s.mu.RLock()
	defer s.mu.RUnlock()
	return s.host
}

func (s *Server) Started() bool {
	s.mu.RLock()
	defer s.mu.RUnlock()
	return s.started
}

func (s *Server) tlsConfig() (*tls.Config, error) {
	cfg := &tls.Config{
		MinVersion:               tls.VersionTLS12,
		PreferServerCipherSuites: true,
		ClientAuth:               tls.RequestClientCert,
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

	ln, host, port, err := s.listen()
	if err != nil {
		s.mu.Unlock()
		return err
	}

	s.host = host
	s.port = port

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

		log.Info("Start serving", "Host", s.host, "Port", port)
		srvErr = srv.ServeTLS(
			ln,
			getCertPath(s.certDir),
			getKeyPath(s.certDir),
		)
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
			log.Error(shutdownErr, "Error shutting down server")
		}

		return srvErr
	}
}

func (s *Server) listen() (ln net.Listener, host string, port int, err error) {
	ln, err = net.Listen("tcp", s.address)
	if err != nil {
		return nil, "", 0, err
	}

	var portString string
	host, portString, err = net.SplitHostPort(ln.Addr().String())
	if err != nil {
		_ = ln.Close()
		return nil, "", 0, err
	}

	port, err = strconv.Atoi(portString)
	if err != nil {
		_ = ln.Close()
		return nil, "", 0, err
	}
	return ln, host, port, nil
}
