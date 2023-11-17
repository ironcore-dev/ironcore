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

package machine

import (
	"context"
	"fmt"
	"os"

	ori "github.com/ironcore-dev/ironcore/ori/apis/machine/v1alpha1"
	"github.com/ironcore-dev/ironcore/orictl-machine/cmd/orictl-machine/orictlmachine/common"
	clicommon "github.com/ironcore-dev/ironcore/orictl/cmd"
	"github.com/ironcore-dev/ironcore/orictl/decoder"
	"github.com/ironcore-dev/ironcore/orictl/renderer"
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

func Command(streams clicommon.Streams, clientFactory common.Factory) *cobra.Command {
	var (
		outputOpts = clientFactory.OutputOptions()
		opts       Options
	)

	cmd := &cobra.Command{
		Use:     "machine",
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

			r, err := outputOpts.RendererOrNil()
			if err != nil {
				return err
			}

			return Run(ctx, streams, client, r, opts)
		},
	}

	opts.AddFlags(cmd.Flags())

	return cmd
}

func Run(ctx context.Context, streams clicommon.Streams, client ori.MachineRuntimeClient, r renderer.Renderer, opts Options) error {
	data, err := clicommon.ReadFileOrReader(opts.Filename, os.Stdin)
	if err != nil {
		return err
	}

	machine := &ori.Machine{}
	if err := decoder.Decode(data, machine); err != nil {
		return err
	}

	res, err := client.CreateMachine(ctx, &ori.CreateMachineRequest{Machine: machine})
	if err != nil {
		return err
	}

	if r != nil {
		return r.Render(res.Machine, streams.Out)
	}
	_, _ = fmt.Fprintf(streams.Out, "Created machine %s\n", res.Machine.Metadata.Id)
	return nil
}
