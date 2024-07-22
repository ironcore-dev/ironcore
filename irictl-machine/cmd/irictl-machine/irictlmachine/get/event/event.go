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

type Options struct {
	Labels         map[string]string
	EventsFromTime string
	EventsToTime   string
}

func (o *Options) AddFlags(fs *pflag.FlagSet) {
	fs.StringToStringVarP(&o.Labels, "labels", "l", o.Labels, "Labels to filter the events by.")
	fs.StringVarP(&o.EventsFromTime, "eventsFromTime", "f", o.EventsFromTime, fmt.Sprintf("Events From Time to filter the events by. In the format of %s", time.DateTime))
	fs.StringVarP(&o.EventsToTime, "eventsToTime", "t", o.EventsToTime, fmt.Sprintf("Events To Time to filter the events by. In the format of %s", time.DateTime))
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
	var filter *iri.EventFilter = &iri.EventFilter{}
	if opts.Labels != nil {
		filter.LabelSelector = opts.Labels
	}
	if opts.EventsFromTime != "" {
		fromDate, err := time.Parse(time.DateTime, opts.EventsFromTime)
		if err != nil {
			return fmt.Errorf("error parsing eventsFromTime: %w", err)
		}
		filter.EventsFromTime = fromDate.Unix()
	}

	if opts.EventsToTime != "" {
		toDate, err := time.Parse(time.DateTime, opts.EventsToTime)
		if err != nil {
			return fmt.Errorf("error parsing eventsToTime: %w", err)
		}
		filter.EventsToTime = toDate.Unix()
	}

	res, err := client.ListEvents(ctx, &iri.ListEventsRequest{Filter: filter})
	if err != nil {
		return fmt.Errorf("error listing events: %w", err)
	}

	return render.Render(res.Events, streams.Out)
}
