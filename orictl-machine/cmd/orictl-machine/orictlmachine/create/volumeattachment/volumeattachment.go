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

package volumeattachment

import (
	"context"
	"fmt"

	ori "github.com/onmetal/onmetal-api/ori/apis/machine/v1alpha1"
	"github.com/onmetal/onmetal-api/orictl-machine/cmd/orictl-machine/orictlmachine/common"
	clicommon "github.com/onmetal/onmetal-api/orictl/cmd"
	"github.com/onmetal/onmetal-api/orictl/decoder"
	"github.com/spf13/cobra"
	"github.com/spf13/pflag"
	ctrl "sigs.k8s.io/controller-runtime"
)

type Options struct {
	Filename  string
	MachineID string
}

func (o *Options) AddFlags(fs *pflag.FlagSet) {
	fs.StringVarP(&o.Filename, "filename", "f", o.Filename, "Path to a file to read.")
	fs.StringVar(&o.MachineID, "machine-id", "", "ID of the machine to attach the volume to.")
}

func (o *Options) MarkFlagsRequired(cmd *cobra.Command) {
	_ = cmd.MarkFlagRequired("machine-id")
}

func Command(streams clicommon.Streams, clientFactory common.Factory) *cobra.Command {
	var (
		opts Options
	)

	cmd := &cobra.Command{
		Use:  "volumeattachment",
		Args: cobra.NoArgs,
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

			return Run(cmd.Context(), streams, client, opts)
		},
	}

	opts.AddFlags(cmd.Flags())
	opts.MarkFlagsRequired(cmd)

	return cmd
}

func Run(ctx context.Context, streams clicommon.Streams, client ori.MachineRuntimeClient, opts Options) error {
	data, err := clicommon.ReadFileOrReader(opts.Filename, streams.In)
	if err != nil {
		return err
	}

	volume := &ori.VolumeAttachment{}
	if err := decoder.Decode(data, volume); err != nil {
		return err
	}

	if _, err := client.CreateVolumeAttachment(ctx, &ori.CreateVolumeAttachmentRequest{
		MachineId: opts.MachineID,
		Volume:    volume,
	}); err != nil {
		return fmt.Errorf("error creating volume attachment: %w", err)
	}

	_, _ = fmt.Fprintf(streams.Out, "Created volume %s attachment\n", volume.Device)
	return nil
}
