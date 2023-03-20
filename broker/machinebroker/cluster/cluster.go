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

package cluster

import (
	"fmt"

	computev1alpha1 "github.com/onmetal/onmetal-api/api/compute/v1alpha1"
	ipamv1alpha1 "github.com/onmetal/onmetal-api/api/ipam/v1alpha1"
	networkingv1alpha1 "github.com/onmetal/onmetal-api/api/networking/v1alpha1"
	storagev1alpha1 "github.com/onmetal/onmetal-api/api/storage/v1alpha1"
	"github.com/onmetal/onmetal-api/broker/common/idgen"
	"k8s.io/apimachinery/pkg/runtime"
	utilruntime "k8s.io/apimachinery/pkg/util/runtime"
	kubernetes "k8s.io/client-go/kubernetes/scheme"
	"k8s.io/client-go/rest"
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

func New(cfg *rest.Config, namespace string, opts Options) (Cluster, error) {
	setOptionsDefaults(&opts)

	c, err := client.New(cfg, client.Options{
		Scheme: scheme,
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
