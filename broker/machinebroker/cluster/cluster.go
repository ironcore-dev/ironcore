// SPDX-FileCopyrightText: 2023 SAP SE or an SAP affiliate company and IronCore contributors
// SPDX-License-Identifier: Apache-2.0

package cluster

import (
	"context"
	"fmt"

	computev1alpha1 "github.com/ironcore-dev/ironcore/api/compute/v1alpha1"
	ipamv1alpha1 "github.com/ironcore-dev/ironcore/api/ipam/v1alpha1"
	networkingv1alpha1 "github.com/ironcore-dev/ironcore/api/networking/v1alpha1"
	storagev1alpha1 "github.com/ironcore-dev/ironcore/api/storage/v1alpha1"
	"github.com/ironcore-dev/ironcore/broker/common/idgen"

	v1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/runtime"
	utilruntime "k8s.io/apimachinery/pkg/util/runtime"
	kubernetes "k8s.io/client-go/kubernetes/scheme"
	"k8s.io/client-go/rest"
	"sigs.k8s.io/controller-runtime/pkg/cache"
	"sigs.k8s.io/controller-runtime/pkg/client"
)

var scheme = runtime.NewScheme()

func init() {
	utilruntime.Must(kubernetes.AddToScheme(scheme))
	utilruntime.Must(computev1alpha1.AddToScheme(scheme))
	utilruntime.Must(networkingv1alpha1.AddToScheme(scheme))
	utilruntime.Must(storagev1alpha1.AddToScheme(scheme))
	utilruntime.Must(ipamv1alpha1.AddToScheme(scheme))
}

type Cluster interface {
	Namespace() string
	Config() *rest.Config
	Client() client.Client
	Scheme() *runtime.Scheme
	IDGen() idgen.IDGen
	MachinePoolName() string
	MachinePoolSelector() map[string]string
}

type cluster struct {
	namespace           string
	config              *rest.Config
	client              client.Client
	scheme              *runtime.Scheme
	idGen               idgen.IDGen
	machinePoolName     string
	machinePoolSelector map[string]string
}

type Options struct {
	IDGen               idgen.IDGen
	MachinePoolName     string
	MachinePoolSelector map[string]string
}

func setOptionsDefaults(o *Options) {
	if o.IDGen == nil {
		o.IDGen = idgen.Default
	}
}

func New(ctx context.Context, cfg *rest.Config, namespace string, opts Options) (Cluster, error) {
	setOptionsDefaults(&opts)

	readCache, err := cache.New(cfg, cache.Options{
		Scheme: scheme,
		DefaultNamespaces: map[string]cache.Config{
			namespace: {},
		},
	})
	if err != nil {
		return nil, fmt.Errorf("error creating cache: %w", err)
	}

	go func() {
		if err := readCache.Start(ctx); err != nil {
			fmt.Printf("Error starting cache: %v\n", err)
		}
	}()
	if !readCache.WaitForCacheSync(ctx) {
		return nil, fmt.Errorf("failed to sync cache")
	}

	c, err := client.New(cfg, client.Options{
		Scheme: scheme,
		Cache: &client.CacheOptions{
			Reader:     readCache,
			DisableFor: []client.Object{&v1.Event{}},
		},
	})
	if err != nil {
		return nil, fmt.Errorf("error creating client: %w", err)
	}

	return &cluster{
		namespace:           namespace,
		config:              cfg,
		client:              c,
		scheme:              scheme,
		idGen:               opts.IDGen,
		machinePoolName:     opts.MachinePoolName,
		machinePoolSelector: opts.MachinePoolSelector,
	}, nil
}

func (c *cluster) Namespace() string {
	return c.namespace
}

func (c *cluster) Config() *rest.Config {
	return c.config
}

func (c *cluster) Client() client.Client {
	return c.client
}

func (c *cluster) Scheme() *runtime.Scheme {
	return c.scheme
}

func (c *cluster) IDGen() idgen.IDGen {
	return c.idGen
}

func (c *cluster) MachinePoolName() string {
	return c.machinePoolName
}

func (c *cluster) MachinePoolSelector() map[string]string {
	return c.machinePoolSelector
}
