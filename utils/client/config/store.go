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
	"fmt"
	"sync"

	clientcmdutil "github.com/ironcore-dev/ironcore/utils/clientcmd"
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/client-go/rest"
	"k8s.io/client-go/tools/clientcmd"
	clientcmdapi "k8s.io/client-go/tools/clientcmd/api"
	"sigs.k8s.io/controller-runtime/pkg/client"
)

type Store interface {
	Get(ctx context.Context) (*rest.Config, error)
	Set(ctx context.Context, cfg *rest.Config) error
}

type MemoryStore struct {
	mu  sync.RWMutex
	cfg *rest.Config
}

func (m *MemoryStore) Get(_ context.Context) (*rest.Config, error) {
	m.mu.RLock()
	defer m.mu.RUnlock()

	if m.cfg == nil {
		return nil, ErrConfigNotFound
	}
	return rest.CopyConfig(m.cfg), nil
}

func (m *MemoryStore) Set(_ context.Context, cfg *rest.Config) error {
	if cfg == nil {
		return fmt.Errorf("must specify cfg")
	}
	m.mu.Lock()
	defer m.mu.Unlock()

	m.cfg = rest.CopyConfig(cfg)
	return nil
}

type FileStore string

func (f FileStore) Get(ctx context.Context) (*rest.Config, error) {
	return FileLoader(f).Load(ctx, nil)
}

func (f FileStore) Set(_ context.Context, cfg *rest.Config) error {
	apiCfg, err := clientcmdutil.RESTConfigToConfig(cfg)
	if err != nil {
		return fmt.Errorf("error converting rest config to config: %w", err)
	}

	return clientcmd.WriteToFile(*apiCfg, string(f))
}

type SecretStore struct {
	client     client.Client
	key        client.ObjectKey
	fieldOwner client.FieldOwner
	field      string
}

func (s *SecretStore) Get(ctx context.Context) (*rest.Config, error) {
	return NewSecretLoader(s.client, s.key, WithField(s.field)).Load(ctx, nil)
}

func (s *SecretStore) Set(ctx context.Context, cfg *rest.Config) error {
	apiCfg, err := clientcmdutil.RESTConfigToConfig(cfg)
	if err != nil {
		return fmt.Errorf("error transforming rest config to api config: %w", err)
	}

	if err := clientcmdapi.FlattenConfig(apiCfg); err != nil {
		return fmt.Errorf("error flattening api config: %w", err)
	}

	kubeconfigData, err := clientcmd.Write(*apiCfg)
	if err != nil {
		return err
	}

	secret := &corev1.Secret{
		TypeMeta: metav1.TypeMeta{
			Kind:       "Secret",
			APIVersion: corev1.SchemeGroupVersion.String(),
		},
		ObjectMeta: metav1.ObjectMeta{
			Namespace: s.key.Namespace,
			Name:      s.key.Name,
		},
		Data: map[string][]byte{
			s.field: kubeconfigData,
		},
	}
	if err := s.client.Patch(ctx, secret, client.Apply, s.fieldOwner, client.ForceOwnership); err != nil {
		return fmt.Errorf("error applying secret %s: %w", s.key, err)
	}
	return nil
}

type SecretStoreOptions struct {
	Field      string
	FieldOwner client.FieldOwner
}

func (o *SecretStoreOptions) ApplyToSecretConfigStore(o2 *SecretStoreOptions) {
	if o.Field != "" {
		o2.Field = o.Field
	}
	if o.FieldOwner != "" {
		o2.FieldOwner = o.FieldOwner
	}
}

func (o *SecretStoreOptions) ApplyOptions(opts []SecretStoreOption) {
	for _, opt := range opts {
		opt.ApplyToSecretConfigStore(o)
	}
}

type SecretStoreOption interface {
	ApplyToSecretConfigStore(o *SecretStoreOptions)
}

type WithField string

func (w WithField) ApplyToSecretConfigStore(o *SecretStoreOptions) {
	o.Field = string(w)
}

func (w WithField) ApplyToSecretLoader(o *SecretLoaderOptions) {
	o.Field = string(w)
}

type WithFieldOwner client.FieldOwner

func (w WithFieldOwner) ApplyToSecretConfigStore(o *SecretStoreOptions) {
	o.FieldOwner = client.FieldOwner(w)
}

type WithOverrides clientcmd.ConfigOverrides

const (
	DefaultSecretKubeconfigField            = "kubeconfig"
	DefaultSecretConfigReadWriterFieldOwner = client.FieldOwner("ironcore.dev/config-read-writer")
)

func setSecretConfigReadWriterOptionsDefaults(o *SecretStoreOptions) {
	if o.Field == "" {
		o.Field = DefaultSecretKubeconfigField
	}
	if o.FieldOwner == "" {
		o.FieldOwner = DefaultSecretConfigReadWriterFieldOwner
	}
}

func NewSecretStore(
	c client.Client,
	key client.ObjectKey,
	opts ...SecretStoreOption,
) *SecretStore {
	o := &SecretStoreOptions{}
	o.ApplyOptions(opts)
	setSecretConfigReadWriterOptionsDefaults(o)

	return &SecretStore{
		client:     c,
		key:        key,
		fieldOwner: o.FieldOwner,
		field:      o.Field,
	}
}
