// SPDX-FileCopyrightText: 2023 SAP SE or an SAP affiliate company and IronCore contributors
// SPDX-License-Identifier: Apache-2.0

package power

import (
	"context"
	"fmt"

	iri "github.com/ironcore-dev/ironcore/iri/apis/machine/v1alpha1"
	"github.com/ironcore-dev/ironcore/irictl-machine/cmd/irictl-machine/irictlmachine/common"
	clicommon "github.com/ironcore-dev/ironcore/irictl/cmd"
	"github.com/spf13/cobra"
	utilruntime "k8s.io/apimachinery/pkg/util/runtime"
	ctrl "sigs.k8s.io/controller-runtime"
)

type Options struct {
	Power     int32
	MachineID string
}

func (o *Options) AddFlags(cmd *cobra.Command) {
	cmd.Flags().Int32Var(&o.Power, "power", o.Power, "The power state to set (0: POWER_ON, 1: POWER_OFF).")
	cmd.Flags().StringVar(&o.MachineID, "machine-id", "", "The machine ID to modify.")
	utilruntime.Must(cmd.MarkFlagRequired("machine-id"))
}

func Command(streams clicommon.Streams, clientFactory common.Factory) *cobra.Command {
	var (
		opts Options
	)

	cmd := &cobra.Command{
		Use: "power",
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

			return Run(ctx, streams, client, opts)
		},
	}

	opts.AddFlags(cmd)

	return cmd
}

func Run(ctx context.Context, streams clicommon.Streams, client iri.MachineRuntimeClient, opts Options) error {
	if _, err := client.UpdateMachinePower(ctx, &iri.UpdateMachinePowerRequest{
		MachineId: opts.MachineID,
		Power:     iri.Power(opts.Power),
	}); err != nil {
		return fmt.Errorf("error updating power state in machine %s: %w", opts.MachineID, err)
	}

	_, _ = fmt.Fprintf(streams.Out, "Updated power state in machine %s\n", opts.MachineID)
	return nil
}
