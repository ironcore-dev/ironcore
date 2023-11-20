// SPDX-FileCopyrightText: 2023 SAP SE or an SAP affiliate company and IronCore contributors
// SPDX-License-Identifier: Apache-2.0

package get

import (
	"github.com/ironcore-dev/ironcore/irictl-volume/cmd/irictl-volume/irictlvolume/common"
	"github.com/ironcore-dev/ironcore/irictl-volume/cmd/irictl-volume/irictlvolume/get/status"
	"github.com/ironcore-dev/ironcore/irictl-volume/cmd/irictl-volume/irictlvolume/get/volume"
	clicommon "github.com/ironcore-dev/ironcore/irictl/cmd"
	"github.com/spf13/cobra"
)

func Command(streams clicommon.Streams, clientFactory common.ClientFactory) *cobra.Command {
	cmd := &cobra.Command{
		Use: "get",
	}

	cmd.AddCommand(
		volume.Command(streams, clientFactory),
		status.Command(streams, clientFactory),
	)

	return cmd
}
