// SPDX-FileCopyrightText: 2023 SAP SE or an SAP affiliate company and IronCore contributors
// SPDX-License-Identifier: Apache-2.0

package detach

import (
	"github.com/ironcore-dev/ironcore/irictl-machine/cmd/irictl-machine/irictlmachine/common"
	"github.com/ironcore-dev/ironcore/irictl-machine/cmd/irictl-machine/irictlmachine/detach/networkinterface"
	"github.com/ironcore-dev/ironcore/irictl-machine/cmd/irictl-machine/irictlmachine/detach/volume"
	clicommon "github.com/ironcore-dev/ironcore/irictl/cmd"
	"github.com/spf13/cobra"
)

func Command(streams clicommon.Streams, clientFactory common.Factory) *cobra.Command {
	cmd := &cobra.Command{
		Use: "detach",
	}

	cmd.AddCommand(
		volume.Command(streams, clientFactory),
		networkinterface.Command(streams, clientFactory),
	)

	return cmd
}
