// Copyright 2023 OnMetal authors
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

	ori "github.com/onmetal/onmetal-api/ori/apis/machine/v1alpha1"
	"github.com/onmetal/onmetal-api/orictl-machine/cmd/orictl-machine/orictlmachine/common"
	clicommon "github.com/onmetal/onmetal-api/orictl/cmd"
	"github.com/spf13/cobra"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
	utilruntime "k8s.io/apimachinery/pkg/util/runtime"
	ctrl "sigs.k8s.io/controller-runtime"
)

type Options struct {
	MachineID string
}

func (o *Options) AddFlags(cmd *cobra.Command) {
	cmd.Flags().StringVar(&o.MachineID, "machine-id", "", "The machine ID to modify.")
	utilruntime.Must(cmd.MarkFlagRequired("machine-id"))
}

func Command(streams clicommon.Streams, clientFactory common.Factory) *cobra.Command {
	var (
		opts Options
	)

	cmd := &cobra.Command{
		Use:     "volume name [names...]",
		Aliases: common.VolumeAliases,
		Args:    cobra.MinimumNArgs(1),
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

			names := args
			return Run(ctx, streams, client, names, opts)
		},
	}

	opts.AddFlags(cmd)

	return cmd
}

func Run(ctx context.Context, streams clicommon.Streams, client ori.MachineRuntimeClient, names []string, opts Options) error {
	for _, name := range names {
		if _, err := client.DetachVolume(ctx, &ori.DetachVolumeRequest{
			MachineId: opts.MachineID,
			Name:      name,
		}); err != nil {
			if status.Code(err) != codes.NotFound {
				return fmt.Errorf("error detaching volume %s from machine %s: %w", name, opts.MachineID, err)
			}
			_, _ = fmt.Fprintf(streams.Out, "Volume %s in machine %s not found\n", name, opts.MachineID)
		} else {
			_, _ = fmt.Fprintf(streams.Out, "Detached volume %s from machine %s\n", name, opts.MachineID)
		}
	}
	return nil
}
