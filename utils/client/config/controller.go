// SPDX-FileCopyrightText: 2023 SAP SE or an SAP affiliate company and IronCore contributors
// SPDX-License-Identifier: Apache-2.0

package config

import (
	"context"
	"crypto/x509"
	"fmt"
	"net/http"
	"sync"
	"time"

	"github.com/go-logr/logr"
	"github.com/ironcore-dev/ironcore/utils/certificate"
	utilrest "github.com/ironcore-dev/ironcore/utils/rest"
	certificatesv1 "k8s.io/api/certificates/v1"
	utilnet "k8s.io/apimachinery/pkg/util/net"
	"k8s.io/apiserver/pkg/server/healthz"
	"k8s.io/client-go/rest"
	"k8s.io/client-go/util/workqueue"
	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/manager"
)

type Controller interface {
	manager.Runnable
	healthz.HealthChecker
	Init(ctx context.Context, force bool) error
	TransportConfig() *rest.Config
	ClientConfig() *rest.Config
}

type NewRESTConfigRotatorFunc func(cfg, bootstrapCfg *rest.Config, opts utilrest.ConfigRotatorOptions) (utilrest.ConfigRotator, error)

type ControllerOptions struct {
	Name                 string
	SignerName           string
	Template             *x509.CertificateRequest
	GetUsages            func(privateKey any) []certificatesv1.KeyUsage
	RequestedDuration    *time.Duration
	LogConstructor       func() logr.Logger
	DialFunc             utilnet.DialFunc
	ForceInitial         bool
	NewRESTConfigRotator NewRESTConfigRotatorFunc
}

func setControllerOptionsDefaults(o *ControllerOptions) {
	if o.LogConstructor == nil {
		log := ctrl.Log.WithName("client").WithName("config").WithValues("controller", o.Name)
		o.LogConstructor = func() logr.Logger {
			return log
		}
	}
	if o.NewRESTConfigRotator == nil {
		o.NewRESTConfigRotator = utilrest.NewConfigRotator
	}
}

type controller struct {
	startedMu sync.Mutex
	started   bool

	name           string
	logConstructor func() logr.Logger
	store          Store

	queue workqueue.TypedRateLimitingInterface[string]

	configRotator utilrest.ConfigRotator
}

func NewController(ctx context.Context, store Store, bootstrapCfg *rest.Config, opts ControllerOptions) (Controller, error) {
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
	setControllerOptionsDefaults(&opts)

	cfg, err := store.Get(ctx)
	if IgnoreErrConfigNotFound(err) != nil {
		return nil, fmt.Errorf("error getting config from store: %w", err)
	}
	if cfg != nil {
		cfg.Dial = opts.DialFunc
	}

	configRotator, err := opts.NewRESTConfigRotator(cfg, bootstrapCfg, utilrest.ConfigRotatorOptions{
		Name:              opts.Name,
		SignerName:        opts.SignerName,
		Template:          opts.Template,
		GetUsages:         opts.GetUsages,
		RequestedDuration: opts.RequestedDuration,
		LogConstructor: func() logr.Logger {
			return opts.LogConstructor().WithName("rest").WithValues("configrotator", opts.Name)
		},
		DialFunc:     opts.DialFunc,
		ForceInitial: opts.ForceInitial,
	})
	if err != nil {
		return nil, fmt.Errorf("error creating rotator: %w", err)
	}

	return &controller{
		name:           opts.Name,
		logConstructor: opts.LogConstructor,
		store:          store,
		configRotator:  configRotator,
	}, nil
}

const workItemKey = "key"

func (c *controller) persist(ctx context.Context) error {
	clientConfig := c.configRotator.ClientConfig()
	if clientConfig == nil {
		return fmt.Errorf("no client config available")
	}

	if err := c.store.Set(ctx, clientConfig); err != nil {
		return fmt.Errorf("error persisting client config: %w", err)
	}
	return nil
}

func (c *controller) processNextWorkItem(ctx context.Context) bool {
	item, shutdown := c.queue.Get()
	if shutdown {
		return false
	}
	defer c.queue.Done(item)

	if err := c.persist(ctx); err != nil {
		c.logConstructor().Error(err, "Error persisting")
		c.queue.AddRateLimited(item)
		return true
	}

	c.queue.Forget(item)
	return true
}

func (c *controller) TransportConfig() *rest.Config {
	return c.configRotator.TransportConfig()
}

func (c *controller) ClientConfig() *rest.Config {
	return c.configRotator.ClientConfig()
}

func (c *controller) Name() string {
	return c.name
}

func (c *controller) Check(req *http.Request) error {
	return c.configRotator.Check(req)
}

func (c *controller) Init(ctx context.Context, force bool) error {
	if err := c.configRotator.Init(ctx, force); err != nil {
		return fmt.Errorf("error running config rotator: %w", err)
	}
	if err := c.persist(ctx); err != nil {
		return fmt.Errorf("error persisting: %w", err)
	}
	return nil
}

func (c *controller) Start(ctx context.Context) error {
	c.startedMu.Lock()
	if c.started {
		c.startedMu.Unlock()
		return fmt.Errorf("controller was already started")
	}

	c.queue = workqueue.NewTypedRateLimitingQueue(workqueue.DefaultTypedControllerRateLimiter[string]())
	go func() {
		<-ctx.Done()
		c.queue.ShutDown()
	}()

	reg := c.configRotator.AddListener(certificate.RotatorListenerFunc(func() {
		c.queue.Add(workItemKey)
	}))
	defer c.configRotator.RemoveListener(reg)

	var wg sync.WaitGroup

	if err := func() error {
		defer c.startedMu.Unlock()

		wg.Add(1)
		go func() {
			defer wg.Done()
			_ = c.configRotator.Start(ctx)
		}()

		wg.Add(1)
		go func() {
			defer wg.Done()
			for c.processNextWorkItem(ctx) {
			}
		}()

		c.started = true
		return nil
	}(); err != nil {
		return err
	}

	<-ctx.Done()
	wg.Wait()
	return nil
}

func SetupControllerWithManager(mgr ctrl.Manager, c Controller) error {
	if c == nil {
		return nil
	}

	if err := mgr.Add(c); err != nil {
		return fmt.Errorf("error adding config controller to manager: %w", err)
	}
	if err := mgr.AddHealthzCheck(c.Name(), c.Check); err != nil {
		return fmt.Errorf("error adding config controller healthz check: %w", err)
	}
	return nil
}
