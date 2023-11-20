// SPDX-FileCopyrightText: 2023 SAP SE or an SAP affiliate company and IronCore contributors
// SPDX-License-Identifier: Apache-2.0

package delete

import (
	"github.com/ironcore-dev/ironcore/irictl-bucket/cmd/irictl-bucket/irictlbucket/common"
	"github.com/ironcore-dev/ironcore/irictl-bucket/cmd/irictl-bucket/irictlbucket/delete/bucket"
	irictlcmd "github.com/ironcore-dev/ironcore/irictl/cmd"
	"github.com/spf13/cobra"
)

func Command(streams irictlcmd.Streams, clientFactory common.ClientFactory) *cobra.Command {
	cmd := &cobra.Command{
		Use: "delete",
	}

	cmd.AddCommand(
		bucket.Command(streams, clientFactory),
	)

	return cmd
}
