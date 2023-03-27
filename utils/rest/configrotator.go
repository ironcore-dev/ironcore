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

package rest

import (
	"context"
	"crypto/tls"
	"crypto/x509"
	"fmt"
	"net"
	"net/http"
	"sync"
	"sync/atomic"
	"time"

	"github.com/go-logr/logr"
	"github.com/onmetal/onmetal-api/utils/certificate"
	certificatesv1 "k8s.io/api/certificates/v1"
	utilnet "k8s.io/apimachinery/pkg/util/net"
	"k8s.io/apimachinery/pkg/util/sets"
	"k8s.io/apiserver/pkg/server/healthz"
	"k8s.io/client-go/rest"
	"k8s.io/client-go/util/workqueue"
	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/manager"
)

func loadConfigs(cfg, bootstrapCfg *rest.Config) (certCfg, clientCfg *rest.Config, initCert *tls.Certificate, err error) {
	if cfg == nil && bootstrapCfg == nil {
		return nil, nil, nil, fmt.Errorf("must specify either cfg or bootstrapCfg")
	}

	if bootstrapCfg == nil || cfg != nil && IsConfigValid(cfg) {
		initCert, _ := CertificateFromConfig(cfg)
		return cfg, rest.CopyConfig(cfg), initCert, nil
	}

	return bootstrapCfg, rest.AnonymousClientConfig(bootstrapCfg), nil, nil
}

type configRotatorListenerHandle struct {
	ConfigRotatorListener
}

type ConfigRotatorListener interface {
	Enqueue()
}

type ConfigRotatorListenerFunc func()

func (f ConfigRotatorListenerFunc) Enqueue() {
	f()
}

type ConfigRotatorListenerRegistration interface{}

type configRotator struct {
	startedMu sync.Mutex
	started   bool

	name           string
	logConstructor func() logr.Logger

	certRotator certificate.Rotator

	queue workqueue.RateLimitingInterface

	closeConns      func()
	baseConfig      *rest.Config
	transportConfig *rest.Config
	clientConfig    atomic.Pointer[rest.Config]

	listenersMu sync.RWMutex
	listeners   sets.Set[*configRotatorListenerHandle]
}

type ConfigRotatorOptions struct {
	Name                  string
	SignerName            string
	Template              *x509.CertificateRequest
	GetUsages             func(privateKey any) []certificatesv1.KeyUsage
	RequestedDuration     *time.Duration
	LogConstructor        func() logr.Logger
	DialFunc              utilnet.DialFunc
	NewCertificateRotator func(opts certificate.RotatorOptions) (certificate.Rotator, error)
	ForceInitial          bool
}

func setRotatorOptionsDefaults(o *ConfigRotatorOptions) {
	if o.LogConstructor == nil {
		o.LogConstructor = func() logr.Logger {
			return ctrl.Log.WithName(o.Name)
		}
	}
	if o.NewCertificateRotator == nil {
		o.NewCertificateRotator = certificate.NewRotator
	}
	if o.DialFunc == nil {
		dialer := &net.Dialer{Timeout: 30 * time.Second, KeepAlive: 30 * time.Second}
		o.DialFunc = dialer.DialContext
	}
}

func NewConfigRotator(cfg, bootstrapCfg *rest.Config, opts ConfigRotatorOptions) (ConfigRotator, error) {
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
	setRotatorOptionsDefaults(&opts)

	certCfg, clientCfg, initCert, err := loadConfigs(cfg, bootstrapCfg)
	if err != nil {
		return nil, err
	}

	r := &configRotator{
		name:           opts.Name,
		logConstructor: opts.LogConstructor,
		baseConfig:     clientCfg,
		listeners:      sets.New[*configRotatorListenerHandle](),
	}

	certRotator, err := opts.NewCertificateRotator(certificate.RotatorOptions{
		Name: opts.Name,
		NewClient: func(cert *tls.Certificate) (client.WithWatch, error) {
			cfg := certCfg
			if cert != nil {
				newCfg, err := ConfigWithCertificate(cfg, cert)
				if err != nil {
					return nil, fmt.Errorf("error creating config with certificate: %w", err)
				}

				cfg = newCfg
			}

			return client.NewWithWatch(cfg, client.Options{})
		},
		LogConstructor: func() logr.Logger {
			return opts.LogConstructor().WithName("certificate").WithValues("rotator", opts.Name)
		},
		SignerName:        opts.SignerName,
		Template:          opts.Template,
		GetUsages:         opts.GetUsages,
		RequestedDuration: opts.RequestedDuration,
		ForceInitial:      opts.ForceInitial,
		InitCertificate:   initCert,
	})
	if err != nil {
		return nil, fmt.Errorf("error creating certificate configRotator: %w", err)
	}

	r.certRotator = certRotator

	transportConfig, closeConns, err := DynamicCertificateConfig(clientCfg, certRotator.Certificate, opts.DialFunc)
	if err != nil {
		return nil, fmt.Errorf("error creating dynamic certificate configuration: %w", err)
	}

	r.transportConfig = transportConfig
	r.closeConns = closeConns
	return r, nil
}

func (r *configRotator) AddListener(listener ConfigRotatorListener) ConfigRotatorListenerRegistration {
	r.listenersMu.Lock()
	defer r.listenersMu.Unlock()

	handle := &configRotatorListenerHandle{listener}
	r.listeners.Insert(handle)
	return handle
}

func (r *configRotator) RemoveListener(reg ConfigRotatorListenerRegistration) {
	r.listenersMu.Lock()
	defer r.listenersMu.Unlock()

	handle, ok := reg.(*configRotatorListenerHandle)
	if !ok {
		return
	}
	r.listeners.Delete(handle)
}

