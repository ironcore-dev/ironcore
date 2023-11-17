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

package bucketclass

import (
	"context"
	"fmt"

	ori "github.com/ironcore-dev/ironcore/ori/apis/bucket/v1alpha1"
	"github.com/ironcore-dev/ironcore/orictl-bucket/cmd/orictl-bucket/orictlbucket/common"
	orictlcmd "github.com/ironcore-dev/ironcore/orictl/cmd"
	"github.com/ironcore-dev/ironcore/orictl/renderer"
	"github.com/spf13/cobra"
	ctrl "sigs.k8s.io/controller-runtime"
)

func Command(streams orictlcmd.Streams, clientFactory common.ClientFactory) *cobra.Command {
	var (
		outputOpts = common.NewOutputOptions()
	)

	cmd := &cobra.Command{
		Use:     "bucketclass",
		Aliases: common.BucketClassAliases,
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

func Run(ctx context.Context, streams orictlcmd.Streams, client ori.BucketRuntimeClient, render renderer.Renderer) error {
	res, err := client.ListBucketClasses(ctx, &ori.ListBucketClassesRequest{})
	if err != nil {
		return fmt.Errorf("error listing bucket classes: %w", err)
	}

	return render.Render(res.BucketClasses, streams.Out)
}
