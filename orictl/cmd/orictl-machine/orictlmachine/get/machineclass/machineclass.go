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

package machineclass

import (
	"context"
	"fmt"

	ori "github.com/onmetal/onmetal-api/ori/apis/machine/v1alpha1"
	clicommon "github.com/onmetal/onmetal-api/orictl/cli/common"
	"github.com/onmetal/onmetal-api/orictl/cmd/orictl-machine/orictlmachine/common"
	"github.com/onmetal/onmetal-api/orictl/renderer"
	"github.com/spf13/cobra"
	ctrl "sigs.k8s.io/controller-runtime"
)

func Command(streams clicommon.Streams, clientFactory common.ClientFactory) *cobra.Command {
	var (
		outputOpts common.OutputOptions
	)

	cmd := &cobra.Command{
		Use:     "machineclass",
		Aliases: common.MachineClassAliases,
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

			render, err := outputOpts.Renderer("table")
			if err != nil {
				return err
			}

			return Run(cmd.Context(), streams, client, render)
		},
	}

	outputOpts.AddFlags(cmd.Flags())

	return cmd
}

func Run(ctx context.Context, streams clicommon.Streams, client ori.MachineRuntimeClient, render renderer.Renderer) error {
	res, err := client.ListMachineClasses(ctx, &ori.ListMachineClassesRequest{})
	if err != nil {
		return fmt.Errorf("error listing machine classes: %w", err)
	}

	return render.Render(res.MachineClasses, streams.Out)
}
