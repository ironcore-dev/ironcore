// Copyright 2023 IronCore authors
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
	"os"

	iri "github.com/ironcore-dev/ironcore/iri/apis/machine/v1alpha1"
	"github.com/ironcore-dev/ironcore/irictl-machine/cmd/irictl-machine/irictlmachine/common"
	clicommon "github.com/ironcore-dev/ironcore/irictl/cmd"
	"github.com/ironcore-dev/ironcore/irictl/decoder"
	"github.com/spf13/cobra"
	utilruntime "k8s.io/apimachinery/pkg/util/runtime"
	ctrl "sigs.k8s.io/controller-runtime"
)

type Options struct {
	Filename  string
	MachineID string
}

func (o *Options) AddFlags(cmd *cobra.Command) {
	cmd.Flags().StringVarP(&o.Filename, "filename", "f", o.Filename, "Path to a file to read.")
	cmd.Flags().StringVar(&o.MachineID, "machine-id", "", "The machine ID to modify.")
	utilruntime.Must(cmd.MarkFlagRequired("machine-id"))
}

func Command(streams clicommon.Streams, clientFactory common.Factory) *cobra.Command {
	var (
		opts Options
	)

	cmd := &cobra.Command{
		Use:     "volume",
		Aliases: common.VolumeAliases,
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
	data, err := clicommon.ReadFileOrReader(opts.Filename, os.Stdin)
	if err != nil {
		return err
	}

	volume := &iri.Volume{}
	if err := decoder.Decode(data, volume); err != nil {
		return err
	}

	if _, err := client.AttachVolume(ctx, &iri.AttachVolumeRequest{Volume: volume, MachineId: opts.MachineID}); err != nil {
		return err
	}

	_, _ = fmt.Fprintf(streams.Out, "Attached volume %s to machine %s\n", volume.Name, opts.MachineID)
	return nil
}