func (r *configRotator) getListeners() []ConfigRotatorListener {
	r.listenersMu.RLock()
	defer r.listenersMu.RUnlock()

	res := make([]ConfigRotatorListener, 0, r.listeners.Len())
	for handle := range r.listeners {
		res = append(res, handle.ConfigRotatorListener)
	}
	return res
}

func (r *configRotator) update(_ context.Context) error {
	r.logConstructor().Info("Config updated, closing connections")
	r.closeConns()

	cert := r.certRotator.Certificate()
	if cert == nil {
		return fmt.Errorf("certificate is not available")
	}

	clientCfg, err := ConfigWithCertificate(r.baseConfig, cert)
	if err != nil {
		return fmt.Errorf("error creating new client config: %w", err)
	}

	r.logConstructor().Info("Updating client certificate")
	r.clientConfig.Store(clientCfg)

	listeners := r.getListeners()
	for _, listener := range listeners {
		listener.Enqueue()
	}
	return nil
}

func (r *configRotator) processNextWorkItem(ctx context.Context) bool {
	item, shutdown := r.queue.Get()
	if shutdown {
		return false
	}
	defer r.queue.Done(item)

	if err := r.update(ctx); err != nil {
		r.logConstructor().Error(err, "Error updating")
		r.queue.AddRateLimited(item)
		return true
	}

	r.queue.Forget(item)
	return true
}

func (r *configRotator) Init(ctx context.Context, force bool) error {
	if err := r.certRotator.Init(ctx, force); err != nil {
		return fmt.Errorf("error initializing certificate rotator: %w", err)
	}
	if err := r.update(ctx); err != nil {
		return fmt.Errorf("error updating: %w", err)
	}
	return nil
}

func (r *configRotator) ClientConfig() *rest.Config {
	return r.clientConfig.Load()
}

func (r *configRotator) TransportConfig() *rest.Config {
	return r.transportConfig
}

const workItemKey = "key"

func (r *configRotator) Start(ctx context.Context) error {
	r.startedMu.Lock()
	if r.started {
		r.startedMu.Unlock()
		return fmt.Errorf("configRotator already started")
	}

	r.queue = workqueue.NewRateLimitingQueue(workqueue.DefaultControllerRateLimiter())
	go func() {
		<-ctx.Done()
		r.queue.ShutDown()
	}()

	reg := r.certRotator.AddListener(certificate.RotatorListenerFunc(func() {
		r.queue.Add(workItemKey)
	}))
	defer r.certRotator.RemoveListener(reg)

	var wg sync.WaitGroup

	if err := func() error {
		defer r.startedMu.Unlock()

		wg.Add(1)
		go func() {
			defer wg.Done()
			_ = r.certRotator.Start(ctx)
		}()

		wg.Add(1)
		go func() {
			defer wg.Done()
			for r.processNextWorkItem(ctx) {
			}
		}()

		r.started = true
		return nil
	}(); err != nil {
		return err
	}

	<-ctx.Done()
	wg.Wait()
	return nil
}

func (r *configRotator) Name() string {
	return r.name
}

func (r *configRotator) Check(req *http.Request) error {
	return r.certRotator.Check(req)
}

type ConfigRotator interface {
	manager.Runnable
	healthz.HealthChecker
	Init(ctx context.Context, force bool) error
	ClientConfig() *rest.Config
	TransportConfig() *rest.Config
	AddListener(listener ConfigRotatorListener) ConfigRotatorListenerRegistration
	RemoveListener(reg ConfigRotatorListenerRegistration)
}

func RequestConfig(
	ctx context.Context,
	certCfg *rest.Config,
	signerName string,
	template *x509.CertificateRequest,
	getUsages func(privateKey any) []certificatesv1.KeyUsage,
	requestedDuration *time.Duration,
) (*rest.Config, error) {
	if certCfg == nil {
		return nil, fmt.Errorf("must specify certCfg")
	}
	if signerName == "" {
		return nil, fmt.Errorf("must specify signerName")
	}
	if template == nil {
		return nil, fmt.Errorf("must specify template")
	}
	if getUsages == nil {
		return nil, fmt.Errorf("must specify getUsages")
	}

	c, err := client.NewWithWatch(certCfg, client.Options{})
	if err != nil {
		return nil, fmt.Errorf("error creating client: %w", err)
	}

	cert, err := certificate.RequestCertificate(ctx, c, signerName, template, getUsages, requestedDuration)
	if err != nil {
		return nil, fmt.Errorf("error requesting certificate: %w", err)
	}

	cfg, err := ConfigWithCertificate(certCfg, cert)
	if err != nil {
		return nil, fmt.Errorf("error generating config with certificate: %w", err)
	}

	return cfg, nil
}

func UseOrRequestConfig(
	ctx context.Context,
	cfg *rest.Config,
	certCfg *rest.Config,
	signerName string,
	template *x509.CertificateRequest,
	getUsages func(privateKey any) []certificatesv1.KeyUsage,
	requestedDuration *time.Duration,
) (resCfg *rest.Config, newConfig bool, err error) {
	if certCfg == nil {
		return nil, false, fmt.Errorf("must specify certCfg")
	}
	if signerName == "" {
		return nil, false, fmt.Errorf("must specify signerName")
	}
	if template == nil {
		return nil, false, fmt.Errorf("must specify template")
	}
	if getUsages == nil {
		return nil, false, fmt.Errorf("must specify getUsages")
	}
	if IsConfigValid(cfg) {
		return cfg, false, nil
	}
	if certCfg == nil {
		return nil, false, fmt.Errorf("cfg is invalid and certCfg is nil")
	}

	newCfg, err := RequestConfig(ctx, certCfg, signerName, template, getUsages, requestedDuration)
	if err != nil {
		return nil, false, fmt.Errorf("error requesting config: %w", err)
	}

	return newCfg, true, nil
}
