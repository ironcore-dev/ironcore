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

package machine

import (
	"context"
	"fmt"
	"os"

	ori "github.com/onmetal/onmetal-api/ori/apis/runtime/v1alpha1"
	"github.com/onmetal/onmetal-api/orictl/cmd/orictl/common"
	"github.com/onmetal/onmetal-api/orictl/decoder"
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

func Command(streams common.Streams, clientFactory common.ClientFactory) *cobra.Command {
	var opts Options

	cmd := &cobra.Command{
		Use:     "machine",
		Aliases: common.MachineAliases,
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

			return Run(ctx, streams, client, opts)
		},
	}

	opts.AddFlags(cmd.Flags())

	return cmd
}

func Run(ctx context.Context, streams common.Streams, client ori.MachineRuntimeClient, opts Options) error {
	data, err := common.ReadFileOrReader(opts.Filename, os.Stdin)
	if err != nil {
		return err
	}

	machineConfig := &ori.MachineConfig{}
	if err := decoder.Decode(data, machineConfig); err != nil {
		return err
	}

	res, err := client.CreateMachine(ctx, &ori.CreateMachineRequest{Config: machineConfig})
	if err != nil {
		return err
	}

	_, _ = fmt.Fprintln(streams.Out, res.Machine.Id)
	return nil
}
