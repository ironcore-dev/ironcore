// SPDX-FileCopyrightText: 2023 SAP SE or an SAP affiliate company and IronCore contributors
// SPDX-License-Identifier: Apache-2.0

package config

import (
	"crypto/x509"
	"crypto/x509/pkix"
	"os"

	storagev1alpha1 "github.com/ironcore-dev/ironcore/api/storage/v1alpha1"
	utilcertificate "github.com/ironcore-dev/ironcore/utils/certificate"
	"github.com/ironcore-dev/ironcore/utils/client/config"
	certificatesv1 "k8s.io/api/certificates/v1"
	"k8s.io/apiserver/pkg/server/egressselector"
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
		GetUsages:      utilcertificate.DefaultKubeAPIServerClientGetUsages,
		NetworkContext: egressselector.ControlPlane.AsNetworkContext(),
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
