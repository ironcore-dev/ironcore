// SPDX-FileCopyrightText: 2023 SAP SE or an SAP affiliate company and IronCore contributors
// SPDX-License-Identifier: Apache-2.0

package volumesnapshot

import (
	"context"
	"fmt"
	"os"

	iri "github.com/ironcore-dev/ironcore/iri/apis/volume/v1alpha1"
	"github.com/ironcore-dev/ironcore/irictl-volume/cmd/irictl-volume/irictlvolume/common"
	clicommon "github.com/ironcore-dev/ironcore/irictl/cmd"
	"github.com/ironcore-dev/ironcore/irictl/decoder"
	"github.com/ironcore-dev/ironcore/irictl/renderer"
	"github.com/spf13/cobra"
	"github.com/spf13/pflag"
	ctrl "sigs.k8s.io/controller-runtime"
)

type Options struct {
	Filename string
}

func (o *Options) AddFlags(fs *pflag.FlagSet) {
	fs.StringVarP(&o.Filename, "filename", "f", o.Filename, "Path to a file to read.")
}

func Command(streams clicommon.Streams, clientFactory common.ClientFactory) *cobra.Command {
	var (
		outputOpts = common.NewOutputOptions()
		opts       Options
	)

	cmd := &cobra.Command{
		Use:     "volumesnapshot",
		Aliases: common.VolumeSnapshotAliases,
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

			r, err := outputOpts.RendererOrNil()
			if err != nil {
				return err
			}

			return Run(ctx, streams, client, r, opts)
		},
	}

	outputOpts.AddFlags(cmd.Flags())
	opts.AddFlags(cmd.Flags())

	return cmd
}

func Run(ctx context.Context, streams clicommon.Streams, client iri.VolumeRuntimeClient, r renderer.Renderer, opts Options) error {
	data, err := clicommon.ReadFileOrReader(opts.Filename, os.Stdin)
	if err != nil {
		return err
	}

	volumeSnapshot := &iri.VolumeSnapshot{}
	if err := decoder.Decode(data, volumeSnapshot); err != nil {
		return err
	}

	res, err := client.CreateVolumeSnapshot(ctx, &iri.CreateVolumeSnapshotRequest{VolumeSnapshot: volumeSnapshot})
	if err != nil {
		return err
	}

	if r != nil {
		return r.Render(res.VolumeSnapshot, streams.Out)
	}

	_, _ = fmt.Fprintf(streams.Out, "Created volume snapshot %s\n", res.VolumeSnapshot.Metadata.Id)
	return nil
}
