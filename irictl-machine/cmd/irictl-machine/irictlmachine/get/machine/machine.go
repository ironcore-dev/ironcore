// SPDX-FileCopyrightText: 2023 SAP SE or an SAP affiliate company and IronCore contributors
// SPDX-License-Identifier: Apache-2.0

package machine

import (
	"context"
	"fmt"

	iri "github.com/ironcore-dev/ironcore/iri/apis/machine/v1alpha1"
	"github.com/ironcore-dev/ironcore/irictl-machine/cmd/irictl-machine/irictlmachine/common"
	clicommon "github.com/ironcore-dev/ironcore/irictl/cmd"
	"github.com/ironcore-dev/ironcore/irictl/renderer"
	"github.com/spf13/cobra"
	"github.com/spf13/pflag"
	ctrl "sigs.k8s.io/controller-runtime"
)

type Options struct {
	Labels map[string]string
}

func (o *Options) AddFlags(fs *pflag.FlagSet) {
	fs.StringToStringVarP(&o.Labels, "labels", "l", o.Labels, "Labels to filter the machines by.")
}

func Command(streams clicommon.Streams, clientFactory common.Factory) *cobra.Command {
	var (
		opts       Options
		outputOpts = clientFactory.OutputOptions()
	)

	cmd := &cobra.Command{
		Use:     "machine name",
		Args:    cobra.MaximumNArgs(1),
		Aliases: common.MachineAliases,
		RunE: func(cmd *cobra.Command, args []string) error {
			ctx := cmd.Context()
			log := ctrl.LoggerFrom(ctx)

			client, cleanup, err := clientFactory.Client()
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

			var name string
			if len(args) > 0 {
				name = args[0]
			}

			return Run(cmd.Context(), streams, client, render, name, opts)
		},
	}

	outputOpts.AddFlags(cmd.Flags())
	opts.AddFlags(cmd.Flags())

	return cmd
}

func Run(
	ctx context.Context,
	streams clicommon.Streams,
	client iri.MachineRuntimeClient,
	render renderer.Renderer,
	name string,
	opts Options,
) error {
	var filter *iri.MachineFilter
	if name != "" || opts.Labels != nil {
		filter = &iri.MachineFilter{
			Id:            name,
			LabelSelector: opts.Labels,
		}
	}

	res, err := client.ListMachines(ctx, &iri.ListMachinesRequest{Filter: filter})
	if err != nil {
		return fmt.Errorf("error listing machines: %w", err)
	}

	return render.Render(res.Machines, streams.Out)
}
