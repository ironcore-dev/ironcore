// Copyright 2022 OnMetal authors
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

package networkinterface

import (
	"context"
	"fmt"

	ori "github.com/onmetal/onmetal-api/ori/apis/runtime/v1alpha1"
	"github.com/onmetal/onmetal-api/orictl/cmd/orictl/common"
	"github.com/onmetal/onmetal-api/orictl/renderer"
	"github.com/spf13/cobra"
	"github.com/spf13/pflag"
	ctrl "sigs.k8s.io/controller-runtime"
)

type Options struct {
	MachineID string
}

func (o *Options) AddFlags(fs *pflag.FlagSet) {
	fs.StringVar(&o.MachineID, "machine-id", "", "Filter network interfaces by machine id.")
}

func Command(streams common.Streams, clientFactory common.ClientFactory) *cobra.Command {
	var (
		opts       Options
		outputOpts common.OutputOptions
	)

	cmd := &cobra.Command{
		Use:     "networkinterface",
		Aliases: common.NetworkInterfaceAliases,
		RunE: func(cmd *cobra.Command, args []string) error {
			ctx := cmd.Context()
			log := ctrl.LoggerFrom(ctx)

			client, cleanup, err := clientFactory(ctx)
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

func Run(ctx context.Context, streams common.Streams, client ori.MachineRuntimeClient, render renderer.Renderer, opts Options) error {
	var filter *ori.NetworkInterfaceFilter
	if opts.MachineID != "" {
		filter = &ori.NetworkInterfaceFilter{
			MachineId: opts.MachineID,
		}
	}

	res, err := client.ListNetworkInterfaces(ctx, &ori.ListNetworkInterfacesRequest{
		Filter: filter,
	})
	if err != nil {
		return fmt.Errorf("error listing network interfaces: %w", err)
	}

	return render.Render(res.NetworkInterfaces, streams.Out)
}
