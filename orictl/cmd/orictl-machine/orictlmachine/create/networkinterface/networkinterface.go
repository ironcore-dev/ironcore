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
	"os"

	ori "github.com/onmetal/onmetal-api/ori/apis/machine/v1alpha1"
	clicommon "github.com/onmetal/onmetal-api/orictl/cli/common"
	"github.com/onmetal/onmetal-api/orictl/cmd/orictl-machine/orictlmachine/common"
	"github.com/onmetal/onmetal-api/orictl/decoder"
	"github.com/spf13/cobra"
	"github.com/spf13/pflag"
	ctrl "sigs.k8s.io/controller-runtime"
)

type Options struct {
	MachineID string
	Filename  string
}

func (o *Options) AddFlags(fs *pflag.FlagSet) {
	fs.StringVar(&o.MachineID, "machine-id", "", "ID of the machine to create the network interface in.")
	fs.StringVarP(&o.Filename, "filename", "f", o.Filename, "Path to a file to read.")
}

func (o *Options) MarkFlagsRequired(cmd *cobra.Command) {
	_ = cmd.MarkFlagRequired("machine-id")
}

func Command(streams clicommon.Streams, clientFactory common.ClientFactory) *cobra.Command {
	var opts Options

	cmd := &cobra.Command{
		Use:     "networkinterface",
		Aliases: common.NetworkInterfaceAliases,
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

			return Run(ctx, streams, client, opts)
		},
	}

	opts.AddFlags(cmd.Flags())
	opts.MarkFlagsRequired(cmd)

	return cmd
}

func Run(ctx context.Context, streams clicommon.Streams, client ori.MachineRuntimeClient, opts Options) error {
	data, err := clicommon.ReadFileOrReader(opts.Filename, os.Stdin)
	if err != nil {
		return err
	}

	networkInterfaceConfig := &ori.NetworkInterfaceConfig{}
	if err := decoder.Decode(data, networkInterfaceConfig); err != nil {
		return err
	}

	_, err = client.CreateNetworkInterface(ctx,
		&ori.CreateNetworkInterfaceRequest{
			MachineId: opts.MachineID,
			Config:    networkInterfaceConfig,
		},
	)
	if err != nil {
		return err
	}

	_, _ = fmt.Fprintln(streams.Out, networkInterfaceConfig.Name)
	return nil
}
