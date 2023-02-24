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

package config_test

import (
	"os"
	"path/filepath"

	"github.com/onmetal/onmetal-api/utils/client/config"
	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/client-go/rest"
	"k8s.io/client-go/tools/clientcmd"
	clientcmdapi "k8s.io/client-go/tools/clientcmd/api"
	"sigs.k8s.io/controller-runtime/pkg/client"
)

var _ = Describe("Store", func() {
	var (
		apiCfg     *clientcmdapi.Config
		apiCfgData []byte
		restCfg    *rest.Config
	)
	BeforeEach(func() {
		apiCfg = &clientcmdapi.Config{
			Clusters: map[string]*clientcmdapi.Cluster{
				"default": {
					Server: "https://foo.example.org",
				},
			},
			AuthInfos: map[string]*clientcmdapi.AuthInfo{
				"default": {
					Username: "foo",
					Password: "bar",
				},
			},
			Contexts: map[string]*clientcmdapi.Context{
				"default": {
					Cluster:  "default",
					AuthInfo: "default",
				},
			},
			CurrentContext: "default",
		}
		var err error
		apiCfgData, err = clientcmd.Write(*apiCfg)
		Expect(err).NotTo(HaveOccurred())

		restCfg = &rest.Config{
			Host:     "https://foo.example.org",
			Username: "foo",
			Password: "bar",
		}
	})

	Context("FileStore", func() {
		var (
			existingFile    string
			nonExistentFile string
		)
		BeforeEach(func() {
			tmpFile, err := os.CreateTemp(GinkgoT().TempDir(), "kubeconfig")
			Expect(err).NotTo(HaveOccurred())
			existingFile = tmpFile.Name()
			Expect(tmpFile.Close()).To(Succeed())

			Expect(clientcmd.WriteToFile(*apiCfg, existingFile)).To(Succeed())

			nonExistentFile = "/definitely/should/never/exist"
		})

		Describe("Get", func() {
			It("should get the config from file", func(ctx SpecContext) {
				cfg, err := config.FileStore(existingFile).Get(ctx)
				Expect(err).NotTo(HaveOccurred())
				Expect(cfg).To(Equal(restCfg))
			})

			It("should return ErrConfigNotFound if a config does not exist", func(ctx SpecContext) {
				_, err := config.FileStore(nonExistentFile).Get(ctx)
				Expect(err).To(MatchError(config.ErrConfigNotFound))
			})
		})

		Describe("Set", func() {
			It("should write the config to the file", func(ctx SpecContext) {
				filename := filepath.Join(GinkgoT().TempDir(), "config")
				store := config.FileStore(filename)
				Expect(store.Set(ctx, restCfg)).To(Succeed())

				data, err := os.ReadFile(filename)
				Expect(err).NotTo(HaveOccurred())
				Expect(clientcmd.RESTConfigFromKubeConfig(data)).To(Equal(restCfg))
			})
		})
	})

	Context("SecretStore", func() {
		var (
			existingSecretKey    client.ObjectKey
			nonExistentSecretKey client.ObjectKey
			ns                   *corev1.Namespace
		)
		BeforeEach(func(ctx SpecContext) {
			ns = &corev1.Namespace{
				ObjectMeta: metav1.ObjectMeta{
					GenerateName: "ns-",
				},
			}
			Expect(k8sClient.Create(ctx, ns)).To(Succeed())
			DeferCleanup(k8sClient.Delete, ns)

			existingSecret := &corev1.Secret{
				ObjectMeta: metav1.ObjectMeta{
					Namespace:    ns.Name,
					GenerateName: "kubeconfig-",
				},
				Data: map[string][]byte{
					config.DefaultSecretKubeconfigField: apiCfgData,
				},
			}
			Expect(k8sClient.Create(ctx, existingSecret)).To(Succeed())

			existingSecretKey = client.ObjectKeyFromObject(existingSecret)
			nonExistentSecretKey = client.ObjectKey{Namespace: ns.Name, Name: "should-definitely-not-exist"}
		})

		Describe("Get", func() {
			It("should load the config from the secret", func(ctx SpecContext) {
				store := config.NewSecretStore(k8sClient, existingSecretKey)
				cfg, err := store.Get(ctx)
				Expect(err).NotTo(HaveOccurred())
				Expect(cfg).To(Equal(restCfg))
			})

			It("should return ErrConfigNotFound if the secret does not exist", func(ctx SpecContext) {
				store := config.NewSecretStore(k8sClient, nonExistentSecretKey)
				_, err := store.Get(ctx)
				Expect(err).To(MatchError(config.ErrConfigNotFound))
			})

			It("should return ErrConfigNotFound if there is no data at the secret field", func(ctx SpecContext) {
				store := config.NewSecretStore(k8sClient, existingSecretKey, config.WithField("other-field"))
				_, err := store.Get(ctx)
				Expect(err).To(MatchError(config.ErrConfigNotFound))
			})
		})

		Describe("Set", func() {
			It("should set the kubeconfig in the secret", func(ctx SpecContext) {
				store := config.NewSecretStore(k8sClient, nonExistentSecretKey)
				Expect(store.Set(ctx, restCfg)).To(Succeed())

				secret := &corev1.Secret{}
				Expect(k8sClient.Get(ctx, existingSecretKey, secret)).To(Succeed())

				Expect(secret.Data).To(HaveKey(config.DefaultSecretKubeconfigField))
				Expect(clientcmd.RESTConfigFromKubeConfig(secret.Data[config.DefaultSecretKubeconfigField])).To(Equal(restCfg))
			})
		})
	})
})
