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

	ori "github.com/onmetal/onmetal-api/ori/apis/machine/v1alpha1"
	clicommon "github.com/onmetal/onmetal-api/orictl/cli/common"
	"github.com/onmetal/onmetal-api/orictl/cmd/orictl-machine/orictlmachine/common"
	"github.com/spf13/cobra"
	"github.com/spf13/pflag"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
	ctrl "sigs.k8s.io/controller-runtime"
)

type Options struct {
	MachineID string
}

func (o *Options) AddFlags(fs *pflag.FlagSet) {
	fs.StringVar(&o.MachineID, "machine-id", "", "ID of the machine to create the network interface in.")
}

func (o *Options) MarkFlagsRequired(cmd *cobra.Command) {
	_ = cmd.MarkFlagRequired("machine-id")
}

func Command(streams clicommon.Streams, clientFactory common.ClientFactory) *cobra.Command {
	var (
		opts Options
	)

	cmd := &cobra.Command{
		Use:     "networkinterface name [names...]",
		Aliases: common.NetworkInterfaceAliases,
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

			names := args

			return Run(cmd.Context(), streams, client, names, opts)
		},
	}

	opts.AddFlags(cmd.Flags())
	opts.MarkFlagsRequired(cmd)

	return cmd
}

func Run(ctx context.Context, streams clicommon.Streams, client ori.MachineRuntimeClient, names []string, opts Options) error {
	for _, name := range names {
		if _, err := client.DeleteNetworkInterface(ctx, &ori.DeleteNetworkInterfaceRequest{
			MachineId:            opts.MachineID,
			NetworkInterfaceName: name,
		}); err != nil {
			if status.Code(err) != codes.NotFound {
				return fmt.Errorf("error deleting machine %s network interface %s: %w", opts.MachineID, name, err)
			}

			_, _ = fmt.Fprintf(streams.Out, "Machine %s network interface %s not found\n", opts.MachineID, name)
		} else {
			_, _ = fmt.Fprintf(streams.Out, "Machine %s network interface %s deleted\n", opts.MachineID, name)
		}
	}
	return nil
}
