// Copyright 2023 IronCore authors
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

package config

import (
	"context"
	"errors"
	"fmt"
	"os"

	corev1 "k8s.io/api/core/v1"
	apierrors "k8s.io/apimachinery/pkg/api/errors"
	"k8s.io/client-go/rest"
	"k8s.io/client-go/tools/clientcmd"
	"sigs.k8s.io/controller-runtime/pkg/client"
)

type Loader interface {
	Load(ctx context.Context, overrides *clientcmd.ConfigOverrides) (*rest.Config, error)
}

type FileLoader string

func (l FileLoader) Load(ctx context.Context, overrides *clientcmd.ConfigOverrides) (*rest.Config, error) {
	apiCfg, err := clientcmd.LoadFromFile(string(l))
	if err != nil {
		if !errors.Is(err, os.ErrNotExist) {
			return nil, fmt.Errorf("error loading api config from file: %w", err)
		}
		return nil, ErrConfigNotFound
	}

	return clientcmd.NewDefaultClientConfig(*apiCfg, overrides).ClientConfig()
}

type SecretLoader struct {
	rd    client.Reader
	key   client.ObjectKey
	field string
}

type SecretLoaderOptions struct {
	Field string
}

func (o *SecretLoaderOptions) ApplyOptions(opts []SecretLoaderOption) {
	for _, opt := range opts {
		opt.ApplyToSecretLoader(o)
	}
}

func (o *SecretLoaderOptions) ApplyToSecretLoader(o2 *SecretLoaderOptions) {
	if o.Field != "" {
		o2.Field = o.Field
	}
}

type SecretLoaderOption interface {
	ApplyToSecretLoader(o *SecretLoaderOptions)
}

func setSecretLoaderOptionsDefaults(o *SecretLoaderOptions) {
	if o.Field == "" {
		o.Field = DefaultSecretKubeconfigField
	}
}

func NewSecretLoader(rd client.Reader, key client.ObjectKey, opts ...SecretLoaderOption) *SecretLoader {
	o := &SecretLoaderOptions{}
	o.ApplyOptions(opts)
	setSecretLoaderOptionsDefaults(o)

	return &SecretLoader{
		rd:    rd,
		key:   key,
		field: o.Field,
	}
}

func (l *SecretLoader) Load(ctx context.Context, overrides *clientcmd.ConfigOverrides) (*rest.Config, error) {
	secret := &corev1.Secret{}
	if err := l.rd.Get(ctx, l.key, secret); err != nil {
		if !apierrors.IsNotFound(err) {
			return nil, fmt.Errorf("error getting secret %s: %w", l.key, err)
		}
		return nil, ErrConfigNotFound
	}

	data, ok := secret.Data[l.field]
	if !ok || len(data) == 0 {
		return nil, ErrConfigNotFound
	}

	apiCfg, err := clientcmd.Load(data)
	if err != nil {
		return nil, fmt.Errorf("error loading api config from secret data: %w", err)
	}

	return clientcmd.NewDefaultClientConfig(*apiCfg, overrides).ClientConfig()
}
