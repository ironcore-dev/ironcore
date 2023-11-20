// SPDX-FileCopyrightText: 2023 SAP SE or an SAP affiliate company and IronCore contributors
// SPDX-License-Identifier: Apache-2.0

package main

import (
	"fmt"
	"os"

	"github.com/ironcore-dev/ironcore/broker/machinebroker/cmd/machinebroker/app"
	ctrl "sigs.k8s.io/controller-runtime"
)

func main() {
	ctx := ctrl.SetupSignalHandler()

	if err := app.Command().ExecuteContext(ctx); err != nil {
		fmt.Println(err.Error())
		os.Exit(1)
	}
}
