// SPDX-FileCopyrightText: 2024 SAP SE or an SAP affiliate company and IronCore contributors
// SPDX-License-Identifier: Apache-2.0

package event

import (
	"context"
	"fmt"
	"time"

	iri "github.com/ironcore-dev/ironcore/iri/apis/machine/v1alpha1"
	"github.com/ironcore-dev/ironcore/irictl-machine/cmd/irictl-machine/irictlmachine/common"
	clicommon "github.com/ironcore-dev/ironcore/irictl/cmd"
	"github.com/ironcore-dev/ironcore/irictl/renderer"
	"github.com/spf13/cobra"
	"github.com/spf13/pflag"
	ctrl "sigs.k8s.io/controller-runtime"
)

const DefaultDuration = 60 * time.Minute

type Options struct {
	Labels   map[string]string
	Duration time.Duration
}

func (o *Options) AddFlags(fs *pflag.FlagSet) {
	fs.StringToStringVarP(&o.Labels, "labels", "l", o.Labels, "Labels to filter the events by.")
	fs.DurationVarP(&o.Duration, "duration", "d", o.Duration, "Duration to filter the events by.")
}

func Command(streams clicommon.Streams, clientFactory common.Factory) *cobra.Command {
	var (
		opts       Options
		outputOpts = clientFactory.OutputOptions()
	)

	cmd := &cobra.Command{
		Use: "events",
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

			return Run(cmd.Context(), streams, client, render, opts)
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
	opts Options,
) error {
	var filter *iri.EventFilter = &iri.EventFilter{
		EventsFromTime: time.Now().Add(-1 * DefaultDuration).Unix(),
		EventsToTime:   time.Now().Unix(),
	}
	if opts.Labels != nil {
		filter.LabelSelector = opts.Labels
	}

	if opts.Duration > 0 {
		filter.EventsFromTime = time.Now().Add(-1 * opts.Duration).Unix()
	}

	res, err := client.ListEvents(ctx, &iri.ListEventsRequest{Filter: filter})
	if err != nil {
		return fmt.Errorf("error listing events: %w", err)
	}

	return render.Render(res.Events, streams.Out)
}
