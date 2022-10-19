/*
 * Copyright (c) 2021 by the OnMetal authors.
 *
 * Licensed under the Apache License, Version 2.0 (the "License");
 * you may not use this file except in compliance with the License.
 * You may obtain a copy of the License at
 *
 *     http://www.apache.org/licenses/LICENSE-2.0
 *
 * Unless required by applicable law or agreed to in writing, software
 * distributed under the License is distributed on an "AS IS" BASIS,
 * WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
 * See the License for the specific language governing permissions and
 * limitations under the License.
 */
// Code generated by client-gen. DO NOT EDIT.

package internalversion

import (
	"net/http"

	"github.com/onmetal/onmetal-api/generated/clientset/internalversion/scheme"
	rest "k8s.io/client-go/rest"
)

type NetworkingInterface interface {
	RESTClient() rest.Interface
	AliasPrefixesGetter
	AliasPrefixRoutingsGetter
	LoadBalancersGetter
	LoadBalancerRoutingsGetter
	NetworksGetter
	NetworkInterfacesGetter
	VirtualIPsGetter
}

// NetworkingClient is used to interact with features provided by the networking.api.onmetal.de group.
type NetworkingClient struct {
	restClient rest.Interface
}

func (c *NetworkingClient) AliasPrefixes(namespace string) AliasPrefixInterface {
	return newAliasPrefixes(c, namespace)
}

func (c *NetworkingClient) AliasPrefixRoutings(namespace string) AliasPrefixRoutingInterface {
	return newAliasPrefixRoutings(c, namespace)
}

func (c *NetworkingClient) LoadBalancers(namespace string) LoadBalancerInterface {
	return newLoadBalancers(c, namespace)
}

func (c *NetworkingClient) LoadBalancerRoutings(namespace string) LoadBalancerRoutingInterface {
	return newLoadBalancerRoutings(c, namespace)
}

func (c *NetworkingClient) Networks(namespace string) NetworkInterface {
	return newNetworks(c, namespace)
}

func (c *NetworkingClient) NetworkInterfaces(namespace string) NetworkInterfaceInterface {
	return newNetworkInterfaces(c, namespace)
}

func (c *NetworkingClient) VirtualIPs(namespace string) VirtualIPInterface {
	return newVirtualIPs(c, namespace)
}

// NewForConfig creates a new NetworkingClient for the given config.
// NewForConfig is equivalent to NewForConfigAndClient(c, httpClient),
// where httpClient was generated with rest.HTTPClientFor(c).
func NewForConfig(c *rest.Config) (*NetworkingClient, error) {
	config := *c
	if err := setConfigDefaults(&config); err != nil {
		return nil, err
	}
	httpClient, err := rest.HTTPClientFor(&config)
	if err != nil {
		return nil, err
	}
	return NewForConfigAndClient(&config, httpClient)
}

// NewForConfigAndClient creates a new NetworkingClient for the given config and http client.
// Note the http client provided takes precedence over the configured transport values.
func NewForConfigAndClient(c *rest.Config, h *http.Client) (*NetworkingClient, error) {
	config := *c
	if err := setConfigDefaults(&config); err != nil {
		return nil, err
	}
	client, err := rest.RESTClientForConfigAndClient(&config, h)
	if err != nil {
		return nil, err
	}
	return &NetworkingClient{client}, nil
}

// NewForConfigOrDie creates a new NetworkingClient for the given config and
// panics if there is an error in the config.
func NewForConfigOrDie(c *rest.Config) *NetworkingClient {
	client, err := NewForConfig(c)
	if err != nil {
		panic(err)
	}
	return client
}

// New creates a new NetworkingClient for the given RESTClient.
func New(c rest.Interface) *NetworkingClient {
	return &NetworkingClient{c}
}

func setConfigDefaults(config *rest.Config) error {
	config.APIPath = "/apis"
	if config.UserAgent == "" {
		config.UserAgent = rest.DefaultKubernetesUserAgent()
	}
	if config.GroupVersion == nil || config.GroupVersion.Group != scheme.Scheme.PrioritizedVersionsForGroup("networking.api.onmetal.de")[0].Group {
		gv := scheme.Scheme.PrioritizedVersionsForGroup("networking.api.onmetal.de")[0]
		config.GroupVersion = &gv
	}
	config.NegotiatedSerializer = scheme.Codecs

	if config.QPS == 0 {
		config.QPS = 5
	}
	if config.Burst == 0 {
		config.Burst = 10
	}

	return nil
}

// RESTClient returns a RESTClient that is used to communicate
// with API server by this client implementation.
func (c *NetworkingClient) RESTClient() rest.Interface {
	if c == nil {
		return nil
	}
	return c.restClient
}
