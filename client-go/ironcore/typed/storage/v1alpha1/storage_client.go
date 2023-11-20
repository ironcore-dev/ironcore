// SPDX-FileCopyrightText: 2023 SAP SE or an SAP affiliate company and IronCore contributors
// SPDX-License-Identifier: Apache-2.0

// Code generated by client-gen. DO NOT EDIT.

package v1alpha1

import (
	"net/http"

	v1alpha1 "github.com/ironcore-dev/ironcore/api/storage/v1alpha1"
	"github.com/ironcore-dev/ironcore/client-go/ironcore/scheme"
	rest "k8s.io/client-go/rest"
)

type StorageV1alpha1Interface interface {
	RESTClient() rest.Interface
	BucketsGetter
	BucketClassesGetter
	BucketPoolsGetter
	VolumesGetter
	VolumeClassesGetter
	VolumePoolsGetter
}

// StorageV1alpha1Client is used to interact with features provided by the storage.ironcore.dev group.
type StorageV1alpha1Client struct {
	restClient rest.Interface
}

func (c *StorageV1alpha1Client) Buckets(namespace string) BucketInterface {
	return newBuckets(c, namespace)
}

func (c *StorageV1alpha1Client) BucketClasses() BucketClassInterface {
	return newBucketClasses(c)
}

func (c *StorageV1alpha1Client) BucketPools() BucketPoolInterface {
	return newBucketPools(c)
}

func (c *StorageV1alpha1Client) Volumes(namespace string) VolumeInterface {
	return newVolumes(c, namespace)
}

func (c *StorageV1alpha1Client) VolumeClasses() VolumeClassInterface {
	return newVolumeClasses(c)
}

func (c *StorageV1alpha1Client) VolumePools() VolumePoolInterface {
	return newVolumePools(c)
}

// NewForConfig creates a new StorageV1alpha1Client for the given config.
// NewForConfig is equivalent to NewForConfigAndClient(c, httpClient),
// where httpClient was generated with rest.HTTPClientFor(c).
func NewForConfig(c *rest.Config) (*StorageV1alpha1Client, error) {
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

// NewForConfigAndClient creates a new StorageV1alpha1Client for the given config and http client.
// Note the http client provided takes precedence over the configured transport values.
func NewForConfigAndClient(c *rest.Config, h *http.Client) (*StorageV1alpha1Client, error) {
	config := *c
	if err := setConfigDefaults(&config); err != nil {
		return nil, err
	}
	client, err := rest.RESTClientForConfigAndClient(&config, h)
	if err != nil {
		return nil, err
	}
	return &StorageV1alpha1Client{client}, nil
}

// NewForConfigOrDie creates a new StorageV1alpha1Client for the given config and
// panics if there is an error in the config.
func NewForConfigOrDie(c *rest.Config) *StorageV1alpha1Client {
	client, err := NewForConfig(c)
	if err != nil {
		panic(err)
	}
	return client
}

// New creates a new StorageV1alpha1Client for the given RESTClient.
func New(c rest.Interface) *StorageV1alpha1Client {
	return &StorageV1alpha1Client{c}
}

func setConfigDefaults(config *rest.Config) error {
	gv := v1alpha1.SchemeGroupVersion
	config.GroupVersion = &gv
	config.APIPath = "/apis"
	config.NegotiatedSerializer = scheme.Codecs.WithoutConversion()

	if config.UserAgent == "" {
		config.UserAgent = rest.DefaultKubernetesUserAgent()
	}

	return nil
}

// RESTClient returns a RESTClient that is used to communicate
// with API server by this client implementation.
func (c *StorageV1alpha1Client) RESTClient() rest.Interface {
	if c == nil {
		return nil
	}
	return c.restClient
}
