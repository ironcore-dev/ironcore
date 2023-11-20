// SPDX-FileCopyrightText: 2023 SAP SE or an SAP affiliate company and IronCore contributors
// SPDX-License-Identifier: Apache-2.0

package main

import (
	"os"

	"github.com/ironcore-dev/ironcore/irictl-machine/cmd/irictl-machine/irictlmachine"
	clicommon "github.com/ironcore-dev/ironcore/irictl/cmd"
	ctrl "sigs.k8s.io/controller-runtime"
)

func main() {
	ctx := ctrl.SetupSignalHandler()
	if err := irictlmachine.Command(clicommon.OSStreams).ExecuteContext(ctx); err != nil {
		ctrl.Log.Error(err, "Error running command")
		os.Exit(1)
	}
}
