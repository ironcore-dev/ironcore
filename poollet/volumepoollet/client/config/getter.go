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

package config

import (
	"crypto/x509"
	"crypto/x509/pkix"
	"os"

	storagev1alpha1 "github.com/onmetal/onmetal-api/api/storage/v1alpha1"
	utilcertificate "github.com/onmetal/onmetal-api/utils/certificate"
	"github.com/onmetal/onmetal-api/utils/client/config"
	certificatesv1 "k8s.io/api/certificates/v1"
	ctrl "sigs.k8s.io/controller-runtime"
)

var log = ctrl.Log.WithName("client").WithName("config")

func NewGetter(volumePoolName string) (*config.Getter, error) {
	return config.NewGetter(config.GetterOptions{
		Name:       "volumepoollet",
		SignerName: certificatesv1.KubeAPIServerClientSignerName,
		Template: &x509.CertificateRequest{
			Subject: pkix.Name{
				CommonName:   storagev1alpha1.VolumePoolCommonName(volumePoolName),
				Organization: []string{storagev1alpha1.VolumePoolsGroup},
			},
		},
		GetUsages: utilcertificate.DefaultKubeAPIServerClientGetUsages,
	})
}

func NewGetterOrDie(volumePoolName string) *config.Getter {
	getter, err := NewGetter(volumePoolName)
	if err != nil {
		log.Error(err, "Error creating getter")
		os.Exit(1)
	}
	return getter
}
