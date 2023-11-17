// Copyright 2022 IronCore authors
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
//      http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.

package volume

import (
	"context"
	"fmt"

	ori "github.com/ironcore-dev/ironcore/ori/apis/volume/v1alpha1"
	"github.com/ironcore-dev/ironcore/orictl-volume/cmd/orictl-volume/orictlvolume/common"
	clicommon "github.com/ironcore-dev/ironcore/orictl/cmd"
	"github.com/ironcore-dev/ironcore/orictl/renderer"
	"github.com/spf13/cobra"
	"github.com/spf13/pflag"
	ctrl "sigs.k8s.io/controller-runtime"
)

type Options struct {
}

func (o *Options) AddFlags(fs *pflag.FlagSet) {
}

func Command(streams clicommon.Streams, clientFactory common.ClientFactory) *cobra.Command {
	var (
		opts       Options
		outputOpts = common.NewOutputOptions()
	)

	cmd := &cobra.Command{
		Use:     "volume",
		Aliases: common.VolumeAliases,
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

			render, err := outputOpts.Renderer("table")
			if err != nil {
				return err
			}

			return Run(cmd.Context(), streams, client, render, opts)
		},
	}

	outputOpts.AddFlags(cmd.Flags())
	opts.AddFlags(cmd.Flags())

	return cmd
}

func Run(ctx context.Context, streams clicommon.Streams, client ori.VolumeRuntimeClient, render renderer.Renderer, opts Options) error {
	res, err := client.ListVolumes(ctx, &ori.ListVolumesRequest{})
	if err != nil {
		return fmt.Errorf("error listing volumes: %w", err)
	}

	return render.Render(res.Volumes, streams.Out)
}
