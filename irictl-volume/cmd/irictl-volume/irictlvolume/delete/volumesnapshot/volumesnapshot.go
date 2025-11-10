// SPDX-FileCopyrightText: 2023 SAP SE or an SAP affiliate company and IronCore contributors
// SPDX-License-Identifier: Apache-2.0

package volumesnapshot

import (
	"context"
	"fmt"

	iri "github.com/ironcore-dev/ironcore/iri/apis/volume/v1alpha1"
	"github.com/ironcore-dev/ironcore/irictl-volume/cmd/irictl-volume/irictlvolume/common"
	clicommon "github.com/ironcore-dev/ironcore/irictl/cmd"
	"github.com/spf13/cobra"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
	ctrl "sigs.k8s.io/controller-runtime"
)

func Command(streams clicommon.Streams, clientFactory common.ClientFactory) *cobra.Command {
	cmd := &cobra.Command{
		Use:     "volumesnapshot id [ids...]",
		Aliases: common.VolumeSnapshotAliases,
		Args:    cobra.MinimumNArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			ctx := cmd.Context()
			log := ctrl.LoggerFrom(ctx)

			client, cleanup, err := clientFactory.New()
			if err != nil {
				return err
			}
			defer func() {
				if err := cleanup(); err != nil {
					log.Error(err, "Error cleaning up")
				}
			}()

			ids := args

			return Run(cmd.Context(), streams, client, ids)
		},
	}

	return cmd
}

func Run(ctx context.Context, streams clicommon.Streams, client iri.VolumeRuntimeClient, ids []string) error {
	for _, id := range ids {
		if _, err := client.DeleteVolumeSnapshot(ctx, &iri.DeleteVolumeSnapshotRequest{
			VolumeSnapshotId: id,
		}); err != nil {
			if status.Code(err) != codes.NotFound {
				return fmt.Errorf("error deleting volume snapshot %s: %w", id, err)
			}

			_, _ = fmt.Fprintf(streams.Out, "VolumeSnapshot %s not found\n", id)
		} else {
			_, _ = fmt.Fprintf(streams.Out, "VolumeSnapshot %s deleted\n", id)
		}
	}
	return nil
}
