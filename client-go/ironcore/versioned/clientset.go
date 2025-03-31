// SPDX-FileCopyrightText: 2023 SAP SE or an SAP affiliate company and IronCore contributors
// SPDX-License-Identifier: Apache-2.0

// Code generated by client-gen. DO NOT EDIT.

package versioned

import (
	fmt "fmt"
	http "net/http"

	computev1alpha1 "github.com/ironcore-dev/ironcore/client-go/ironcore/versioned/typed/compute/v1alpha1"
	corev1alpha1 "github.com/ironcore-dev/ironcore/client-go/ironcore/versioned/typed/core/v1alpha1"
	ipamv1alpha1 "github.com/ironcore-dev/ironcore/client-go/ironcore/versioned/typed/ipam/v1alpha1"
	networkingv1alpha1 "github.com/ironcore-dev/ironcore/client-go/ironcore/versioned/typed/networking/v1alpha1"
	storagev1alpha1 "github.com/ironcore-dev/ironcore/client-go/ironcore/versioned/typed/storage/v1alpha1"
	discovery "k8s.io/client-go/discovery"
	rest "k8s.io/client-go/rest"
	flowcontrol "k8s.io/client-go/util/flowcontrol"
)

type Interface interface {
	Discovery() discovery.DiscoveryInterface
	ComputeV1alpha1() computev1alpha1.ComputeV1alpha1Interface
	CoreV1alpha1() corev1alpha1.CoreV1alpha1Interface
	IpamV1alpha1() ipamv1alpha1.IpamV1alpha1Interface
	NetworkingV1alpha1() networkingv1alpha1.NetworkingV1alpha1Interface
	StorageV1alpha1() storagev1alpha1.StorageV1alpha1Interface
}

// Clientset contains the clients for groups.
type Clientset struct {
	*discovery.DiscoveryClient
	computeV1alpha1    *computev1alpha1.ComputeV1alpha1Client
	coreV1alpha1       *corev1alpha1.CoreV1alpha1Client
	ipamV1alpha1       *ipamv1alpha1.IpamV1alpha1Client
	networkingV1alpha1 *networkingv1alpha1.NetworkingV1alpha1Client
	storageV1alpha1    *storagev1alpha1.StorageV1alpha1Client
}

// ComputeV1alpha1 retrieves the ComputeV1alpha1Client
func (c *Clientset) ComputeV1alpha1() computev1alpha1.ComputeV1alpha1Interface {
	return c.computeV1alpha1
}

// CoreV1alpha1 retrieves the CoreV1alpha1Client
func (c *Clientset) CoreV1alpha1() corev1alpha1.CoreV1alpha1Interface {
	return c.coreV1alpha1
}

// IpamV1alpha1 retrieves the IpamV1alpha1Client
func (c *Clientset) IpamV1alpha1() ipamv1alpha1.IpamV1alpha1Interface {
	return c.ipamV1alpha1
}

// NetworkingV1alpha1 retrieves the NetworkingV1alpha1Client
func (c *Clientset) NetworkingV1alpha1() networkingv1alpha1.NetworkingV1alpha1Interface {
	return c.networkingV1alpha1
}

// StorageV1alpha1 retrieves the StorageV1alpha1Client
func (c *Clientset) StorageV1alpha1() storagev1alpha1.StorageV1alpha1Interface {
	return c.storageV1alpha1
}

// Discovery retrieves the DiscoveryClient
func (c *Clientset) Discovery() discovery.DiscoveryInterface {
	if c == nil {
		return nil
	}
	return c.DiscoveryClient
}

// NewForConfig creates a new Clientset for the given config.
// If config's RateLimiter is not set and QPS and Burst are acceptable,
// NewForConfig will generate a rate-limiter in configShallowCopy.
// NewForConfig is equivalent to NewForConfigAndClient(c, httpClient),
// where httpClient was generated with rest.HTTPClientFor(c).
func NewForConfig(c *rest.Config) (*Clientset, error) {
	configShallowCopy := *c

	if configShallowCopy.UserAgent == "" {
		configShallowCopy.UserAgent = rest.DefaultKubernetesUserAgent()
	}

	// share the transport between all clients
	httpClient, err := rest.HTTPClientFor(&configShallowCopy)
	if err != nil {
		return nil, err
	}

	return NewForConfigAndClient(&configShallowCopy, httpClient)
}

// NewForConfigAndClient creates a new Clientset for the given config and http client.
// Note the http client provided takes precedence over the configured transport values.
// If config's RateLimiter is not set and QPS and Burst are acceptable,
// NewForConfigAndClient will generate a rate-limiter in configShallowCopy.
func NewForConfigAndClient(c *rest.Config, httpClient *http.Client) (*Clientset, error) {
	configShallowCopy := *c
	if configShallowCopy.RateLimiter == nil && configShallowCopy.QPS > 0 {
		if configShallowCopy.Burst <= 0 {
			return nil, fmt.Errorf("burst is required to be greater than 0 when RateLimiter is not set and QPS is set to greater than 0")
		}
		configShallowCopy.RateLimiter = flowcontrol.NewTokenBucketRateLimiter(configShallowCopy.QPS, configShallowCopy.Burst)
	}

	var cs Clientset
	var err error
	cs.computeV1alpha1, err = computev1alpha1.NewForConfigAndClient(&configShallowCopy, httpClient)
	if err != nil {
		return nil, err
	}
	cs.coreV1alpha1, err = corev1alpha1.NewForConfigAndClient(&configShallowCopy, httpClient)
	if err != nil {
		return nil, err
	}
	cs.ipamV1alpha1, err = ipamv1alpha1.NewForConfigAndClient(&configShallowCopy, httpClient)
	if err != nil {
		return nil, err
	}
	cs.networkingV1alpha1, err = networkingv1alpha1.NewForConfigAndClient(&configShallowCopy, httpClient)
	if err != nil {
		return nil, err
	}
	cs.storageV1alpha1, err = storagev1alpha1.NewForConfigAndClient(&configShallowCopy, httpClient)
	if err != nil {
		return nil, err
	}

	cs.DiscoveryClient, err = discovery.NewDiscoveryClientForConfigAndClient(&configShallowCopy, httpClient)
	if err != nil {
		return nil, err
	}
	return &cs, nil
}

// NewForConfigOrDie creates a new Clientset for the given config and
// panics if there is an error in the config.
func NewForConfigOrDie(c *rest.Config) *Clientset {
	cs, err := NewForConfig(c)
	if err != nil {
		panic(err)
	}
	return cs
}

// New creates a new Clientset for the given RESTClient.
func New(c rest.Interface) *Clientset {
	var cs Clientset
	cs.computeV1alpha1 = computev1alpha1.New(c)
	cs.coreV1alpha1 = corev1alpha1.New(c)
	cs.ipamV1alpha1 = ipamv1alpha1.New(c)
	cs.networkingV1alpha1 = networkingv1alpha1.New(c)
	cs.storageV1alpha1 = storagev1alpha1.New(c)

	cs.DiscoveryClient = discovery.NewDiscoveryClient(c)
	return &cs
}
