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

	ori "github.com/onmetal/onmetal-api/ori/apis/compute/v1alpha1"
	"github.com/onmetal/onmetal-api/orictl/cmd/orictl/compute/common"
	"github.com/spf13/cobra"
	"github.com/spf13/pflag"
	ctrl "sigs.k8s.io/controller-runtime"
)

type Options struct {
	MachineID string
	IPs       []string
	VirtualIP string
}

func (o *Options) AddFlags(fs *pflag.FlagSet) {
	fs.StringVar(&o.MachineID, "machine-id", "", "ID of the machine to create the network interface in.")
	fs.StringSliceVar(&o.IPs, "ips", o.IPs, "IPs to set on the network interface.")
	fs.StringVar(&o.VirtualIP, "virtual-ip", o.VirtualIP, "Virtual ip to set on the network interface")
}

func (o *Options) MarkFlagsRequired(cmd *cobra.Command) {
	_ = cmd.MarkFlagRequired("machine-id")
	_ = cmd.MarkFlagRequired("ips")
}

func Command(streams common.Streams, clientFactory common.ClientFactory) *cobra.Command {
	var opts Options

	cmd := &cobra.Command{
		Use:     "networkinterface name",
		Aliases: common.NetworkInterfaceAliases,
		Args:    cobra.ExactArgs(1),
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

			name := args[0]

			return Run(ctx, streams, client, name, opts)
		},
	}

	opts.AddFlags(cmd.Flags())
	opts.MarkFlagsRequired(cmd)

	return cmd
}

func Run(ctx context.Context, streams common.Streams, client ori.MachineRuntimeClient, networkInterfaceName string, opts Options) error {
	var virtualIP *ori.VirtualIPConfig
	if opts.VirtualIP != "" {
		virtualIP = &ori.VirtualIPConfig{
			Ip: opts.VirtualIP,
		}
	}

	_, err := client.UpdateNetworkInterface(ctx,
		&ori.UpdateNetworkInterfaceRequest{
			MachineId:            opts.MachineID,
			NetworkInterfaceName: networkInterfaceName,
			Ips:                  opts.IPs,
			VirtualIp:            virtualIP,
		},
	)
	if err != nil {
		return err
	}

	_, _ = fmt.Fprintf(streams.Out, "Updated machine %s network interface %s\n", opts.MachineID, networkInterfaceName)
	return nil
}
