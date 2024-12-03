// SPDX-FileCopyrightText: 2023 SAP SE or an SAP affiliate company and IronCore contributors
// SPDX-License-Identifier: Apache-2.0

package config

import (
	"os"

	"github.com/ironcore-dev/ironcore/utils/client/config"
	"k8s.io/apiserver/pkg/server/egressselector"
	ctrl "sigs.k8s.io/controller-runtime"
)

var log = ctrl.Log.WithName("client").WithName("config")

func NewGetter() (*config.BrokerGetter, error) {
	return config.NewBrokerGetter(config.GetterOptions{
		Name:           "bucketbroker",
		NetworkContext: egressselector.ControlPlane.AsNetworkContext(),
	})
}

func NewGetterOrDie() *config.BrokerGetter {
	getter, err := NewGetter()
	if err != nil {
		log.Error(err, "Error creating getter")
		os.Exit(1)
	}
	return getter
}
