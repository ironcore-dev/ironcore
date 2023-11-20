// SPDX-FileCopyrightText: 2023 SAP SE or an SAP affiliate company and IronCore contributors
// SPDX-License-Identifier: Apache-2.0

package main

import (
	"os"

	"github.com/ironcore-dev/ironcore/internal/app/apiserver"
	genericapiserver "k8s.io/apiserver/pkg/server"
	"k8s.io/component-base/cli"
)

func main() {
	ctx := genericapiserver.SetupSignalContext()
	options := apiserver.NewIronCoreAPIServerOptions()
	cmd := apiserver.NewCommandStartIronCoreAPIServer(ctx, options)
	code := cli.Run(cmd)
	os.Exit(code)
}
